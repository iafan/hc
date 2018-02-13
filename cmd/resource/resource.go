package resource

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/raff/godet"

	"github.com/iafan/hc/lib"
)

// Command implements 'load' command
type Command struct {
	host lib.Host

	blockedURLsParam string
	blockedURLs      []string
	resourceMatch    string
	matchMode        string
	matchIsContains  bool
	matchIsPrefix    bool
	matchIsRegex     bool
	url              string
	wait             time.Duration

	reMatch *regexp.Regexp
}

// GetDescription implements Command.GetDescription
func (c *Command) GetDescription() string {
	return "Load resource in the context of a page and return its content"
}

// ShowHelp implements Command.ShowHelp
func (c *Command) ShowHelp() {
	os.Stderr.WriteString(`Description:

	Load resource (defined by its URL mask) in the context of a page
	and return its content

Usage:

	hc resource [options] <URL> <resource-URL-mask>
	hc resource --help

Available options:

`)

	flag.PrintDefaults()
}

// Init implements Command.Init
func (c *Command) Init(host lib.Host) {
	c.host = host

	flag.StringVar(&c.matchMode, "match", "exact", "Match mode to use ('contains', 'exact', 'prefix' or 'regexp')")
	flag.BoolVar(&c.matchIsRegex, "match-regexp", false, "Match using a regular expression")
	flag.DurationVar(&c.wait, "wait", 500*time.Millisecond, "Extra time to wait before capturing data")

	flag.StringVar(
		&c.blockedURLsParam,
		"blocked-urls",
		"",
		"Comma-separated list of file masks to block from loading",
	)
}

// Validate implements Command.Validate
func (c *Command) Validate(args []string) {
	if c.blockedURLsParam != "" {
		c.blockedURLs = strings.Split(c.blockedURLsParam, ",")
	}

	if len(args) != 2 {
		os.Stderr.WriteString("Usage: hc resource [options] <URL> <resource-URL-mask>\n")
		os.Stderr.WriteString("       hc resource --help\n")
		os.Exit(2)
	}

	c.url = args[0]
	c.resourceMatch = args[1]

	switch c.matchMode {
	case "exact":
		break
	case "prefix":
		c.matchIsPrefix = true
	case "contains":
		c.matchIsContains = true
	case "regexp":
		c.matchIsRegex = true
	default:
		os.Stderr.WriteString(fmt.Sprintf(
			"Unknown match mode: '%s'. Available modes: 'contains', 'exact', 'prefix' or 'regexp'\n",
			c.matchMode,
		))
		os.Exit(2)
	}

	if c.matchIsRegex {
		var err error
		c.reMatch, err = regexp.Compile(c.resourceMatch)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf(
				"Failed to compile the regular expression: %s\n",
				err,
			))
			os.Exit(2)
		}
	}
}

// Run implements Command.Run
func (c *Command) Run(outfile *os.File) (err error) {
	remote, err := c.host.ConnectToRemote()
	if err != nil {
		return
	}
	defer c.host.DisconnectFromRemote()

	verbose := c.host.GetVerbose()

	// block resource loading
	if len(c.blockedURLs) > 0 {
		err = remote.SetBlockedURLs(c.blockedURLs...)
		if err != nil {
			return
		}
	}

	// create new tab
	/*
		_, err = remote.NewTab(c.url)
		if err != nil {
			return
		}
	*/
	//_, err = remote.Navigate(c.url)

	/*
		err = remote.ActivateTab(tab)
		if err != nil {
			return
		}
	*/

	remote.NetworkEvents(true)
	if err != nil {
		return
	}

	_, err = remote.Navigate(c.url)

	status := make(chan bool, 2)
	result := false

	exitErr := fmt.Errorf("Request timed out")

	go func() {
		time.Sleep(c.host.GetDeadline())
		status <- false
	}()

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		resp := params["response"].(map[string]interface{})
		respURL := resp["url"].(string)
		if verbose {
			mime := resp["mimeType"].(string)
			log.Printf("Loaded %s (%s)", respURL, mime)
		}

		matched := false
		if c.matchIsRegex {
			matched = c.reMatch.MatchString(respURL)
		} else if c.matchIsPrefix {
			matched = strings.HasPrefix(respURL, c.resourceMatch)
		} else if c.matchIsContains {
			matched = strings.Contains(respURL, c.resourceMatch)
		} else {
			// exact match
			matched = respURL == c.resourceMatch
		}

		if matched {
			time.Sleep(c.wait)

			p := make(godet.Params)
			p["requestId"] = params["requestId"].(string)

			res, err := remote.SendRequest(
				"Network.getResponseBody",
				p,
			)

			if res["body"] == nil {
				exitErr = fmt.Errorf("Failed to fetch the resource (internal error)")
				status <- false
				return
			}

			if err != nil {
				log.Printf("Network.getResponseBody Error: %s", err)
			} else {
				isBase64Encoded := res["base64Encoded"] != nil && res["base64Encoded"].(bool)
				if isBase64Encoded {
					buf := make([]byte, 8192)
					r := strings.NewReader(res["body"].(string))
					b64 := base64.NewDecoder(base64.StdEncoding, r)
					for {
						n, err := b64.Read(buf)
						if n > 0 {
							_, err2 := outfile.Write(buf[:n])
							if err != nil {
								exitErr = fmt.Errorf("Write error: %v", err2)
								status <- false
								return
							}
						}
						if err != nil && err != io.EOF {
							exitErr = fmt.Errorf("Read error: %v", err)
							status <- false
							return
						}
						if n == 0 {
							break
						}
					}
				} else {
					outfile.WriteString(res["body"].(string))
				}
			}

			status <- true
		}
	})

	result = <-status

	if !result {
		return exitErr
	}

	return
}

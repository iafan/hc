package eval

import (
	"flag"
	"fmt"
	"os"
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
	url              string
	stopEvent        string
	evalStr          string
	wait             time.Duration
}

// GetDescription implements Command.GetDescription
func (c *Command) GetDescription() string {
	return "Load a specific page, evaluate an expression and print its result"
}

// ShowHelp implements Command.ShowHelp
func (c *Command) ShowHelp() {
	os.Stderr.WriteString(`Description:

	Load a specific page, wait for a specific page lifecycle event,
	evaluate a JavaScript expression and print its result

Usage:

	hc eval [options] <URL> <JavaScript-expression>
	hc eval --help

Available options:

`)

	flag.PrintDefaults()
}

// Init implements Command.Init
func (c *Command) Init(host lib.Host) {
	c.host = host

	flag.StringVar(&c.stopEvent, "stop-event", "networkIdle", "Event to stop upon")
	flag.DurationVar(&c.wait, "wait", 500*time.Millisecond, "Extra time to wait before running the script")

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
		os.Stderr.WriteString("Usage: hc eval [options] <URL> <JavaScript-expression>\n")
		os.Stderr.WriteString("       hc eval --help\n")
		os.Exit(2)
	}

	c.url = args[0]
	c.evalStr = args[1]
}

// Run implements Command.Run
func (c *Command) Run(outfile *os.File) (err error) {
	remote, err := c.host.ConnectToRemote()
	if err != nil {
		return
	}
	defer c.host.DisconnectFromRemote()

	// block resource loading
	if len(c.blockedURLs) > 0 {
		err = remote.SetBlockedURLs(c.blockedURLs...)
		if err != nil {
			return
		}
	}

	remote.PageEvents(true)
	if err != nil {
		return
	}

	tabID, err := remote.Navigate(c.url)

	status := make(chan bool, 2)
	result := false

	go func() {
		time.Sleep(c.host.GetDeadline())
		status <- false
	}()

	remote.CallbackEvent("Page.lifecycleEvent", func(params godet.Params) {
		if params["name"] == c.stopEvent && params["frameId"] == tabID {
			time.Sleep(c.wait)
			status <- true
		}
	})

	result = <-status

	if !result {
		return fmt.Errorf("Request timed out")
	}

	// evaluate Javascript expression in existing context
	res, err := remote.EvaluateWrap(c.evalStr)
	if err != nil {
		return err
	}

	outfile.WriteString(fmt.Sprintf("%v", res))

	return
}

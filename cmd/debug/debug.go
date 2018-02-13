package debug

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iafan/hc/lib"
)

// Command implements 'load' command
type Command struct {
	host lib.Host

	blockedURLsParam string
	blockedURLs      []string
	url              string
}

// GetDescription implements Command.GetDescription
func (c *Command) GetDescription() string {
	return "Peek into page and DOM events as the page loads"
}

// ShowHelp implements Command.ShowHelp
func (c *Command) ShowHelp() {
	os.Stderr.WriteString(`Description:

	Get debugging information on various events as the page loads.
	This is useful to understand the inner workings of Chrome
	and to use gathered information when running other commands
	(e.g. 'hc load-resource')

Usage:

	hc debug [options] <URL>
	hc debug --help

Available options:

`)

	flag.PrintDefaults()
}

// Init implements Command.Init
func (c *Command) Init(host lib.Host) {
	c.host = host

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

	if len(args) != 1 {
		os.Stderr.WriteString("Usage: hc debug [options] <URL>\n")
		os.Stderr.WriteString("       hc debug --help\n")
		os.Exit(2)
	}

	c.url = args[0]
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

	remote.AllEvents(true)
	if err != nil {
		return
	}

	_, err = remote.Navigate(c.url)

	status := make(chan bool, 2)
	result := false

	go func() {
		time.Sleep(c.host.GetDeadline())
		status <- false
	}()

	result = <-status

	if !result {
		return fmt.Errorf("Request timed out")
	}

	return
}

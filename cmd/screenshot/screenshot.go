package screenshot

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/raff/godet"

	"github.com/iafan/hc/lib"
	"github.com/iafan/hc/lib/util"
)

// Command implements 'load' command
type Command struct {
	host lib.Host

	blockedURLsParam string
	blockedURLs      []string
	url              string
	stopEvent        string
	initialWidth     int
	initialHeight    int
	maxWidth         int
	maxHeight        int
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
	flag.DurationVar(&c.wait, "wait", 500*time.Millisecond, "Extra time to wait before making the screenshot")
	flag.IntVar(&c.initialWidth, "initial-width", 1024, "Initial viewport width to render the page")
	flag.IntVar(&c.initialHeight, "initial-height", 768, "Initial viewport height to render the page")
	flag.IntVar(&c.maxWidth, "max-width", 0, "Maximum screenshot width (0 = no maximum)")
	flag.IntVar(&c.maxHeight, "max-height", 0, "Maximum screenshot height (0 = no maximum)")

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
		os.Stderr.WriteString("Usage: hc screenshot [options] <URL>\n")
		os.Stderr.WriteString("       hc screenshot --help\n")
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

	err = util.SetDeviceMetricsOverride(remote, c.initialWidth, c.initialHeight, 1, false, false)
	if err != nil {
		return
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

	res, err := remote.EvaluateWrap("return document.documentElement.scrollWidth")
	if err != nil {
		return
	}
	width := int(res.(float64))

	res, err = remote.EvaluateWrap("return document.documentElement.scrollHeight")
	if err != nil {
		return
	}
	height := int(res.(float64))

	if c.host.GetVerbose() {
		log.Printf("Document size: %dx%d", width, height)
	}

	if c.maxWidth > 0 && width > c.maxWidth {
		width = c.maxWidth
	}

	if c.maxHeight > 0 && height > c.maxHeight {
		height = c.maxHeight
	}

	if c.host.GetVerbose() {
		log.Printf("Screenshot size: %dx%d", width, height)
	}

	err = remote.SetVisibleSize(width, height)
	if err != nil {
		return
	}

	err = util.SetDeviceMetricsOverride(remote, width, height, 1, false, false)
	if err != nil {
		return
	}

	bytes, err := remote.CaptureScreenshot("png", 0, true)
	if err != nil {
		return
	}
	outfile.Write(bytes)

	return
}

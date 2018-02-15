package html

import (
	"flag"
	"os"

	"github.com/iafan/hc/cmd/eval"
	"github.com/iafan/hc/lib"
)

// Command implements 'load' command
type Command struct {
	host lib.Host
	eval *eval.Command
}

// GetDescription implements Command.GetDescription
func (c *Command) GetDescription() string {
	return "Load and render HTML page content"
}

// ShowHelp implements Command.ShowHelp
func (c *Command) ShowHelp() {
	os.Stderr.WriteString(`Description:

	Load a specific page and return its HTML content

Usage:

	hc html [options] <URL>
	hc html --help

Available options:

`)

	flag.PrintDefaults()
}

// Init implements Command.Init
func (c *Command) Init(host lib.Host) {
	c.host = host

	c.eval = &eval.Command{}
	c.eval.Init(host)
}

// Validate implements Command.Validate
func (c *Command) Validate(args []string) {
	if len(args) != 1 {
		os.Stderr.WriteString("Usage: hc html [options] <URL>\n")
		os.Stderr.WriteString("       hc html --help\n")
		os.Exit(2)
	}

	c.eval.Validate([]string{args[0], "return document.documentElement.outerHTML"})
}

// Run implements Command.Run
func (c *Command) Run(outfile *os.File) (err error) {
	return c.eval.Run(outfile)
}

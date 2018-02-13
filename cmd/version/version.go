package version

import (
	"fmt"
	"os"

	"github.com/iafan/hc/lib"
)

// Command implements 'version' command
type Command struct {
}

// GetDescription implements Command.GetDescription
func (c *Command) GetDescription() string {
	return "Show the program version"
}

// ShowHelp implements Command.ShowHelp
func (c *Command) ShowHelp() {
	os.Stderr.WriteString(`Description:

	Show the program version

Usage:

	hc version
`)
}

// Init implements Command.Init
func (c *Command) Init(host lib.Host) {
	// nothing
}

// Validate implements Command.Validate
func (c *Command) Validate(args []string) {
	// nothing
}

// Run implements Command.Run
func (c *Command) Run(outfile *os.File) (err error) {
	outfile.WriteString(fmt.Sprintf("hc version %s", lib.GetVersion()))
	return
}

package lib

import (
	"os"
	"time"

	"github.com/raff/godet"
)

// Host defines an interface for command host
type Host interface {
	ConnectToRemote() (*godet.RemoteDebugger, error)
	DisconnectFromRemote() error
	GetDeadline() time.Duration
	GetVerbose() bool
}

// Command defines an interface for pluggable commands
type Command interface {
	GetDescription() string
	ShowHelp()
	Init(host Host)
	Validate(args []string)
	Run(outfile *os.File) error
}

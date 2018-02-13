package host

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/raff/godet"

	"github.com/iafan/hc/lib"
)

// CommandHost is a host for other commands
type CommandHost struct {
	showHelp        bool
	verbose         bool
	verboseDevTools bool
	//chromeHost      string
	dockerImage string
	deadline    time.Duration
	remote      *godet.RemoteDebugger
	commands    map[string]lib.Command

	containerName string
}

// ConnectToRemote implements Host.ConnectToRemote
func (h *CommandHost) ConnectToRemote() (remote *godet.RemoteDebugger, err error) {
	return h.ConnectToNewDockerContainer()
}

// DisconnectFromRemote implements Host.Disconnect
func (h *CommandHost) DisconnectFromRemote() (err error) {
	return h.DisconnectAndRemoveDockerContainer()
}

// GetDeadline implements Host.GetDeadline
func (h *CommandHost) GetDeadline() time.Duration {
	return h.deadline
}

// GetVerbose implements Host.GetVerbose
func (h *CommandHost) GetVerbose() bool {
	return h.verbose
}

// GetShowHelp implements Host.GetShowHelp
func (h *CommandHost) GetShowHelp() bool {
	return h.showHelp
}

// ListCommands renders a formatted list of registered commands
func (h *CommandHost) ListCommands() {
	os.Stderr.WriteString("Available commands:")

	max := 0
	for cmd := range h.commands {
		l := len(cmd)
		if l > max {
			max = l
		}
	}

	var keys []string
	for k := range h.commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := range keys {
		cmd := keys[i]
		handler := h.commands[cmd]
		os.Stderr.WriteString(fmt.Sprintf(
			"    %s%s    %s\n",
			cmd,
			strings.Repeat(" ", max-len(cmd)),
			handler.GetDescription(),
		))
	}
}

// SetHandler registers a new command name and its handler
func (h *CommandHost) SetHandler(cmdName string, handler lib.Command) {
	h.commands[cmdName] = handler
}

// GetHandler return a command handler by the provided command name
func (h *CommandHost) GetHandler(cmdName string) lib.Command {
	return h.commands[cmdName]
}

// Init specifies common command-line flags to parse
func (h *CommandHost) Init(cmdName string) {
	flag.BoolVar(&h.showHelp, "help", false, "Show help")
	flag.BoolVar(&h.verbose, "verbose", cmdName == "debug", "Show verbose messages")
	flag.BoolVar(&h.verboseDevTools, "verbose-devtools", cmdName == "debug", "Show verbose DevTools protocol messages")
	//flag.StringVar(&h.chromeHost, "host", /*"localhost:9222"*/, "Headless Chrome hostname to connect to")
	flag.StringVar(&h.dockerImage, "docker-image", "justinribeiro/chrome-headless", "Docker image to use to spin up a temporary container")
	flag.DurationVar(&h.deadline, "deadline", 30*time.Second, "Deadline")
}

// New returns an initialized command host instance
func New() *CommandHost {
	return &CommandHost{
		commands: make(map[string]lib.Command),
	}
}

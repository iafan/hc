package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/iafan/hc/cmd/debug"
	"github.com/iafan/hc/cmd/eval"
	"github.com/iafan/hc/cmd/html"
	"github.com/iafan/hc/cmd/resource"
	"github.com/iafan/hc/cmd/screenshot"
	"github.com/iafan/hc/cmd/version"
	"github.com/iafan/hc/host"
	"github.com/iafan/hc/lib/util"
)

func main() {
	var err error
	var args = os.Args[1:]

	var host = host.New()
	host.SetHandler("debug", &debug.Command{})
	host.SetHandler("eval", &eval.Command{})
	host.SetHandler("html", &html.Command{})
	host.SetHandler("resource", &resource.Command{})
	host.SetHandler("screenshot", &screenshot.Command{})
	host.SetHandler("version", &version.Command{})

	aliases := make(map[string]string)
	aliases["d"] = "debug"
	aliases["e"] = "eval"
	aliases["h"] = "html"
	aliases["r"] = "resource"
	aliases["s"] = "screenshot"
	aliases["v"] = "version"

	if len(args) == 0 || args[0] == "--help" {
		os.Stderr.WriteString("Usage: hc <command> [--help] [parameters]\n\n")
		host.ListCommands()
		os.Exit(0)
	}

	cmdName := aliases[args[0]]
	if cmdName == "" {
		cmdName = args[0]
	}

	handler := host.GetHandler(cmdName)
	if handler == nil {
		fmt.Printf("Unknown command: %s\n\n", args[0])
		host.ListCommands()
		os.Exit(2)
	}

	// setup common flags to parse
	host.Init(cmdName)

	// init handler and setup its own flags to parse
	handler.Init(host)

	// init common flags
	filename := ""
	flag.StringVar(&filename, "output-file", "", "File to write output to; if not provided, will write to STDOUT")

	flag.CommandLine.Parse(args[1:])

	if host.GetShowHelp() {
		handler.ShowHelp()
		os.Exit(0)
	}

	// validate input data
	handler.Validate(flag.Args())

	file := os.Stdout

	useFile := filename != ""

	if useFile {
		filename = util.ExpandMacros(filename)
		if host.GetVerbose() {
			log.Printf("Opening [%s] for writing", filename)
		}

		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Failed to open [%s] file for writing: %v", filename, err))
			os.Exit(4)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			os.Stderr.WriteString("\n")
			host.RequestInterrupt()
		}
	}()

	// run the command
	err = handler.Run(file)

	var err2 error
	if useFile {
		err2 = file.Close()
	}

	// First, stop on command execution error,
	// and only then check for file.Close() error
	util.StopOnError(err)
	if useFile {
		util.StopOnError(err2)
	}
}

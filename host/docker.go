package host

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/raff/godet"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getSeccompFilePath() (filename string) {
	// test if chrome.json file exists next to the binary

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		filename = filepath.Join(dir, "chrome.json")
		if fileExists(filename) {
			return
		}
	}

	// test if chrome.json file exists in the sources (if installed via `go get`)

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	filename = filepath.Join(
		gopath, "src", "github.com", "iafan", "hc", "chrome.json",
	)
	if fileExists(filename) {
		return
	}

	return ""
}

// ConnectToNewDockerContainer connects to a new docker container
// and returns a connected instance of *godet.RemoteDebugger
func (h *CommandHost) ConnectToNewDockerContainer() (remote *godet.RemoteDebugger, err error) {
	if h.remote != nil {
		return h.remote, nil
	}

	seccompFile := getSeccompFilePath()
	if seccompFile == "" {
		err = fmt.Errorf("chrome.json can not be found next to hc binary or in default source location")
		return
	}

	if h.verbose {
		log.Printf("Creating a container from %s image", h.dockerImage)
	}

	cmd := exec.Command(
		"docker", "run", "-d",
		"-p", "127.0.0.1::9222",
		"--security-opt", fmt.Sprintf("seccomp=%s", seccompFile),
		h.dockerImage,
	)

	if h.verbose {
		log.Printf("Command: %+v", cmd.Args)
	}
	bytes, err := cmd.Output()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	h.containerName = strings.TrimSpace(string(bytes))

	cmd = exec.Command("docker", "port", h.containerName)
	if h.verbose {
		log.Printf("Command: %+v", cmd.Args)
	}
	bytes, err = cmd.Output()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	dbgHost := strings.TrimPrefix(strings.TrimSpace(string(bytes)), "9222/tcp -> ")

	if h.verbose {
		log.Printf("Connecting to %s", dbgHost /*h.chromeHost*/)
	}

	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)

		remote, err = godet.Connect(
			dbgHost, /*h.chromeHost*/
			h.verboseDevTools,
		)
		if err == nil {
			h.remote = remote
			return
		}
	}
	return
}

// DisconnectAndRemoveDockerContainer dicsonnects from a headless Chrome
// instance, and then stops and removes the temporary Docker container
func (h *CommandHost) DisconnectAndRemoveDockerContainer() (err error) {
	if h.remote != nil {
		if h.verbose {
			log.Printf("Disconnecting")
		}
		err = h.remote.Close()
		h.remote = nil

		if h.verbose {
			log.Printf("Removing docker container")
		}
		cmd := exec.Command("docker", "rm", "--force", h.containerName)
		err = cmd.Run()
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
	}
	return
}

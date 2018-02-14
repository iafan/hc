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

// ConnectToNewDockerContainer connects to a new docker container
// and returns a connected instance of *godet.RemoteDebugger
func (h *CommandHost) ConnectToNewDockerContainer() (remote *godet.RemoteDebugger, err error) {
	if h.remote != nil {
		return h.remote, nil
	}

	seccompFile := filepath.Join(
		os.Getenv("GOPATH"), "src", "github.com", "iafan", "hc", "chrome.json",
	)

	_, err = os.Stat(seccompFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("File [%s] does not exist", seccompFile)
		}
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

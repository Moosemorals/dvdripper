package main

import (
	"io"
	"log"
	"os/exec"
)

// A thin wrapper around os.exec
func startCmd(path string, args ...string) (cmd *exec.Cmd, stdout io.ReadCloser, err error) {
	cmd = exec.Command(path, args...)

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	return
}

func main() {
	log.Print("Starting")

	startServer()

}

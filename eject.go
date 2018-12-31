package main

import (
	"os/exec"
)

func eject() error {
	return exec.Command("/usr/bin/eject").Run()
}

func (c *client) doEject() error {
	if err := eject(); err != nil {
		return err
	}
	c.send("eject-success", nil)
	return nil
}

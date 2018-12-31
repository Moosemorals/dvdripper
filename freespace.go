package main

import (
	"log"
	"time"

	"golang.org/x/sys/unix"
)

// FSData holds total and free space for a drive
type FSData struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

func getFreespace(path string) (*FSData, error) {
	var buf unix.Statfs_t

	if err := unix.Statfs(path, &buf); err != nil {
		return nil, err
	}

	return &FSData{
		Total: uint64(buf.Bsize) * buf.Blocks,
		Free:  uint64(buf.Bsize) * buf.Bavail,
	}, nil
}

func sendFreespace(c *client) error {
	fs, err := getFreespace("wwwroot/rips")
	if err != nil {
		log.Printf("WARN: Can't get free space")
		return err
	}

	c.send("freespace", fs)
	return nil
}

func (c *client) pushFreespace(shutdown <-chan bool) {
	if err := sendFreespace(c); err != nil {
		return
	}

	tick := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-shutdown:
			tick.Stop()
			return
		case <-tick.C:
			if err := sendFreespace(c); err != nil {
				return
			}
		}
	}
}

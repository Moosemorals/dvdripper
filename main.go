package main

import (
	"io"
	"log"
	"os/exec"
)

// DVDTrack holds details about an indivdual track
type DVDTrack struct {
	// Title is an index into the disc
	Title int
	// Length is the length of the title in hh:mm:ss.ms format
	Length string
	// Chapters is a count of the chapters in the title
	Chapters int
	// Cells is the count of the cells in the title
	Cells int
	// Streams is the count of audio streams in the title
	Streams int
	// Subpictures is the count of (probably) subtitles
	Subpictures int
}

// DVD holds track information about a DVD
type DVD struct {
	// Title (id) of the disk
	Title string
	// LongestTrack is the index of the longest track
	LongestTrack int
	// Titles is a slice of all the titles on the disk
	Titles []DVDTrack
	// ParseOK is a bool that's true if there were no parsing errors
	ParseOK bool
}

// DVDProgress reports the progress of the rip
type DVDProgress struct {
	// Bytes is the number of bytes written to disk
	Bytes int
	// Percent is the (rough) percent of the track ripped
	Percent float64
}

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

	err := mplayer(3, "/tmp/junk.mpg")
	if err != nil {
		log.Fatalf("Can't run mplayer: %v", err)
	}
}

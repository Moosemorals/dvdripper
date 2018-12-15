package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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

func lsdvd() DVD {
	cmd, stdout, err := startCmd("/usr/bin/lsdvd")

	if err != nil {
		log.Fatal(err)
	}

	reDiskTitle := regexp.MustCompile(`^Disc Title: (.+)$`)
	reTitle := regexp.MustCompile(`^Title: (\d\d), Length: (\d\d:\d\d:\d\d.\d\d\d) Chapters: (\d\d), Cells: (\d\d), Audio streams: (\d\d), Subpictures: (\d\d)$`)
	reLongestTrack := regexp.MustCompile(`^Longest track: (\d\d)$`)

	disk := DVD{
		ParseOK: true,
	}

	scanner := bufio.NewScanner(stdout)
scan:
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case reTitle.MatchString(line):
			match := reTitle.FindStringSubmatch(line)
			trackLength := match[2]
			numberFields := append(match[1:2], match[3:]...)
			numbers := []int{}
			for _, s := range numberFields {
				n, err := strconv.Atoi(s)
				if err != nil {
					disk.ParseOK = false
					log.Printf("WARN: Can't parse line %s: %v", line, err)
					continue scan
				}
				numbers = append(numbers, n)
			}

			disk.Titles = append(disk.Titles, DVDTrack{
				Title:       numbers[0],
				Length:      trackLength,
				Chapters:    numbers[1],
				Cells:       numbers[2],
				Streams:     numbers[3],
				Subpictures: numbers[4],
			})

		case reDiskTitle.MatchString(line):
			match := reDiskTitle.FindStringSubmatch(line)
			disk.Title = match[1]
		case reLongestTrack.MatchString(line):
			match := reLongestTrack.FindStringSubmatch(line)
			disk.LongestTrack, err = strconv.Atoi(match[1])
			if err != nil {
				disk.ParseOK = false
				log.Printf("WARN: Can't convert %s to int: %v", match[1], err)
				continue scan
			}
		default:
			log.Printf("Unknown line %s", line)
			disk.ParseOK = false
			continue scan
		}
	}

	// Wait for the command to finish.
	cmd.Wait()

	return disk
}

func startMplayer(path string, args ...string) (cmd *exec.Cmd, stderr io.ReadCloser, err error) {
	cmd = exec.Command(path, args...)

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	return
}

// These two functions are adappted from the go library scanlines
// under a modified BSD licence
// See https://golang.org/src/bufio/scan.go?s=11799:11877#L335

// dropCR drops a terminal \r fron the data
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// Found a newline
		return i + 1, dropCR(data), nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// Found a newline
		return i + 1, dropCR(data), nil
	}

	// Request more data
	return 0, nil, nil
}

func mplayer(track int, dest string) error {

	src := fmt.Sprintf("dvd://%d", track)

	cmd, stdout, err := startCmd("/usr/bin/mplayer", "-quiet", "-nocache", "-dumpstream", src, "-dumpfile", dest)

	if err != nil {
		return err
	}

	out := json.NewEncoder(os.Stdout)

	dumpRE := regexp.MustCompile(`^dump: (\d+) bytes written \(~(\d+\.\d+)%\)$`)

	scanner := bufio.NewScanner(stdout)
	scanner.Split(scanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if dumpRE.MatchString(line) {
			match := dumpRE.FindStringSubmatch(line)

			bytes, err := strconv.Atoi(match[1])
			if err != nil {
				log.Printf("WARN: %+v", err)
				continue
			}

			percent, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				log.Printf("WARN: %+v", err)
				continue
			}

			progress := DVDProgress{
				Bytes:   bytes,
				Percent: percent,
			}

			out.Encode(progress)
			fmt.Println()
		}
	}

	cmd.Wait()
	return nil
}

func diskToJSON() {
	disk := lsdvd()
	json.NewEncoder(os.Stdout).Encode(disk)
}

func main() {
	log.Print("Starting")

	err := mplayer(3, "/tmp/junk.mpg")
	if err != nil {
		log.Fatalf("Can't run mplayer: %v", err)
	}
}

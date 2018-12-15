package main

import (
	"bufio"
	"encoding/json"
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
}

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

	disk := DVD{}

	scanner := bufio.NewScanner(stdout)
scan:
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case reTitle.MatchString(line):
			match := reTitle.FindStringSubmatch(line)

			trackLength := match[2]

			numberFields := append(match[1:2], match[3:]...)

			log.Printf("Number fileds %+v", numberFields)

			numbers := []int{}
			for _, s := range numberFields {
				n, err := strconv.Atoi(s)
				if err != nil {
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
				log.Printf("WARN: Can't convert %s to int: %v", match[1], err)
			}
		default:
			log.Printf("Unknown line %s", line)
		}

	}

	cmd.Wait()

	return disk

}

func main() {
	log.Print("Starting")
	disk := lsdvd()

	json.NewEncoder(os.Stdout).Encode(disk)

}

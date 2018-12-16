package main

import (
	"bufio"
	"log"
	"regexp"
	"strconv"
)

// DVDTrack holds details about an indivdual track
type DVDTrack struct {
	// ID is the title number
	ID int `json:"id"`
	// Length is the length of the title in hh:mm:ss.ms format
	Length string `json:"length"`
	// Chapters is a count of the chapters in the title
	Chapters int `json:"chapters"`
	// Cells is the count of the cells in the title
	Cells int `json:"cells"`
	// Streams is the count of audio streams in the title
	Streams int `json:"streams"`
	// Subpictures is the count of (probably) subtitles
	Subpictures int `json:"subpictures"`
}

// DVD holds track information about a DVD
type DVD struct {
	// Title (id) of the disk
	DiskID string `json:"diskId"`
	// LongestTrack is the index of the longest track
	LongestTrack int `json:"longestTrack"`
	// Titles is a slice of all the titles on the disk
	Tracks []DVDTrack `json:"tracks"`
	// ParseOK is a bool that's true if there were no parsing errors
	ParsedOK bool `json:"parsedOK"`
}

func (disk *DVD) scan() {
	cmd, stdout, err := startCmd("/usr/bin/lsdvd")

	if err != nil {
		log.Fatal(err)
	}

	reDiskID := regexp.MustCompile(`^Disc Title: (.+)$`)
	reTitle := regexp.MustCompile(`^Title: (\d\d), Length: (\d\d:\d\d:\d\d.\d\d\d) Chapters: (\d\d), Cells: (\d\d), Audio streams: (\d\d), Subpictures: (\d\d)$`)
	reLongestTrack := regexp.MustCompile(`^Longest track: (\d\d)$`)

	disk.ParsedOK = true

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
					disk.ParsedOK = false
					log.Printf("WARN: Can't parse line %s: %v", line, err)
					continue scan
				}
				numbers = append(numbers, n)
			}

			disk.Tracks = append(disk.Tracks, DVDTrack{
				ID:          numbers[0],
				Length:      trackLength,
				Chapters:    numbers[1],
				Cells:       numbers[2],
				Streams:     numbers[3],
				Subpictures: numbers[4],
			})

		case reDiskID.MatchString(line):
			match := reDiskID.FindStringSubmatch(line)
			disk.DiskID = match[1]
		case reLongestTrack.MatchString(line):
			match := reLongestTrack.FindStringSubmatch(line)
			disk.LongestTrack, err = strconv.Atoi(match[1])
			if err != nil {
				disk.ParsedOK = false
				log.Printf("WARN: Can't convert %s to int: %v", match[1], err)
				continue scan
			}
		default:
			log.Printf("Unknown line %s", line)
			disk.ParsedOK = false
			continue scan
		}
	}

	// Wait for the command to finish.
	cmd.Wait()
}

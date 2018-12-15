package main

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"regexp"
)

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

func lsdvd() {
	cmd, stdout, err := startCmd("/usr/bin/lsdvd")

	if err != nil {
		log.Fatal(err)
	}

	reDiskTitle := regexp.MustCompile(`^Disc Title: (.+)$`)
	reTitle := regexp.MustCompile(`^Title: (\d\d), Length: (\d\d:\d\d:\d\d.\d\d\d) Chapters: (\d\d), Cells: (\d\d), Audio streams: (\d\d), Subpictures: (\d\d)$`)
	reLongestTrack := regexp.MustCompile(`^Longest track: (\d\d)$`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case reTitle.MatchString(line):
			match := reTitle.FindStringSubmatch(line)
			log.Printf("Title (%s) - Chapters: %s", match[1], match[3])
		case reDiskTitle.MatchString(line):
			match := reDiskTitle.FindStringSubmatch(line)
			log.Printf("Disk Title = %s", match[1])
		case reLongestTrack.MatchString(line):
			match := reLongestTrack.FindStringSubmatch(line)
			log.Printf("Longest track is %s", match[1])
		default:
			log.Printf("Unknown line %s", line)
		}

	}

	cmd.Wait()

}

func main() {
	log.Print("Starting")
	lsdvd()
}

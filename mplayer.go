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

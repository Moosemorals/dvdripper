package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
)

// DVDProgress reports the progress of the rip
type DVDProgress struct {
	Track   int     `json:"track"`
	Bytes   int     `json:"bytes"`
	Percent float64 `json:"percent"`
}

type mplayer struct {
	progress chan DVDProgress
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
	// Added by me
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// Found a newline
		return i + 1, dropCR(data), nil
	}

	// Request more data
	return 0, nil, nil
}

func (m *mplayer) rip(track int, dest string) error {
	src := fmt.Sprintf("dvd://%d", track)

	cmd, stdout, err := startCmd("/usr/bin/mplayer", "-quiet", "-nocache", "-dumpstream", src, "-dumpfile", dest)

	if err != nil {
		return err
	}

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
			m.progress <- DVDProgress{
				Track:   track,
				Bytes:   bytes,
				Percent: percent,
			}

		}
	}

	cmd.Wait()

	close(m.progress)

	return nil
}

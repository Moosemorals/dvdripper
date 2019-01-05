package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

// CmdPacket is the data structure for talking to/from the remote end
type CmdPacket struct {
	Command string          `json:"cmd"`
	Payload json.RawMessage `json:"payload"`
}

// RipTrack holds details about tracks to be ripped
type RipTrack struct {
	Track    int    `json:"track"`
	Filename string `json:"filename"`
}

type client struct {
	conn      *websocket.Conn
	out       chan CmdPacket
	control   chan bool
	interrupt chan bool
}

func (c *client) writeHandler() {
	for {
		select {
		case cmd := <-c.out:
			out, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			json.NewEncoder(out).Encode(cmd)
			out.Close()
		case <-c.control:
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8096,
	WriteBufferSize: 8096,
}

func buildErrorResponse(msg string) (result CmdPacket) {
	result.Command = "error"
	j, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ERROR: Can't convert error to json: %v", err)
		return
	}
	result.Payload = j
	return
}

func (c *client) send(cmd string, payload interface{}) error {
	j, err := json.Marshal(payload)
	if err != nil {
		log.Printf("WARN: Can't convert %v to json: %v", payload, err)
		return err
	}

	c.out <- CmdPacket{
		Command: cmd,
		Payload: j,
	}

	return nil
}

func (c *client) doScan() error {
	disk := DVD{}
	disk.scan()

	err := c.send("scan", disk)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) doInterupt() error {
	c.interrupt <- true
	return nil
}

func (c *client) doTidy() error {
	const base = "wwwroot/rips"

	dir, err := os.Open(base)
	if err != nil {
		return err
	}

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	log.Println("DEBUG: tidy")

	for _, n := range names {
		log.Printf("DEBUG: deleting %s", n)
		if err := os.Remove(base + "/" + n); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) doRip(payload json.RawMessage) error {

	tracks := []RipTrack{}

	err := json.Unmarshal(payload, &tracks)
	if err != nil {
		return err
	}

	log.Print("Ripping:", tracks)
	for _, track := range tracks {
		m := mplayer{
			progress: make(chan DVDProgress),
		}

		if err := c.send("rip-started", track); err != nil {
			return err
		}

		go m.rip(track.Track, "wwwroot/rips/"+track.Filename)

	outer:
		for {
			select {
			case update, ok := <-m.progress:
				if !ok {
					break outer
				}
				if err := c.send("rip-progress", update); err != nil {
					return err
				}
			case <-c.interrupt:
				m.interrupt <- true
				if err := c.send("rip-interrupted", track); err != nil {
					return err
				}
				return nil
			}
		}

		if err := c.send("rip-completed", track); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) readHandler(in io.Reader) error {
	raw, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	log.Printf("Got packet %s", raw)

	var cmd CmdPacket
	err = json.Unmarshal(raw, &cmd)
	if err != nil {
		return err
	}

	switch cmd.Command {
	case "scan":
		err = c.doScan()
	case "rip":
		err = c.doRip(cmd.Payload)
	case "interrupt":
		err = c.doInterupt()
	case "eject":
		err = c.doEject()
	case "tidy":
		err = c.doTidy()
	default:
		c.out <- buildErrorResponse("Unknown command: " + cmd.Command)
	}

	if err != nil {
		return err
	}
	return nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Websocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket: Upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	client := client{
		conn:      conn,
		out:       make(chan CmdPacket),
		control:   make(chan bool),
		interrupt: make(chan bool),
	}

	defer close(client.control)
	go client.writeHandler()

	shutdown := make(chan bool)
	go client.pushFreespace(shutdown)
	defer func() { shutdown <- true }()

	for {
		messageType, in, err := conn.NextReader()
		if err != nil {
			log.Printf("INFO: Client closed")
			client.control <- true
			return
		}

		switch messageType {
		case websocket.TextMessage:
			err := client.readHandler(in)
			if err != nil {
				log.Print("WARN: Client read error", err)
				client.control <- true
				return
			}
		}

	}
}

func startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", wsHandler)
	mux.Handle("/", http.FileServer(http.Dir("wwwroot")))

	log.Print(http.ListenAndServe(":8080", mux))
}

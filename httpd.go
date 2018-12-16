package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// OutboundResponse is the response to the client
type OutboundResponse struct {
	// Message
	Message string          `json:"message"`
	Payload json.RawMessage `json:"payload"`
}

type client struct {
	conn    *websocket.Conn
	out     chan OutboundResponse
	control chan bool
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

func buildErrorResponse(msg string) (result OutboundResponse) {
	result.Message = "error"
	j, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ERROR: Can't convert error to json: %v", err)
		return
	}
	result.Payload = j
	return
}

func (c *client) readHandler(in io.Reader) error {
	raw, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	cmd := string(raw)
	switch cmd {
	case "scan":
		disk := DVD{}
		disk.scan()

		j, err := json.Marshal(disk)
		if err != nil {
			log.Printf("WARN: Can't convert %v to json: %v", disk, err)
			return nil
		}

		c.out <- OutboundResponse{
			Message: "scan",
			Payload: j,
		}
	default:
		c.out <- buildErrorResponse("Unknown command: " + cmd)
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
		conn:    conn,
		out:     make(chan OutboundResponse),
		control: make(chan bool),
	}

	go client.writeHandler()

	for {
		messageType, in, err := conn.NextReader()
		if err != nil {
			log.Printf("INFO: Client closed")
			client.control <- true
			return
		}

		log.Printf("Got a message from client type %d", messageType)

		switch messageType {
		case websocket.TextMessage:
			err := client.readHandler(in)
			if err != nil {
				log.Printf("WARN: Client read error")
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

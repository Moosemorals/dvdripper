package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// InboundCommand holds commands from the client
type InboundCommand string

// OutboundResponse is the response to the client
type OutboundResponse struct {
	// Message
	Message string
	Payload json.RawMessage
}

type client struct {
	conn    *websocket.Conn
	out     chan OutboundResponse
	control chan bool
}

func (c *client) handler() {
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

	go client.handler()

	for {
		var cmd InboundCommand
		if err := conn.ReadJSON(&cmd); err != nil {
			log.Printf("ERROR: Can't parse json from client")
			return
		}

		switch cmd {
		case "Scan":
			disk := DVD{}
			disk.scan()

			j, err := json.Marshal(disk)
			if err != nil {
				log.Printf("WARN: Can't convert %v to json: %v", disk, err)
				continue
			}

			client.out <- OutboundResponse{
				Message: "scan",
				Payload: j,
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

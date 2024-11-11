package sse

import (
	"log"
)

// It keeps a list of clients those are currently attached
// and broadcasting events to those clients.
type Stream struct {
	// Events are pushed to this channel by the main events-gathering routine
	Messages chan string

	// New client connections
	NewClients chan chan string

	// Closed client connections
	ClosedClients chan chan string

	// Total client connections
	TotalClients map[chan string]bool
}

// New event messages are broadcast to all registered client connection channels
type ClientChan chan string

// Initialize event and Start procnteessing requests
func NewStream() (stream *Stream) {
	stream = &Stream{
		Messages:      make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}

	go stream.listen()

	return
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (stream *Stream) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Messages:
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (stream *Stream) Publish(message string) {
	stream.Messages <- message
}

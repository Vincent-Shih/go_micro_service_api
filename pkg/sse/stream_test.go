package sse_test

import (
	"go_micro_service_api/pkg/sse"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddClient(t *testing.T) {
	stream := sse.NewStream()

	client := make(sse.ClientChan)
	stream.NewClients <- client

	assert.True(t, stream.TotalClients[client])
	assert.Equal(t, 1, len(stream.TotalClients))
}

func TestDeleteClient(t *testing.T) {
	stream := sse.NewStream()

	client := make(sse.ClientChan)
	stream.NewClients <- client
	stream.ClosedClients <- client
	// TODO: better way for the goroutine to finish
	time.Sleep(10 * time.Millisecond)

	assert.Empty(t, stream.NewClients)
	assert.Empty(t, stream.ClosedClients)
	assert.Empty(t, stream.TotalClients)
}

func TestPublish(t *testing.T) {
	stream := sse.NewStream()

	client1 := make(sse.ClientChan)
	client2 := make(sse.ClientChan)
	stream.NewClients <- client1
	stream.NewClients <- client2

	stream.Publish("test")

	assert.Equal(t, "test", <-client1)
	assert.Equal(t, "test", <-client2)
}

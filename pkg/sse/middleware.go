package sse

import (
	"github.com/gin-gonic/gin"
)

const ContextKey = "clientChan"

func StreamMiddleware(stream *Stream) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		c.Set(ContextKey, clientChan)

		c.Next()
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func GetChannel(c *gin.Context) ClientChan {
	v, ok := c.Get(ContextKey)
	if !ok {
		return nil
	}

	clientChan, ok := v.(ClientChan)
	if !ok {
		return nil
	}

	return clientChan
}

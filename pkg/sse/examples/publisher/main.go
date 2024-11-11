package main

import (
	"fmt"
	"go_micro_service_api/pkg/sse"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	s := sse.NewStream()
	r := gin.Default()

	// publisher
	go func() {
		for {
			time.Sleep(time.Second * 1)
			now := time.Now().Format("2006-01-02 15:04:05")
			currentTime := fmt.Sprintf("The Current Time Is %v", now)

			// Send current time to clients message channel
			s.Messages <- currentTime
		}
	}()

	// publish endpoint
	r.GET("/sse", sse.HeadersMiddleware(), sse.StreamMiddleware(s), func(c *gin.Context) {
		clientChan := sse.GetChannel(c)
		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	})

	// if you want to test through browser
	// put it here, so no CORS for this demo code
	r.StaticFile("/", "./index.html")

	r.Run(":8080")
}

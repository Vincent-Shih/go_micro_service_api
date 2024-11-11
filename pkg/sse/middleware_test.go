package sse_test

import (
	"bufio"
	"context"
	"go_micro_service_api/pkg/helper"
	"go_micro_service_api/pkg/sse"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHeadersMiddleware(t *testing.T) {
	w := helper.CreateTestResponseRecorder()
	c, _ := gin.CreateTestContext(w)

	sse.HeadersMiddleware()(c)

	header := w.Header()
	assert.Equal(t, "text/event-stream", header.Get("Content-Type"))
	assert.Equal(t, "no-cache", header.Get("Cache-Control"))
	assert.Equal(t, "keep-alive", header.Get("Connection"))
	assert.Equal(t, "chunked", header.Get("Transfer-Encoding"))
}

func TestStreamMiddleware(t *testing.T) {
	stream := sse.NewStream()
	w := helper.CreateTestResponseRecorder()
	c, _ := gin.CreateTestContext(w)

	sse.StreamMiddleware(stream)(c)

	// Check if the client channel is set
	clientChan := sse.GetChannel(c)
	assert.NotNil(t, clientChan)
}

func TestServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	stream := sse.NewStream()
	router := gin.New()
	gin.SetMode(gin.TestMode)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				stream.Messages <- "Hello"
			}
		}
	}()

	router.GET("/sse", sse.HeadersMiddleware(), sse.StreamMiddleware(stream), func(c *gin.Context) {
		// stopping after 2 messages
		count := 0

		clientChan := sse.GetChannel(c)
		c.Stream(func(w io.Writer) bool {
			count++
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				if count <= 2 {
					return true
				}
				cancel()
				c.Abort()
			}
			return false
		})
	})

	// w := helper.CreateTestResponseRecorder()
	// router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sse", nil))
	go router.Run(":33333")
	res, err := http.Get("http://localhost:33333/sse")
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	// defer w.Result().Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	scanner := bufio.NewScanner(res.Body)
	scanner.Scan()
	assert.Equal(t, []byte("event:message"), scanner.Bytes())
	scanner.Scan()
	assert.Equal(t, []byte("data:Hello"), scanner.Bytes())
}

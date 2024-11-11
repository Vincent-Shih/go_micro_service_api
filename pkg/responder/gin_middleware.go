package responder

import (
	"go_micro_service_api/pkg/cus_err"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// _responseKey is the key to store the response in the gin context.
var _responseKey = "response"

// GinResponser is a middleware that handles the response in a unified way.
// It processes both errors and successful responses, ensuring a consistent
// format across all API endpoints.
//
// Usage:
//
//	router := gin.Default()
//	router.Use(responder.GinResponser())
//
// The middleware performs the following actions:
//  1. Handles any errors that occurred during request processing.
//  2. Sends the response data if it was set in the context.
//  3. Returns a 404 Not Found if no response or error was set.
func GinResponser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		span := trace.SpanContextFromContext(c.Request.Context())
		traceID := span.TraceID().String()

		// Handle errors
		if len(c.Errors) > 0 {
			// Get the last error from context
			// if the error is a CusError, return the error response
			// else return the 500 error response.
			err := c.Errors.Last().Err
			if kgsErr, ok := err.(*cus_err.CusError); ok {
				c.JSON(kgsErr.HttpCode(), Error(kgsErr).toGinH(traceID))
			} else {
				res := UnknownError(err)
				c.JSON(res.HttpCode(), res.toGinH(traceID))
			}
			c.Abort()
			return

		}

		// If the response has been written, just return.
		if c.Writer.Written() {
			c.Abort()
			return
		}

		// Handle response data
		if res, exists := c.Get(_responseKey); exists {
			if r, ok := res.(*Response); ok {
				c.JSON(r.HttpCode(), r.toGinH(traceID))
				c.Abort()
				return
			}
		}

		// Return 404 if no response.
		notFoundErr := cus_err.New(cus_err.ResponseNotFound, "no response")
		c.JSON(http.StatusNotFound, Error(notFoundErr).toGinH(traceID))
	}
}

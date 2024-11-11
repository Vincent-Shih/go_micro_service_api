package security_test

import (
	"go_micro_service_api/frontend_api/internal/config"
	"go_micro_service_api/frontend_api/internal/middleware/security"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Host: config.Host{
			ServiceDomains: []string{"https://example.com", "https://example2.com"},
		},
	}
	r := gin.New()
	r.Use(security.NewCORSMiddleware(cfg))
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
	}{
		{
			name:           "origin is in the ACAO list",
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
		},
		{
			name:           "origin is not in the ACAO list",
			origin:         "https://example3.com",
			expectedOrigin: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodOptions, "/", nil)
			req.Header.Set("Origin", tt.origin)
			r.ServeHTTP(w, req)

			header := w.Result().Header

			assert.Equal(t, tt.expectedOrigin, header.Get("Access-Control-Allow-Origin"))
		})
	}
}

package rate_limiter

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

// errorHandler is a simple middleware that handles errors for this test.
func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		cusErr, ok := err.Err.(*cus_err.CusError)
		if ok {
			c.Status(cusErr.HttpCode())
			return
		} else {
			c.Status(http.StatusInternalServerError)
			return
		}

	}
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(mock redismock.ClientMock)
		options        []Option
		expectedStatus int
	}{
		{
			name: "Allow request within limit",
			setupMock: func(mock redismock.ClientMock) {
				mock.ExpectIncr("test:rate_limit:127.0.0.1").SetVal(1)
				mock.ExpectExpire("test:rate_limit:127.0.0.1", time.Second).SetVal(true)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Block request exceeding limit",
			setupMock: func(mock redismock.ClientMock) {
				mock.ExpectIncr("test:rate_limit:127.0.0.1").SetVal(101)
				mock.ExpectExpire("test:rate_limit:127.0.0.1", time.Second).SetVal(true)
			},
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "Allow request for bypassed IP",
			setupMock: func(mock redismock.ClientMock) {
				// No Redis calls expected
			},
			options:        []Option{WithByPassIPs("127.0.0.1")},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Allow request for bypassed path",
			setupMock: func(mock redismock.ClientMock) {
				// No Redis calls expected
			},
			options:        []Option{WithByPassPaths("/test")},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			tt.setupMock(mock)

			r := gin.New()
			r.Use(errorHandler())
			r.Use(RateLimitMiddleware("test", db, tt.options...))

			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345" // Set IP address
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

		})
	}
}

func TestIsBypassPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		bypassPaths []string
		expected    bool
	}{
		{
			name:        "Exact match",
			path:        "/api/v1/user",
			bypassPaths: []string{"/api/v1/user"},
			expected:    true,
		},
		{
			name:        "Wildcard match",
			path:        "/api/v1/user/profile",
			bypassPaths: []string{"/api/v1/user/*"},
			expected:    true,
		},
		{
			name:        "No match",
			path:        "/api/v1/product",
			bypassPaths: []string{"/api/v1/user", "/api/v1/order/*"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBypassPath(tt.path, tt.bypassPaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestCheckRateLimit(t *testing.T) {
	ctx := context.Background()

	t.Run("Within rate limit", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cfg := &config{
			redisClient: db,
			maxRequests: 5,
			interval:    time.Minute,
		}
		key := "test_key"

		mock.ExpectIncr(key).SetVal(1)
		mock.ExpectExpire(key, time.Minute).SetVal(true)

		allowed, err := checkRateLimit(ctx, cfg, key)

		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Exceeds rate limit", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cfg := &config{
			redisClient: db,
			maxRequests: 5,
			interval:    time.Minute,
		}
		key := "test_key"

		mock.ExpectIncr(key).SetVal(6)

		allowed, err := checkRateLimit(ctx, cfg, key)

		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Redis error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cfg := &config{
			redisClient: db,
			maxRequests: 5,
			interval:    time.Minute,
		}
		key := "test_key"

		mock.ExpectIncr(key).SetErr(assert.AnError)

		allowed, err := checkRateLimit(ctx, cfg, key)

		assert.Error(t, err)
		assert.False(t, allowed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Expire error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		cfg := &config{
			redisClient: db,
			maxRequests: 5,
			interval:    time.Minute,
		}
		key := "test_key"

		mock.ExpectIncr(key).SetVal(1)
		mock.ExpectExpire(key, time.Minute).SetErr(assert.AnError)
		mock.ExpectDecr(key).SetVal(0)

		allowed, err := checkRateLimit(ctx, cfg, key)

		assert.Error(t, err)
		assert.False(t, allowed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

package rate_limiter

import (
	"context"
	"fmt"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Constants for default values and Redis key prefix
const (
	_defaultInterval    = time.Second
	_defaultMaxRequests = 100
	_redisKeyPrefix     = "rate_limit"
)

// config holds the configuration for the rate limiter
type config struct {
	redisClient *redis.Client
	interval    time.Duration
	maxRequests int64
	byPassIPs   map[string]struct{}
	byPassPaths []string
}

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithInterval sets the interval at which the rate limiter should reset the request count.
func WithInterval(interval time.Duration) Option {
	return optionFunc(func(c *config) {
		c.interval = interval
	})
}

func WithMaxRequests(maxRequests int64) Option {
	return optionFunc(func(c *config) {
		c.maxRequests = maxRequests
	})
}

func WithByPassIPs(ips ...string) Option {
	return optionFunc(func(c *config) {
		c.byPassIPs = make(map[string]struct{})
		for _, ip := range ips {
			c.byPassIPs[ip] = struct{}{}
		}
	})
}

func WithByPassPaths(paths ...string) Option {
	return optionFunc(func(c *config) {
		c.byPassPaths = paths
	})
}

// RateLimitMiddleware creates a Gin middleware for rate limiting.
//
// Parameters:
//   - serviceName: A unique identifier for the pod (used in Redis key generation)
//   - redisClient: A pointer to a redis.Client for storing rate limit data
//   - opts: Optional configuration options
//
// Returns:
//   - gin.HandlerFunc: A middleware function for Gin
//
// The middleware checks each request against the rate limit. If the limit is exceeded,
// it aborts the request with a "Too Many Requests" error. IP addresses and paths can be
// configured to bypass the rate limit check.
// Example:
//
//		r := gin.New()
//		r.Use(RateLimitMiddleware("service_name",
//			redisClient,
//			WithInterval(time.Second),
//			WithMaxRequests(100),
//			WithByPassIPs("127.0.0.1"),
//			WithByPassPaths("/api/v1/public/*"),
//		 ))
//	 // Skip...
func RateLimitMiddleware(serviceName string, redisClient *redis.Client, opts ...Option) gin.HandlerFunc {
	cfg := &config{
		redisClient: redisClient,
		interval:    _defaultInterval,
		maxRequests: _defaultMaxRequests,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	if cfg.redisClient == nil {
		log.Fatal("redis client is nil")
	}

	return func(c *gin.Context) {
		// Get context from the request
		ctx := c.Request.Context()

		// Start a new trace span
		ctx, span := cus_otel.StartTrace(ctx)
		defer span.End()

		ip := c.ClientIP()
		key := fmt.Sprintf("%s:%s:%s", serviceName, _redisKeyPrefix, ip)

		// Check if the IP is in the bypass list
		if _, ok := cfg.byPassIPs[c.ClientIP()]; ok {
			c.Next()
			return
		}

		// Check if the path is in the bypass list
		if isBypassPath(c.FullPath(), cfg.byPassPaths) {
			c.Next()
			return
		}

		// Increment the request count and set the expiration time
		allowed, err := checkRateLimit(ctx, cfg, key)
		if err != nil {
			kgsErr := cus_err.New(cus_err.InternalServerError, err.Error())
			cus_otel.Error(ctx, kgsErr.Error())
			c.Abort()
			return
		}

		// If the request count exceeds the limit, return a error
		if !allowed {
			kgsErr := cus_err.New(cus_err.TooManyRequests, "Rate limit exceeded")
			cus_otel.Error(ctx, kgsErr.Error())
			c.Errors = append(c.Errors, c.Error(kgsErr))
			c.Abort()
			return
		}

		c.Next()
	}
}

// isBypassPath checks if the given path should bypass authentication.
//
// This function determines whether a given path matches any of the bypass patterns.
// It supports exact matches and wildcard matches using the "*" character.
//
// Parameters:
//   - path: The path to check against the bypass list.
//   - bypass: A slice of string patterns representing paths that should bypass authentication.
//
// Returns:
//   - bool: true if the path matches any bypass pattern, false otherwise.
//
// Examples:
//
//	isBypassPath("/api/v1/user", []string{"/api/v1/user"}) // Returns: true
//	isBypassPath("/api/v1/user/get", []string{"/api/v1/user/*"}) // Returns: true
//	isBypassPath("/api/v1/user/get", []string{"/api/v1/user"}) // Returns: false
//	isBypassPath("/api/v1/public/status", []string{"/api/v1/public/*"}) // Returns: true
func isBypassPath(path string, bypass []string) bool {
	for _, pattern := range bypass {
		if pattern == path {
			return true
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}

// checkRateLimit checks if the current request is within the rate limit.
//
// This function increments the request count in Redis and checks if it's within the limit.
// It only sets the expiration time if the key is new.
//
// Parameters:
//   - ctx: The context for the Redis operations
//   - cfg: The rate limiter configuration
//   - key: The Redis key for the current rate limit bucket
//
// Returns:
//   - bool: true if the request is allowed, false if it exceeds the rate limit
//   - error: any error that occurred during the Redis operations
func checkRateLimit(ctx context.Context, cfg *config, key string) (bool, error) {
	// Use INCR command and get the new value
	newCount, err := cfg.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// If this is the first request (newCount == 1), set the expiration
	if newCount == 1 {
		err = cfg.redisClient.Expire(ctx, key, cfg.interval).Err()
		if err != nil {
			// If setting expiration fails, we should decrement the counter to maintain consistency
			cfg.redisClient.Decr(ctx, key)
			return false, err
		}
	}

	// Check if the new count exceeds the limit
	return newCount <= cfg.maxRequests, nil
}

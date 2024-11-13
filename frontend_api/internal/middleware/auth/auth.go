package auth

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// userInfoKey is the key used to store user information in the Gin context.
const userInfoKey = "userInfo"

// Filter is a function that determines if a request should be allowed to pass through
// (*gin.context) has the FullPath(),we can make good use of it
type Filter func(*gin.Context) bool

// AuthClient is an interface for validating access tokens.
type AuthClient interface {
	ValidToken(ctx context.Context, token string) (*UserInfo, *cus_err.CusError)
}

type config struct {
	authClient AuthClient
	filter     []Filter
	bypass     []string
}

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithFilter adds one or more Filter functions to the middleware configuration.
// when filter return true, then skip the middleware
func WithFilter(f ...Filter) Option {
	return optionFunc(func(c *config) {
		c.filter = append(c.filter, f...)
	})
}

// ByPassPath adds paths that should bypass authentication.
//
// It allows the caller to specify one or more paths that should not be subject
// to authentication checks. This function supports both exact matches and
// wildcard patterns using the "*" character.
//
// Parameters:
//   - paths: A variadic list of string patterns representing paths to bypass.
//
// Returns:
//   - Option: A function that modifies the middleware configuration.
//
// Usage:
//
//	middleware.AuthMiddleware(
//	    oauthClient,
//	    middleware.ByPassPath("/api/v1/public/*"),
//	    middleware.ByPassPath("/health", "/metrics"),
//	)
//
// Examples:
//   - "/api/v1/public/*" bypasses all paths that start with "/api/v1/public/"
//   - "/api/v1/user" only bypasses the exact path "/api/v1/user"
//   - "/api/v1/user/:id" bypasses "/api/v1/user/:id" (useful with path parameters)
func ByPassPath(paths ...string) Option {
	return optionFunc(func(c *config) {
		c.bypass = append(c.bypass, paths...)
	})
}

// AuthMiddleware creates a Gin middleware for OAuth authentication.
//
// It validates the access token for each request, bypassing specified paths and
// applying custom filters if configured.
//
// Parameters:
//   - authClient: An AuthClient implementation for validating access tokens.
//   - opts: Optional configuration options for the middleware.
//
// Returns:
//   - gin.HandlerFunc: A Gin middleware function.
//
// Usage:
//
//	// Can use auth middleware with bypass path or filter
//	router.Use(
//		middleware.AuthMiddleware(oauthClient,
//		middleware.ByPassPath("api/v1/oauth/*"), // Optional
//		middleware.WithFilter(func(c *gin.Context) bool { // Optional
//			// Custom filter logic
//			return true
//		}),
//	))
func AuthMiddleware(oauthClient AuthClient, opts ...Option) gin.HandlerFunc {
	if oauthClient == nil {
		log.Fatal("oauth client is nil")
	}

	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	cfg.authClient = oauthClient

	return func(c *gin.Context) {
		// If any of the bypass path matches the current path, then skip the middleware
		if isBypassPath(c.FullPath(), cfg.bypass) {
			c.Next()
			return
		}

		// If any of the filter returns true, then skip the middleware
		for _, f := range cfg.filter {
			if f(c) {
				c.Next()
				return
			}
		}

		// Get access token from Authorization header
		var token string
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) == 2 {
				token = splitToken[1]
			}
		}

		// If the token is empty, return an error
		if token == "" {
			cusErr := cus_err.New(cus_err.MissingAccessToken, "Access token not found in header", nil)
			cus_otel.Warn(c.Request.Context(), cusErr.Error())
			_ = c.Error(cusErr)
			c.Abort()
			return
		}

		// Validate the token
		userInfo, err := cfg.authClient.ValidToken(c.Request.Context(), token)
		if err != nil {
			cus_otel.Warn(c.Request.Context(), err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		if userInfo == nil {
			cusErr := cus_err.New(cus_err.InternalServerError, "UserInfo is nil")
			cus_otel.Warn(c.Request.Context(), cusErr.Error())
			_ = c.Error(cusErr)
			c.Abort()
			return
		}

		// Set the user info in the context
		c.Set(userInfoKey, userInfo)

		// Continue to the next middleware
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
//	isBypassPath("/api/v1/pubic/status/1", []string{"/api/v1/*/status"}) // Returns: true
func isBypassPath(path string, bypass []string) bool {
	trie := NewTrie()
	for _, pattern := range bypass {
		trie.Insert(pattern)
	}
	return trie.Search(path)
}

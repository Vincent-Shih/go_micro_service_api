package auth

import (
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/enum"

	"github.com/gin-gonic/gin"
)

// permConfig is the configuration for the permission guard middleware.
type permConfig struct {
	needPerms []enum.Permission // List of permissions required to access the endpoint.
}

type PermOption interface {
	apply(*permConfig)
}

type permOptionFunc func(*permConfig)

func (p permOptionFunc) apply(c *permConfig) {
	p(c)
}

// WithPerms is a middleware option that specifies the permissions required to access the endpoint.
func WithPerms(perms ...enum.Permission) PermOption {
	return permOptionFunc(func(c *permConfig) {
		c.needPerms = perms
	})
}

// Guard is a middleware that checks if the user has the required permissions to access the endpoint.
// If the user does not have the required permissions or roles, the middleware returns an error.
//
// Parameters:
//   - opts: A variadic list of PermOption functions that specify the required permissions and roles.
//
// Returns:
//   - gin.HandlerFunc: A middleware function that checks the user's permissions and roles.
//
// Validation Rules:
//  1. WithPerms: The user must have ALL the specified permissions for it to return true.
//  2. If both WithPerms and WithRoles are used, BOTH conditions must be met for it to return true.
//
// Usage:
//
//	// Example 1: Protect an endpoint
//	router.GET("/protected",
//		auth.Guard(auth.WithPerms(auth.Deposit)),
//		handler)
//
//	// Example 2: Protect a group of endpoints
//	g := router.Group("/protected",
//		auth.Guard(
//			auth.WithPerms(auth.Deposit),
//			auth.WithRoles(auth.Admin))
//	)
func Guard(opts ...PermOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start trace
		ctx, span := cus_otel.StartTrace(c.Request.Context())
		defer span.End()

		cfg := &permConfig{}
		for _, opt := range opts {
			opt.apply(cfg)
		}

		// Get the user information from the context
		userInfo, ok := GetUserInfo(c)
		if !ok {
			kgsErr := cus_err.New(cus_err.InternalServerError, "failed to get user info from gin.")
			cus_otel.Error(ctx, kgsErr.Error())
			_ = c.Error(kgsErr)
			c.Abort()
			return
		}

		// Check if the user has the required permissions
		if len(cfg.needPerms) > 0 && !userInfo.HasPermission(cfg.needPerms...) {
			kgsErr := cus_err.New(cus_err.NoPermission, "user does not have the required permissions.")
			cus_otel.Error(ctx, kgsErr.Error())
			_ = c.Error(kgsErr)
			c.Abort()
			return
		}

		c.Next()

	}
}

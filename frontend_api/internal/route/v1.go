package route

import (
	v1_handler "go_micro_service_api/frontend_api/internal/api/v1"
	"go_micro_service_api/frontend_api/internal/config"
	"go_micro_service_api/frontend_api/internal/middleware/auth"
	"go_micro_service_api/pkg/helper"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// RouteV1 is a struct implementing the Route interface
// Version 1 of the API
type RouteV1 struct {
	authClient    auth.AuthClient // This client is used to validate token for the middleware
	authHandler   *v1_handler.AuthHandler
	userHandler   *v1_handler.UserHandler
	verifyHandler *v1_handler.VerifyHandler
	cfg           *config.Config
	// Add other handlers here
}

var _ Route = (*RouteV1)(nil)

// NewRouteV1Set creates a new fx.Option for the RouteV1 module
func NewRouteV1Set() fx.Option {
	return fx.Module("route-v1",
		fx.Provide(
			newRouteV1,
			v1_handler.NewAuthHandler,
			v1_handler.NewUserHandler,
			v1_handler.NewVerifyHandler,
			helper.NewMachineID,
			helper.NewSnowflake,
			// Add other handlers here
		),
	)
}

// newRouteV1 creates a new RouteV1 instance for the Route interface
func newRouteV1(
	cfg *config.Config,
	authClient auth.AuthClient,
	authHandler *v1_handler.AuthHandler,
	userHandler *v1_handler.UserHandler,
	verifyHandler *v1_handler.VerifyHandler,
) Route {
	return &RouteV1{
		authClient:    authClient,
		authHandler:   authHandler,
		userHandler:   userHandler,
		verifyHandler: verifyHandler,

		cfg: cfg,
	}
}

func (r *RouteV1) RegisterRoutes(g *gin.Engine) {
	g.Use(auth.AuthMiddleware(
		r.authClient,
		auth.ByPassPath("/swagger/v1/*any", "/api/v1/auth", "/api/v1/auth/*"),
		// auth.WithFilter(func(c *gin.Context) bool {
		// 	host := c.Request.Host
		// 	return host == r.cfg.ServiceUrl
		// }),
	))

	v1 := g.Group("/api/v1")
	addSwaggerRouters(g)
	r.addAuthRoutes(v1)
	r.addUserRoutes(v1)
}

func (r *RouteV1) addAuthRoutes(g *gin.RouterGroup) {
	auth := g.Group("/auth")
	auth.GET("", r.authHandler.ClientAuth)
}

func (r *RouteV1) addUserRoutes(g *gin.RouterGroup) {
	auth := g.Group("/users")
	auth.POST("", r.userHandler.CreateUser)
	auth.GET("/verificationCode", r.verifyHandler.RegisterVerification)
	auth.POST("/verification", r.verifyHandler.Verification)
	auth.GET("/existence", r.userHandler.CheckUserExistence)
	auth.POST("/login", r.authHandler.Login)
}

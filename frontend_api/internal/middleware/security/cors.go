package security

import (
	"go_micro_service_api/frontend_api/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewCORSMiddleware(envCfg *config.Config) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = envCfg.ServiceDomains
	config.AllowCredentials = true
	config.AllowHeaders = append(config.AllowHeaders, "client_id", "withcredentials", "Authorization")

	return cors.New(config)
}

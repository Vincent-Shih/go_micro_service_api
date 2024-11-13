package main

import (
	_ "go_micro_service_api/frontend_api/docs"
	"go_micro_service_api/frontend_api/internal/config"
	"go_micro_service_api/frontend_api/internal/infrastructure/grpc_client"
	httpserver "go_micro_service_api/frontend_api/internal/infrastructure/http_server"
	"go_micro_service_api/frontend_api/internal/infrastructure/redis_initializer"
	"go_micro_service_api/frontend_api/internal/route"
	"net/http"

	"go.uber.org/fx"
)

// @title           Sample Frontend API
// @version         0.1
// @description     This is the Sample Frontend API
// @termsOfService  http://swagger.io/terms/
// @contact.name    Sample
// @contact.url     http://www.swagger.io/support
// @license.name    Belong to Sample
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath        /api
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @schemes         http https
// @externalDocs.description  	OpenAPI
// @externalDocs.url          	https://swagger.io/resources/open-api/
func main() {

	fx.New(
		grpc_client.NewGrpcClientSet(),
		route.NewRouteV1Set(),
		fx.Provide(
			config.NewConfig,
			httpserver.NewHttpServer,
			redis_initializer.NewRedisClient,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()

}

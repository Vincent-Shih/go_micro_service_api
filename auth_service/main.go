package main

import (
	"go_micro_service_api/auth_service/internal/application"
	"go_micro_service_api/auth_service/internal/domain/repository"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/infrastructure/db_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/grpc_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/redis_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/token_helper"
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/pkg/req_analyzer"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func main() {
	fx.New(
		db_impl.NewEntDbFx(),
		fx.Provide(
			grpc_impl.NewGrpcServer,
			application.NewAuthService,
			application.NewClientService,
			application.NewUserService,
			service.NewAuthService,
			service.NewClientService,
			service.NewUserService,
			fx.Annotate(
				ent_impl.NewClientRepoImpl,
				fx.As(new(repository.ClientRepo)),
			),
			fx.Annotate(
				ent_impl.NewUserRepoImpl,
				fx.As(new(repository.UserRepo)),
			),
			fx.Annotate(
				token_helper.NewJwtToken,
				fx.As(new(token_helper.TokenHelper)),
			),
			fx.Annotate(
				redis_cache.NewRedisCache,
				fx.As(new(db.Cache)),
			),
			redis_impl.NewRedisClient,
			req_analyzer.NewReqAnalyzer,
		),
		fx.Invoke(func(server *grpc.Server) {}),
	).Run()

}

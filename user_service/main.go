package main

import (
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/user_service/internal/application"
	"go_micro_service_api/user_service/internal/domain/service"
	"go_micro_service_api/user_service/internal/infrastructure/db_impl"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/user_service/internal/infrastructure/grpc_impl"
	"go_micro_service_api/user_service/internal/infrastructure/redis_impl"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func main() {
	fx.New(
		db_impl.NewEntDbFx(),
		fx.Provide(
			grpc_impl.NewGrpcServer,
			application.NewUserService,
			application.NewVerifyService,
			service.NewUserService,
			service.NewVerifyService,
			ent_impl.NewUserRepo,
			redis_impl.NewVerifyRepo,
			fx.Annotate(
				redis_cache.NewRedisCache,
				fx.As(new(db.Cache)),
			),
			redis_impl.NewRedisClient,
		),
		fx.Invoke(
			func(server *grpc.Server) {},
		),
	).Run()

}

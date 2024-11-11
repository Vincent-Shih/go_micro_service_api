package grpc_client

import (
	"go_micro_service_api/frontend_api/internal/middleware/auth"

	"go.uber.org/fx"
)

func NewGrpcClientSet() fx.Option {
	return fx.Module("grpc-client",
		fx.Provide(
			NewAuthClient,
			NewUserClient,
			fx.Annotate(
				func(a *AuthClient) auth.AuthClient { return a },
			),
			// Add other clients here
		),
	)
}

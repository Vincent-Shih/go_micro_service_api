package grpc_impl

import (
	"context"
	"fmt"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	otelgrpc "go_micro_service_api/pkg/cus_otel/grpc"
	"go_micro_service_api/user_service/internal/application"
	"go_micro_service_api/user_service/internal/config"
	"log"
	"net"

	"go_micro_service_api/pkg/pb/gen/user"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func NewGrpcServer(lc fx.Lifecycle, userService *application.UserService, verifyService *application.VerifyService) *grpc.Server {
	// Get config
	cfg := config.GetConfig()

	// New grpc server
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.TracingMiddleware(otelgrpc.RoleServer)),
		grpc.ChainUnaryInterceptor(
			cus_err.ErrorInterceptor,
			// Any other interceptors can be added here
		),
		grpc.StreamInterceptor(
			cus_err.StreamErrorInterceptor,
			// Any other interceptors can be added here
		),
	)

	var shutdown func(context.Context) error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Init cus_otel
			_shutdown, err := cus_otel.InitTelemetry(ctx, cfg.Host.ServiceName, cfg.OtelUrl)
			if err != nil {
				return err
			}
			shutdown = _shutdown

			// Listen the port
			lis, err := net.Listen("tcp", cfg.ServiceUrl)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			// Register the service
			go func() {
				user.RegisterUserServiceServer(s, userService)
				user.RegisterVerifyServiceServer(s, verifyService)
				if err := s.Serve(lis); err != nil {
					cus_otel.Error(ctx, "failed to serve", cus_otel.NewField("error", err))
				}
			}()

			cus_otel.Info(ctx, fmt.Sprintf("gRPC server started at %s", cfg.ServiceUrl))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := shutdown(ctx)
			if err != nil {
				return err
			}

			cus_otel.Info(ctx, "gRPC server shut down gracefully")
			return nil
		},
	})

	return s
}

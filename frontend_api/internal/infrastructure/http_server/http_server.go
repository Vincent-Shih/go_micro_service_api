package httpserver

import (
	"context"
	"fmt"
	"go_micro_service_api/frontend_api/internal/config"
	"go_micro_service_api/pkg/helper"
	"os"
	"path/filepath"

	"go_micro_service_api/frontend_api/internal/middleware/security"
	"go_micro_service_api/frontend_api/internal/route"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	otelgin "go_micro_service_api/pkg/cus_otel/gin"
	"go_micro_service_api/pkg/rate_limiter"
	"go_micro_service_api/pkg/responder"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// NewHttpServer creates a new http server
func NewHttpServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	route route.Route,
	redisClient *redis.Client,
) *http.Server {

	// New gin server
	srv := http.Server{
		Addr: cfg.ServiceUrl,
	}

	var shutdown func(context.Context) error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Init cus_otel
			_shutdown, err := cus_otel.InitTelemetry(ctx, cfg.Host.ServiceName, cfg.OtelUrl)
			if err != nil {
				return err
			}
			shutdown = _shutdown

			if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
				v.RegisterValidation("one_alpha", helper.ContainsAtLeastOneAlpha)
				v.RegisterValidation("one_num", helper.ContainsAtLeastOneNum)
			}

			r := gin.New()
			// Register the middleware
			r.Use(otelgin.TracingMiddleware(cfg.ServiceName))
			r.Use(responder.GinResponser())
			r.Use(rate_limiter.RateLimitMiddleware(cfg.Host.ServiceName, redisClient,
				rate_limiter.WithInterval(time.Duration(cfg.Host.RateLimitIntervalSecs)*time.Second),
				rate_limiter.WithMaxRequests(int64(cfg.Host.RateLimitMaxRequests)),
			))

			r.Use(security.NewCORSMiddleware(cfg))

			// Register the routes
			route.RegisterRoutes(r)

			// Replace the handler
			srv.Handler = r

			go startService(ctx, r, cfg)

			cus_otel.Info(ctx, fmt.Sprintf("http server started at %s", cfg.ServiceUrl))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := shutdown(ctx)
			if err != nil {
				cus_otel.Error(ctx, "Error shutting down http server", cus_otel.NewField("error", err))
				return err
			}
			cus_otel.Info(ctx, "http server shut down gracefully")
			return nil
		},
	})

	return &srv
}

// startService starts the HTTP service using the provided Gin engine and configuration.
// It supports both TLS and non-TLS modes based on the configuration.
func startService(ctx context.Context, r *gin.Engine, cfg *config.Config) {
	if cfg.Host.EnableTLS {
		crtPath, cusErr := absPath(cfg.Host.CertFilePath)
		if cusErr != nil {
			cus_otel.Error(ctx, cusErr.Error())
			return
		}
		keyPath, cusErr := absPath(cfg.Host.KeyFilePath)
		if cusErr != nil {
			cus_otel.Error(ctx, cusErr.Error())
			return
		}
		if err := r.RunTLS(cfg.ServiceUrl, crtPath, keyPath); err != nil {
			cus_otel.Error(ctx, "Error starting http server", cus_otel.NewField("error", err))
		}
	} else {
		if err := r.Run(cfg.ServiceUrl); err != nil {
			cus_otel.Error(ctx, "Error starting http server", cus_otel.NewField("error", err))
		}
	}
}

// absPath returns the absolute path of the given path
func absPath(path string) (string, *cus_err.CusError) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "failed to get current working directory", err)
		return "", cusErr
	}

	absPath := filepath.Join(workDir, path)
	return absPath, nil
}

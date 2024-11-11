package redis_initializer

import (
	"context"
	"go_micro_service_api/frontend_api/internal/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client and returns it.
func NewRedisClient(cfg *config.Config) *redis.Client {

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.RedisUrl,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MinIdleConns:    cfg.MinIdle,
		MaxIdleConns:    cfg.MaxIdle,
		MaxActiveConns:  cfg.MaxActive,
		ConnMaxLifetime: time.Duration(cfg.ConnTimeout) * time.Second,
	})

	// Ping Redis
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	return rdb
}

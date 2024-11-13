package redis_cache

import (
	"context"
	"encoding/json"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

var _ db.Cache = (*RedisCache)(nil)

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get value from Redis
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		var errCode cus_err.CusCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = cus_err.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = cus_err.InternalServerError
		}

		cusErr := cus_err.New(errCode, "Failed to get key", err)
		cus_otel.Error(ctx, cusErr.Error())
		return "", cusErr
	}

	return val, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) *cus_err.CusError {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Set value in Redis
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		// Return InternalServerError error.
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to set key", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	return nil
}

func (r *RedisCache) GetObject(ctx context.Context, key string, dest any) *cus_err.CusError {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get value from Redis
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		// When key not found return ResourceNotFound error.
		if err == redis.Nil {
			cusErr := cus_err.New(cus_err.ResourceNotFound, "Failed to get key", err)
			cus_otel.Error(ctx, cusErr.Error())
			return cusErr
		}
		// Otherwise, return InternalServerError error.
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to get key", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	// Unmarshal value to dest
	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		// Return InternalServerError error.
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to unmarshal value", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	return nil
}

func (r *RedisCache) SetObject(ctx context.Context, key string, value any, expiration time.Duration) *cus_err.CusError {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Marshal value
	val, err := json.Marshal(value)
	if err != nil {
		// Return InternalServerError error.
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to marshal value", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	// Set value in Redis
	err = r.client.Set(ctx, key, val, expiration).Err()
	if err != nil {
		// Return InternalServerError error.
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to set key", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	return nil
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) *cus_err.CusError {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Delete key from Redis
	result, err := r.client.Del(ctx, keys...).Result()
	if err != nil {
		// Return InternalServerError error.
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to delete key", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	// If no key was deleted, return ResourceNotFound error
	if result == 0 {
		cusErr := cus_err.New(cus_err.ResourceNotFound, "Key not found", nil)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	return nil
}

// GetHash retreive whole hash object by key
// - dest: struct field needs redis tag
func (r *RedisCache) GetHash(ctx context.Context, key string, dest any) *cus_err.CusError {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	res := r.client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		var errCode cus_err.CusCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = cus_err.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = cus_err.InternalServerError
		}

		cusErr := cus_err.New(errCode, "Failed to get hash", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	if err := res.Scan(dest); err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to scan hash", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	return nil
}

// GetHashFields retreive hash fields by key
func (r *RedisCache) GetHashFields(ctx context.Context, key string, fields ...string) ([]interface{}, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	res, err := r.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		var errCode cus_err.CusCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = cus_err.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = cus_err.InternalServerError
		}

		cusErr := cus_err.New(errCode, "Failed to get hash", err)
		cus_otel.Error(ctx, cusErr.Error())
		return make([]interface{}, 0), cusErr
	}

	return res, nil
}

// SetHash set hash object by key
// - expiration: if assign 0, then no expiration on it
func (r *RedisCache) SetHash(ctx context.Context, expiration time.Duration, key string, values ...interface{}) *cus_err.CusError {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	if err := r.client.HSet(ctx, key, values...).Err(); err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to set hash", err)
		cus_otel.Error(ctx, cusErr.Error())
		return cusErr
	}

	if expiration > 0 {
		if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
			cusErr := cus_err.New(cus_err.InternalServerError, "Failed to set expiration", err)
			cus_otel.Error(ctx, cusErr.Error())
			return cusErr
		}
	}

	return nil
}

// Incr increment value of key by 1
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	res, err := r.client.Incr(ctx, key).Result()

	if err != nil {
		var errCode cus_err.CusCode
		if err == redis.Nil {
			// When key not found return ResourceNotFound error.
			errCode = cus_err.ResourceNotFound
		} else {
			// Otherwise, return InternalServerError error.
			errCode = cus_err.InternalServerError
		}

		cusErr := cus_err.New(errCode, "Failed to increment", err)
		cus_otel.Error(ctx, cusErr.Error())
		return 0, cusErr
	}

	return res, nil
}

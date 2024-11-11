package db

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, *cus_err.CusError)
	Set(ctx context.Context, key string, value string, expiration time.Duration) *cus_err.CusError
	GetObject(ctx context.Context, key string, dest any) *cus_err.CusError
	SetObject(ctx context.Context, key string, value any, expiration time.Duration) *cus_err.CusError
	Delete(ctx context.Context, keys ...string) *cus_err.CusError
	GetHash(ctx context.Context, key string, dest any) *cus_err.CusError
	GetHashFields(ctx context.Context, key string, fields ...string) ([]interface{}, *cus_err.CusError)
	SetHash(ctx context.Context, expiration time.Duration, key string, values ...interface{}) *cus_err.CusError
	Incr(ctx context.Context, key string) (int64, *cus_err.CusError)
}

package token_helper

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
)

type TokenHelper interface {
	Create(ctx context.Context, secret string, claims map[string]any) (string, *cus_err.CusError)
	Validate(ctx context.Context, token string, secret string) (map[string]any, *cus_err.CusError)
	GetPayload(ctx context.Context, token string) (map[string]any, *cus_err.CusError)
}

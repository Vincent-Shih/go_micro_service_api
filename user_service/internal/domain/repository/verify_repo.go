package repository

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/user_service/internal/domain/vo"
)

type VerifyRepo interface {
	RegisterVerification(ctx context.Context, session *vo.VerificationSession) *cus_err.CusError
	Verification(ctx context.Context, session *vo.VerificationSession) (bool, *cus_err.CusError)
}

package service

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/user_service/internal/domain/repository"
	"go_micro_service_api/user_service/internal/domain/vo"
)

type VerifyService struct {
	verifyRepo repository.VerifyRepo
}

func NewVerifyService(verifyRepo repository.VerifyRepo) *VerifyService {
	return &VerifyService{
		verifyRepo: verifyRepo,
	}
}

func (s *VerifyService) RegisterVerification(ctx context.Context, values ...string) (*vo.VerificationSession, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// generate code
	session := &vo.VerificationSession{}
	session.NextCode().NextToken(ctx, values...)

	// store code
	err := s.verifyRepo.RegisterVerification(ctx, session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func (s *VerifyService) Verification(ctx context.Context, session *vo.VerificationSession) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	res, err := s.verifyRepo.Verification(ctx, session)
	if err != nil {
		return false, err
	}

	return res, nil
}

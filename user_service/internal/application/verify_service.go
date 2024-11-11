package application

import (
	"context"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/pb/gen/user"
	"go_micro_service_api/user_service/internal/domain/service"
	"go_micro_service_api/user_service/internal/domain/vo"
)

type VerifyService struct {
	user.VerifyServiceServer
	verifyService *service.VerifyService
}

var _ user.VerifyServiceServer = (*VerifyService)(nil)

func NewVerifyService(verifyService *service.VerifyService) *VerifyService {
	return &VerifyService{
		verifyService: verifyService,
	}
}

func (s *VerifyService) RegisterVerification(ctx context.Context, req *user.RegisterVerificationRequest) (*user.RegisterVerificationResponse, error) {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	session, err := s.verifyService.RegisterVerification(ctx, req.GetType(), req.GetEmail(), req.GetCountryCode(), req.GetMobileNumber())
	if err != nil {
		return &user.RegisterVerificationResponse{}, err
	}

	return &user.RegisterVerificationResponse{
		VerificationCodePrefix: session.Prefix,
		VerificationCode:       session.Code,
		VerificationCodeToken:  session.Token,
	}, nil
}

func (s *VerifyService) Verification(ctx context.Context, req *user.VerificationRequest) (*user.VerificationResponse, error) {
	_, span := cus_otel.StartTrace(ctx)
	defer span.End()

	session := vo.NewVerificationSession(
		req.GetType(),
		req.GetVerificationCodePrefix(),
		req.GetVerificationCode(),
		req.GetVerificationCodeToken(),
	)

	res, err := s.verifyService.Verification(ctx, session)
	if err != nil {
		return &user.VerificationResponse{}, err
	}

	return &user.VerificationResponse{
		Result: res,
	}, nil
}

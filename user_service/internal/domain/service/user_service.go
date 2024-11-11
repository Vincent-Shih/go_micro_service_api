package service

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/user_service/internal/domain/aggregate"
	"go_micro_service_api/user_service/internal/domain/repository"
	"go_micro_service_api/user_service/internal/domain/vo"
)

type UserService struct {
	userRepo repository.UserRepo
}

func NewUserService(userRepo repository.UserRepo) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateProfile(ctx context.Context, user *aggregate.User) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.CreateProfile(ctx, user)
}

func (s *UserService) GetProfile(ctx context.Context, user *aggregate.User, keys []int) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.GetProfile(ctx, user, keys)
}

func (s *UserService) GetProfileFromOAuth(ctx context.Context, user *aggregate.User, session *vo.OAuthSession) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.GetProfileFromOAuth(ctx, user, session)
}

func (s *UserService) CheckMobileExistence(ctx context.Context, countryCode string, mobileNumber string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.CheckMobileExistence(ctx, countryCode, mobileNumber)
}

func (s *UserService) CheckEmailExistence(ctx context.Context, email string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.CheckEmailExistence(ctx, email)
}

func (s *UserService) IsAccountExist(ctx context.Context, account string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.IsAccountExist(ctx, account)
}

func (s *UserService) GetLoginUserInfo(ctx context.Context, identifiers map[int]string) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	uid, err := s.userRepo.GetUserIdByProfile(ctx, identifiers)
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	u, err := s.userRepo.GetProfile(
		ctx,
		&aggregate.User{ID: int64(uid)},
		[]int{enum.ProfileKey.Email.ID, enum.ProfileKey.CountryCode.ID, enum.ProfileKey.MobileNumber.ID},
	)
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	return u, nil
}

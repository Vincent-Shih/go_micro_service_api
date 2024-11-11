package application

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/pb/gen/user"
	"go_micro_service_api/user_service/internal/domain/aggregate"
	"go_micro_service_api/user_service/internal/domain/entity"
	"go_micro_service_api/user_service/internal/domain/service"
	"go_micro_service_api/user_service/internal/domain/vo"
)

type UserService struct {
	user.UserServiceServer
	userService *service.UserService
	db          db.Database
}

var _ user.UserServiceServer = (*UserService)(nil)

func NewUserService(userService *service.UserService, db db.Database) *UserService {
	return &UserService{
		userService: userService,
		db:          db,
	}
}

func (s *UserService) CreateProfile(ctx context.Context, req *user.CreateProfileRequest) (*user.CreateProfileResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	u := &aggregate.User{
		ID: req.GetId(),
		Profile: entity.Profile{
			Account:      req.GetAccount(),
			Email:        req.GetEmail(),
			CountryCode:  req.GetCountryCode(),
			MobileNumber: req.GetMobileNumber(),
		},
	}

	ctx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.userService.CreateProfile(ctx, u)
	if err != nil {
		_, rollbackErr := s.db.Rollback(ctx)
		if rollbackErr != nil {
			cus_otel.Error(ctx, rollbackErr.Error())
			err = rollbackErr
		}
		return nil, err
	}

	// Commit the transaction
	_, commitErr := s.db.Commit(ctx)
	if commitErr != nil {
		cus_otel.Error(ctx, commitErr.Error())
		return nil, commitErr
	}

	return &user.CreateProfileResponse{}, nil
}

func (s *UserService) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	u := &aggregate.User{
		ID:      req.GetId(),
		Profile: entity.Profile{},
	}

	u, err := s.userService.GetProfile(ctx, u, nil)
	if err != nil {
		return nil, err
	}

	return &user.GetProfileResponse{
		Email:        u.Profile.Email,
		CountryCode:  u.Profile.CountryCode,
		MobileNumber: u.Profile.MobileNumber,
	}, nil
}

func (s *UserService) GetProfileFromOAuth(ctx context.Context, req *user.GetProfileFromOAuthRequest) (*user.GetProfileFromOAuthResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	value := req.GetProvider()
	provider, err := enum.OAuthProviderFromString(value)
	if err != nil {
		return nil, err
	}

	session := &vo.OAuthSession{
		Provider:    provider,
		AccessToken: req.GetAccessToken(),
	}

	u := &aggregate.User{
		Profile: entity.Profile{
			ThirdParties: entity.ThridParties{
				Google:  entity.ThirdParty{},
				Meta:    entity.ThirdParty{},
				Twitter: entity.ThirdParty{},
				LINE:    entity.ThirdParty{},
			},
		},
	}

	u, err = s.userService.GetProfileFromOAuth(ctx, u, session)
	if err != nil {
		return nil, err
	}

	switch session.Provider {
	case enum.OAuthProvider.Google:
		return &user.GetProfileFromOAuthResponse{
			OpenID: u.Profile.ThirdParties.Google.ID,
		}, nil
	case enum.OAuthProvider.Meta:
		return &user.GetProfileFromOAuthResponse{
			OpenID: u.Profile.ThirdParties.Meta.ID,
		}, nil
	case enum.OAuthProvider.Twitter:
		return &user.GetProfileFromOAuthResponse{
			OpenID: u.Profile.ThirdParties.Twitter.ID,
		}, nil
	case enum.OAuthProvider.LINE:
		return &user.GetProfileFromOAuthResponse{
			OpenID: u.Profile.ThirdParties.LINE.ID,
		}, nil
	default:
		return &user.GetProfileFromOAuthResponse{}, nil
	}
}

func (s *UserService) CheckMobileExistence(ctx context.Context, req *user.MobileExistenceRequest) (*user.ExistenceResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	existe, err := s.userService.CheckMobileExistence(ctx, req.GetCountryCode(), req.GetMobileNumber())
	if err != nil {
		return nil, err
	}

	return &user.ExistenceResponse{
		Exist: existe,
	}, nil
}

func (s *UserService) CheckEmailExistence(ctx context.Context, req *user.EmailExistenceRequest) (*user.ExistenceResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	existe, err := s.userService.CheckEmailExistence(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}

	return &user.ExistenceResponse{
		Exist: existe,
	}, nil
}

func (s *UserService) IsAccountExist(ctx context.Context, req *user.IsAccountExistRequest) (*user.ExistenceResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	exist, err := s.userService.IsAccountExist(ctx, req.GetAccount())
	if err != nil {
		return nil, err
	}

	return &user.ExistenceResponse{
		Exist: exist,
	}, nil
}

func (s *UserService) GetLoginUserInfo(ctx context.Context, req *user.GetLoginUserInfoRequest) (*user.GetLoginUserInfoResponse, error) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	identifiers := map[int]string{}
	switch req.LoginType {
	case enum.LoginTypes.Account.String:
		identifiers[enum.ProfileKey.Account.ID] = req.Account
	case enum.LoginTypes.Email.String:
		identifiers[enum.ProfileKey.Email.ID] = req.Email
	case enum.LoginTypes.MobileNumber.String:
		identifiers[enum.ProfileKey.CountryCode.ID] = req.CountryCode
		identifiers[enum.ProfileKey.MobileNumber.ID] = req.MobileNumber
	default:
		return nil, cus_err.New(cus_err.InvalidArgument, "invalid login type")
	}

	u, err := s.userService.GetLoginUserInfo(ctx, identifiers)
	if err != nil {
		return nil, err
	}

	return &user.GetLoginUserInfoResponse{
		UserId:       u.ID,
		Email:        u.Profile.Email,
		Account:      u.Profile.Account,
		CountryCode:  u.Profile.CountryCode,
		MobileNumber: u.Profile.MobileNumber,
	}, nil
}

package application

import (
	"context"
	"fmt"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/pb/gen/auth"
	"go_micro_service_api/pkg/req_analyzer"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	authService   *service.AuthService
	clientService *service.ClientService
	userService   *service.UserService
	db            db.Database
	reqAnalyzer   req_analyzer.ReqAnalyzer
}

func NewAuthService(
	authService *service.AuthService,
	clientService *service.ClientService,
	userService *service.UserService,
	db db.Database,
	reqAnalyzer req_analyzer.ReqAnalyzer) *AuthService {
	return &AuthService{
		authService:   authService,
		clientService: clientService,
		userService:   userService,
		db:            db,
		reqAnalyzer:   reqAnalyzer,
	}
}

func (s *AuthService) ClientAuth(ctx context.Context, req *auth.ClientAuthRequest) (res *auth.AuthResponse, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := s.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}

	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := s.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := s.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Create client token
	result, cusErr := s.authService.CreateClientToken(ctx, req.ClientId)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.AuthResponse{
		AccessToken:     result.Token,
		TokenExpireSecs: int64(result.TokenExpireSecs),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.AuthResponse, error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	var loginErr *cus_err.CusError

	// Begin transaction
	ctx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		// If there is an error, rollback the transaction
		if loginErr != nil &&
			loginErr.Code().Int() != cus_err.AccountPasswordError &&
			loginErr.Code().Int() != cus_err.AccountLocked {
			_, rollbackErr := s.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				loginErr = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := s.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Login
	result, loginErr := s.authService.Login(ctx, req.AccessToken, req.UserId, req.Password, false)

	// Get user ip and user agent info
	ipInfo := s.reqAnalyzer.GetIpInfo(ctx, req.Ip)
	userAgentInfo := s.reqAnalyzer.GetUserAgentInfo(ctx, req.UserAgent)
	// check if last login record is unusual
	isUnusual, err := s.authService.IsLoginRecordUnusual(ctx, req.UserId, ipInfo, userAgentInfo)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to check login unusualness", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// add new login record
	// If login failed, there will be error message
	errMsg := ""
	if loginErr != nil {
		errMsg = loginErr.Error()
	}
	loginRecord := &entity.LoginRecord{
		Browser:     userAgentInfo.Browser,
		BrowserVer:  userAgentInfo.BrowserVer,
		Ip:          ipInfo.Ip,
		Os:          userAgentInfo.OS,
		Platform:    userAgentInfo.Platform,
		Country:     ipInfo.Country,
		CountryCode: ipInfo.CountryCode,
		City:        ipInfo.City,
		Asp:         ipInfo.Asp,
		IsMobile:    userAgentInfo.IsMobile,
		IsSuccess:   loginErr == nil,
		ErrMessage:  errMsg,
	}
	_, err = s.authService.AddLoginRecord(ctx, req.UserId, loginRecord)
	if err != nil {
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// If login failed, return the error
	if loginErr != nil {
		cus_otel.Error(ctx, loginErr.Error())
		return nil, loginErr
	}

	if isUnusual {
		cusErr := cus_err.New(cus_err.UnusualLogin, fmt.Sprintf("Unusual login detected for user %d", req.UserId))
		cus_otel.Warn(ctx, cusErr.Error())
		return nil, cusErr
	}

	return &auth.AuthResponse{
		AccessToken:     result.Token,
		TokenExpireSecs: int64(result.TokenExpireSecs),
	}, nil
}

func (s *AuthService) ValidToken(ctx context.Context, req *auth.ValidTokenRequest) (res *auth.ValidTokenResponse, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	res = &auth.ValidTokenResponse{}

	// Validate token
	payload, cusErr := s.authService.ValidateToken(ctx, req.AccessToken)
	if cusErr != nil {
		return nil, cusErr
	}

	res.ClientId = payload.ClientId
	res.MerchantId = payload.MerchantId
	res.UserAccount = payload.Account
	res.UserId = payload.UserId

	if payload.RoleId == nil {
		return res, nil
	}

	role, cusErr := s.clientService.FindRole(ctx, payload.ClientId, *payload.RoleId)
	if cusErr != nil {
		if cus_err.ResourceNotFound == cusErr.Code().Int() {
			return res, nil
		}
		return nil, cusErr
	}

	res.Role = &auth.Role{
		RoleId:   role.Id,
		RoleName: role.Name,
		PermIds:  role.GetPermissionIds(),
	}

	return res, nil
}

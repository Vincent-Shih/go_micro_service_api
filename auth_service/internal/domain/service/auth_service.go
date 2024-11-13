package service

import (
	"context"
	"fmt"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/repository"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/auth_service/internal/infrastructure/token_helper"
	"go_micro_service_api/pkg/cus_crypto"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/req_analyzer"
	"time"
)

// AuthService handles authentication-related operations.
type AuthService struct {
	clientRepo  repository.ClientRepo
	userRepo    repository.UserRepo
	tokenHelper token_helper.TokenHelper
	cache       db.Cache
	crypto      cus_crypto.CusCrypto
}

const TokenPrefix = "token"

func NewAuthService(
	clientRepo repository.ClientRepo,
	userRepo repository.UserRepo,
	cache db.Cache,
	helper token_helper.TokenHelper) *AuthService {
	return &AuthService{
		clientRepo:  clientRepo,
		userRepo:    userRepo,
		tokenHelper: helper,
		cache:       cache,
		crypto:      cus_crypto.New(),
	}
}

// CreateClientToken creates a client token for the given client ID.
func (a *AuthService) CreateClientToken(ctx context.Context, clientId int64) (*vo.Token, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get client info
	client, err := a.clientRepo.Find(ctx, clientId)
	if err != nil {
		return nil, err
	}

	// Check client ia active
	if !client.Active {
		err = cus_err.New(cus_err.ClientInactive, fmt.Sprintf("Client id: %v is not active", client.Id))
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Create token
	payload := vo.NewTokenPayload(client.MerchantId, client.Id)
	token, err := a.tokenHelper.Create(ctx, client.Secret, payload.ToMap())
	if err != nil {
		return nil, err
	}

	// Cache token
	key := fmt.Sprintf("%s:%s", TokenPrefix, token)
	err = a.cache.Set(
		ctx,
		key,
		token,
		time.Second*time.Duration(client.TokenExpireSecs),
	)
	if err != nil {
		return nil, err
	}

	return &vo.Token{
		Token:           token,
		TokenExpireSecs: client.TokenExpireSecs,
	}, nil
}

// Login authenticates a user with the provided token (containing 'cid'), user ID, and password.
// It returns a new token upon successful login.
// When login is successful , the old token is going to delete from cache.
// Regardless of success or failure, a login record is created.
func (a *AuthService) Login(ctx context.Context, token string, userId int64, password string, forceLogin bool) (*vo.LoginTokenList, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find user by id
	user, err := a.userRepo.Find(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Validate client token
	claims, err := a.GetTokenPayload(ctx, token)
	if err != nil {
		return nil, err
	}

	// Get client, this is going to find from cache first and then from database
	client, err := a.clientRepo.Find(ctx, claims.ClientId)
	if err != nil {
		return nil, err
	}

	// Validate token
	_, err = a.tokenHelper.Validate(ctx, token, client.Secret)
	if err != nil {
		return nil, err
	}

	// Check token is in cache
	key := ""
	if claims.UserId != nil {
		key = fmt.Sprintf("%s:%d", TokenPrefix, *claims.UserId)
	} else {
		key = fmt.Sprintf("%s:%s", TokenPrefix, token)
	}
	cacheToken, err := a.cache.Get(ctx, key)
	if err != nil || cacheToken != token {
		err = cus_err.New(cus_err.TokenExpired, "Token is expired")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Check client is active
	if !client.Active {
		err = cus_err.New(cus_err.ClientInactive, fmt.Sprintf("Client id: %v is not active", client.Id))
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Check user status
	if user.Status != enum.UserStatusType.Active {
		err = cus_err.New(cus_err.AccountLocked, fmt.Sprintf("User id: %v is not active", user.Status))
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Check user password is correct
	if forceLogin || !a.crypto.CompareHashAndPassword(ctx, user.Password, password) {
		// Increase login error count
		user.PasswordFailTimes += 1

		// If password failed time is more than client login failed times, then lock user

		fmt.Println(client.LoginFailedTimes)
		if user.PasswordFailTimes >= client.LoginFailedTimes {

			user.Status = enum.UserStatusType.Locked
		}

		// Create password error with remaining times
		err := cus_err.New(cus_err.AccountPasswordError, "Invalid password").
			WithData(map[string]interface{}{
				"errorCount":    user.PasswordFailTimes,
				"totalAttempts": client.LoginFailedTimes,
			})
		cus_otel.Error(ctx, err.Error())

		// Update user
		_, updateErr := a.userRepo.Update(ctx, user)
		if updateErr != nil {
			cus_otel.Error(ctx, updateErr.Error())
			return nil, err
		}
		return &vo.LoginTokenList{
			Token:           token,
			TokenExpireSecs: client.TokenExpireSecs,
			ErrorCount:      user.PasswordFailTimes,
			TotalAttempts:   client.LoginFailedTimes,
		}, err
	}
	if user.PasswordFailTimes != 0 {
		// PasswordFailedTime is reset to 0
		user.PasswordFailTimes = 0
		// Update
		_, updateErr := a.userRepo.Update(ctx, user)
		if updateErr != nil {
			cus_otel.Error(ctx, updateErr.Error())
			return nil, err
		}
	}

	// TODO 不會做 等問人role是要怎樣做的
	// Get user role.
	// If user role not found(no role) it just continue to create token
	role, loginErr := user.Role(ctx)
	if loginErr != nil && loginErr.Code().Int() != cus_err.ResourceNotFound {
		cus_otel.Warn(ctx, loginErr.Error())
		return nil, nil
	}

	// Create new token
	opts := []vo.TokenPayloadOption{
		vo.WithUserId(user.Id),
		vo.WithAccount(user.Account),
	}
	if role != nil {
		opts = append(opts, vo.WithRoleId(role.Id))
	}
	payload := vo.NewTokenPayload(client.MerchantId, client.Id, opts...)
	newToken, loginErr := a.tokenHelper.Create(ctx, client.Secret, payload.ToMap())
	if loginErr != nil {
		return nil, loginErr
	}

	// Delete old token from cache
	loginErr = a.cache.Delete(ctx, key)
	if loginErr != nil {
		cusErr := cus_err.New(cus_err.TokenExpired, "Token is expired", loginErr)
		cus_otel.Error(ctx, cusErr.Error())
		loginErr = cusErr
		return nil, loginErr
	}

	// Cache token
	key = fmt.Sprintf("%s:%d", TokenPrefix, user.Id)
	loginErr = a.cache.Set(ctx, key, newToken, time.Second*time.Duration(client.TokenExpireSecs))
	if loginErr != nil {
		return nil, loginErr
	}

	return &vo.LoginTokenList{
		Token:           newToken,
		TokenExpireSecs: client.TokenExpireSecs,
	}, nil
}

// ValidateToken validates the given token.
func (a *AuthService) ValidateToken(ctx context.Context, token string) (*vo.TokenPayload, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get token payload
	claims, err := a.GetTokenPayload(ctx, token)
	if err != nil {
		return nil, err
	}

	// Get client, this is going to find from cache first and then from database
	client, err := a.clientRepo.Find(ctx, claims.ClientId)
	if err != nil {
		return nil, err
	}

	// Validate token
	_, err = a.tokenHelper.Validate(ctx, token, client.Secret)
	if err != nil {
		return nil, err
	}

	// Check client is active
	if !client.Active {
		err = cus_err.New(cus_err.ClientInactive, fmt.Sprintf("Client id: %v is not active", client.Id))
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	key := ""
	if claims.UserId != nil {
		key = fmt.Sprintf("%s:%d", TokenPrefix, *claims.UserId)
	} else {
		key = fmt.Sprintf("%s:%s", TokenPrefix, token)
	}

	// Get token from cache
	cacheToken, err := a.cache.Get(ctx, key)
	if err != nil {
		err = cus_err.New(cus_err.TokenExpired, "Token is expired")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Check the cache token is the same as the given token
	if cacheToken != token {
		err = cus_err.New(cus_err.TokenExpired, "Token is expired")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	return &claims, nil
}

// GetTokenPayload extracts the payload information from the given token.
//
// Parameters:
//   - ctx: context for tracing and potential cancellation
//   - token: the token string to be parsed
//
// Returns:
//   - vo.TokenPayload: the extracted token payload
//   - *cus_err.CusError: an error if parsing fails; nil on success
//
// Security note: Ensure that the token's legitimacy and validity are verified
// elsewhere before relying on this payload for sensitive operations.
func (a *AuthService) GetTokenPayload(ctx context.Context, token string) (vo.TokenPayload, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	tp := vo.TokenPayload{}

	// Get payload from token
	claims, err := a.tokenHelper.GetPayload(ctx, token)
	if err != nil {
		return tp, err
	}

	// Get TokenPayload from claims
	tp, err = vo.ToTokenPayload(ctx, claims)
	if err != nil {
		return tp, err
	}

	return tp, nil
}

func (a *AuthService) AddLoginRecord(ctx context.Context, userId int64, record *entity.LoginRecord) (*entity.LoginRecord, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	return a.userRepo.AddLoginRecord(ctx, userId, record)
}

func (a *AuthService) IsLoginRecordUnusual(
	ctx context.Context,
	userId int64,
	ipInfo req_analyzer.IPInfo,
	userAgentInfo req_analyzer.UserAgentInfo,
) (bool, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// get last login record
	lastRecord, err := a.userRepo.GetLastLoginRecord(ctx, userId)
	if err != nil && err.Code().Int() != cus_err.ResourceNotFound {
		cusErr := cus_err.New(cus_err.InternalServerError, "Failed to get last login record", err)
		cus_otel.Error(ctx, cusErr.Error())
		return false, cusErr
	}

	// and check if the login is unusual
	if lastRecord != nil {
		return lastRecord.City != ipInfo.City && lastRecord.Browser != userAgentInfo.Browser, nil
	}

	return false, nil
}

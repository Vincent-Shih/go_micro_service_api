package service

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/repository"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/pkg/cus_crypto"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
)

type UserService struct {
	clientRepo repository.ClientRepo
	userRepo   repository.UserRepo
	crypto     cus_crypto.CusCrypto
}

func NewUserService(clientRepo repository.ClientRepo, userRepo repository.UserRepo) *UserService {
	return &UserService{
		userRepo:   userRepo,
		clientRepo: clientRepo,
		crypto:     cus_crypto.New(),
	}
}

func (u *UserService) CreateUser(ctx context.Context, clientId int64, userInfo vo.UserInfo) (*aggregate.User, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find client
	_, err := u.clientRepo.Find(ctx, clientId)
	if err != nil {
		return nil, err
	}

	// Check the parameters
	if userInfo.Account == "" {
		err = cus_err.New(cus_err.AccountPasswordError, "account is required")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	if userInfo.Id == 0 {
		err = cus_err.New(cus_err.AccountPasswordError, "id is required")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	if userInfo.Status == 0 {
		err = cus_err.New(cus_err.AccountPasswordError, "status is required")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Password could be empty ,if not empty, hash with Md5 , and encode with base64
	if userInfo.Password != "" {
		hashPwd, err := u.crypto.HashPassword(ctx, userInfo.Password)
		if err != nil {
			return nil, err
		}
		userInfo.Password = hashPwd
	}

	// Create user
	user := &aggregate.User{
		Id:       userInfo.Id,
		Account:  userInfo.Account,
		Status:   userInfo.Status,
		Password: userInfo.Password,
	}

	user, err = u.userRepo.Create(ctx, clientId, user)

	return user, err
}

func (u *UserService) UpdateUser(ctx context.Context, userInfo vo.UserInfo) (*aggregate.User, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find user
	user, err := u.userRepo.Find(ctx, userInfo.Id)
	if err != nil {
		return nil, err
	}

	// If account is not empty, update it
	if userInfo.Account == "" || userInfo.Status == 0 {
		err = cus_err.New(cus_err.AccountPasswordError, "user account and status are required")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	user.Account = userInfo.Account
	user.Status = userInfo.Status

	if userInfo.Password != "" {
		secretByte := u.crypto.HashMD5(ctx, userInfo.Password)
		user.Password = u.crypto.EncodeBase64(ctx, secretByte)
	}

	// Update user
	user, err = u.userRepo.Update(ctx, user)

	return user, err
}

func (u *UserService) GetUser(ctx context.Context, userId int64) (*aggregate.User, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find user
	user, err := u.userRepo.Find(ctx, userId)

	return user, err
}

func (u *UserService) CheckAccountExistence(ct context.Context, Account string) (bool, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ct)
	defer span.End()

	// Check account existence
	exist, err := u.userRepo.CheckAccountExistence(ctx, Account)

	return exist, err
}

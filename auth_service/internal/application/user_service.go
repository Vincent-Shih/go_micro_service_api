package application

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/pb/gen/auth"
)

type UserService struct {
	auth.UnimplementedUserServiceServer
	userService *service.UserService
	db          db.Database
}

func NewUserService(userService *service.UserService, db db.Database) *UserService {
	return &UserService{
		userService: userService,
		db:          db,
	}
}

func (u *UserService) CreateUser(ctx context.Context, req *auth.CreateUserRequest) (res *auth.Empty, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := u.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}

	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := u.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := u.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Convert user status
	userStatus, cusErr := enum.UserStatusFromInt(int(req.Status))
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	userInfo := vo.UserInfo{
		Id:       req.Id,
		Account:  req.Account,
		Password: req.Password,
		Status:   userStatus,
	}

	// Create user
	_, cusErr = u.userService.CreateUser(ctx, req.ClientId, userInfo)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.Empty{}, nil
}

func (u *UserService) UpdateUser(ctx context.Context, req *auth.UpdateUserRequest) (res *auth.Empty, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := u.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}

	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := u.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := u.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Convert user status
	userStatus, cusErr := enum.UserStatusFromInt(int(req.Status))
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	userInfo := vo.UserInfo{
		Id:       req.Id,
		Account:  req.Account,
		Password: req.Password,
		Status:   userStatus,
	}

	// Update user
	_, cusErr = u.userService.UpdateUser(ctx, userInfo)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.Empty{}, nil
}

func (u *UserService) CheckAccountExistence(ctx context.Context, req *auth.AccountExistenceRequest) (res *auth.ExistenceResponse, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Check account existence
	exist, cusErr := u.userService.CheckAccountExistence(ctx, req.Account)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.ExistenceResponse{
		Existence: exist,
	}, nil
}

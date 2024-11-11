package repository

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/pkg/cus_err"
)

type UserRepo interface {
	Find(ctx context.Context, id int64) (*aggregate.User, *cus_err.CusError)
	Create(ctx context.Context, clientId int64, user *aggregate.User) (*aggregate.User, *cus_err.CusError)
	Update(ctx context.Context, user *aggregate.User) (*aggregate.User, *cus_err.CusError)
	AddLoginRecord(ctx context.Context, userId int64, loginRecord *entity.LoginRecord) (*entity.LoginRecord, *cus_err.CusError)
	BindRole(ctx context.Context, userId int64, roleId int64) (*aggregate.User, *cus_err.CusError)
	CheckAccountExistence(ctx context.Context, account string) (bool, *cus_err.CusError)
	GetLastLoginRecord(ctx context.Context, userId int64) (*entity.LoginRecord, *cus_err.CusError)
}

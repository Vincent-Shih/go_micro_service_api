package repository

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/pkg/cus_err"
)

type ClientRepo interface {
	Create(ctx context.Context, client *aggregate.Client) (*aggregate.Client, *cus_err.CusError)
	Find(ctx context.Context, id int64) (*aggregate.Client, *cus_err.CusError)
	Update(ctx context.Context, client *aggregate.Client) (*aggregate.Client, *cus_err.CusError)
	BindSystemRoles(ctx context.Context, clientId int64, sysRoles ...entity.Role) *cus_err.CusError
	CreateRoles(ctx context.Context, clientId int64, roles ...entity.Role) ([]entity.Role, *cus_err.CusError)
	DeleteRoles(ctx context.Context, clientId int64, roleIds ...int64) *cus_err.CusError
	UpdateRoles(ctx context.Context, clientId int64, roles ...entity.Role) ([]entity.Role, *cus_err.CusError)
	FindRole(ctx context.Context, clientId int64, roleId int64) (*entity.Role, *cus_err.CusError)
}

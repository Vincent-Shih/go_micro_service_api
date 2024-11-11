package aggregate

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/enum"
)

type User struct {
	Id                    int64                                                              // Primary key for the user
	Account               string                                                             // Unique name for each user
	Password              string                                                             // Hashed password
	PasswordFailTimes     int                                                                // Number of times the user has failed to login
	Status                enum.UserStatus                                                    // Status of the user
	clientLoader          func(ctx context.Context) (*Client, *cus_err.CusError)             // Lazy loader for the client
	roleLoader            func(ctx context.Context) (*entity.Role, *cus_err.CusError)        // Lazy loader for the role
	lastLoginRecordLoader func(ctx context.Context) (*entity.LoginRecord, *cus_err.CusError) // Lazy loader for the last login record
}

func (u *User) SetRoleLoader(loader func(ctx context.Context) (*entity.Role, *cus_err.CusError)) {
	u.roleLoader = loader
}

func (u *User) Role(ctx context.Context) (*entity.Role, *cus_err.CusError) {
	return u.roleLoader(ctx)
}

func (u *User) SetLoginRecordLoader(loader func(ctx context.Context) (*entity.LoginRecord, *cus_err.CusError)) {
	u.lastLoginRecordLoader = loader
}

func (u *User) LoginRecord(ctx context.Context) (*entity.LoginRecord, *cus_err.CusError) {
	return u.lastLoginRecordLoader(ctx)
}

func (u *User) SetClientLoader(loader func(ctx context.Context) (*Client, *cus_err.CusError)) {
	u.clientLoader = loader
}

func (u *User) Client(ctx context.Context) (*Client, *cus_err.CusError) {
	return u.clientLoader(ctx)
}

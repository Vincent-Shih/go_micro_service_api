package repository

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/user_service/internal/domain/aggregate"
	"go_micro_service_api/user_service/internal/domain/vo"
)

type UserRepo interface {
	CreateProfile(ctx context.Context, u *aggregate.User) (*aggregate.User, *cus_err.CusError)
	GetProfile(ctx context.Context, u *aggregate.User, keys []int) (*aggregate.User, *cus_err.CusError)
	GetProfileFromOAuth(ctx context.Context, u *aggregate.User, session *vo.OAuthSession) (*aggregate.User, *cus_err.CusError)
	CheckMobileExistence(ctx context.Context, mobileNumber string, countryCode string) (bool, *cus_err.CusError)
	CheckEmailExistence(ctx context.Context, email string) (bool, *cus_err.CusError)
	IsAccountExist(ctx context.Context, account string) (bool, *cus_err.CusError)
	GetUserIdByProfile(ctx context.Context, mapping map[int]string) (int, *cus_err.CusError)
}

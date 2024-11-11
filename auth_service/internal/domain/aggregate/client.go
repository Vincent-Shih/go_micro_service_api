package aggregate

import (
	"context"

	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/enum"
)

type Client struct {
	Id               int64
	MerchantId       int64
	ClientType       enum.Client
	Secret           string
	Active           bool
	TokenExpireSecs  int
	LoginFailedTimes int
	rolesLoader      func(ctx context.Context) (*map[int64]entity.Role, *cus_err.CusError)
}

func (c *Client) Roles(ctx context.Context) (*map[int64]entity.Role, *cus_err.CusError) {
	return c.rolesLoader(ctx)
}

func (c *Client) SetRolesLoader(loader func(ctx context.Context) (*map[int64]entity.Role, *cus_err.CusError)) {
	c.rolesLoader = loader
}

package application

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/pb/gen/auth"
)

type ClientService struct {
	auth.UnimplementedClientServiceServer
	clientService *service.ClientService
	db            db.Database
}

func NewClientService(clientService *service.ClientService, db db.Database) *ClientService {
	return &ClientService{
		clientService: clientService,
		db:            db,
	}
}

func (c *ClientService) CreateClient(ctx context.Context, req *auth.CreateClientRequest) (res *auth.Empty, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := c.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}
	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := c.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := c.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Convert client type
	clientType, cusErr := enum.ClientTypeFromId(int(req.ClientType))
	if cusErr != nil {
		return nil, cusErr
	}

	// Map request to client info
	clientInfo := vo.ClientInfo{
		Id:               req.ClientId,
		MerchantId:       req.MerchantId,
		ClientType:       clientType,
		LoginFailedTimes: int(req.LoginFailedTimes),
		TokenExpireSecs:  int(req.TokenExpireSecs),
		Active:           req.IsActive,
	}

	// Create client
	_, cusErr = c.clientService.CreateClient(ctx, clientInfo)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.Empty{}, nil
}

func (c *ClientService) UpdateClient(ctx context.Context, req *auth.UpdateClientRequest) (res *auth.Empty, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := c.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}
	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := c.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := c.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Map request to client info
	clientInfo := vo.ClientInfo{
		Id:               req.ClientId,
		LoginFailedTimes: int(req.LoginFailedTimes),
		TokenExpireSecs:  int(req.TokenExpireSecs),
		Active:           req.IsActive,
	}

	// Update client
	_, cusErr = c.clientService.UpdateClient(ctx, clientInfo)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.Empty{}, nil
}

func (c *ClientService) CreateRole(ctx context.Context, req *auth.CreateRoleRequest) (res *auth.Role, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := c.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}
	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := c.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := c.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Get permissions
	perms := make([]enum.Permission, 0)
	for _, id := range req.PermIds {
		perm, cusErr := enum.PermissionById(id)
		if cusErr != nil {
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}
		perms = append(perms, perm)
	}

	// Map request to role
	role := entity.Role{
		Name:        req.RoleName,
		Permissions: perms,
	}

	// Create role
	createdRoles, cusErr := c.clientService.CreateRoles(ctx, req.ClientId, role)
	if cusErr != nil {
		return nil, cusErr
	}

	// Map role to response
	res = &auth.Role{
		RoleId:     createdRoles[0].Id,
		RoleName:   createdRoles[0].Name,
		PermIds:    createdRoles[0].GetPermissionIds(),
		ClientType: int32(createdRoles[0].ClientType.Id),
		IsSystem:   createdRoles[0].IsSystem(),
	}

	return res, nil
}

func (c *ClientService) UpdateRole(ctx context.Context, req *auth.UpdateRoleRequest) (res *auth.Role, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := c.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}
	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := c.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := c.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Get permissions
	perms := make([]enum.Permission, 0)
	for _, id := range req.PermIds {
		perm, cusErr := enum.PermissionById(id)
		if cusErr != nil {
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}
		perms = append(perms, perm)
	}

	// Map request to role
	role := entity.Role{
		Id:          req.RoleId,
		Name:        req.RoleName,
		Permissions: perms,
	}

	// Update role
	updatedRoles, cusErr := c.clientService.UpdateRoles(ctx, req.ClientId, role)
	if cusErr != nil {
		return nil, cusErr
	}

	// Map role to response
	res = &auth.Role{
		RoleId:     updatedRoles[0].Id,
		RoleName:   updatedRoles[0].Name,
		PermIds:    updatedRoles[0].GetPermissionIds(),
		ClientType: int32(updatedRoles[0].ClientType.Id),
		IsSystem:   updatedRoles[0].IsSystem(),
	}

	return res, nil
}

func (c *ClientService) DeleteRole(ctx context.Context, req *auth.DeleteRoleRequest) (res *auth.Empty, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, cusErr := c.db.Begin(ctx)
	if cusErr != nil {
		return nil, cusErr
	}
	defer func() {
		// If there is an error, rollback the transaction
		if err != nil {
			_, rollbackErr := c.db.Rollback(ctx)
			if rollbackErr != nil {
				cus_otel.Error(ctx, rollbackErr.Error())
				err = rollbackErr
			}
			return
		}

		// Commit the transaction
		_, commitErr := c.db.Commit(ctx)
		if commitErr != nil {
			cus_otel.Error(ctx, commitErr.Error())
			err = commitErr
		}
	}()

	// Delete role
	cusErr = c.clientService.DeleteRoles(ctx, req.ClientId, req.RoleId)
	if cusErr != nil {
		return nil, cusErr
	}

	return &auth.Empty{}, nil
}

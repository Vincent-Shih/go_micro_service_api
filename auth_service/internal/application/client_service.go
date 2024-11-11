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
	ctx, kgsErr := c.db.Begin(ctx)
	if kgsErr != nil {
		return nil, kgsErr
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
	clientType, kgsErr := enum.ClientTypeFromId(int(req.ClientType))
	if kgsErr != nil {
		return nil, kgsErr
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
	_, kgsErr = c.clientService.CreateClient(ctx, clientInfo)
	if kgsErr != nil {
		return nil, kgsErr
	}

	return &auth.Empty{}, nil
}

func (c *ClientService) UpdateClient(ctx context.Context, req *auth.UpdateClientRequest) (res *auth.Empty, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, kgsErr := c.db.Begin(ctx)
	if kgsErr != nil {
		return nil, kgsErr
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
	_, kgsErr = c.clientService.UpdateClient(ctx, clientInfo)
	if kgsErr != nil {
		return nil, kgsErr
	}

	return &auth.Empty{}, nil
}

func (c *ClientService) CreateRole(ctx context.Context, req *auth.CreateRoleRequest) (res *auth.Role, err error) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Begin transaction
	ctx, kgsErr := c.db.Begin(ctx)
	if kgsErr != nil {
		return nil, kgsErr
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
		perm, kgsErr := enum.PermissionById(id)
		if kgsErr != nil {
			cus_otel.Error(ctx, kgsErr.Error())
			return nil, kgsErr
		}
		perms = append(perms, perm)
	}

	// Map request to role
	role := entity.Role{
		Name:        req.RoleName,
		Permissions: perms,
	}

	// Create role
	createdRoles, kgsErr := c.clientService.CreateRoles(ctx, req.ClientId, role)
	if kgsErr != nil {
		return nil, kgsErr
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
	ctx, kgsErr := c.db.Begin(ctx)
	if kgsErr != nil {
		return nil, kgsErr
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
		perm, kgsErr := enum.PermissionById(id)
		if kgsErr != nil {
			cus_otel.Error(ctx, kgsErr.Error())
			return nil, kgsErr
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
	updatedRoles, kgsErr := c.clientService.UpdateRoles(ctx, req.ClientId, role)
	if kgsErr != nil {
		return nil, kgsErr
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
	ctx, kgsErr := c.db.Begin(ctx)
	if kgsErr != nil {
		return nil, kgsErr
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
	kgsErr = c.clientService.DeleteRoles(ctx, req.ClientId, req.RoleId)
	if kgsErr != nil {
		return nil, kgsErr
	}

	return &auth.Empty{}, nil
}

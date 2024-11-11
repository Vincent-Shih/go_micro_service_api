package service

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/repository"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/pkg/cus_crypto"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/enum"
)

type ClientService struct {
	clientRepo repository.ClientRepo
	crypto     cus_crypto.CusCrypto
}

func NewClientService(clientRepo repository.ClientRepo) *ClientService {
	return &ClientService{
		clientRepo: clientRepo,
		crypto:     cus_crypto.New(),
	}
}

func (c *ClientService) CreateClient(ctx context.Context, clientInfo vo.ClientInfo) (*aggregate.Client, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Generate random secret
	secretByte, err := c.crypto.GenerateRandomSecret(ctx, 32)
	if err != nil {
		return nil, err
	}
	secret := c.crypto.EncodeHex(ctx, secretByte)

	client := &aggregate.Client{
		Id:               clientInfo.Id,
		MerchantId:       clientInfo.MerchantId,
		ClientType:       clientInfo.ClientType,
		LoginFailedTimes: clientInfo.LoginFailedTimes,
		TokenExpireSecs:  clientInfo.TokenExpireSecs,
		Secret:           secret,
		Active:           clientInfo.Active,
	}

	// Create client
	client, err = c.clientRepo.Create(ctx, client)
	if err != nil {
		return nil, err
	}

	// Create default roles for the client
	var roles []entity.Role
	switch clientInfo.ClientType {
	case enum.ClientType.Frontend:
		roles = entity.AllFrontendRoles
	case enum.ClientType.Backend:
		roles = entity.AllBackendRoles
	default:
		err = cus_err.New(cus_err.AccountPasswordError, "invalid client type")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Bind system roles to the client
	err = c.clientRepo.BindSystemRoles(ctx, client.Id, roles...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ClientService) UpdateClient(ctx context.Context, clientInfo vo.ClientInfo) (*aggregate.Client, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find client
	client, err := c.clientRepo.Find(ctx, clientInfo.Id)
	if err != nil {
		return nil, err
	}

	// Check the parameters
	if clientInfo.LoginFailedTimes == 0 ||
		clientInfo.TokenExpireSecs == 0 {
		err = cus_err.New(cus_err.AccountPasswordError, "login failed times and token expire seconds are required")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	client.LoginFailedTimes = clientInfo.LoginFailedTimes
	client.TokenExpireSecs = clientInfo.TokenExpireSecs
	client.Active = clientInfo.Active

	// Update client
	client, err = c.clientRepo.Update(ctx, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ClientService) CreateRoles(ctx context.Context, clientId int64, roles ...entity.Role) ([]entity.Role, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find client
	client, err := c.clientRepo.Find(ctx, clientId)
	if err != nil {
		return nil, err
	}

	// Create a new slice for modified roles
	modifiedRoles := make([]entity.Role, len(roles))
	for i, role := range roles {
		modifiedRole := role // Create a copy of the role
		modifiedRole.ClientType = client.ClientType
		modifiedRoles[i] = modifiedRole
	}

	// Create role
	newRoles, err := c.clientRepo.CreateRoles(ctx, clientId, modifiedRoles...)
	if err != nil {
		return nil, err
	}

	return newRoles, nil
}

func (c *ClientService) DeleteRoles(ctx context.Context, clientId int64, roleIds ...int64) *cus_err.CusError {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Define a function to filter out system roles
	filterRoleIds := func(roleIds []int64) (*[]int64, *cus_err.CusError) {
		// Get client by id
		client, err := c.clientRepo.Find(ctx, clientId)
		if err != nil {
			return nil, err
		}
		roles, err := client.Roles(ctx)
		if err != nil {
			return nil, err
		}
		if roles == nil {
			err = cus_err.New(cus_err.AccountPasswordError, "Client roles not found")
			return nil, err
		}

		// Check if the role is not a system role, system roles cannot be deleted
		var ids []int64
		for _, roleId := range roleIds {
			if val, exists := (*roles)[roleId]; exists && !val.IsSystem() {
				ids = append(ids, roleId)
			}
		}
		return &ids, nil
	}

	ids, err := filterRoleIds(roleIds)
	if err != nil {
		return err
	}

	// Delete roles
	err = c.clientRepo.DeleteRoles(ctx, clientId, *ids...)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClientService) UpdateRoles(ctx context.Context, clientId int64, roles ...entity.Role) ([]entity.Role, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Update role
	role, err := c.clientRepo.UpdateRoles(ctx, clientId, roles...)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (c *ClientService) FindRole(ctx context.Context, clientId int64, roleId int64) (*entity.Role, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Find role
	role, err := c.clientRepo.FindRole(ctx, clientId, roleId)
	if err != nil {
		return nil, err
	}

	return role, nil
}

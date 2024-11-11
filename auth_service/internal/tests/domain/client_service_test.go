package tests_domain

import (
	"context"
	"fmt"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/auth_service/internal/tests"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/pkg/enum"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClientService() (clientService *service.ClientService, db db.Database, cache db.Cache, closeFunc func()) {
	db = tests.NewMemoryDB()
	redis, closeFunc := tests.NewMemoryRedis()
	cache = redis_cache.NewRedisCache(redis)
	clientRepo := ent_impl.NewClientRepoImpl(db, cache)

	return service.NewClientService(clientRepo), db, cache, closeFunc
}

func TestCreateClient(t *testing.T) {
	clientService, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	t.Run("Create Client Success", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		client := vo.ClientInfo{
			Id:               123456789,
			MerchantId:       111111111,
			ClientType:       enum.ClientType.Frontend,
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 5,
		}

		created, err := clientService.CreateClient(ctx, client)
		assert.Nil(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, int64(123456789), created.Id)
		assert.Equal(t, int64(111111111), created.MerchantId)
		assert.Equal(t, true, created.Active)
		assert.Equal(t, 3600, created.TokenExpireSecs)
		assert.Equal(t, 5, created.LoginFailedTimes)

		roles, err := created.Roles(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(*roles), len(entity.AllFrontendRoles))

		// Check redis cache
		key := fmt.Sprintf("%s:%d", ent_impl.ClientInfoPrefix, created.Id)
		cacheClient := &aggregate.Client{}
		err = cache.GetObject(ctx, key, cacheClient)
		assert.Nil(t, err)
		assert.Equal(t, created.Id, cacheClient.Id)
		assert.Equal(t, created.MerchantId, cacheClient.MerchantId)
		assert.Equal(t, created.Active, cacheClient.Active)
		assert.Equal(t, created.TokenExpireSecs, cacheClient.TokenExpireSecs)
		assert.Equal(t, created.LoginFailedTimes, cacheClient.LoginFailedTimes)
	})

	t.Run("Create Client Without loginFailed time", func(t *testing.T) {
		client := vo.ClientInfo{
			Id:               123456789,
			MerchantId:       111111111,
			ClientType:       enum.ClientType.Frontend,
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 5,
		}

		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		client.LoginFailedTimes = 0
		created, err := clientService.CreateClient(ctx, client)
		assert.Nil(t, created)
		assert.NotNil(t, err)
		assert.Equal(t, err.Code().Int(), cus_err.AccountPasswordError)
	})

	t.Run("Create Client Without TokenExpireSecs", func(t *testing.T) {
		client := vo.ClientInfo{
			Id:               123456789,
			MerchantId:       111111111,
			ClientType:       enum.ClientType.Frontend,
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 5,
		}

		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		client.TokenExpireSecs = 0
		created, err := clientService.CreateClient(ctx, client)
		assert.Nil(t, created)
		assert.NotNil(t, err)
		assert.Equal(t, err.Code().Int(), cus_err.AccountPasswordError)
	})

	t.Run("Create Client Without MerchantId", func(t *testing.T) {
		client := vo.ClientInfo{
			Id:               123456789,
			MerchantId:       111111111,
			ClientType:       enum.ClientType.Frontend,
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 5,
		}

		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		client.MerchantId = 0
		created, err := clientService.CreateClient(ctx, client)
		assert.Nil(t, created)
		assert.NotNil(t, err)
		assert.Equal(t, err.Code().Int(), cus_err.AccountPasswordError)
	})

	t.Run("Create Client Without Id", func(t *testing.T) {
		client := vo.ClientInfo{
			Id:               123456789,
			MerchantId:       111111111,
			ClientType:       enum.ClientType.Frontend,
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 5,
		}

		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		client.Id = 0
		created, err := clientService.CreateClient(ctx, client)
		assert.Nil(t, created)
		assert.NotNil(t, err)
		assert.Equal(t, err.Code().Int(), cus_err.AccountPasswordError)
	})

	t.Run("Create Client Without ClientType", func(t *testing.T) {
		client := vo.ClientInfo{
			Id:               123456789,
			MerchantId:       111111111,
			ClientType:       enum.ClientType.Frontend,
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 5,
		}

		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		client.ClientType = enum.Client{}
		created, err := clientService.CreateClient(ctx, client)
		assert.Nil(t, created)
		assert.NotNil(t, err)
		assert.Equal(t, err.Code().Int(), cus_err.AccountPasswordError)
	})
}

func TestUpdateClient(t *testing.T) {
	clientService, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               123456789,
		MerchantId:       111111111,
		ClientType:       enum.ClientType.Frontend,
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)

	require.Nil(t, err)

	created, err := clientService.CreateClient(ctx, clientInfo)
	assert.Nil(t, err)
	assert.NotNil(t, created)

	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Update Client Success", func(t *testing.T) {
		// Begin a transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyClientInfo := &vo.ClientInfo{}
		_ = copier.CopyWithOption(copyClientInfo, clientInfo, copier.Option{IgnoreEmpty: true})

		// Update client
		copyClientInfo.TokenExpireSecs = 7200
		copyClientInfo.LoginFailedTimes = 10

		updated, err := clientService.UpdateClient(ctx, *copyClientInfo)
		assert.Nil(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, int64(123456789), updated.Id)
		assert.Equal(t, int64(111111111), updated.MerchantId)
		assert.Equal(t, true, updated.Active)
		assert.Equal(t, 7200, updated.TokenExpireSecs)
		assert.Equal(t, 10, updated.LoginFailedTimes)

		// Check redis cache
		key := fmt.Sprintf("%s:%d", ent_impl.ClientInfoPrefix, updated.Id)
		cacheClient := &aggregate.Client{}
		err = cache.GetObject(ctx, key, cacheClient)
		assert.Nil(t, err)
		assert.Equal(t, updated.Id, cacheClient.Id)
		assert.Equal(t, updated.MerchantId, cacheClient.MerchantId)
		assert.Equal(t, updated.Active, cacheClient.Active)
		assert.Equal(t, updated.TokenExpireSecs, cacheClient.TokenExpireSecs)
		assert.Equal(t, updated.LoginFailedTimes, cacheClient.LoginFailedTimes)
	})

	t.Run("Update Client Without TokenExpireSecs", func(t *testing.T) {
		// Begin a transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyClientInfo := &vo.ClientInfo{}
		_ = copier.CopyWithOption(copyClientInfo, clientInfo, copier.Option{IgnoreEmpty: true})

		// Update client
		copyClientInfo.TokenExpireSecs = 0

		updated, err := clientService.UpdateClient(ctx, *copyClientInfo)
		assert.Nil(t, updated)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})

	t.Run("Update Client Without LoginFailedTimes", func(t *testing.T) {
		// Begin a transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyClientInfo := &vo.ClientInfo{}
		e := copier.CopyWithOption(copyClientInfo, clientInfo, copier.Option{IgnoreEmpty: true})
		require.Nil(t, e)

		// Update client
		copyClientInfo.LoginFailedTimes = 0

		updated, err := clientService.UpdateClient(ctx, *copyClientInfo)
		assert.Nil(t, updated)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})

	t.Run("Update Client Without MerchantId", func(t *testing.T) {
		// Begin a transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyClientInfo := &vo.ClientInfo{}
		e := copier.CopyWithOption(copyClientInfo, clientInfo, copier.Option{IgnoreEmpty: true})
		require.Nil(t, e)

		// Update client
		copyClientInfo.MerchantId = 0

		// MerchantId is not allowed to be updated
		// so the error should be nil
		updated, err := clientService.UpdateClient(ctx, *copyClientInfo)
		assert.Nil(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, int64(123456789), updated.Id)
		assert.Equal(t, int64(111111111), updated.MerchantId)
		assert.Equal(t, true, updated.Active)
		assert.Equal(t, 3600, updated.TokenExpireSecs)
	})
}

func TestCreateRoles(t *testing.T) {
	clientService, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               123456789,
		MerchantId:       111111111,
		ClientType:       enum.ClientType.Frontend,
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	customRoles := []entity.Role{
		{
			Name: "CustomRole1",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
			},
			ClientType: enum.ClientType.Frontend,
		},
		{
			Name: "CustomRole2",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
				enum.PermissionType.Withdraw,
			},
			ClientType: enum.ClientType.Frontend,
		},
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	defer func() {
		_, rollbackErr := db.Rollback(ctx)
		require.Nil(t, rollbackErr)
	}()

	client, err := clientService.CreateClient(ctx, clientInfo)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	t.Run("Create Roles Success", func(t *testing.T) {
		// Create roles
		created, err := clientService.CreateRoles(ctx, client.Id, customRoles...)

		assert.Nil(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, len(customRoles), len(created))
		assert.Equal(t, customRoles[0].Name, (created)[0].Name)
		assert.Equal(t, customRoles[1].Name, (created)[1].Name)
		assert.Equal(t, customRoles[0].Permissions, (created)[0].Permissions)
		assert.Equal(t, customRoles[1].Permissions, (created)[1].Permissions)

		// Check redis cache
		for _, role := range created {
			key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, client.Id, role.Id)
			cacheRole := &entity.Role{}
			err = cache.GetObject(ctx, key, cacheRole)
			assert.Nil(t, err)
			assert.Equal(t, role.Id, cacheRole.Id)
			assert.Equal(t, role.Name, cacheRole.Name)
			assert.Equal(t, role.Permissions, cacheRole.Permissions)
		}

	})

	t.Run("Create Roles with non-existent client", func(t *testing.T) {
		// Create roles
		created, err := clientService.CreateRoles(ctx, 0, customRoles...)

		assert.Nil(t, created)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})

	t.Run("Create Roles with empty roles", func(t *testing.T) {
		// Create roles
		created, err := clientService.CreateRoles(ctx, client.Id)

		assert.NotNil(t, created)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(created))
	})
}

func TestDeleteRoles(t *testing.T) {
	clientService, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               123456789,
		MerchantId:       111111111,
		ClientType:       enum.ClientType.Frontend,
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	customRoles := []entity.Role{
		{
			Name: "CustomRole1",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
			},
			ClientType: enum.ClientType.Frontend,
		},
		{
			Name: "CustomRole2",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
				enum.PermissionType.Withdraw,
			},
			ClientType: enum.ClientType.Frontend,
		},
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	defer func() {
		_, rollbackErr := db.Rollback(ctx)
		require.Nil(t, rollbackErr)
	}()

	client, err := clientService.CreateClient(ctx, clientInfo)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	roles, err := clientService.CreateRoles(ctx, client.Id, customRoles...)
	assert.Nil(t, err)
	assert.NotNil(t, roles)

	t.Run("Delete Roles Success", func(t *testing.T) {
		// Delete roles
		err := clientService.DeleteRoles(ctx, client.Id, roles[0].Id, roles[1].Id)

		assert.Nil(t, err)

		// Check redis cache
		for _, role := range roles {
			key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, client.Id, role.Id)
			cacheRole := &entity.Role{}
			err = cache.GetObject(ctx, key, cacheRole)
			assert.NotNil(t, err)
			assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
		}
	})

	t.Run("Delete Roles with non-existent client", func(t *testing.T) {
		// Delete roles
		err := clientService.DeleteRoles(ctx, 0, roles[0].Id, roles[1].Id)

		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})

	t.Run("Delete Roles with non-existent roles", func(t *testing.T) {
		// Delete roles
		err := clientService.DeleteRoles(ctx, client.Id,
			roles[0].Id, roles[1].Id, 0)

		// The error should be nil , only delete the existing roles
		assert.Nil(t, err)

		// Check redis cache
		for _, role := range roles {
			key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, client.Id, role.Id)
			cacheRole := &entity.Role{}
			err = cache.GetObject(ctx, key, cacheRole)
			assert.NotNil(t, err)
			assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
		}
	})
}

func TestFindRole(t *testing.T) {
	clientService, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               123456789,
		MerchantId:       111111111,
		ClientType:       enum.ClientType.Frontend,
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	customRoles := []entity.Role{
		{
			Name: "CustomRole1",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
			},
			ClientType: enum.ClientType.Frontend,
		},
		{
			Name: "CustomRole2",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
				enum.PermissionType.Withdraw,
			},
			ClientType: enum.ClientType.Frontend,
		},
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	defer func() {
		_, rollbackErr := db.Rollback(ctx)
		require.Nil(t, rollbackErr)
	}()

	client, err := clientService.CreateClient(ctx, clientInfo)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	roles, err := clientService.CreateRoles(ctx, client.Id, customRoles...)
	assert.Nil(t, err)
	assert.NotNil(t, roles)

	t.Run("Find Role Success", func(t *testing.T) {
		// Find role
		role, err := clientService.FindRole(ctx, client.Id, roles[0].Id)

		assert.Nil(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, roles[0].Id, role.Id)
		assert.Equal(t, roles[0].Name, role.Name)
		assert.Equal(t, roles[0].Permissions, role.Permissions)

		// Check redis cache
		key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, client.Id, role.Id)
		cacheRole := &entity.Role{}
		err = cache.GetObject(ctx, key, cacheRole)
		assert.Nil(t, err)
		assert.Equal(t, role.Id, cacheRole.Id)
		assert.Equal(t, role.Name, cacheRole.Name)
		assert.Equal(t, role.Permissions, cacheRole.Permissions)
	})

	t.Run("Find Role with non-existent client", func(t *testing.T) {
		// Find role
		role, err := clientService.FindRole(ctx, 0, roles[0].Id)

		assert.Nil(t, role)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})
}

func TestUpdateRoles(t *testing.T) {
	clientService, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               123456789,
		MerchantId:       111111111,
		ClientType:       enum.ClientType.Frontend,
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	customRoles := []entity.Role{
		{
			Name: "CustomRole1",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
			},
			ClientType: enum.ClientType.Frontend,
		},
		{
			Name: "CustomRole2",
			Permissions: []enum.Permission{
				enum.PermissionType.PlayGame,
				enum.PermissionType.Withdraw,
			},
			ClientType: enum.ClientType.Frontend,
		},
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	defer func() {
		_, rollbackErr := db.Rollback(ctx)
		require.Nil(t, rollbackErr)
	}()

	client, err := clientService.CreateClient(ctx, clientInfo)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	roles, err := clientService.CreateRoles(ctx, client.Id, customRoles...)
	assert.Nil(t, err)
	assert.NotNil(t, roles)

	t.Run("Update Role Success", func(t *testing.T) {
		// Update role
		role := &entity.Role{
			Id:          roles[0].Id,
			Name:        "UpdatedRole",
			Permissions: []enum.Permission{enum.PermissionType.Withdraw},
		}

		updated, err := clientService.UpdateRoles(ctx, client.Id, *role)

		assert.Nil(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, role.Id, updated[0].Id)
		assert.Equal(t, role.Name, updated[0].Name)
		assert.Equal(t, role.Permissions, updated[0].Permissions)

		// Check redis cache
		key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, client.Id, updated[0].Id)
		cacheRole := &entity.Role{}
		err = cache.GetObject(ctx, key, cacheRole)
		assert.Nil(t, err)
	})

	t.Run("Update Role with non-existent client", func(t *testing.T) {
		// Update role
		role := &entity.Role{
			Id:          roles[0].Id,
			Name:        "UpdatedRole",
			Permissions: []enum.Permission{enum.PermissionType.Withdraw},
		}

		updated, err := clientService.UpdateRoles(ctx, 0, *role)

		assert.Nil(t, updated)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})

	t.Run("Update Role with non-existent role", func(t *testing.T) {
		// Update role
		role := &entity.Role{
			Id:          0,
			Name:        "UpdatedRole",
			Permissions: []enum.Permission{enum.PermissionType.Withdraw},
		}

		updated, err := clientService.UpdateRoles(ctx, client.Id, *role)

		assert.Nil(t, updated)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})
}

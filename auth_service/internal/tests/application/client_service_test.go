package application_tests

import (
	"context"
	"fmt"
	"go_micro_service_api/auth_service/internal/application"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/tests"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/pb/gen/auth"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClientService() (clientApp *application.ClientService, db db.Database, cache db.Cache, closeFunc func()) {
	db = tests.NewMemoryDB()
	redis, closeFunc := tests.NewMemoryRedis()
	cache = redis_cache.NewRedisCache(redis)
	clientRepo := ent_impl.NewClientRepoImpl(db, cache)
	clientService := service.NewClientService(clientRepo)
	clientApp = application.NewClientService(clientService, db)
	return clientApp, db, cache, closeFunc
}

func TestCreateClient(t *testing.T) {
	clientApp, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               12345,
		MerchantId:       11111,
		ClientType:       enum.ClientType.Frontend,
		LoginFailedTimes: 3,
		TokenExpireSecs:  3600,
	}

	t.Run("CreateClient", func(t *testing.T) {
		// Create a client
		req := &auth.CreateClientRequest{
			ClientId:         int64(clientInfo.Id),
			MerchantId:       int64(clientInfo.MerchantId),
			ClientType:       int32(clientInfo.ClientType.Id),
			LoginFailedTimes: int32(clientInfo.LoginFailedTimes),
			TokenExpireSecs:  int64(clientInfo.TokenExpireSecs),
		}

		_, e := clientApp.CreateClient(ctx, req)
		require.Nil(t, e)

		entdb, ok := db.GetConn(ctx).(*ent.Client)
		require.True(t, ok)

		// Find the client in db
		entity, e := entdb.AuthClient.Get(ctx, clientInfo.Id)
		require.Nil(t, e)
		assert.Equal(t, clientInfo.Id, entity.ID)
		assert.Equal(t, clientInfo.MerchantId, entity.MerchantID)
		assert.Equal(t, clientInfo.ClientType.Id, entity.ClientType)
		assert.Equal(t, clientInfo.LoginFailedTimes, entity.LoginFailedTimes)
		assert.Equal(t, clientInfo.TokenExpireSecs, entity.TokenExpireSecs)

		// Find the client in cache
		key := fmt.Sprintf("%s:%d", ent_impl.ClientInfoPrefix, clientInfo.Id)
		aggregateClient := &aggregate.Client{}
		KgsErr := cache.GetObject(ctx, key, aggregateClient)
		require.Nil(t, KgsErr)
		assert.Equal(t, clientInfo.Id, aggregateClient.Id)
		assert.Equal(t, clientInfo.MerchantId, aggregateClient.MerchantId)
		assert.Equal(t, clientInfo.ClientType, aggregateClient.ClientType)
		assert.Equal(t, clientInfo.LoginFailedTimes, aggregateClient.LoginFailedTimes)
		assert.Equal(t, clientInfo.TokenExpireSecs, aggregateClient.TokenExpireSecs)
	})

	t.Run("CreateClientError", func(t *testing.T) {
		// Create a client
		req := &auth.CreateClientRequest{
			ClientId:         1111,
			MerchantId:       2,
			ClientType:       int32(clientInfo.ClientType.Id),
			LoginFailedTimes: int32(clientInfo.LoginFailedTimes),
			TokenExpireSecs:  int64(clientInfo.TokenExpireSecs),
		}

		// Create the same client again
		_, e := clientApp.CreateClient(ctx, req)
		require.Nil(t, e)

		// Create the same client again
		_, e = clientApp.CreateClient(ctx, req)
		kgsErr, ok := e.(*cus_err.CusError)
		require.NotNil(t, e)
		require.True(t, ok)
		assert.Equal(t, cus_err.ResourceIsExist, kgsErr.Code().Int())
	})

	t.Run("CreateClientInvalidClientType", func(t *testing.T) {
		// Create a client
		req := &auth.CreateClientRequest{
			ClientId:         int64(clientInfo.Id),
			MerchantId:       int64(clientInfo.MerchantId),
			ClientType:       9999,
			LoginFailedTimes: int32(clientInfo.LoginFailedTimes),
			TokenExpireSecs:  int64(clientInfo.TokenExpireSecs),
		}

		_, e := clientApp.CreateClient(ctx, req)
		kgsErr, ok := e.(*cus_err.CusError)
		require.NotNil(t, e)
		require.True(t, ok)
		assert.Equal(t, cus_err.AccountPasswordError, kgsErr.Code().Int())
	})
}

func TestUpdateClient(t *testing.T) {
	clientApp, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               12345,
		MerchantId:       11111,
		ClientType:       enum.ClientType.Frontend,
		LoginFailedTimes: 3,
		TokenExpireSecs:  3600,
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create a client
	client, e := tx.AuthClient.Create().
		SetID(clientInfo.Id).
		SetMerchantID(clientInfo.MerchantId).
		SetClientType(clientInfo.ClientType.Id).
		SetLoginFailedTimes(clientInfo.LoginFailedTimes).
		SetTokenExpireSecs(clientInfo.TokenExpireSecs).
		SetActive(true).
		SetSecret("secret").
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, client)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("UpdateClient", func(t *testing.T) {
		// Update a client
		req := &auth.UpdateClientRequest{
			ClientId:         int64(clientInfo.Id),
			LoginFailedTimes: 4,
			TokenExpireSecs:  7200,
			IsActive:         false,
		}

		_, e := clientApp.UpdateClient(ctx, req)
		require.Nil(t, e)

		entdb, ok := db.GetConn(ctx).(*ent.Client)
		require.True(t, ok)

		// Find the client in db
		entity, e := entdb.AuthClient.Get(ctx, clientInfo.Id)
		require.Nil(t, e)
		assert.Equal(t, clientInfo.Id, entity.ID)
		assert.Equal(t, clientInfo.MerchantId, entity.MerchantID)
		assert.Equal(t, clientInfo.ClientType.Id, entity.ClientType)
		assert.Equal(t, req.LoginFailedTimes, int32(entity.LoginFailedTimes))
		assert.Equal(t, req.TokenExpireSecs, int64(entity.TokenExpireSecs))
		assert.Equal(t, req.IsActive, entity.Active)

		// Find the client in cache
		key := fmt.Sprintf("%s:%d", ent_impl.ClientInfoPrefix, clientInfo.Id)
		aggregateClient := &aggregate.Client{}
		KgsErr := cache.GetObject(ctx, key, aggregateClient)
		require.Nil(t, KgsErr)
		assert.Equal(t, clientInfo.Id, aggregateClient.Id)
		assert.Equal(t, clientInfo.MerchantId, aggregateClient.MerchantId)
		assert.Equal(t, clientInfo.ClientType, aggregateClient.ClientType)
		assert.Equal(t, req.LoginFailedTimes, int32(aggregateClient.LoginFailedTimes))
		assert.Equal(t, req.TokenExpireSecs, int64(aggregateClient.TokenExpireSecs))
		assert.Equal(t, req.IsActive, aggregateClient.Active)
	})

	t.Run("Update non-exist client", func(t *testing.T) {
		// Update a client
		req := &auth.UpdateClientRequest{
			ClientId:         1111,
			LoginFailedTimes: 4,
			TokenExpireSecs:  7200,
			IsActive:         false,
		}

		_, e := clientApp.UpdateClient(ctx, req)
		kgsErr, ok := e.(*cus_err.CusError)
		require.NotNil(t, e)
		require.True(t, ok)
		assert.Equal(t, cus_err.ResourceNotFound, kgsErr.Code().Int())
	})

	t.Run("Update client invalid argument", func(t *testing.T) {
		// Update a client
		req := &auth.UpdateClientRequest{
			ClientId:         int64(clientInfo.Id),
			LoginFailedTimes: 0,
			TokenExpireSecs:  0,
			IsActive:         false,
		}

		_, e := clientApp.UpdateClient(ctx, req)
		kgsErr, ok := e.(*cus_err.CusError)
		require.NotNil(t, e)
		require.True(t, ok)
		assert.Equal(t, cus_err.AccountPasswordError, kgsErr.Code().Int())
	})
}

func TestCreateRole(t *testing.T) {
	clientApp, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create a client
	client, e := tx.AuthClient.Create().
		SetID(12345).
		SetMerchantID(1111).
		SetClientType(enum.ClientType.Frontend.Id).
		SetLoginFailedTimes(3).
		SetTokenExpireSecs(3600).
		SetActive(true).
		SetSecret("secret").
		AddRoleIDs(1, 2, 3).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, client)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("CreateRole", func(t *testing.T) {
		// Create a role
		req := &auth.CreateRoleRequest{
			ClientId: 12345,
			RoleName: "test",
			PermIds:  []int64{1, 2, 3},
		}

		res, e := clientApp.CreateRole(ctx, req)
		require.Nil(t, e)

		entdb, ok := db.GetConn(ctx).(*ent.Client)
		require.True(t, ok)

		// Find the role in db
		role, e := entdb.Role.Get(ctx, res.RoleId)
		require.Nil(t, e)
		assert.Equal(t, res.RoleId, role.ID)
		assert.Equal(t, res.RoleName, role.Name)

		// Find the role in cache
		key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, req.ClientId, res.RoleId)
		aggregateRole := &entity.Role{}
		KgsErr := cache.GetObject(ctx, key, aggregateRole)
		require.Nil(t, KgsErr)
		assert.Equal(t, res.RoleId, aggregateRole.Id)
		assert.Equal(t, res.RoleName, aggregateRole.Name)
	})

	t.Run("Create Role With Wrong PermIds", func(t *testing.T) {
		// Create a role
		req := &auth.CreateRoleRequest{
			ClientId: 12345,
			RoleName: "test",
			PermIds:  []int64{101, 102},
		}

		// Create the same role again
		_, e := clientApp.CreateRole(ctx, req)
		assert.NotNil(t, e)
		kgsErr, ok := e.(*cus_err.CusError)
		require.True(t, ok)
		assert.Equal(t, cus_err.AccountPasswordError, kgsErr.Code().Int())
	})
}

func TestUpdateRole(t *testing.T) {
	clientApp, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create a client
	client, e := tx.AuthClient.Create().
		SetID(12345).
		SetMerchantID(1111).
		SetClientType(enum.ClientType.Frontend.Id).
		SetLoginFailedTimes(3).
		SetTokenExpireSecs(3600).
		SetActive(true).
		SetSecret("secret").
		AddRoleIDs(1, 2, 3).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, client)

	// Create a role
	role, e := tx.Role.Create().
		SetID(123).
		AddAuthClientIDs(12345).
		SetName("test").
		SetPermissions([]enum.Permission{
			enum.PermissionType.PlayGame,
		}).
		SetIsSystem(false).
		SetClientType(enum.ClientType.Frontend.Id).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, role)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("UpdateRole", func(t *testing.T) {
		// Update a role
		req := &auth.UpdateRoleRequest{
			ClientId: 12345,
			RoleId:   123,
			RoleName: "test2",
			PermIds:  []int64{1, 2},
		}

		_, e := clientApp.UpdateRole(ctx, req)
		require.Nil(t, e)

		entdb, ok := db.GetConn(ctx).(*ent.Client)
		require.True(t, ok)

		// Find the role in db
		role, e := entdb.Role.Get(ctx, req.RoleId)
		require.Nil(t, e)
		assert.Equal(t, req.RoleId, role.ID)
		assert.Equal(t, req.RoleName, role.Name)
		assert.Equal(t, 2, len(role.Permissions))

		// Find the role in cache
		key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, req.ClientId, req.RoleId)
		aggregateRole := &entity.Role{}
		KgsErr := cache.GetObject(ctx, key, aggregateRole)
		require.Nil(t, KgsErr)
		assert.Equal(t, req.RoleId, aggregateRole.Id)
		assert.Equal(t, req.RoleName, aggregateRole.Name)
		assert.Equal(t, 2, len(aggregateRole.Permissions))
	})

	t.Run("UpdateRoleWithWrongPermIds", func(t *testing.T) {
		// Update a role
		req := &auth.UpdateRoleRequest{
			ClientId: 12345,
			RoleId:   123,
			RoleName: "test2",
			PermIds:  []int64{101, 102},
		}

		_, e := clientApp.UpdateRole(ctx, req)
		assert.NotNil(t, e)
		kgsErr, ok := e.(*cus_err.CusError)
		require.True(t, ok)
		assert.Equal(t, cus_err.AccountPasswordError, kgsErr.Code().Int())
	})

	t.Run("UpdateRoleWithWrongRoleId", func(t *testing.T) {
		// Update a role
		req := &auth.UpdateRoleRequest{
			ClientId: 12345,
			RoleId:   999,
			RoleName: "test2",
			PermIds:  []int64{1, 2},
		}

		_, e := clientApp.UpdateRole(ctx, req)
		assert.NotNil(t, e)
		kgsErr, ok := e.(*cus_err.CusError)
		require.True(t, ok)
		assert.Equal(t, cus_err.ResourceNotFound, kgsErr.Code().Int())
	})
}

func TestDeleteRole(t *testing.T) {
	clientApp, db, cache, closeFunc := setupClientService()
	defer closeFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create a client
	client, e := tx.AuthClient.Create().
		SetID(12345).
		SetMerchantID(1111).
		SetClientType(enum.ClientType.Frontend.Id).
		SetLoginFailedTimes(3).
		SetTokenExpireSecs(3600).
		SetActive(true).
		SetSecret("secret").
		AddRoleIDs(1, 2, 3).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, client)

	// Create a role
	role, e := tx.Role.Create().
		SetID(123).
		AddAuthClientIDs(12345).
		SetName("test").
		SetPermissions([]enum.Permission{
			enum.PermissionType.PlayGame,
		}).
		SetIsSystem(false).
		SetClientType(enum.ClientType.Frontend.Id).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, role)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("DeleteRole", func(t *testing.T) {
		// Delete a role
		req := &auth.DeleteRoleRequest{
			ClientId: 12345,
			RoleId:   123,
		}

		res, e := clientApp.DeleteRole(ctx, req)
		require.Nil(t, e)
		require.NotNil(t, res)

		entdb, ok := db.GetConn(ctx).(*ent.Client)
		require.True(t, ok)

		// Find the role in db
		role, e := entdb.Role.Get(ctx, req.RoleId)
		require.NotNil(t, e)
		assert.Nil(t, role)

		// Find the role in cache
		key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, req.ClientId, req.RoleId)
		aggregateRole := &entity.Role{}
		KgsErr := cache.GetObject(ctx, key, aggregateRole)
		require.NotNil(t, KgsErr)
		assert.Empty(t, aggregateRole)
	})

	t.Run("DeleteRoleWithWrongRoleId", func(t *testing.T) {
		// Delete a role
		req := &auth.DeleteRoleRequest{
			ClientId: 12345,
			RoleId:   999,
		}

		// Should not return error, but not any role deleted
		_, e := clientApp.DeleteRole(ctx, req)
		assert.Nil(t, e)

		// Find the role in db
		entdb, ok := db.GetConn(ctx).(*ent.Client)
		require.True(t, ok)
		role, e := entdb.Role.Get(ctx, req.RoleId)
		require.NotNil(t, e)
		assert.Nil(t, role)

		// Find the role in cache
		key := fmt.Sprintf("%s:%d:%d", ent_impl.RolePrefix, req.ClientId, req.RoleId)
		aggregateRole := &entity.Role{}
		KgsErr := cache.GetObject(ctx, key, aggregateRole)
		require.NotNil(t, KgsErr)
		assert.Empty(t, aggregateRole)

	})
}

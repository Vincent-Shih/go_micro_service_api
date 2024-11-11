package ent_impl

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/infrastructure/db_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent/migrate"
	"go_micro_service_api/pkg/cus_err"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/pkg/enum"
	"log"
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/alicebob/miniredis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// newMemoryCache create a new RedisCache instance with an in-memory cache
func newMemoryRedis() (client *redis.Client, closeFunc func()) {

	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("failed to open memory cache: %v", err)
	}

	client = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() { mr.Close() }
}

// newMemoryDB creates a new EntDB instance with an in-memory database
func newMemoryClientRepoImpl(t *testing.T) (*ClientRepoImpl, func()) {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&_fk=1")
	require.Nil(t, err)

	// Run the auto migration tool
	ctx := context.Background()
	err = client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true))
	require.Nil(t, err)

	// Seed the database with system roles
	r := make([]entity.Role, 0)
	r = append(r, entity.AllFrontendRoles...)
	r = append(r, entity.AllBackendRoles...)
	sysRoles, err := client.Role.MapCreateBulk(r, func(c *ent.RoleCreate, i int) {
		c.SetID(r[i].Id)
		c.SetName(r[i].Name)
		c.SetIsSystem(true)
		c.SetPermissions(r[i].Permissions)
		c.SetClientType(r[i].ClientType.Id)
	}).Save(ctx)
	require.Nil(t, err)
	require.NotNil(t, sysRoles)

	db := db_impl.NewEntDb(client).(*db_impl.EntDB)

	redisClient, closeFunc := newMemoryRedis()
	cache := redis_cache.NewRedisCache(redisClient)

	return NewClientRepoImpl(db, cache), closeFunc
}

func TestFindClient(t *testing.T) {
	repo, closeFunc := newMemoryClientRepoImpl(t)
	defer closeFunc()

	ctx := context.Background()

	// Begin transaction
	ctx, err := repo.db.Begin(ctx)
	require.Nil(t, err)

	// Create a test client
	testClient := &aggregate.Client{
		Id:               1,
		ClientType:       enum.ClientType.Frontend,
		MerchantId:       1,
		Secret:           "test_secret",
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	createdClient, err := repo.Create(ctx, testClient)
	require.Nil(t, err)
	require.NotNil(t, createdClient)

	// Test Find method
	t.Run("Find existing client", func(t *testing.T) {
		foundClient, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)
		assert.NotNil(t, foundClient)
		assert.Equal(t, createdClient.Id, foundClient.Id)
		assert.Equal(t, createdClient.ClientType, foundClient.ClientType)
		assert.Equal(t, createdClient.MerchantId, foundClient.MerchantId)
		assert.Equal(t, createdClient.Secret, foundClient.Secret)
		assert.Equal(t, createdClient.Active, foundClient.Active)
		assert.Equal(t, createdClient.TokenExpireSecs, foundClient.TokenExpireSecs)
		assert.Equal(t, createdClient.LoginFailedTimes, foundClient.LoginFailedTimes)

		roles, err := foundClient.Roles(ctx)
		assert.Equal(t, err.Code().Int(), cus_err.ResourceNotFound)
		assert.Nil(t, roles)
	})

	t.Run("Find non-existing client", func(t *testing.T) {
		foundClient, err := repo.Find(ctx, 9999)
		assert.Error(t, err)
		assert.Nil(t, foundClient)
		assert.Equal(t, err.Code().Int(), cus_err.ResourceNotFound)
		assert.Contains(t, err.Error(), "client not found")
	})
}

func TestCreate(t *testing.T) {
	repo, closFunc := newMemoryClientRepoImpl(t)
	defer closFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := repo.db.Begin(ctx)
	require.Nil(t, err)

	t.Run("Create new client", func(t *testing.T) {
		newClient := &aggregate.Client{
			Id:               1,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       1,
			Secret:           "new_secret",
			Active:           true,
			TokenExpireSecs:  7200,
			LoginFailedTimes: 3,
		}

		createdClient, err := repo.Create(ctx, newClient)
		assert.Nil(t, err)
		assert.NotNil(t, createdClient)
		assert.Equal(t, newClient.Id, createdClient.Id)
		assert.Equal(t, newClient.ClientType, createdClient.ClientType)
		assert.Equal(t, newClient.MerchantId, createdClient.MerchantId)
		assert.Equal(t, newClient.Secret, createdClient.Secret)
		assert.Equal(t, newClient.Active, createdClient.Active)
		assert.Equal(t, newClient.TokenExpireSecs, createdClient.TokenExpireSecs)
		assert.Equal(t, newClient.LoginFailedTimes, createdClient.LoginFailedTimes)

		// Verify the client was actually created
		foundClient, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)
		assert.NotNil(t, foundClient)
		assert.Equal(t, createdClient.Id, foundClient.Id)
	})

	t.Run("Create client with existing type for merchant", func(t *testing.T) {
		existingClient := &aggregate.Client{
			Id:               2,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       2,
			Secret:           "existing_secret",
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 3,
		}

		_, err := repo.Create(ctx, existingClient)
		assert.Nil(t, err)

		newClient := &aggregate.Client{
			Id:               3,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       2,
			Secret:           "new_secret",
			Active:           true,
			TokenExpireSecs:  7200,
			LoginFailedTimes: 3,
		}

		createdClient, err := repo.Create(ctx, newClient)
		assert.Error(t, err)
		assert.Nil(t, createdClient)
		assert.Contains(t, err.Error(), "client already exists for merchant")
	})

	t.Run("Create client with invalid parameters", func(t *testing.T) {
		invalidClient := &aggregate.Client{
			Id:               4,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       3,
			Secret:           "", // Invalid: empty secret
			Active:           true,
			TokenExpireSecs:  0, // Invalid: zero token expire seconds
			LoginFailedTimes: 0, // Invalid: zero login failed times
		}

		createdClient, err := repo.Create(ctx, invalidClient)
		assert.Error(t, err)
		assert.Nil(t, createdClient)
		assert.Contains(t, err.Error(), "client secret is required")
	})
}

func TestUpdate(t *testing.T) {
	repo, closFunc := newMemoryClientRepoImpl(t)
	defer closFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := repo.db.Begin(ctx)
	require.Nil(t, err)

	// Create a test client
	testClient := &aggregate.Client{
		Id:               1,
		ClientType:       enum.ClientType.Frontend,
		MerchantId:       1,
		Secret:           "old_secret",
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 5,
	}

	createdClient, err := repo.Create(ctx, testClient)
	require.Nil(t, err)
	require.NotNil(t, createdClient)

	t.Run("Update existing client", func(t *testing.T) {
		updatedClient := &aggregate.Client{
			Id:               createdClient.Id,
			ClientType:       createdClient.ClientType,
			MerchantId:       createdClient.MerchantId,
			Secret:           "new_secret",
			Active:           false,
			TokenExpireSecs:  7200,
			LoginFailedTimes: 1,
		}

		result, err := repo.Update(ctx, updatedClient)
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updatedClient.Id, result.Id)
		assert.Equal(t, testClient.Secret, result.Secret) // Secret should not be updated
		assert.Equal(t, updatedClient.Active, result.Active)
		assert.Equal(t, updatedClient.TokenExpireSecs, result.TokenExpireSecs)
		assert.Equal(t, updatedClient.LoginFailedTimes, result.LoginFailedTimes)
	})

	t.Run("Update non-existing client", func(t *testing.T) {
		nonExistingClient := &aggregate.Client{
			Id:               9999,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       1,
			Secret:           "new_secret",
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 0,
		}

		result, err := repo.Update(ctx, nonExistingClient)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "client not found")
		assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
	})
}

func TestCreateRoles(t *testing.T) {
	repo, closFunc := newMemoryClientRepoImpl(t)
	defer closFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := repo.db.Begin(ctx)
	require.Nil(t, err)

	// Create a test client
	testClient := &aggregate.Client{
		Id:               1,
		ClientType:       enum.ClientType.Frontend,
		MerchantId:       1,
		Secret:           "test_secret",
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 3,
	}

	createdClient, err := repo.Create(ctx, testClient)
	require.Nil(t, err)
	require.NotNil(t, createdClient)

	t.Run("Create roles for existing client", func(t *testing.T) {
		roles := []entity.Role{
			{
				Name: "Admin",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
					enum.PermissionType.Withdraw,
				},
				ClientType: enum.ClientType.Frontend,
			},
			{
				Name: "User",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
				},
				ClientType: enum.ClientType.Frontend,
			},
		}

		createdRoles, err := repo.CreateRoles(ctx, createdClient.Id, roles...)
		assert.Nil(t, err)
		assert.NotNil(t, createdRoles)
		assert.Len(t, createdRoles, 2)
		assert.Equal(t, "Admin", (createdRoles)[0].Name)
		assert.Equal(t, roles[0].Permissions, createdRoles[0].Permissions)
		assert.Equal(t, "User", (createdRoles)[1].Name)
		assert.Equal(t, roles[1].Permissions, createdRoles[1].Permissions)
		assert.Equal(t, roles[0].ClientType, createdRoles[0].ClientType)
		assert.Equal(t, roles[1].ClientType, createdRoles[1].ClientType)
	})

	t.Run("Create roles for non-existing client", func(t *testing.T) {
		roles := []entity.Role{
			{
				Name: "TestRole",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
				},
			},
		}

		createdRoles, err := repo.CreateRoles(ctx, 9999, roles...)
		assert.Error(t, err)
		assert.Nil(t, createdRoles)
		assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
	})
}

func TestBindSystemRoles(t *testing.T) {
	repo, closFunc := newMemoryClientRepoImpl(t)
	defer closFunc()

	ctx := context.Background()

	// Begin a transaction
	ctx, err := repo.db.Begin(ctx)
	require.Nil(t, err)

	// Create a test client
	testClient := &aggregate.Client{
		Id:               1,
		ClientType:       enum.ClientType.Frontend,
		MerchantId:       1,
		Secret:           "test_secret",
		Active:           true,
		TokenExpireSecs:  3600,
		LoginFailedTimes: 3,
	}

	createdClient, err := repo.Create(ctx, testClient)
	require.Nil(t, err)
	require.NotNil(t, createdClient)

	t.Run("Bind system roles to existing client", func(t *testing.T) {
		err := repo.BindSystemRoles(ctx, createdClient.Id, entity.AllFrontendRoles...)
		assert.Nil(t, err)

		// Verify roles are bound
		clientWithRoles, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)
		assert.NotNil(t, clientWithRoles)

		roles, err := clientWithRoles.Roles(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, roles)
		assert.Len(t, *roles, len(entity.AllFrontendRoles))
	})

	t.Run("Bind system roles to non-existing client", func(t *testing.T) {
		err := repo.BindSystemRoles(ctx, 9999, entity.AllFrontendRoles...)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})
}

func TestDeleteRoles(t *testing.T) {
	repo, closeFunc := newMemoryClientRepoImpl(t)
	defer closeFunc()

	// Helper function to setup test environment
	setup := func(t *testing.T) (context.Context, *aggregate.Client, []entity.Role) {
		ctx := context.Background()

		// Begin a new transaction for each test
		ctx, err := repo.db.Begin(ctx)
		require.Nil(t, err)

		// Create a test client
		testClient := &aggregate.Client{
			Id:               1,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       1,
			Secret:           "test_secret",
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 3,
		}

		createdClient, err := repo.Create(ctx, testClient)
		require.Nil(t, err)
		require.NotNil(t, createdClient)

		// Bind system roles
		err = repo.BindSystemRoles(ctx, createdClient.Id, entity.AllFrontendRoles...)
		require.Nil(t, err)

		// Create custom roles
		roles := []entity.Role{
			{
				Name: "Admin", Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
					enum.PermissionType.Withdraw,
				},
				ClientType: enum.ClientType.Frontend,
			},
			{
				Name: "User",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
				},
				ClientType: enum.ClientType.Frontend,
			},
		}
		createdRoles, err := repo.CreateRoles(ctx, createdClient.Id, roles...)
		require.Nil(t, err)
		require.NotNil(t, createdRoles)

		return ctx, createdClient, createdRoles
	}

	t.Run("Delete all roles for existing client", func(t *testing.T) {
		ctx, createdClient, createdRoles := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		ids := make([]int64, len(createdRoles))
		for i, role := range createdRoles {
			ids[i] = role.Id
		}

		err := repo.DeleteRoles(ctx, createdClient.Id, ids...)
		assert.Nil(t, err)

		// Verify roles are deleted
		clientWithRoles, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)
		assert.NotNil(t, clientWithRoles)

		roles, err := clientWithRoles.Roles(ctx)
		assert.Nil(t, err)

		// System roles should not be deleted
		for _, role := range *roles {
			assert.True(t, role.IsSystem())
		}

		// Custom roles should be deleted
		for _, createdRole := range createdRoles {
			assert.NotContains(t, *roles, createdRole)
		}
	})

	t.Run("Delete all roles for non-existing client", func(t *testing.T) {
		ctx, _, createdRoles := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		ids := make([]int64, len(createdRoles))
		for i, role := range createdRoles {
			ids[i] = role.Id
		}

		//  Is should not err , but no roles will be deleted
		err := repo.DeleteRoles(ctx, 9999, ids...)
		assert.Nil(t, err)

	})

	t.Run("Delete non-existing roles for existing client", func(t *testing.T) {
		ctx, createdClient, createdRoles := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		err := repo.DeleteRoles(ctx, createdClient.Id, 999)
		assert.Nil(t, err)

		// Verify roles are not deleted
		clientWithRoles, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)
		assert.NotNil(t, clientWithRoles)

		roleMap, err := clientWithRoles.Roles(ctx)
		assert.Nil(t, err)

		// Check if the original roles are still present
		for _, role := range createdRoles {
			_, exists := (*roleMap)[role.Id]
			assert.True(t, exists, "Role %s should exist", role.Name)
		}

		// Check if system roles are present
		for _, systemRole := range entity.AllFrontendRoles {
			_, exists := (*roleMap)[systemRole.Id]
			assert.True(t, exists, "System role %s should exist", systemRole.Name)
		}

		// Check if the non-existing role is not present
		_, exists := (*roleMap)[999]
		assert.False(t, exists, "Role with ID 999 should not exist")
	})

	t.Run("Delete system roles for existing client", func(t *testing.T) {
		ctx, createdClient, _ := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		err := repo.DeleteRoles(ctx, createdClient.Id, entity.FrontendRoles.Guest.Id)
		assert.Nil(t, err)
		roleMap, err := createdClient.Roles(ctx)
		require.Nil(t, err)

		// Check if system roles are present
		for _, systemRole := range entity.AllFrontendRoles {
			_, exists := (*roleMap)[systemRole.Id]
			assert.True(t, exists, "System role %s should exist", systemRole.Name)
		}
	})
}

func TestUpdateRoles(t *testing.T) {
	repo, closFunc := newMemoryClientRepoImpl(t)
	defer closFunc()

	// Helper function to setup test environment
	setup := func(t *testing.T) (context.Context, *aggregate.Client, []entity.Role) {
		ctx := context.Background()

		// Begin a new transaction for each test
		ctx, err := repo.db.Begin(ctx)
		require.Nil(t, err)

		// Create a test client
		testClient := &aggregate.Client{
			Id:               1,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       1,
			Secret:           "test_secret",
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 3,
		}

		createdClient, err := repo.Create(ctx, testClient)
		require.Nil(t, err)
		require.NotNil(t, createdClient)

		// Bind system roles
		err = repo.BindSystemRoles(ctx, createdClient.Id, entity.AllFrontendRoles...)
		require.Nil(t, err)

		// Create custom roles
		roles := []entity.Role{
			{
				Name: "Test1",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
					enum.PermissionType.Withdraw,
				},
				ClientType: enum.ClientType.Frontend,
			},
			{
				Name: "Test2",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
				},
				ClientType: enum.ClientType.Frontend,
			},
		}
		createdRoles, err := repo.CreateRoles(ctx, createdClient.Id, roles...)
		require.Nil(t, err)
		require.NotNil(t, createdRoles)

		return ctx, createdClient, createdRoles
	}

	findRoleByName := func(roleMap *map[int64]entity.Role, name string) (entity.Role, bool) {
		for _, role := range *roleMap {
			if role.Name == name {
				return role, true
			}
		}
		return entity.Role{}, false
	}

	t.Run("Update roles for existing client", func(t *testing.T) {
		ctx, createdClient, createdRoles := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		// Get the client's roles
		roleMap, err := createdClient.Roles(ctx)
		require.Nil(t, err)

		// Find the role which we want to update
		role, exists := findRoleByName(roleMap, "Test1")
		require.True(t, exists)
		role.Permissions = []enum.Permission{enum.PermissionType.Deposit}
		role.ClientType = enum.ClientType.Backend //  Change the client type should not affect the role

		// Update the role
		updated, err := repo.UpdateRoles(ctx, createdClient.Id, role)
		assert.Nil(t, err)
		assert.NotNil(t, updated)
		assert.Len(t, updated, 1)

		// Verify roles are updated
		clientWithRoles, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)

		// Verify the updated role is present
		roleMap, err = clientWithRoles.Roles(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(*roleMap), len(entity.AllFrontendRoles)+len(createdRoles))
		role, exists = findRoleByName(roleMap, "Test1")
		assert.True(t, exists)
		assert.Equal(t, []enum.Permission{enum.PermissionType.Deposit}, role.Permissions)
		assert.Equal(t, enum.ClientType.Frontend, role.ClientType)
	})

	t.Run("Update non-existing roles for existing client", func(t *testing.T) {
		ctx, createdClient, createdRoles := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		nonExistingRole := entity.Role{
			Name:        "NonExistingRole",
			Permissions: []enum.Permission{enum.PermissionType.Deposit},
		}

		updated, err := repo.UpdateRoles(ctx, createdClient.Id, nonExistingRole)
		assert.NotNil(t, err)
		assert.Nil(t, updated)

		// Verify roles are not updated
		clientWithRoles, err := repo.Find(ctx, createdClient.Id)
		assert.Nil(t, err)

		// Verify the original roles are still present
		roleMap, err := clientWithRoles.Roles(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(*roleMap), len(entity.AllFrontendRoles)+len(createdRoles))
	})

	t.Run("Update roles for non-existing client", func(t *testing.T) {
		ctx, _, _ := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		role := entity.Role{
			Name:        "TestRole",
			Permissions: []enum.Permission{enum.PermissionType.Deposit},
		}

		updated, err := repo.UpdateRoles(ctx, 9999, role)
		assert.NotNil(t, err)
		assert.Nil(t, updated)

		// Verify roles are not updated
		clientWithRoles, err := repo.Find(ctx, 9999)
		assert.NotNil(t, err)
		assert.Nil(t, clientWithRoles)
	})
}

func TestFindRole(t *testing.T) {
	repo, closFunc := newMemoryClientRepoImpl(t)
	defer closFunc()

	// Helper function to setup test environment
	setup := func(t *testing.T) (context.Context, *aggregate.Client, []entity.Role) {
		ctx := context.Background()

		// Begin a new transaction for each test
		ctx, err := repo.db.Begin(ctx)
		require.Nil(t, err)

		// Create a test client
		testClient := &aggregate.Client{
			Id:               1,
			ClientType:       enum.ClientType.Frontend,
			MerchantId:       1,
			Secret:           "test_secret",
			Active:           true,
			TokenExpireSecs:  3600,
			LoginFailedTimes: 3,
		}

		createdClient, err := repo.Create(ctx, testClient)
		require.Nil(t, err)
		require.NotNil(t, createdClient)

		// Bind system roles
		err = repo.BindSystemRoles(ctx, createdClient.Id, entity.AllFrontendRoles...)
		require.Nil(t, err)

		// Create custom roles
		roles := []entity.Role{
			{
				Name: "Test1",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
					enum.PermissionType.Withdraw,
				},
				ClientType: enum.ClientType.Frontend,
			},
			{
				Name: "Test2",
				Permissions: []enum.Permission{
					enum.PermissionType.Deposit,
				},
				ClientType: enum.ClientType.Frontend,
			},
		}
		createdRoles, err := repo.CreateRoles(ctx, createdClient.Id, roles...)
		require.Nil(t, err)
		require.NotNil(t, createdRoles)

		return ctx, createdClient, createdRoles
	}

	t.Run("Find existing role for existing client", func(t *testing.T) {
		ctx, createdClient, createdRoles := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		foundRole, err := repo.FindRole(ctx, createdClient.Id, createdRoles[0].Id)
		assert.Nil(t, err)
		assert.NotNil(t, foundRole)
		assert.Equal(t, createdRoles[0].Id, foundRole.Id)
		assert.Equal(t, createdRoles[0].Name, foundRole.Name)
		assert.Equal(t, createdRoles[0].Permissions, foundRole.Permissions)
	})

	t.Run("Find non-existing role for existing client", func(t *testing.T) {
		ctx, createdClient, _ := setup(t)
		defer func() {
			_, rollbackErr := repo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		foundRole, err := repo.FindRole(ctx, createdClient.Id, 9999)
		assert.NotNil(t, err)
		assert.Nil(t, foundRole)
		assert.Contains(t, err.Error(), "role not found")
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})
}

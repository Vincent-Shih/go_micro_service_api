package ent_impl

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/infrastructure/db_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent/migrate"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/enum"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMemoryUserRepoImpl(t *testing.T) *UserRepoImpl {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&_fk=1")
	require.Nil(t, err)

	// Run migration
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

	db := db_impl.NewEntDb(client)
	return NewUserRepoImpl(db)
}

func TestFindUser(t *testing.T) {
	userRepo := newMemoryUserRepoImpl(t)

	setup := func(t *testing.T) (context.Context, *ent.AuthClient, *ent.User) {
		ctx := context.Background()

		// Begin a transaction
		ctx, kgsErr := userRepo.db.Begin(ctx)
		require.Nil(t, kgsErr)

		tx, ok := userRepo.db.GetTx(ctx).(*ent.Tx)
		require.True(t, ok)

		// Create the test client
		entClient, err := tx.AuthClient.Create().
			SetID(1).
			SetActive(true).
			SetClientType(enum.ClientType.Frontend.Id).
			SetLoginFailedTimes(5).
			SetSecret("secret").
			SetMerchantID(1).
			SetTokenExpireSecs(3600).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, entClient)

		// Create the test user
		entUser, err := tx.User.Create().
			SetID(1).
			SetStatus(enum.UserStatusType.Active.Int()).
			SetAccount("testUser").
			SetPassword("password").
			SetPasswordFailTimes(3).
			SetAuthClientsID(entClient.ID).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, entUser)

		return ctx, entClient, entUser
	}

	t.Run("FindUser", func(t *testing.T) {
		ctx, testClient, testUser := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user, kgsErr := userRepo.Find(ctx, testUser.ID)

		assert.Nil(t, kgsErr)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.Id)
		assert.Equal(t, testUser.Status, int(user.Status))
		assert.Equal(t, testUser.Account, user.Account)
		assert.Equal(t, testUser.Password, user.Password)

		client, kgsErr := user.Client(ctx)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, client)
		assert.Equal(t, testClient.ID, client.Id)
		assert.Equal(t, testClient.Active, client.Active)
		assert.Equal(t, testClient.ClientType, client.ClientType.Id)
		assert.Equal(t, testClient.LoginFailedTimes, client.LoginFailedTimes)
		assert.Equal(t, testClient.Secret, client.Secret)
		assert.Equal(t, testClient.MerchantID, client.MerchantId)
		assert.Equal(t, testClient.TokenExpireSecs, client.TokenExpireSecs)
	})

	t.Run("FindUser_NotFound", func(t *testing.T) {
		ctx, _, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user, kgsErr := userRepo.Find(ctx, 2)

		require.NotNil(t, kgsErr)
		require.Nil(t, user)
		require.Equal(t, cus_err.ResourceNotFound, kgsErr.Code().Int())
	})
}

func TestCreateUser(t *testing.T) {
	userRepo := newMemoryUserRepoImpl(t)

	setup := func(t *testing.T) (context.Context, *ent.AuthClient, *ent.User) {
		ctx := context.Background()

		// Begin a transaction
		ctx, kgsErr := userRepo.db.Begin(ctx)
		require.Nil(t, kgsErr)

		tx, ok := userRepo.db.GetTx(ctx).(*ent.Tx)
		require.True(t, ok)

		// Create the test client
		entClient, err := tx.AuthClient.Create().
			SetID(1).
			SetActive(true).
			SetClientType(enum.ClientType.Frontend.Id).
			SetLoginFailedTimes(5).
			SetSecret("secret").
			SetMerchantID(1).
			SetTokenExpireSecs(3600).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, entClient)

		// Create the test user
		entUser, err := tx.User.Create().
			SetID(1).
			SetStatus(enum.UserStatusType.Active.Int()).
			SetAccount("testUser").
			SetPassword("password").
			SetPasswordFailTimes(3).
			SetAuthClientsID(entClient.ID).
			Save(ctx)

		require.Nil(t, err)
		require.NotNil(t, entUser)

		return ctx, entClient, entUser
	}

	t.Run("CreateUser", func(t *testing.T) {
		ctx, testClient, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user := &aggregate.User{
			Id:       2,
			Status:   enum.UserStatusType.Active,
			Account:  "testUser2",
			Password: "password",
		}

		user, kgsErr := userRepo.Create(ctx, testClient.ID, user)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, user)
		assert.Equal(t, int64(2), user.Id)
		assert.Equal(t, enum.UserStatusType.Active, user.Status)
		assert.Equal(t, "testUser2", user.Account)
		assert.Equal(t, "password", user.Password)

		client, kgsErr := user.Client(ctx)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, client)
		assert.Equal(t, testClient.ID, client.Id)
		assert.Equal(t, testClient.Active, client.Active)
		assert.Equal(t, testClient.ClientType, client.ClientType.Id)
	})

	t.Run("CreateUser Duplicate ID", func(t *testing.T) {
		ctx, testClient, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user := &aggregate.User{
			Id:       1,
			Status:   enum.UserStatusType.Active,
			Account:  "testUser2",
			Password: "password",
		}

		user, kgsErr := userRepo.Create(ctx, testClient.ID, user)
		assert.NotNil(t, kgsErr)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.InternalServerError, kgsErr.Code().Int())
	})

	t.Run("CreateUser Duplicate Account", func(t *testing.T) {
		ctx, testClient, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user := &aggregate.User{
			Id:       1,
			Status:   enum.UserStatusType.Active,
			Account:  "testUser",
			Password: "password",
		}

		user, kgsErr := userRepo.Create(ctx, testClient.ID, user)
		assert.NotNil(t, kgsErr)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.InternalServerError, kgsErr.Code().Int())
	})

	t.Run("CreateUser ClientNotFound", func(t *testing.T) {
		ctx, _, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user := &aggregate.User{
			Id:       2,
			Status:   enum.UserStatusType.Active,
			Account:  "testUser2",
			Password: "password",
		}

		user, kgsErr := userRepo.Create(ctx, 2, user)
		assert.NotNil(t, kgsErr)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.ResourceNotFound, kgsErr.Code().Int())
	})
}

func TestUpdateUser(t *testing.T) {
	userRepo := newMemoryUserRepoImpl(t)

	setup := func(t *testing.T) (context.Context, *ent.AuthClient, *ent.User) {
		ctx := context.Background()

		// Begin a transaction
		ctx, kgsErr := userRepo.db.Begin(ctx)
		require.Nil(t, kgsErr)

		tx, ok := userRepo.db.GetTx(ctx).(*ent.Tx)
		require.True(t, ok)

		// Create the test client
		entClient, err := tx.AuthClient.Create().
			SetID(1).
			SetActive(true).
			SetClientType(enum.ClientType.Frontend.Id).
			SetLoginFailedTimes(5).
			SetSecret("secret").
			SetMerchantID(1).
			SetTokenExpireSecs(3600).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, entClient)

		// Create the test user
		entUser, err := tx.User.Create().
			SetID(1).
			SetStatus(enum.UserStatusType.Active.Int()).
			SetAccount("testUser").
			SetPassword("password").
			SetPasswordFailTimes(3).
			SetAuthClientsID(entClient.ID).
			Save(ctx)

		require.Nil(t, err)
		require.NotNil(t, entUser)

		return ctx, entClient, entUser
	}

	t.Run("UpdateUser", func(t *testing.T) {
		ctx, testClient, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user := &aggregate.User{
			Id:       1,
			Status:   enum.UserStatusType.Locked,
			Account:  "testUser2",
			Password: "password2",
		}

		user, kgsErr := userRepo.Update(ctx, user)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, user)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, enum.UserStatusType.Locked, user.Status)
		assert.Equal(t, "testUser2", user.Account)
		assert.Equal(t, "password2", user.Password)

		client, kgsErr := user.Client(ctx)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, client)
		assert.Equal(t, testClient.ID, client.Id)
		assert.Equal(t, testClient.Active, client.Active)
		assert.Equal(t, testClient.ClientType, client.ClientType.Id)
	})

	t.Run("UpdateUser_NotFound", func(t *testing.T) {
		ctx, _, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user := &aggregate.User{
			Id:       2,
			Status:   enum.UserStatusType.Locked,
			Account:  "testUser2",
			Password: "password2",
		}

		user, kgsErr := userRepo.Update(ctx, user)
		assert.NotNil(t, kgsErr)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.ResourceNotFound, kgsErr.Code().Int())

	})

	t.Run("UpdateUser_DuplicateAccount", func(t *testing.T) {
		ctx, testClient, testUser := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		// Create another user
		user := &aggregate.User{
			Id:       2,
			Status:   enum.UserStatusType.Locked,
			Account:  "testUser2",
			Password: "password2",
		}
		user, kgsErr := userRepo.Create(ctx, testClient.ID, user)
		require.Nil(t, kgsErr)
		require.NotNil(t, user)
		require.Equal(t, int64(2), user.Id)
		require.Equal(t, enum.UserStatusType.Locked, user.Status)
		require.Equal(t, "testUser2", user.Account)

		// Update with duplicate account
		user.Account = testUser.Account
		user, kgsErr = userRepo.Update(ctx, user)
		assert.NotNil(t, kgsErr)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.InternalServerError, kgsErr.Code().Int())
	})
}

func TestAddLoginRecord(t *testing.T) {
	userRepo := newMemoryUserRepoImpl(t)

	setup := func(t *testing.T) (context.Context, *ent.AuthClient, *ent.User) {
		ctx := context.Background()

		// Begin a transaction
		ctx, kgsErr := userRepo.db.Begin(ctx)
		require.Nil(t, kgsErr)

		tx, ok := userRepo.db.GetTx(ctx).(*ent.Tx)
		require.True(t, ok)

		// Create the test client
		entClient, err := tx.AuthClient.Create().
			SetID(1).
			SetActive(true).
			SetClientType(enum.ClientType.Frontend.Id).
			SetLoginFailedTimes(5).
			SetSecret("secret").
			SetMerchantID(1).
			SetTokenExpireSecs(3600).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, entClient)

		// Create the test user
		entUser, err := tx.User.Create().
			SetID(1).
			SetStatus(enum.UserStatusType.Active.Int()).
			SetAccount("testUser").
			SetPassword("password").
			SetPasswordFailTimes(3).
			SetAuthClientsID(entClient.ID).
			Save(ctx)

		require.Nil(t, err)
		require.NotNil(t, entUser)

		return ctx, entClient, entUser
	}

	t.Run("AddLoginRecord", func(t *testing.T) {
		ctx, _, testUser := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		loginRecord := &entity.LoginRecord{
			Browser:   "Chrome",
			Ip:        "127.0.0.1",
			Os:        "Windows",
			Country:   "TW",
			City:      "Taipei",
			IsSuccess: true,
		}

		createdRecord, kgsErr := userRepo.AddLoginRecord(ctx, testUser.ID, loginRecord)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, loginRecord)
		assert.Equal(t, createdRecord.Browser, loginRecord.Browser)
		assert.Equal(t, createdRecord.Ip, loginRecord.Ip)
		assert.Equal(t, createdRecord.Os, loginRecord.Os)
		assert.Equal(t, createdRecord.Country, loginRecord.Country)
		assert.Equal(t, createdRecord.City, loginRecord.City)
		assert.Equal(t, createdRecord.IsSuccess, loginRecord.IsSuccess)
	})
}

func TestBindRole(t *testing.T) {
	userRepo := newMemoryUserRepoImpl(t)

	setup := func(t *testing.T) (context.Context, *ent.AuthClient, *ent.User, *ent.Role) {
		ctx := context.Background()

		// Begin a transaction
		ctx, kgsErr := userRepo.db.Begin(ctx)
		require.Nil(t, kgsErr)

		tx, ok := userRepo.db.GetTx(ctx).(*ent.Tx)
		require.True(t, ok)

		// Create the test client
		entClient, err := tx.AuthClient.Create().
			SetID(1).
			SetActive(true).
			SetClientType(enum.ClientType.Frontend.Id).
			SetLoginFailedTimes(5).
			SetSecret("secret").
			SetMerchantID(1).
			SetTokenExpireSecs(3600).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, entClient)

		// Create the test user
		entUser, err := tx.User.Create().
			SetID(1).
			SetStatus(enum.UserStatusType.Active.Int()).
			SetAccount("testUser").
			SetPassword("password").
			SetPasswordFailTimes(3).
			SetAuthClientsID(entClient.ID).
			Save(ctx)
		require.Nil(t, err)
		require.NotNil(t, entUser)

		// Create the test role
		role := entity.Role{
			Name:        "TestRole",
			Permissions: []enum.Permission{enum.PermissionType.Deposit},
			ClientType:  enum.ClientType.Frontend,
		}

		createdRole, err := tx.Role.Create().
			SetName(role.Name).
			SetIsSystem(role.IsSystem()).
			SetPermissions(role.Permissions).
			AddAuthClients(entClient).
			SetClientType(role.ClientType.Id).
			Save(ctx)

		require.Nil(t, err)
		require.NotNil(t, entUser)

		return ctx, entClient, entUser, createdRole
	}

	t.Run("BindRole", func(t *testing.T) {
		ctx, _, testUser, testRole := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		// Bind role
		user, kgsErr := userRepo.BindRole(ctx, testUser.ID, testRole.ID)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.Id)
		assert.Equal(t, testUser.Account, user.Account)
		assert.Equal(t, testUser.Status, int(user.Status))
		assert.Equal(t, testUser.Password, user.Password)
		assert.Equal(t, testUser.PasswordFailTimes, user.PasswordFailTimes)

		// Get user role
		role, kgsErr := user.Role(ctx)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, role)
		assert.Equal(t, role.Id, testRole.ID)
		assert.Equal(t, role.Name, testRole.Name)
		assert.Equal(t, role.Permissions, testRole.Permissions)
		assert.Equal(t, role.IsSystem(), testRole.IsSystem)

		// Bind other role
		user, kgsErr = userRepo.BindRole(ctx, testUser.ID, 1)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, user)

		// Get user role
		role, kgsErr = user.Role(ctx)
		assert.Nil(t, kgsErr)
		assert.NotNil(t, role)
		assert.Equal(t, role.Id, entity.FrontendRoles.Guest.Id)
		assert.Equal(t, role.Name, entity.FrontendRoles.Guest.Name)
		assert.Equal(t, role.Permissions, entity.FrontendRoles.Guest.Permissions)
		assert.Equal(t, role.IsSystem(), entity.FrontendRoles.Guest.IsSystem())
	})

	t.Run("BindRole_NotFound", func(t *testing.T) {
		ctx, _, testUser, _ := setup(t)
		defer func() {
			_, rollbackErr := userRepo.db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		// Bind role
		user, kgsErr := userRepo.BindRole(ctx, testUser.ID, 9999)
		assert.NotNil(t, kgsErr)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.ResourceNotFound, kgsErr.Code().Int())
	})
}

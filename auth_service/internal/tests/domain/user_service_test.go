package tests_domain

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/tests"
	"go_micro_service_api/pkg/cus_crypto"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/pkg/enum"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserService() (userService *service.UserService, db db.Database, cache db.Cache, closeFunc func()) {
	db = tests.NewMemoryDB()
	redis, closeFunc := tests.NewMemoryRedis()
	cache = redis_cache.NewRedisCache(redis)
	clientRepo := ent_impl.NewClientRepoImpl(db, cache)
	userRepo := ent_impl.NewUserRepoImpl(db)

	return service.NewUserService(clientRepo, userRepo), db, cache, closeFunc
}

func TestCreateUser(t *testing.T) {
	userService, db, _, closeFunc := setupUserService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               12345,
		MerchantId:       11111,
		ClientType:       enum.ClientType.Frontend,
		LoginFailedTimes: 3,
		TokenExpireSecs:  3600,
	}

	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)
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

	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	userInfo := vo.UserInfo{
		Id:       1,
		Account:  "test",
		Password: "password",
		Status:   enum.UserStatusType.Active,
	}

	t.Run("Create user successfully", func(t *testing.T) {

		// Begin the transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
		assert.Nil(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userInfo.Id, user.Id)
		assert.Equal(t, userInfo.Account, user.Account)
		assert.Equal(t, userInfo.Status, user.Status)

		crypto := cus_crypto.New()
		require.Nil(t, err)
		assert.True(t, crypto.CompareHashAndPassword(context.Background(), user.Password, userInfo.Password))
		assert.Equal(t, 0, user.PasswordFailTimes)

		// Check the user client
		client, err := user.Client(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, clientInfo.Id, client.Id)
		assert.Equal(t, clientInfo.MerchantId, client.MerchantId)
		assert.Equal(t, clientInfo.ClientType, client.ClientType)
		assert.Equal(t, clientInfo.LoginFailedTimes, client.LoginFailedTimes)
		assert.Equal(t, clientInfo.TokenExpireSecs, client.TokenExpireSecs)
	})

	t.Run("Create user with empty account", func(t *testing.T) {
		// Begin the transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		userInfo.Account = ""
		user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
		assert.NotNil(t, err)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})

	t.Run("Create user with empty id", func(t *testing.T) {
		// Begin the transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		userInfo.Account = "test"
		userInfo.Id = 0
		user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
		assert.NotNil(t, err)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})

	t.Run("Create user with empty status", func(t *testing.T) {
		// Begin the transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		userInfo.Id = 1
		userInfo.Status = 0
		user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
		assert.NotNil(t, err)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})

	t.Run("Create user with empty password", func(t *testing.T) {
		// Begin the transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		userInfo.Status = enum.UserStatusType.Active
		userInfo.Password = ""
		user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
		assert.Nil(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userInfo.Id, user.Id)
		assert.Equal(t, userInfo.Account, user.Account)
		assert.Equal(t, userInfo.Status, user.Status)
		assert.Equal(t, "", user.Password)
		assert.Equal(t, 0, user.PasswordFailTimes)
	})
}

func TestUpdateUser(t *testing.T) {
	userService, db, _, closeFunc := setupUserService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               12345,
		MerchantId:       11111,
		ClientType:       enum.ClientType.Frontend,
		LoginFailedTimes: 3,
		TokenExpireSecs:  3600,
	}

	userInfo := vo.UserInfo{
		Id:       1,
		Account:  "test",
		Password: "password",
		Status:   enum.UserStatusType.Active,
	}

	// Begin the transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create the client
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

	// Create the user
	user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
	require.Nil(t, err)
	require.NotNil(t, user)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Update user successfully", func(t *testing.T) {
		// Begin the transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyUserInfo := &vo.UserInfo{}
		e := copier.CopyWithOption(copyUserInfo, userInfo, copier.Option{IgnoreEmpty: true})
		require.Nil(t, e)

		copyUserInfo.Account = "newTest"
		copyUserInfo.Password = "newPassword"
		copyUserInfo.Status = enum.UserStatusType.Locked

		user, err := userService.UpdateUser(ctx, *copyUserInfo)
		assert.Nil(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, copyUserInfo.Id, user.Id)
		assert.Equal(t, copyUserInfo.Account, user.Account)
		assert.Equal(t, copyUserInfo.Status, user.Status)
		assert.Equal(t, "FKiLnS9SxVtfvPnF2cEYdQ==", user.Password)
		assert.Equal(t, 0, user.PasswordFailTimes)
	})

	t.Run("Update user with empty account", func(t *testing.T) {
		// Begin the transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyUserInfo := &vo.UserInfo{}
		e := copier.CopyWithOption(copyUserInfo, userInfo, copier.Option{IgnoreEmpty: true})
		require.Nil(t, e)

		copyUserInfo.Account = ""
		user, err := userService.UpdateUser(ctx, *copyUserInfo)
		assert.NotNil(t, err)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})

	t.Run("Update user with empty id", func(t *testing.T) {
		// Begin the transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyUserInfo := &vo.UserInfo{}
		e := copier.CopyWithOption(copyUserInfo, userInfo, copier.Option{IgnoreEmpty: true})
		require.Nil(t, e)

		copyUserInfo.Account = "test"
		copyUserInfo.Id = 0

		user, err := userService.UpdateUser(ctx, *copyUserInfo)
		assert.NotNil(t, err)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})

	t.Run("Update user with empty status", func(t *testing.T) {
		// Begin the transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		copyUserInfo := &vo.UserInfo{}
		e := copier.CopyWithOption(copyUserInfo, userInfo, copier.Option{IgnoreEmpty: true})
		require.Nil(t, e)

		copyUserInfo.Id = 1
		copyUserInfo.Status = 0

		user, err := userService.UpdateUser(ctx, *copyUserInfo)
		assert.NotNil(t, err)
		assert.Nil(t, user)
		assert.Equal(t, cus_err.AccountPasswordError, err.Code().Int())
	})
}

func TestGetUser(t *testing.T) {
	userService, db, _, closeFunc := setupUserService()
	defer closeFunc()

	ctx := context.Background()

	clientInfo := vo.ClientInfo{
		Id:               12345,
		MerchantId:       11111,
		ClientType:       enum.ClientType.Frontend,
		LoginFailedTimes: 3,
		TokenExpireSecs:  3600,
	}

	userInfo := vo.UserInfo{
		Id:       1,
		Account:  "test",
		Password: "password",
		Status:   enum.UserStatusType.Active,
	}

	// Begin the transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create the client
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

	// Create the user
	user, err := userService.CreateUser(ctx, clientInfo.Id, userInfo)
	require.Nil(t, err)
	require.NotNil(t, user)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Get user successfully", func(t *testing.T) {
		// Begin the transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		crypto := cus_crypto.New()
		getUser, err := userService.GetUser(ctx, userInfo.Id)
		assert.Nil(t, err)
		assert.NotNil(t, getUser)
		assert.Equal(t, userInfo.Id, getUser.Id)
		assert.Equal(t, userInfo.Account, getUser.Account)
		assert.Equal(t, userInfo.Status, getUser.Status)
		assert.True(t, crypto.CompareHashAndPassword(context.Background(), getUser.Password, userInfo.Password))
		assert.Equal(t, 0, getUser.PasswordFailTimes)
	})

	t.Run("Get user with empty id", func(t *testing.T) {
		// Begin the transaction
		ctx, err = db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		getUser, err := userService.GetUser(ctx, 0)
		assert.NotNil(t, err)
		assert.Nil(t, getUser)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})
}

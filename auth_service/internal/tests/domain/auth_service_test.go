package tests_domain

import (
	"context"
	"fmt"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/repository"
	"go_micro_service_api/auth_service/internal/domain/service"
	"go_micro_service_api/auth_service/internal/domain/vo"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/infrastructure/token_helper"
	"go_micro_service_api/auth_service/internal/tests"
	"go_micro_service_api/pkg/cus_crypto"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/pkg/enum"
	"testing"

	"github.com/jinzhu/copier"
	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthService() (authService *service.AuthService, clientRepo repository.ClientRepo, db db.Database, cache db.Cache, closeFunc func()) {
	db = tests.NewMemoryDB()
	redis, closeFunc := tests.NewMemoryRedis()
	userRepo := ent_impl.NewUserRepoImpl(db)
	cache = redis_cache.NewRedisCache(redis)
	clientRepo = ent_impl.NewClientRepoImpl(db, cache)
	tokenHelper := token_helper.NewJwtToken()

	return service.NewAuthService(clientRepo, userRepo, cache, tokenHelper), clientRepo, db, cache, closeFunc
}

func TestCreateClientToken(t *testing.T) {
	authService, clientRepo, db, cache, closeFunc := setupAuthService()
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

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Create Client Token Success", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, err := db.Rollback(ctx)
			require.Nil(t, err)
		}()

		token, err := authService.CreateClientToken(ctx, clientInfo.Id)
		assert.Nil(t, err)
		assert.NotNil(t, token)
		assert.NotEmpty(t, token)

		// Check if the token is in the cache
		key := fmt.Sprintf("%s:%s", service.TokenPrefix, token.Token)
		cacheToken, err := cache.Get(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, token.Token, cacheToken)
	})

	t.Run("Create Client Token With no exists client", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, err := db.Rollback(ctx)
			require.Nil(t, err)
		}()

		token, err := authService.CreateClientToken(ctx, 999)
		assert.Empty(t, token)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})

	t.Run("Create Client With Inactive Client", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, err := db.Rollback(ctx)
			require.Nil(t, err)
		}()

		// Deep copy the client , set active to false
		copyClient := &aggregate.Client{}
		_ = copier.CopyWithOption(copyClient, client, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		copyClient.Active = false

		// Update the client
		_, kgsErr := clientRepo.Update(ctx, copyClient)
		require.Nil(t, kgsErr)

		token, err := authService.CreateClientToken(ctx, clientInfo.Id)
		assert.Empty(t, token)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ClientInactive, err.Code().Int())

		// Check if the token is in the cache
		key := fmt.Sprintf("%s:%d", service.TokenPrefix, clientInfo.Id)
		cacheToken, err := cache.Get(ctx, key)
		assert.Empty(t, cacheToken)
		assert.NotNil(t, err)
	})

	t.Run("Create Client With no exists client", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, err := db.Rollback(ctx)
			require.Nil(t, err)
		}()

		token, err := authService.CreateClientToken(ctx, 999)
		assert.Empty(t, token)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.ResourceNotFound, err.Code().Int())
	})
}

func TestGetTokenPayload(t *testing.T) {
	authService, _, db, _, closeFunc := setupAuthService()
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

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create the client
	_, e := tx.AuthClient.Create().
		SetID(clientInfo.Id).
		SetMerchantID(clientInfo.MerchantId).
		SetClientType(clientInfo.ClientType.Id).
		SetLoginFailedTimes(clientInfo.LoginFailedTimes).
		SetTokenExpireSecs(clientInfo.TokenExpireSecs).
		SetActive(true).
		SetSecret("secret").
		Save(ctx)
	require.Nil(t, e)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Get Token Payload Success", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, err := db.Rollback(ctx)
			require.Nil(t, err)
		}()

		result, err := authService.CreateClientToken(ctx, clientInfo.Id)
		require.Nil(t, err)

		payload, err := authService.GetTokenPayload(ctx, result.Token)
		assert.Nil(t, err)
		assert.NotNil(t, payload)
		assert.Equal(t, clientInfo.Id, payload.ClientId)
	})

	t.Run("Get Token Payload With no exists token", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, err := db.Rollback(ctx)
			require.Nil(t, err)
		}()

		payload, err := authService.GetTokenPayload(ctx, "noExistsToken")
		assert.Empty(t, payload)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.Unauthorized, err.Code().Int())
	})
}

func TestLogin(t *testing.T) {
	authService, _, db, cache, closeFunc := setupAuthService()
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

	user := &aggregate.User{
		Id:       123456789,
		Account:  "account",
		Password: "password",
		Status:   enum.UserStatusType.Active,
	}

	// Begin a transaction
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)

	// Get the transaction
	tx, ok := db.GetTx(ctx).(*ent.Tx)
	require.True(t, ok)

	// Create the client
	_, e := tx.AuthClient.Create().
		SetID(clientInfo.Id).
		SetMerchantID(clientInfo.MerchantId).
		SetClientType(clientInfo.ClientType.Id).
		SetLoginFailedTimes(clientInfo.LoginFailedTimes).
		SetTokenExpireSecs(clientInfo.TokenExpireSecs).
		SetActive(clientInfo.Active).
		SetSecret("secret").
		Save(ctx)
	require.Nil(t, e)

	// Create a user
	crypto := cus_crypto.New()
	pwd, err := crypto.HashPassword(ctx, user.Password)
	require.Nil(t, err)

	_, e = tx.User.Create().
		SetID(user.Id).
		SetAccount(user.Account).
		SetPassword(pwd).
		SetPasswordFailTimes(0).
		SetStatus(enum.UserStatusType.Active.Int()).
		SetRolesID(1).
		Save(ctx)
	require.Nil(t, e)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	// create a Auth Token
	cToken, err := authService.CreateClientToken(ctx, clientInfo.Id)
	require.Nil(t, err)

	t.Run("Login Success", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		token, err := authService.Login(ctx, cToken.Token, user.Id, user.Password, false)
		assert.Nil(t, err)
		assert.NotNil(t, token)

		// Check if the token is in the cache
		key := fmt.Sprintf("%s:%d", service.TokenPrefix, user.Id)
		cacheToken, err := cache.Get(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, token.Token, cacheToken)
	})

	t.Run("Duplicate login With  u token", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		// create a Auth Token
		cToken, err := authService.CreateClientToken(ctx, clientInfo.Id)
		require.Nil(t, err)

		// First time
		token, err := authService.Login(ctx, cToken.Token, user.Id, user.Password, false)
		assert.Nil(t, err)
		assert.NotNil(t, token)
		assert.NotEmpty(t, token)

		// Use the same cToken to login, should not success
		empty, err := authService.Login(ctx, cToken.Token, user.Id, user.Password, false)
		assert.Empty(t, empty)
		assert.NotNil(t, err)

		// Use the new token to login ,should success
		token, err = authService.Login(ctx, token.Token, user.Id, user.Password, false)
		assert.Nil(t, err)
		assert.NotNil(t, token)
		assert.NotEmpty(t, token)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		// Begin a transaction
		ctx, err := db.Begin(ctx)
		require.Nil(t, err)
		defer func() {
			_, rollbackErr := db.Rollback(ctx)
			require.Nil(t, rollbackErr)
		}()

		// create a Auth Token
		cToken, err := authService.CreateClientToken(ctx, clientInfo.Id)
		require.Nil(t, err)

		token, err := authService.Login(ctx, cToken.Token, user.Id, "wrongpassword", false)
		assert.Empty(t, token)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.WrongPassword, err.Code().Int())

		// Decrease the login failed times
		for i := 0; i < 5; i++ {
			token, err := authService.Login(ctx, cToken.Token, user.Id, "wrongpassword", false)
			assert.Empty(t, token)
			assert.NotNil(t, err)
		}

		// The account should be locked
		token, err = authService.Login(ctx, cToken.Token, user.Id, user.Password, false)
		assert.Empty(t, token)
		assert.NotNil(t, err)
		assert.Equal(t, cus_err.AccountLocked, err.Code().Int())
	})
}

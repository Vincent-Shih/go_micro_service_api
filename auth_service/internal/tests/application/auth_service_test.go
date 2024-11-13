package application_tests

import (
	"context"
	"go_micro_service_api/auth_service/internal/application"
	domainService "go_micro_service_api/auth_service/internal/domain/service"
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
	"go_micro_service_api/pkg/pb/gen/auth"
	"go_micro_service_api/pkg/req_analyzer"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthApplication() (authApp *application.AuthService, db db.Database, cache db.Cache, closeFunc func()) {
	db = tests.NewMemoryDB()
	redis, closeFunc := tests.NewMemoryRedis()
	cache = redis_cache.NewRedisCache(redis)
	clientRepo := ent_impl.NewClientRepoImpl(db, cache)
	userRepo := ent_impl.NewUserRepoImpl(db)
	tokenHelper := token_helper.NewJwtToken()

	authService := domainService.NewAuthService(clientRepo, userRepo, cache, tokenHelper)
	clientService := domainService.NewClientService(clientRepo)
	userService := domainService.NewUserService(clientRepo, userRepo)
	reqAnalyzer := req_analyzer.NewReqAnalyzer()
	authApp = application.NewAuthService(authService, clientService, userService, db, reqAnalyzer)

	return authApp, db, cache, closeFunc
}

func TestClientAuth(t *testing.T) {
	authApp, db, _, closeFunc := setupAuthApplication()
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

	t.Run("Client Auth", func(t *testing.T) {
		req := &auth.ClientAuthRequest{
			ClientId: clientInfo.Id,
		}

		res, err := authApp.ClientAuth(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Client Auth Fail", func(t *testing.T) {
		req := &auth.ClientAuthRequest{
			ClientId: 0,
		}

		res, err := authApp.ClientAuth(ctx, req)
		assert.NotNil(t, err)
		assert.Nil(t, res)
	})
}

func TestLogin(t *testing.T) {
	authApp, db, _, closeFunc := setupAuthApplication()
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
		Id:       12345,
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

	// Create a user
	crypto := cus_crypto.New()
	hashPwd, err := crypto.HashPassword(ctx, userInfo.Password)
	require.Nil(t, err)

	user, e := tx.User.Create().
		SetID(userInfo.Id).
		SetAccount(userInfo.Account).
		SetPassword(hashPwd).
		SetPasswordFailTimes(0).
		SetStatus(userInfo.Status.Int()).
		SetRolesID(1).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, user)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Login", func(t *testing.T) {
		// Create a c token
		cToken, err := authApp.ClientAuth(ctx, &auth.ClientAuthRequest{
			ClientId: clientInfo.Id,
		})
		assert.Nil(t, err)
		assert.NotNil(t, cToken)

		req := &auth.LoginRequest{
			UserId:      userInfo.Id,
			AccessToken: cToken.AccessToken,
			Password:    "password",
		}

		res, err := authApp.Login(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Login without cToken", func(t *testing.T) {
		req := &auth.LoginRequest{
			UserId:      userInfo.Id,
			AccessToken: "",
			Password:    "password",
		}

		res, err := authApp.Login(ctx, req)
		assert.NotNil(t, err)
		assert.Nil(t, res)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		// Create a c token
		cToken, err := authApp.ClientAuth(ctx, &auth.ClientAuthRequest{
			ClientId: clientInfo.Id,
		})
		assert.Nil(t, err)
		assert.NotNil(t, cToken)

		req := &auth.LoginRequest{
			UserId:      userInfo.Id,
			AccessToken: cToken.AccessToken,
			Password:    "111",
		}

		res, err := authApp.Login(ctx, req)
		cusErr := err.(*cus_err.CusError)
		assert.NotNil(t, err)
		assert.Nil(t, res)
		assert.Equal(t, cus_err.WrongPassword, cusErr.Code().Int())
	})
}

func TestValidToken(t *testing.T) {
	authApp, db, _, closeFunc := setupAuthApplication()
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
		Id:       12345,
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

	// Create a user
	crypto := cus_crypto.New()
	hashPwd, err := crypto.HashPassword(ctx, userInfo.Password)
	require.Nil(t, err)

	user, e := tx.User.Create().
		SetID(userInfo.Id).
		SetAccount(userInfo.Account).
		SetPassword(hashPwd).
		SetPasswordFailTimes(0).
		SetStatus(userInfo.Status.Int()).
		SetRolesID(1).
		Save(ctx)
	require.Nil(t, e)
	require.NotNil(t, user)

	// Commit the transaction
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	t.Run("Valid cToken", func(t *testing.T) {
		// Create a c token
		cToken, err := authApp.ClientAuth(ctx, &auth.ClientAuthRequest{
			ClientId: clientInfo.Id,
		})
		assert.Nil(t, err)
		assert.NotNil(t, cToken)

		req := &auth.ValidTokenRequest{
			AccessToken: cToken.AccessToken,
		}

		res, err := authApp.ValidToken(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, clientInfo.Id, res.ClientId)
		assert.Equal(t, clientInfo.MerchantId, res.MerchantId)
		assert.Nil(t, res.UserAccount)
		assert.Nil(t, res.UserId)
		assert.Nil(t, res.Role)
	})

	t.Run("Valid uToken missing role", func(t *testing.T) {
		// Create a c token
		cToken, err := authApp.ClientAuth(ctx, &auth.ClientAuthRequest{
			ClientId: clientInfo.Id,
		})
		assert.Nil(t, err)
		assert.NotNil(t, cToken)

		// Create a u token
		uToken, err := authApp.Login(ctx, &auth.LoginRequest{
			UserId:      userInfo.Id,
			AccessToken: cToken.AccessToken,
			Password:    "password",
		})
		assert.Nil(t, err)

		req := &auth.ValidTokenRequest{
			AccessToken: uToken.AccessToken,
		}

		res, err := authApp.ValidToken(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, clientInfo.Id, res.ClientId)
		assert.Equal(t, clientInfo.MerchantId, res.MerchantId)
		assert.Equal(t, userInfo.Account, *res.UserAccount)
		assert.Equal(t, userInfo.Id, *res.UserId)
		assert.Nil(t, res.Role)
	})
}

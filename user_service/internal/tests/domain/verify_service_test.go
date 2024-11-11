package domain_test

// import (
// 	"context"
// 	"go_micro_service_api/pkg/db"
// 	redis_cache "go_micro_service_api/pkg/db/redis"
// 	"go_micro_service_api/user_service/internal/domain/repository"
// 	"go_micro_service_api/user_service/internal/domain/service"
// 	"go_micro_service_api/user_service/internal/infrastructure/redis_impl"
// 	"go_micro_service_api/user_service/internal/tests"
// 	"testing"

// 	"github.com/go-playground/validator/v10"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func setupVerifyService() (verifyService *service.VerifyService, verifyRepo repository.VerifyRepo, cache db.Cache, closeFunc func()) {
// 	redis, closeFunc := tests.NewMemoryRedis()
// 	cache = redis_cache.NewRedisCache(redis)
// 	verifyRepo = redis_impl.NewVerifyRepo(cache)

// 	return service.NewVerifyService(verifyRepo), verifyRepo, cache, closeFunc
// }

// func TestRegisterVerification(t *testing.T) {
// 	service, _, cache, _ := setupVerifyService()
// 	ctx := context.Background()

// 	t.Run("success session", func(t *testing.T) {
// 		session, err := service.RegisterVerification(ctx, "test@gmail.com")
// 		require.NoError(t, err)
// 		require.NotNil(t, session)

// 		validate := validator.New()
// 		assert.NoError(t, validate.Var(session.Code, "required,len=6,number"))
// 		assert.NoError(t, validate.Var(session.Prefix, "required,len=3,alpha"))
// 		assert.NoError(t, validate.Var(session.Token, "required,alphanum"))
// 	})

// 	t.Run("notify lock exist", func(t *testing.T) {
// 		session, err := service.RegisterVerification(ctx, "test@gmail.com")
// 		require.NoError(t, err)
// 		require.NotNil(t, session)

// 		_, err = cache.Get(ctx, session.GetNotifyLockRedisKey())
// 		assert.NoError(t, err)
// 	})
// }

// // func TestVerification(t *testing.T) {
// // 	service, _, _, _ := setupVerifyService()
// // 	ctx := context.Background()

// // 	tcs := []struct {
// // 		name    string
// // 		wantErr bool
// // 	}{}
// // 	for _, tc := range tcs {
// // 		t.Run(tc.name, func(t *testing.T) {
// // 			session
// // 			err := service.Verification(ctx, session)
// // 			if tc.wantErr {
// // 				require.Error(t, err)
// // 			} else {
// // 				require.NoError(t, err)
// // 			}
// // 		})
// // 	}
// // }

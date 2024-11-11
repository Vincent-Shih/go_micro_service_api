package application_test

// import (
// 	"go_micro_service_api/pkg/db"
// 	redis_cache "go_micro_service_api/pkg/db/redis"
// 	"go_micro_service_api/user_service/internal/application"
// 	domainService "go_micro_service_api/user_service/internal/domain/service"
// 	"go_micro_service_api/user_service/internal/infrastructure/ent_impl"
// 	"go_micro_service_api/user_service/internal/tests"
// 	"testing"
// )

// func setUpVerifyApplication() (verifyApp *application.VerifyService, db db.Database, cache db.Cache, closeFunc func()) {
// 	db = tests.NewMemoryDB()
// 	redis, closeFunc := tests.NewMemoryRedis()
// 	cache = redis_cache.NewRedisCache(redis)
// 	verifyRepo := ent_impl.NewVerifyRepo(cache)

// 	verifyService := domainService.NewVerifyService(verifyRepo)
// 	verifyApp = application.NewVerifyService(verifyService)

// 	return verifyApp, db, cache, closeFunc
// }

// func TestRegisterVerification(t *testing.T) {
// 	app, _, cache, closeFunc := setUpVerifyApplication()
// 	defer closeFunc()

// }

// func TestVerification(t *testing.T) {
// 	_, _, _, closeFunc := setUpVerifyApplication()
// 	defer closeFunc()

// }

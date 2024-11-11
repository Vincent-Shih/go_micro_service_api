package application_test

import (
	"go_micro_service_api/pkg/db"
	redis_cache "go_micro_service_api/pkg/db/redis"
	"go_micro_service_api/user_service/internal/application"
	domainService "go_micro_service_api/user_service/internal/domain/service"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/user_service/internal/tests"
	"testing"
)

func setUpUserApplication() (userApp *application.UserService, db db.Database, cache db.Cache, closeFunc func()) {
	db = tests.NewMemoryDB()
	redis, closeFunc := tests.NewMemoryRedis()
	cache = redis_cache.NewRedisCache(redis)
	userRepo := ent_impl.NewUserRepo(db)

	userService := domainService.NewUserService(userRepo)
	userApp = application.NewUserService(userService, db)

	return userApp, db, cache, closeFunc
}

func TestCreateProfile(t *testing.T) {
	_, _, _, closeFunc := setUpUserApplication()
	defer closeFunc()

}

func TestFindProfile(t *testing.T) {

}

func TestCheckEmailExistence(t *testing.T) {

}

func TestCheckMobileNumberExistence(t *testing.T) {

}

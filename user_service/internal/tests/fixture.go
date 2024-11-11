package tests

import (
	"context"
	"go_micro_service_api/user_service/internal/infrastructure/db_impl"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/migrate"
	"log"

	"github.com/alicebob/miniredis"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

// newMemoryDB create a new EntDB instance with an in-memory database
func NewMemoryDB() *db_impl.EntDB {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&_fk=1")
	if err != nil {
		log.Fatalf("failed to open memory database: %v", err)
	}

	// Run the auto migration tool.
	ctx := context.Background()
	err = client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true))
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	return db_impl.NewEntDb(client).(*db_impl.EntDB)
}

// newMemoryCache create a new RedisCache instance with an in-memory cache
func NewMemoryRedis() (client *redis.Client, closeFunc func()) {

	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("failed to open memory cache: %v", err)
	}

	client = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() { mr.Close() }
}

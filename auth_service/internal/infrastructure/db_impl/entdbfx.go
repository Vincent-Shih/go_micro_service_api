package db_impl

import (
	"context"
	"fmt"
	"go_micro_service_api/auth_service/internal/config"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent/migrate"
	"go_micro_service_api/pkg/db"
	"log"
	"time"

	"go.uber.org/fx"
)

const roleAutoIncrementValue = 1000

func NewDriver() ent.Option {
	cfg := config.GetConfig()
	driver := db.NewDriver(
		cfg.DB.User,
		cfg.DB.Pass,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	// extra configurations
	db := driver.DB()
	db.SetMaxIdleConns(cfg.DB.MaxIdle)
	db.SetMaxOpenConns(cfg.DB.MaxConn)
	db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnLife) * time.Second)

	return ent.Driver(driver)
}

func NewClient(driver ent.Option) *ent.Client {
	return ent.NewClient(driver)
}

func AutoMigrate(client *ent.Client) {
	if !config.GetConfig().DB.AutoMigrate {
		return
	}

	ctx := context.Background()

	// Auto migrate
	if err := client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		log.Fatalf("failed to create schema resources: %v", err)
	}

	// Seed default roles
	allSystemRoles := append(entity.AllBackendRoles, entity.AllFrontendRoles...)
	err := client.Role.MapCreateBulk(allSystemRoles, func(c *ent.RoleCreate, i int) {
		c.SetID(allSystemRoles[i].Id).
			SetName(allSystemRoles[i].Name).
			SetIsSystem(allSystemRoles[i].IsSystem()).
			SetPermissions(allSystemRoles[i].Permissions).
			SetClientType(allSystemRoles[i].ClientType.Id)
	}).OnConflictColumns("id").UpdateNewValues().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to seed system roles: %v", err)
	}

	// Set the auto increment value for the roles table
	// in order to avoid conflicts with the system roles
	// frontend roles are assigned starting 1~100
	// backend roles are assigned starting 101~200
	// custom roles are assigned starting `roleAutoIncrementValue`
	_, err = client.ExecContext(ctx, fmt.Sprintf("ALTER SEQUENCE roles_id_seq RESTART WITH %d", roleAutoIncrementValue))
	if err != nil {
		log.Fatalf("failed to set auto increment value for roles table: %v", err)
	}
}

func NewEntDbFx() fx.Option {
	return fx.Module("ent",
		fx.Provide(NewDriver, NewClient, NewEntDb),
		fx.Invoke(AutoMigrate),
	)
}

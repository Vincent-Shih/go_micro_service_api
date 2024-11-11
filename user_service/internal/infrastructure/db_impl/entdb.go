package db_impl

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent"

	_ "github.com/lib/pq"
)

// txKey is a type used as a key for storing transaction in context
type txKey struct{}

// NewTxKey creates a new txKey instance
func NewTxKey() txKey {
	return txKey{}
}

// EntDB implements the db.Database interface using ent ORM
type EntDB struct {
	client *ent.Client
}

var _ db.Database = (*EntDB)(nil)

// NewEntDb creates and initializes a new EntDB instance
//
// It reads database configuration, establishes a connection to the database,
// sets up connection pool, and optionally performs auto migration.
//
// Returns:
//   - db.Database: An interface that can be used to interact with the database
//
// Panics if it fails to connect to the database or create schema resources (when auto-migrate is enabled)
func NewEntDb(client *ent.Client) db.Database {
	return &EntDB{client: client}
}

func (e *EntDB) GetConn(ctx context.Context) any {
	return e.client
}

func (e *EntDB) GetTx(ctx context.Context) any {
	return ctx.Value(txKey{})
}

func (e *EntDB) GetClient(ctx context.Context) any {
	if tx, ok := e.GetTx(ctx).(*ent.Tx); ok {
		return tx.Client()
	} else {
		return e.GetConn(ctx).(*ent.Client)
	}
}

func (e *EntDB) Begin(ctx context.Context) (context.Context, *cus_err.CusError) {
	// Check if ent client is initialized
	if e.client == nil {
		kgsErr := cus_err.New(cus_err.InternalServerError, "ent client not found", nil)
		cus_otel.Error(ctx, kgsErr.Error())
		return nil, kgsErr
	}

	tx, err := e.client.Tx(ctx)
	if err != nil {
		kgsErr := cus_err.New(cus_err.InternalServerError, "failed to start transaction", err)
		cus_otel.Error(ctx, kgsErr.Error())
		return nil, kgsErr
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

func (e *EntDB) Commit(ctx context.Context) (context.Context, *cus_err.CusError) {
	tx, ok := ctx.Value(txKey{}).(*ent.Tx)
	if !ok {
		kgsErr := cus_err.New(cus_err.InternalServerError, "transaction not found in context", nil)
		cus_otel.Error(ctx, kgsErr.Error())
		return ctx, kgsErr
	}

	if err := tx.Commit(); err != nil {
		kgsErr := cus_err.New(cus_err.InternalServerError, "failed to commit transaction", err)
		cus_otel.Error(ctx, kgsErr.Error())
		return ctx, kgsErr
	}

	return context.WithValue(ctx, txKey{}, nil), nil
}

func (e *EntDB) Rollback(ctx context.Context) (context.Context, *cus_err.CusError) {
	tx, ok := ctx.Value(txKey{}).(*ent.Tx)
	if !ok {
		kgsErr := cus_err.New(cus_err.InternalServerError, "transaction not found in context", nil)
		cus_otel.Error(ctx, kgsErr.Error())
		return ctx, kgsErr
	}

	if err := tx.Rollback(); err != nil {
		kgsErr := cus_err.New(cus_err.InternalServerError, "failed to rollback transaction", err)
		cus_otel.Error(ctx, kgsErr.Error())
		return ctx, kgsErr
	}

	return context.WithValue(ctx, txKey{}, nil), nil
}

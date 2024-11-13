package db_impl_test

import (
	"context"
	"log"
	"testing"

	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/user_service/internal/infrastructure/db_impl"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/profile"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMemoryDB create a new EntDB instance with an in-memory database
func newMemoryDB() *db_impl.EntDB {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&_fk=1")
	if err != nil {
		log.Fatalf("failed to open memory database: %v", err)
	}

	return db_impl.NewEntDb(client).(*db_impl.EntDB)
}

// TestEntDB_GetConn tests the GetConn method of EntDB
func TestEntDB_GetConn(t *testing.T) {
	ctx := context.Background()
	db := newMemoryDB()
	client := db.GetConn(ctx).(*ent.Client)
	defer client.Close()

	assert.NotNil(t, db.GetConn(ctx))

}

// TestEntDB_GetTx tests the GetTx method of EntDB
func TestEntDB_GetTx(t *testing.T) {
	ctx := context.Background()
	db := newMemoryDB()
	defer db.GetConn(ctx).(*ent.Client).Close()

	ctx = context.WithValue(context.Background(), db_impl.NewTxKey(), "test_tx")

	assert.Equal(t, "test_tx", db.GetTx(ctx))
}

// TestEntDB_GetTx_NoTx tests the GetTx method of EntDB when there is no transaction in the context
func TestEntDB_Begin(t *testing.T) {
	ctx := context.Background()
	db := newMemoryDB()
	client := db.GetConn(ctx).(*ent.Client)
	defer client.Close()

	ctx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to start transaction: %v", err)
	}

	tx := db.GetTx(ctx)
	assert.IsType(t, (*ent.Tx)(nil), tx)

}

// TestEntDB_Begin_Error tests the Begin method of EntDB when the client is not initialized
func TestEntDB_Begin_Error(t *testing.T) {
	ctx := context.Background()
	db := &db_impl.EntDB{} // Empty EntDB instance

	_, err := db.Begin(ctx) // Should not panic

	assert.Error(t, err)
	assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
}

// TestEntDB_Begin tests the Begin method of EntDB when the transaction fails to start
func TestEntDB_Commit(t *testing.T) {
	ctx := context.Background()
	db := newMemoryDB()
	client := db.GetConn(ctx).(*ent.Client)
	defer client.Close()

	// Auto migrate within the transaction
	err := client.Schema.Create(ctx)
	assert.Nil(t, err)

	// Start a transaction for the entire test
	ctx, cusErr := db.Begin(ctx)
	assert.Nil(t, cusErr)

	// Get transaction from context
	tx := db.GetTx(ctx).(*ent.Tx)

	// Create a frontend within the same transaction
	_, err = tx.Profile.Create().
		SetUserID(1).
		SetKey(enum.ProfileKey.Email.ID).
		SetValue("test@gmail.com").
		Save(ctx)
	assert.Nil(t, err)

	// Commit the transaction
	ctx, cusErr = db.Commit(ctx)
	assert.Nil(t, cusErr)

	// Check if the user was created
	row, err := client.Profile.Query().Where(profile.UserIDEQ(1), profile.KeyEQ(enum.ProfileKey.Email.ID)).Only(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, row)
	assert.Equal(t, "test@gmail.com", row.Value)
}

// TestEntDB_Commit_Error tests the Commit method of EntDB when the transaction is not found in the context
func TestEntDB_Commit_Error(t *testing.T) {
	ctx := context.Background()
	db := newMemoryDB()
	client := db.GetConn(ctx).(*ent.Client)
	defer client.Close()

	_, err := db.Commit(ctx)

	assert.Error(t, err)
	assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
}

// TestEntDB_Commit_Error tests the Commit method of EntDB when the transaction fails to commit
func TestEntDB_Rollback(t *testing.T) {
	ctx := context.Background()
	db := newMemoryDB()
	client := db.GetConn(ctx).(*ent.Client)
	defer client.Close()

	// Auto migrate
	migrateErr := client.Schema.Create(context.Background())
	require.Nil(t, migrateErr)

	//  Start transaction
	ctx, err := db.Begin(ctx)
	assert.Nil(t, err)

	// Get transaction from context
	tx := db.GetTx(ctx).(*ent.Tx)

	_, e := tx.Profile.Create().
		SetUserID(1).
		SetKey(enum.ProfileKey.Email.ID).
		SetValue("test@gmail.com").
		Save(ctx)
	assert.Nil(t, e)

	// Commit transaction
	ctx, err = db.Rollback(ctx)
	assert.Nil(t, err)

	// User should not be created
	_, _err := client.Profile.Query().Where(profile.UserIDEQ(1), profile.KeyEQ(enum.ProfileKey.Email.ID)).Only(ctx)
	assert.Error(t, _err)
}

// TestEntDB_Rollback_NoTx tests the Rollback method of EntDB when the transaction is not found in the context
func TestEntDB_Rollback_NoTx(t *testing.T) {
	db := &db_impl.EntDB{}
	ctx := context.Background()

	_, err := db.Rollback(ctx)

	assert.Error(t, err)
	assert.Equal(t, cus_err.InternalServerError, err.Code().Int())
}

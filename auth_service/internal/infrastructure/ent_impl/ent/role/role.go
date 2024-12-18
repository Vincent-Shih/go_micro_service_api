// Code generated by ent, DO NOT EDIT.

package role

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the role type in the database.
	Label = "role"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldPermissions holds the string denoting the permissions field in the database.
	FieldPermissions = "permissions"
	// FieldIsSystem holds the string denoting the is_system field in the database.
	FieldIsSystem = "is_system"
	// FieldClientType holds the string denoting the client_type field in the database.
	FieldClientType = "client_type"
	// EdgeUsers holds the string denoting the users edge name in mutations.
	EdgeUsers = "users"
	// EdgeAuthClients holds the string denoting the auth_clients edge name in mutations.
	EdgeAuthClients = "auth_clients"
	// Table holds the table name of the role in the database.
	Table = "roles"
	// UsersTable is the table that holds the users relation/edge.
	UsersTable = "users"
	// UsersInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UsersInverseTable = "users"
	// UsersColumn is the table column denoting the users relation/edge.
	UsersColumn = "user_roles"
	// AuthClientsTable is the table that holds the auth_clients relation/edge. The primary key declared below.
	AuthClientsTable = "auth_client_roles"
	// AuthClientsInverseTable is the table name for the AuthClient entity.
	// It exists in this package in order to avoid circular dependency with the "authclient" package.
	AuthClientsInverseTable = "auth_clients"
)

// Columns holds all SQL columns for role fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldName,
	FieldPermissions,
	FieldIsSystem,
	FieldClientType,
}

var (
	// AuthClientsPrimaryKey and AuthClientsColumn2 are the table columns denoting the
	// primary key for the auth_clients relation (M2M).
	AuthClientsPrimaryKey = []string{"auth_client_id", "role_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultIsSystem holds the default value on creation for the "is_system" field.
	DefaultIsSystem bool
)

// OrderOption defines the ordering options for the Role queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByIsSystem orders the results by the is_system field.
func ByIsSystem(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsSystem, opts...).ToFunc()
}

// ByClientType orders the results by the client_type field.
func ByClientType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldClientType, opts...).ToFunc()
}

// ByUsersCount orders the results by users count.
func ByUsersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newUsersStep(), opts...)
	}
}

// ByUsers orders the results by users terms.
func ByUsers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUsersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByAuthClientsCount orders the results by auth_clients count.
func ByAuthClientsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAuthClientsStep(), opts...)
	}
}

// ByAuthClients orders the results by auth_clients terms.
func ByAuthClients(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAuthClientsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newUsersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UsersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, UsersTable, UsersColumn),
	)
}
func newAuthClientsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AuthClientsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, AuthClientsTable, AuthClientsPrimaryKey...),
	)
}

// Code generated by ent, DO NOT EDIT.

package loginrecord

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the loginrecord type in the database.
	Label = "login_record"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldBrowser holds the string denoting the browser field in the database.
	FieldBrowser = "browser"
	// FieldBrowserVer holds the string denoting the browser_ver field in the database.
	FieldBrowserVer = "browser_ver"
	// FieldIP holds the string denoting the ip field in the database.
	FieldIP = "ip"
	// FieldOs holds the string denoting the os field in the database.
	FieldOs = "os"
	// FieldPlatform holds the string denoting the platform field in the database.
	FieldPlatform = "platform"
	// FieldCountry holds the string denoting the country field in the database.
	FieldCountry = "country"
	// FieldCountryCode holds the string denoting the country_code field in the database.
	FieldCountryCode = "country_code"
	// FieldCity holds the string denoting the city field in the database.
	FieldCity = "city"
	// FieldAsp holds the string denoting the asp field in the database.
	FieldAsp = "asp"
	// FieldIsMobile holds the string denoting the is_mobile field in the database.
	FieldIsMobile = "is_mobile"
	// FieldIsSuccess holds the string denoting the is_success field in the database.
	FieldIsSuccess = "is_success"
	// FieldErrMessage holds the string denoting the err_message field in the database.
	FieldErrMessage = "err_message"
	// EdgeUsers holds the string denoting the users edge name in mutations.
	EdgeUsers = "users"
	// Table holds the table name of the loginrecord in the database.
	Table = "login_records"
	// UsersTable is the table that holds the users relation/edge.
	UsersTable = "login_records"
	// UsersInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UsersInverseTable = "users"
	// UsersColumn is the table column denoting the users relation/edge.
	UsersColumn = "user_login_records"
)

// Columns holds all SQL columns for loginrecord fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldBrowser,
	FieldBrowserVer,
	FieldIP,
	FieldOs,
	FieldPlatform,
	FieldCountry,
	FieldCountryCode,
	FieldCity,
	FieldAsp,
	FieldIsMobile,
	FieldIsSuccess,
	FieldErrMessage,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "login_records"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"user_login_records",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
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
)

// OrderOption defines the ordering options for the LoginRecord queries.
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

// ByBrowser orders the results by the browser field.
func ByBrowser(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBrowser, opts...).ToFunc()
}

// ByBrowserVer orders the results by the browser_ver field.
func ByBrowserVer(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBrowserVer, opts...).ToFunc()
}

// ByIP orders the results by the ip field.
func ByIP(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIP, opts...).ToFunc()
}

// ByOs orders the results by the os field.
func ByOs(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOs, opts...).ToFunc()
}

// ByPlatform orders the results by the platform field.
func ByPlatform(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPlatform, opts...).ToFunc()
}

// ByCountry orders the results by the country field.
func ByCountry(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCountry, opts...).ToFunc()
}

// ByCountryCode orders the results by the country_code field.
func ByCountryCode(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCountryCode, opts...).ToFunc()
}

// ByCity orders the results by the city field.
func ByCity(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCity, opts...).ToFunc()
}

// ByAsp orders the results by the asp field.
func ByAsp(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAsp, opts...).ToFunc()
}

// ByIsMobile orders the results by the is_mobile field.
func ByIsMobile(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsMobile, opts...).ToFunc()
}

// ByIsSuccess orders the results by the is_success field.
func ByIsSuccess(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsSuccess, opts...).ToFunc()
}

// ByErrMessage orders the results by the err_message field.
func ByErrMessage(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldErrMessage, opts...).ToFunc()
}

// ByUsersField orders the results by users field.
func ByUsersField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUsersStep(), sql.OrderByField(field, opts...))
	}
}
func newUsersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UsersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, UsersTable, UsersColumn),
	)
}

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("account").Unique(),
		field.String("password"),
		field.Int("password_fail_times"),
		field.Int("status"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("auth_clients", AuthClient.Type).Ref("users").Unique(),
		edge.To("login_records", LoginRecord.Type),
		edge.To("roles", Role.Type).Unique(),
	}
}

package schema

import (
	"go_micro_service_api/pkg/enum"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Mixin of the Role.
func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name"),
		field.JSON("permissions", []enum.Permission{}),
		field.Bool("is_system").Default(false),
		field.Int("client_type"),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("roles"),
		edge.From("auth_clients", AuthClient.Type).Ref("roles"),
	}
}

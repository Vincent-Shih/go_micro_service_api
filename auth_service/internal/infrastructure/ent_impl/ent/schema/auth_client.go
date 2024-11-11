package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AuthClient holds the schema definition for the AuthClient entity.
type AuthClient struct {
	ent.Schema
}

// Mixin of the AuthClient.
func (AuthClient) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the AuthClient.
func (AuthClient) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int("client_type"),
		field.Int64("merchant_id"),
		field.String("secret"),
		field.Bool("active"),
		field.Int("token_expire_secs"),
		field.Int("login_failed_times"),
	}
}

// Edges of the AuthClient.
func (AuthClient) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("roles", Role.Type),
	}
}

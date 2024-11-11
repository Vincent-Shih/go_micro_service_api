package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// LoginRecord holds the schema definition for the LoginRecord entity.
type LoginRecord struct {
	ent.Schema
}

// Mixin of the LoginRecord.
func (LoginRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the LoginRecord.
func (LoginRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("browser"),
		field.String("browser_ver"),
		field.String("ip"),
		field.String("os"),
		field.String("platform"),
		field.String("country"),
		field.String("country_code"),
		field.String("city"),
		field.String("asp"),
		field.Bool("is_mobile"),
		field.Bool("is_success"),
		field.String("err_message"),
	}
}

// Edges of the LoginRecord.
func (LoginRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).Ref("login_records").Unique(),
	}
}

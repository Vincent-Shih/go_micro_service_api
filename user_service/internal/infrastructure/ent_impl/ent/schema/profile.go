package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Profile holds the schema definition for the Profile entity.
type Profile struct {
	ent.Schema
}

// Mixin of the Merchant.
func (Profile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the Profile.
func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("key"),
		field.String("value"),
		field.Bool("is_active").Default(true),
	}
}

// Edges of the Profile.
func (Profile) Edges() []ent.Edge {
	return nil
}

// Indexes of the Profile.
func (Profile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "key"),
		index.Fields("user_id", "key", "is_active"),
	}
}

func (Profile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// entsql.Annotation{Table: "profiles"},
		entsql.WithComments(true),
		schema.Comment("User profile, in the format of key-value pair"),
	}
}

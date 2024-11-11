package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type TimeMixin struct {
	mixin.Schema
}

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(func() time.Time {
				return time.Now().UTC()
			}),
		field.Time("updated_at").
			UpdateDefault(func() time.Time {
				return time.Now().UTC()
			}).
			Default(func() time.Time {
				return time.Now().UTC()
			}),
	}
}

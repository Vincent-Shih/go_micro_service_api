// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/predicate"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/profile"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// ProfileUpdate is the builder for updating Profile entities.
type ProfileUpdate struct {
	config
	hooks    []Hook
	mutation *ProfileMutation
}

// Where appends a list predicates to the ProfileUpdate builder.
func (pu *ProfileUpdate) Where(ps ...predicate.Profile) *ProfileUpdate {
	pu.mutation.Where(ps...)
	return pu
}

// SetUpdatedAt sets the "updated_at" field.
func (pu *ProfileUpdate) SetUpdatedAt(t time.Time) *ProfileUpdate {
	pu.mutation.SetUpdatedAt(t)
	return pu
}

// SetUserID sets the "user_id" field.
func (pu *ProfileUpdate) SetUserID(i int) *ProfileUpdate {
	pu.mutation.ResetUserID()
	pu.mutation.SetUserID(i)
	return pu
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (pu *ProfileUpdate) SetNillableUserID(i *int) *ProfileUpdate {
	if i != nil {
		pu.SetUserID(*i)
	}
	return pu
}

// AddUserID adds i to the "user_id" field.
func (pu *ProfileUpdate) AddUserID(i int) *ProfileUpdate {
	pu.mutation.AddUserID(i)
	return pu
}

// SetKey sets the "key" field.
func (pu *ProfileUpdate) SetKey(i int) *ProfileUpdate {
	pu.mutation.ResetKey()
	pu.mutation.SetKey(i)
	return pu
}

// SetNillableKey sets the "key" field if the given value is not nil.
func (pu *ProfileUpdate) SetNillableKey(i *int) *ProfileUpdate {
	if i != nil {
		pu.SetKey(*i)
	}
	return pu
}

// AddKey adds i to the "key" field.
func (pu *ProfileUpdate) AddKey(i int) *ProfileUpdate {
	pu.mutation.AddKey(i)
	return pu
}

// SetValue sets the "value" field.
func (pu *ProfileUpdate) SetValue(s string) *ProfileUpdate {
	pu.mutation.SetValue(s)
	return pu
}

// SetNillableValue sets the "value" field if the given value is not nil.
func (pu *ProfileUpdate) SetNillableValue(s *string) *ProfileUpdate {
	if s != nil {
		pu.SetValue(*s)
	}
	return pu
}

// SetIsActive sets the "is_active" field.
func (pu *ProfileUpdate) SetIsActive(b bool) *ProfileUpdate {
	pu.mutation.SetIsActive(b)
	return pu
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (pu *ProfileUpdate) SetNillableIsActive(b *bool) *ProfileUpdate {
	if b != nil {
		pu.SetIsActive(*b)
	}
	return pu
}

// Mutation returns the ProfileMutation object of the builder.
func (pu *ProfileUpdate) Mutation() *ProfileMutation {
	return pu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pu *ProfileUpdate) Save(ctx context.Context) (int, error) {
	pu.defaults()
	return withHooks(ctx, pu.sqlSave, pu.mutation, pu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pu *ProfileUpdate) SaveX(ctx context.Context) int {
	affected, err := pu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pu *ProfileUpdate) Exec(ctx context.Context) error {
	_, err := pu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pu *ProfileUpdate) ExecX(ctx context.Context) {
	if err := pu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pu *ProfileUpdate) defaults() {
	if _, ok := pu.mutation.UpdatedAt(); !ok {
		v := profile.UpdateDefaultUpdatedAt()
		pu.mutation.SetUpdatedAt(v)
	}
}

func (pu *ProfileUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(profile.Table, profile.Columns, sqlgraph.NewFieldSpec(profile.FieldID, field.TypeInt))
	if ps := pu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pu.mutation.UpdatedAt(); ok {
		_spec.SetField(profile.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := pu.mutation.UserID(); ok {
		_spec.SetField(profile.FieldUserID, field.TypeInt, value)
	}
	if value, ok := pu.mutation.AddedUserID(); ok {
		_spec.AddField(profile.FieldUserID, field.TypeInt, value)
	}
	if value, ok := pu.mutation.Key(); ok {
		_spec.SetField(profile.FieldKey, field.TypeInt, value)
	}
	if value, ok := pu.mutation.AddedKey(); ok {
		_spec.AddField(profile.FieldKey, field.TypeInt, value)
	}
	if value, ok := pu.mutation.Value(); ok {
		_spec.SetField(profile.FieldValue, field.TypeString, value)
	}
	if value, ok := pu.mutation.IsActive(); ok {
		_spec.SetField(profile.FieldIsActive, field.TypeBool, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, pu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{profile.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	pu.mutation.done = true
	return n, nil
}

// ProfileUpdateOne is the builder for updating a single Profile entity.
type ProfileUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ProfileMutation
}

// SetUpdatedAt sets the "updated_at" field.
func (puo *ProfileUpdateOne) SetUpdatedAt(t time.Time) *ProfileUpdateOne {
	puo.mutation.SetUpdatedAt(t)
	return puo
}

// SetUserID sets the "user_id" field.
func (puo *ProfileUpdateOne) SetUserID(i int) *ProfileUpdateOne {
	puo.mutation.ResetUserID()
	puo.mutation.SetUserID(i)
	return puo
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (puo *ProfileUpdateOne) SetNillableUserID(i *int) *ProfileUpdateOne {
	if i != nil {
		puo.SetUserID(*i)
	}
	return puo
}

// AddUserID adds i to the "user_id" field.
func (puo *ProfileUpdateOne) AddUserID(i int) *ProfileUpdateOne {
	puo.mutation.AddUserID(i)
	return puo
}

// SetKey sets the "key" field.
func (puo *ProfileUpdateOne) SetKey(i int) *ProfileUpdateOne {
	puo.mutation.ResetKey()
	puo.mutation.SetKey(i)
	return puo
}

// SetNillableKey sets the "key" field if the given value is not nil.
func (puo *ProfileUpdateOne) SetNillableKey(i *int) *ProfileUpdateOne {
	if i != nil {
		puo.SetKey(*i)
	}
	return puo
}

// AddKey adds i to the "key" field.
func (puo *ProfileUpdateOne) AddKey(i int) *ProfileUpdateOne {
	puo.mutation.AddKey(i)
	return puo
}

// SetValue sets the "value" field.
func (puo *ProfileUpdateOne) SetValue(s string) *ProfileUpdateOne {
	puo.mutation.SetValue(s)
	return puo
}

// SetNillableValue sets the "value" field if the given value is not nil.
func (puo *ProfileUpdateOne) SetNillableValue(s *string) *ProfileUpdateOne {
	if s != nil {
		puo.SetValue(*s)
	}
	return puo
}

// SetIsActive sets the "is_active" field.
func (puo *ProfileUpdateOne) SetIsActive(b bool) *ProfileUpdateOne {
	puo.mutation.SetIsActive(b)
	return puo
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (puo *ProfileUpdateOne) SetNillableIsActive(b *bool) *ProfileUpdateOne {
	if b != nil {
		puo.SetIsActive(*b)
	}
	return puo
}

// Mutation returns the ProfileMutation object of the builder.
func (puo *ProfileUpdateOne) Mutation() *ProfileMutation {
	return puo.mutation
}

// Where appends a list predicates to the ProfileUpdate builder.
func (puo *ProfileUpdateOne) Where(ps ...predicate.Profile) *ProfileUpdateOne {
	puo.mutation.Where(ps...)
	return puo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (puo *ProfileUpdateOne) Select(field string, fields ...string) *ProfileUpdateOne {
	puo.fields = append([]string{field}, fields...)
	return puo
}

// Save executes the query and returns the updated Profile entity.
func (puo *ProfileUpdateOne) Save(ctx context.Context) (*Profile, error) {
	puo.defaults()
	return withHooks(ctx, puo.sqlSave, puo.mutation, puo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (puo *ProfileUpdateOne) SaveX(ctx context.Context) *Profile {
	node, err := puo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (puo *ProfileUpdateOne) Exec(ctx context.Context) error {
	_, err := puo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (puo *ProfileUpdateOne) ExecX(ctx context.Context) {
	if err := puo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (puo *ProfileUpdateOne) defaults() {
	if _, ok := puo.mutation.UpdatedAt(); !ok {
		v := profile.UpdateDefaultUpdatedAt()
		puo.mutation.SetUpdatedAt(v)
	}
}

func (puo *ProfileUpdateOne) sqlSave(ctx context.Context) (_node *Profile, err error) {
	_spec := sqlgraph.NewUpdateSpec(profile.Table, profile.Columns, sqlgraph.NewFieldSpec(profile.FieldID, field.TypeInt))
	id, ok := puo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Profile.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := puo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, profile.FieldID)
		for _, f := range fields {
			if !profile.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != profile.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := puo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := puo.mutation.UpdatedAt(); ok {
		_spec.SetField(profile.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := puo.mutation.UserID(); ok {
		_spec.SetField(profile.FieldUserID, field.TypeInt, value)
	}
	if value, ok := puo.mutation.AddedUserID(); ok {
		_spec.AddField(profile.FieldUserID, field.TypeInt, value)
	}
	if value, ok := puo.mutation.Key(); ok {
		_spec.SetField(profile.FieldKey, field.TypeInt, value)
	}
	if value, ok := puo.mutation.AddedKey(); ok {
		_spec.AddField(profile.FieldKey, field.TypeInt, value)
	}
	if value, ok := puo.mutation.Value(); ok {
		_spec.SetField(profile.FieldValue, field.TypeString, value)
	}
	if value, ok := puo.mutation.IsActive(); ok {
		_spec.SetField(profile.FieldIsActive, field.TypeBool, value)
	}
	_node = &Profile{config: puo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, puo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{profile.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	puo.mutation.done = true
	return _node, nil
}

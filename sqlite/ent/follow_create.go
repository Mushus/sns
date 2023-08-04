// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/Mushus/activitypub/sqlite/ent/follow"
)

// FollowCreate is the builder for creating a Follow entity.
type FollowCreate struct {
	config
	mutation *FollowMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetFromID sets the "fromID" field.
func (fc *FollowCreate) SetFromID(s string) *FollowCreate {
	fc.mutation.SetFromID(s)
	return fc
}

// SetToID sets the "toID" field.
func (fc *FollowCreate) SetToID(s string) *FollowCreate {
	fc.mutation.SetToID(s)
	return fc
}

// SetID sets the "id" field.
func (fc *FollowCreate) SetID(s string) *FollowCreate {
	fc.mutation.SetID(s)
	return fc
}

// Mutation returns the FollowMutation object of the builder.
func (fc *FollowCreate) Mutation() *FollowMutation {
	return fc.mutation
}

// Save creates the Follow in the database.
func (fc *FollowCreate) Save(ctx context.Context) (*Follow, error) {
	return withHooks(ctx, fc.sqlSave, fc.mutation, fc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (fc *FollowCreate) SaveX(ctx context.Context) *Follow {
	v, err := fc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fc *FollowCreate) Exec(ctx context.Context) error {
	_, err := fc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fc *FollowCreate) ExecX(ctx context.Context) {
	if err := fc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fc *FollowCreate) check() error {
	if _, ok := fc.mutation.FromID(); !ok {
		return &ValidationError{Name: "fromID", err: errors.New(`ent: missing required field "Follow.fromID"`)}
	}
	if _, ok := fc.mutation.ToID(); !ok {
		return &ValidationError{Name: "toID", err: errors.New(`ent: missing required field "Follow.toID"`)}
	}
	return nil
}

func (fc *FollowCreate) sqlSave(ctx context.Context) (*Follow, error) {
	if err := fc.check(); err != nil {
		return nil, err
	}
	_node, _spec := fc.createSpec()
	if err := sqlgraph.CreateNode(ctx, fc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Follow.ID type: %T", _spec.ID.Value)
		}
	}
	fc.mutation.id = &_node.ID
	fc.mutation.done = true
	return _node, nil
}

func (fc *FollowCreate) createSpec() (*Follow, *sqlgraph.CreateSpec) {
	var (
		_node = &Follow{config: fc.config}
		_spec = sqlgraph.NewCreateSpec(follow.Table, sqlgraph.NewFieldSpec(follow.FieldID, field.TypeString))
	)
	_spec.OnConflict = fc.conflict
	if id, ok := fc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := fc.mutation.FromID(); ok {
		_spec.SetField(follow.FieldFromID, field.TypeString, value)
		_node.FromID = value
	}
	if value, ok := fc.mutation.ToID(); ok {
		_spec.SetField(follow.FieldToID, field.TypeString, value)
		_node.ToID = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Follow.Create().
//		SetFromID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.FollowUpsert) {
//			SetFromID(v+v).
//		}).
//		Exec(ctx)
func (fc *FollowCreate) OnConflict(opts ...sql.ConflictOption) *FollowUpsertOne {
	fc.conflict = opts
	return &FollowUpsertOne{
		create: fc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Follow.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (fc *FollowCreate) OnConflictColumns(columns ...string) *FollowUpsertOne {
	fc.conflict = append(fc.conflict, sql.ConflictColumns(columns...))
	return &FollowUpsertOne{
		create: fc,
	}
}

type (
	// FollowUpsertOne is the builder for "upsert"-ing
	//  one Follow node.
	FollowUpsertOne struct {
		create *FollowCreate
	}

	// FollowUpsert is the "OnConflict" setter.
	FollowUpsert struct {
		*sql.UpdateSet
	}
)

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Follow.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(follow.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *FollowUpsertOne) UpdateNewValues() *FollowUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(follow.FieldID)
		}
		if _, exists := u.create.mutation.FromID(); exists {
			s.SetIgnore(follow.FieldFromID)
		}
		if _, exists := u.create.mutation.ToID(); exists {
			s.SetIgnore(follow.FieldToID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Follow.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *FollowUpsertOne) Ignore() *FollowUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *FollowUpsertOne) DoNothing() *FollowUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the FollowCreate.OnConflict
// documentation for more info.
func (u *FollowUpsertOne) Update(set func(*FollowUpsert)) *FollowUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&FollowUpsert{UpdateSet: update})
	}))
	return u
}

// Exec executes the query.
func (u *FollowUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for FollowCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *FollowUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *FollowUpsertOne) ID(ctx context.Context) (id string, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: FollowUpsertOne.ID is not supported by MySQL driver. Use FollowUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *FollowUpsertOne) IDX(ctx context.Context) string {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// FollowCreateBulk is the builder for creating many Follow entities in bulk.
type FollowCreateBulk struct {
	config
	builders []*FollowCreate
	conflict []sql.ConflictOption
}

// Save creates the Follow entities in the database.
func (fcb *FollowCreateBulk) Save(ctx context.Context) ([]*Follow, error) {
	specs := make([]*sqlgraph.CreateSpec, len(fcb.builders))
	nodes := make([]*Follow, len(fcb.builders))
	mutators := make([]Mutator, len(fcb.builders))
	for i := range fcb.builders {
		func(i int, root context.Context) {
			builder := fcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*FollowMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, fcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = fcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, fcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, fcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (fcb *FollowCreateBulk) SaveX(ctx context.Context) []*Follow {
	v, err := fcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fcb *FollowCreateBulk) Exec(ctx context.Context) error {
	_, err := fcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fcb *FollowCreateBulk) ExecX(ctx context.Context) {
	if err := fcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Follow.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.FollowUpsert) {
//			SetFromID(v+v).
//		}).
//		Exec(ctx)
func (fcb *FollowCreateBulk) OnConflict(opts ...sql.ConflictOption) *FollowUpsertBulk {
	fcb.conflict = opts
	return &FollowUpsertBulk{
		create: fcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Follow.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (fcb *FollowCreateBulk) OnConflictColumns(columns ...string) *FollowUpsertBulk {
	fcb.conflict = append(fcb.conflict, sql.ConflictColumns(columns...))
	return &FollowUpsertBulk{
		create: fcb,
	}
}

// FollowUpsertBulk is the builder for "upsert"-ing
// a bulk of Follow nodes.
type FollowUpsertBulk struct {
	create *FollowCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Follow.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(follow.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *FollowUpsertBulk) UpdateNewValues() *FollowUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(follow.FieldID)
			}
			if _, exists := b.mutation.FromID(); exists {
				s.SetIgnore(follow.FieldFromID)
			}
			if _, exists := b.mutation.ToID(); exists {
				s.SetIgnore(follow.FieldToID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Follow.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *FollowUpsertBulk) Ignore() *FollowUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *FollowUpsertBulk) DoNothing() *FollowUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the FollowCreateBulk.OnConflict
// documentation for more info.
func (u *FollowUpsertBulk) Update(set func(*FollowUpsert)) *FollowUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&FollowUpsert{UpdateSet: update})
	}))
	return u
}

// Exec executes the query.
func (u *FollowUpsertBulk) Exec(ctx context.Context) error {
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the FollowCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for FollowCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *FollowUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

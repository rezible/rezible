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
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/subscription"
	"github.com/twohundreds/rezible/ent/team"
	"github.com/twohundreds/rezible/ent/user"
)

// SubscriptionCreate is the builder for creating a Subscription entity.
type SubscriptionCreate struct {
	config
	mutation *SubscriptionMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetEvent sets the "event" field.
func (sc *SubscriptionCreate) SetEvent(s string) *SubscriptionCreate {
	sc.mutation.SetEvent(s)
	return sc
}

// SetActive sets the "active" field.
func (sc *SubscriptionCreate) SetActive(b bool) *SubscriptionCreate {
	sc.mutation.SetActive(b)
	return sc
}

// SetNillableActive sets the "active" field if the given value is not nil.
func (sc *SubscriptionCreate) SetNillableActive(b *bool) *SubscriptionCreate {
	if b != nil {
		sc.SetActive(*b)
	}
	return sc
}

// SetID sets the "id" field.
func (sc *SubscriptionCreate) SetID(u uuid.UUID) *SubscriptionCreate {
	sc.mutation.SetID(u)
	return sc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (sc *SubscriptionCreate) SetNillableID(u *uuid.UUID) *SubscriptionCreate {
	if u != nil {
		sc.SetID(*u)
	}
	return sc
}

// SetUserID sets the "user" edge to the User entity by ID.
func (sc *SubscriptionCreate) SetUserID(id uuid.UUID) *SubscriptionCreate {
	sc.mutation.SetUserID(id)
	return sc
}

// SetNillableUserID sets the "user" edge to the User entity by ID if the given value is not nil.
func (sc *SubscriptionCreate) SetNillableUserID(id *uuid.UUID) *SubscriptionCreate {
	if id != nil {
		sc = sc.SetUserID(*id)
	}
	return sc
}

// SetUser sets the "user" edge to the User entity.
func (sc *SubscriptionCreate) SetUser(u *User) *SubscriptionCreate {
	return sc.SetUserID(u.ID)
}

// SetTeamID sets the "team" edge to the Team entity by ID.
func (sc *SubscriptionCreate) SetTeamID(id uuid.UUID) *SubscriptionCreate {
	sc.mutation.SetTeamID(id)
	return sc
}

// SetNillableTeamID sets the "team" edge to the Team entity by ID if the given value is not nil.
func (sc *SubscriptionCreate) SetNillableTeamID(id *uuid.UUID) *SubscriptionCreate {
	if id != nil {
		sc = sc.SetTeamID(*id)
	}
	return sc
}

// SetTeam sets the "team" edge to the Team entity.
func (sc *SubscriptionCreate) SetTeam(t *Team) *SubscriptionCreate {
	return sc.SetTeamID(t.ID)
}

// SetIncidentID sets the "incident" edge to the Incident entity by ID.
func (sc *SubscriptionCreate) SetIncidentID(id uuid.UUID) *SubscriptionCreate {
	sc.mutation.SetIncidentID(id)
	return sc
}

// SetNillableIncidentID sets the "incident" edge to the Incident entity by ID if the given value is not nil.
func (sc *SubscriptionCreate) SetNillableIncidentID(id *uuid.UUID) *SubscriptionCreate {
	if id != nil {
		sc = sc.SetIncidentID(*id)
	}
	return sc
}

// SetIncident sets the "incident" edge to the Incident entity.
func (sc *SubscriptionCreate) SetIncident(i *Incident) *SubscriptionCreate {
	return sc.SetIncidentID(i.ID)
}

// Mutation returns the SubscriptionMutation object of the builder.
func (sc *SubscriptionCreate) Mutation() *SubscriptionMutation {
	return sc.mutation
}

// Save creates the Subscription in the database.
func (sc *SubscriptionCreate) Save(ctx context.Context) (*Subscription, error) {
	sc.defaults()
	return withHooks(ctx, sc.sqlSave, sc.mutation, sc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sc *SubscriptionCreate) SaveX(ctx context.Context) *Subscription {
	v, err := sc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sc *SubscriptionCreate) Exec(ctx context.Context) error {
	_, err := sc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sc *SubscriptionCreate) ExecX(ctx context.Context) {
	if err := sc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sc *SubscriptionCreate) defaults() {
	if _, ok := sc.mutation.Active(); !ok {
		v := subscription.DefaultActive
		sc.mutation.SetActive(v)
	}
	if _, ok := sc.mutation.ID(); !ok {
		v := subscription.DefaultID()
		sc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sc *SubscriptionCreate) check() error {
	if _, ok := sc.mutation.Event(); !ok {
		return &ValidationError{Name: "event", err: errors.New(`ent: missing required field "Subscription.event"`)}
	}
	if _, ok := sc.mutation.Active(); !ok {
		return &ValidationError{Name: "active", err: errors.New(`ent: missing required field "Subscription.active"`)}
	}
	return nil
}

func (sc *SubscriptionCreate) sqlSave(ctx context.Context) (*Subscription, error) {
	if err := sc.check(); err != nil {
		return nil, err
	}
	_node, _spec := sc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	sc.mutation.id = &_node.ID
	sc.mutation.done = true
	return _node, nil
}

func (sc *SubscriptionCreate) createSpec() (*Subscription, *sqlgraph.CreateSpec) {
	var (
		_node = &Subscription{config: sc.config}
		_spec = sqlgraph.NewCreateSpec(subscription.Table, sqlgraph.NewFieldSpec(subscription.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = sc.conflict
	if id, ok := sc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := sc.mutation.Event(); ok {
		_spec.SetField(subscription.FieldEvent, field.TypeString, value)
		_node.Event = value
	}
	if value, ok := sc.mutation.Active(); ok {
		_spec.SetField(subscription.FieldActive, field.TypeBool, value)
		_node.Active = value
	}
	if nodes := sc.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   subscription.UserTable,
			Columns: []string{subscription.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.subscription_user = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := sc.mutation.TeamIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   subscription.TeamTable,
			Columns: []string{subscription.TeamColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.subscription_team = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := sc.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   subscription.IncidentTable,
			Columns: []string{subscription.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.subscription_incident = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Subscription.Create().
//		SetEvent(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SubscriptionUpsert) {
//			SetEvent(v+v).
//		}).
//		Exec(ctx)
func (sc *SubscriptionCreate) OnConflict(opts ...sql.ConflictOption) *SubscriptionUpsertOne {
	sc.conflict = opts
	return &SubscriptionUpsertOne{
		create: sc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Subscription.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (sc *SubscriptionCreate) OnConflictColumns(columns ...string) *SubscriptionUpsertOne {
	sc.conflict = append(sc.conflict, sql.ConflictColumns(columns...))
	return &SubscriptionUpsertOne{
		create: sc,
	}
}

type (
	// SubscriptionUpsertOne is the builder for "upsert"-ing
	//  one Subscription node.
	SubscriptionUpsertOne struct {
		create *SubscriptionCreate
	}

	// SubscriptionUpsert is the "OnConflict" setter.
	SubscriptionUpsert struct {
		*sql.UpdateSet
	}
)

// SetEvent sets the "event" field.
func (u *SubscriptionUpsert) SetEvent(v string) *SubscriptionUpsert {
	u.Set(subscription.FieldEvent, v)
	return u
}

// UpdateEvent sets the "event" field to the value that was provided on create.
func (u *SubscriptionUpsert) UpdateEvent() *SubscriptionUpsert {
	u.SetExcluded(subscription.FieldEvent)
	return u
}

// SetActive sets the "active" field.
func (u *SubscriptionUpsert) SetActive(v bool) *SubscriptionUpsert {
	u.Set(subscription.FieldActive, v)
	return u
}

// UpdateActive sets the "active" field to the value that was provided on create.
func (u *SubscriptionUpsert) UpdateActive() *SubscriptionUpsert {
	u.SetExcluded(subscription.FieldActive)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Subscription.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(subscription.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SubscriptionUpsertOne) UpdateNewValues() *SubscriptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(subscription.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Subscription.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *SubscriptionUpsertOne) Ignore() *SubscriptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SubscriptionUpsertOne) DoNothing() *SubscriptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SubscriptionCreate.OnConflict
// documentation for more info.
func (u *SubscriptionUpsertOne) Update(set func(*SubscriptionUpsert)) *SubscriptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SubscriptionUpsert{UpdateSet: update})
	}))
	return u
}

// SetEvent sets the "event" field.
func (u *SubscriptionUpsertOne) SetEvent(v string) *SubscriptionUpsertOne {
	return u.Update(func(s *SubscriptionUpsert) {
		s.SetEvent(v)
	})
}

// UpdateEvent sets the "event" field to the value that was provided on create.
func (u *SubscriptionUpsertOne) UpdateEvent() *SubscriptionUpsertOne {
	return u.Update(func(s *SubscriptionUpsert) {
		s.UpdateEvent()
	})
}

// SetActive sets the "active" field.
func (u *SubscriptionUpsertOne) SetActive(v bool) *SubscriptionUpsertOne {
	return u.Update(func(s *SubscriptionUpsert) {
		s.SetActive(v)
	})
}

// UpdateActive sets the "active" field to the value that was provided on create.
func (u *SubscriptionUpsertOne) UpdateActive() *SubscriptionUpsertOne {
	return u.Update(func(s *SubscriptionUpsert) {
		s.UpdateActive()
	})
}

// Exec executes the query.
func (u *SubscriptionUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SubscriptionCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SubscriptionUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *SubscriptionUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: SubscriptionUpsertOne.ID is not supported by MySQL driver. Use SubscriptionUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *SubscriptionUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// SubscriptionCreateBulk is the builder for creating many Subscription entities in bulk.
type SubscriptionCreateBulk struct {
	config
	err      error
	builders []*SubscriptionCreate
	conflict []sql.ConflictOption
}

// Save creates the Subscription entities in the database.
func (scb *SubscriptionCreateBulk) Save(ctx context.Context) ([]*Subscription, error) {
	if scb.err != nil {
		return nil, scb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scb.builders))
	nodes := make([]*Subscription, len(scb.builders))
	mutators := make([]Mutator, len(scb.builders))
	for i := range scb.builders {
		func(i int, root context.Context) {
			builder := scb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SubscriptionMutation)
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
					_, err = mutators[i+1].Mutate(root, scb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = scb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, scb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scb *SubscriptionCreateBulk) SaveX(ctx context.Context) []*Subscription {
	v, err := scb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scb *SubscriptionCreateBulk) Exec(ctx context.Context) error {
	_, err := scb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scb *SubscriptionCreateBulk) ExecX(ctx context.Context) {
	if err := scb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Subscription.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SubscriptionUpsert) {
//			SetEvent(v+v).
//		}).
//		Exec(ctx)
func (scb *SubscriptionCreateBulk) OnConflict(opts ...sql.ConflictOption) *SubscriptionUpsertBulk {
	scb.conflict = opts
	return &SubscriptionUpsertBulk{
		create: scb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Subscription.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scb *SubscriptionCreateBulk) OnConflictColumns(columns ...string) *SubscriptionUpsertBulk {
	scb.conflict = append(scb.conflict, sql.ConflictColumns(columns...))
	return &SubscriptionUpsertBulk{
		create: scb,
	}
}

// SubscriptionUpsertBulk is the builder for "upsert"-ing
// a bulk of Subscription nodes.
type SubscriptionUpsertBulk struct {
	create *SubscriptionCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Subscription.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(subscription.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SubscriptionUpsertBulk) UpdateNewValues() *SubscriptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(subscription.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Subscription.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *SubscriptionUpsertBulk) Ignore() *SubscriptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SubscriptionUpsertBulk) DoNothing() *SubscriptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SubscriptionCreateBulk.OnConflict
// documentation for more info.
func (u *SubscriptionUpsertBulk) Update(set func(*SubscriptionUpsert)) *SubscriptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SubscriptionUpsert{UpdateSet: update})
	}))
	return u
}

// SetEvent sets the "event" field.
func (u *SubscriptionUpsertBulk) SetEvent(v string) *SubscriptionUpsertBulk {
	return u.Update(func(s *SubscriptionUpsert) {
		s.SetEvent(v)
	})
}

// UpdateEvent sets the "event" field to the value that was provided on create.
func (u *SubscriptionUpsertBulk) UpdateEvent() *SubscriptionUpsertBulk {
	return u.Update(func(s *SubscriptionUpsert) {
		s.UpdateEvent()
	})
}

// SetActive sets the "active" field.
func (u *SubscriptionUpsertBulk) SetActive(v bool) *SubscriptionUpsertBulk {
	return u.Update(func(s *SubscriptionUpsert) {
		s.SetActive(v)
	})
}

// UpdateActive sets the "active" field to the value that was provided on create.
func (u *SubscriptionUpsertBulk) UpdateActive() *SubscriptionUpsertBulk {
	return u.Update(func(s *SubscriptionUpsert) {
		s.UpdateActive()
	})
}

// Exec executes the query.
func (u *SubscriptionUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the SubscriptionCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SubscriptionCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SubscriptionUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

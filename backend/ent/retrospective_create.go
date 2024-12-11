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
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivediscussion"
)

// RetrospectiveCreate is the builder for creating a Retrospective entity.
type RetrospectiveCreate struct {
	config
	mutation *RetrospectiveMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetDocumentName sets the "document_name" field.
func (rc *RetrospectiveCreate) SetDocumentName(s string) *RetrospectiveCreate {
	rc.mutation.SetDocumentName(s)
	return rc
}

// SetState sets the "state" field.
func (rc *RetrospectiveCreate) SetState(r retrospective.State) *RetrospectiveCreate {
	rc.mutation.SetState(r)
	return rc
}

// SetID sets the "id" field.
func (rc *RetrospectiveCreate) SetID(u uuid.UUID) *RetrospectiveCreate {
	rc.mutation.SetID(u)
	return rc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (rc *RetrospectiveCreate) SetNillableID(u *uuid.UUID) *RetrospectiveCreate {
	if u != nil {
		rc.SetID(*u)
	}
	return rc
}

// SetIncidentID sets the "incident" edge to the Incident entity by ID.
func (rc *RetrospectiveCreate) SetIncidentID(id uuid.UUID) *RetrospectiveCreate {
	rc.mutation.SetIncidentID(id)
	return rc
}

// SetNillableIncidentID sets the "incident" edge to the Incident entity by ID if the given value is not nil.
func (rc *RetrospectiveCreate) SetNillableIncidentID(id *uuid.UUID) *RetrospectiveCreate {
	if id != nil {
		rc = rc.SetIncidentID(*id)
	}
	return rc
}

// SetIncident sets the "incident" edge to the Incident entity.
func (rc *RetrospectiveCreate) SetIncident(i *Incident) *RetrospectiveCreate {
	return rc.SetIncidentID(i.ID)
}

// AddDiscussionIDs adds the "discussions" edge to the RetrospectiveDiscussion entity by IDs.
func (rc *RetrospectiveCreate) AddDiscussionIDs(ids ...uuid.UUID) *RetrospectiveCreate {
	rc.mutation.AddDiscussionIDs(ids...)
	return rc
}

// AddDiscussions adds the "discussions" edges to the RetrospectiveDiscussion entity.
func (rc *RetrospectiveCreate) AddDiscussions(r ...*RetrospectiveDiscussion) *RetrospectiveCreate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return rc.AddDiscussionIDs(ids...)
}

// Mutation returns the RetrospectiveMutation object of the builder.
func (rc *RetrospectiveCreate) Mutation() *RetrospectiveMutation {
	return rc.mutation
}

// Save creates the Retrospective in the database.
func (rc *RetrospectiveCreate) Save(ctx context.Context) (*Retrospective, error) {
	rc.defaults()
	return withHooks(ctx, rc.sqlSave, rc.mutation, rc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (rc *RetrospectiveCreate) SaveX(ctx context.Context) *Retrospective {
	v, err := rc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rc *RetrospectiveCreate) Exec(ctx context.Context) error {
	_, err := rc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rc *RetrospectiveCreate) ExecX(ctx context.Context) {
	if err := rc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (rc *RetrospectiveCreate) defaults() {
	if _, ok := rc.mutation.ID(); !ok {
		v := retrospective.DefaultID()
		rc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (rc *RetrospectiveCreate) check() error {
	if _, ok := rc.mutation.DocumentName(); !ok {
		return &ValidationError{Name: "document_name", err: errors.New(`ent: missing required field "Retrospective.document_name"`)}
	}
	if _, ok := rc.mutation.State(); !ok {
		return &ValidationError{Name: "state", err: errors.New(`ent: missing required field "Retrospective.state"`)}
	}
	if v, ok := rc.mutation.State(); ok {
		if err := retrospective.StateValidator(v); err != nil {
			return &ValidationError{Name: "state", err: fmt.Errorf(`ent: validator failed for field "Retrospective.state": %w`, err)}
		}
	}
	return nil
}

func (rc *RetrospectiveCreate) sqlSave(ctx context.Context) (*Retrospective, error) {
	if err := rc.check(); err != nil {
		return nil, err
	}
	_node, _spec := rc.createSpec()
	if err := sqlgraph.CreateNode(ctx, rc.driver, _spec); err != nil {
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
	rc.mutation.id = &_node.ID
	rc.mutation.done = true
	return _node, nil
}

func (rc *RetrospectiveCreate) createSpec() (*Retrospective, *sqlgraph.CreateSpec) {
	var (
		_node = &Retrospective{config: rc.config}
		_spec = sqlgraph.NewCreateSpec(retrospective.Table, sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = rc.conflict
	if id, ok := rc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := rc.mutation.DocumentName(); ok {
		_spec.SetField(retrospective.FieldDocumentName, field.TypeString, value)
		_node.DocumentName = value
	}
	if value, ok := rc.mutation.State(); ok {
		_spec.SetField(retrospective.FieldState, field.TypeEnum, value)
		_node.State = value
	}
	if nodes := rc.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   retrospective.IncidentTable,
			Columns: []string{retrospective.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.incident_retrospective = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := rc.mutation.DiscussionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   retrospective.DiscussionsTable,
			Columns: []string{retrospective.DiscussionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussion.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Retrospective.Create().
//		SetDocumentName(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.RetrospectiveUpsert) {
//			SetDocumentName(v+v).
//		}).
//		Exec(ctx)
func (rc *RetrospectiveCreate) OnConflict(opts ...sql.ConflictOption) *RetrospectiveUpsertOne {
	rc.conflict = opts
	return &RetrospectiveUpsertOne{
		create: rc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Retrospective.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (rc *RetrospectiveCreate) OnConflictColumns(columns ...string) *RetrospectiveUpsertOne {
	rc.conflict = append(rc.conflict, sql.ConflictColumns(columns...))
	return &RetrospectiveUpsertOne{
		create: rc,
	}
}

type (
	// RetrospectiveUpsertOne is the builder for "upsert"-ing
	//  one Retrospective node.
	RetrospectiveUpsertOne struct {
		create *RetrospectiveCreate
	}

	// RetrospectiveUpsert is the "OnConflict" setter.
	RetrospectiveUpsert struct {
		*sql.UpdateSet
	}
)

// SetDocumentName sets the "document_name" field.
func (u *RetrospectiveUpsert) SetDocumentName(v string) *RetrospectiveUpsert {
	u.Set(retrospective.FieldDocumentName, v)
	return u
}

// UpdateDocumentName sets the "document_name" field to the value that was provided on create.
func (u *RetrospectiveUpsert) UpdateDocumentName() *RetrospectiveUpsert {
	u.SetExcluded(retrospective.FieldDocumentName)
	return u
}

// SetState sets the "state" field.
func (u *RetrospectiveUpsert) SetState(v retrospective.State) *RetrospectiveUpsert {
	u.Set(retrospective.FieldState, v)
	return u
}

// UpdateState sets the "state" field to the value that was provided on create.
func (u *RetrospectiveUpsert) UpdateState() *RetrospectiveUpsert {
	u.SetExcluded(retrospective.FieldState)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Retrospective.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(retrospective.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *RetrospectiveUpsertOne) UpdateNewValues() *RetrospectiveUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(retrospective.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Retrospective.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *RetrospectiveUpsertOne) Ignore() *RetrospectiveUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *RetrospectiveUpsertOne) DoNothing() *RetrospectiveUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the RetrospectiveCreate.OnConflict
// documentation for more info.
func (u *RetrospectiveUpsertOne) Update(set func(*RetrospectiveUpsert)) *RetrospectiveUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&RetrospectiveUpsert{UpdateSet: update})
	}))
	return u
}

// SetDocumentName sets the "document_name" field.
func (u *RetrospectiveUpsertOne) SetDocumentName(v string) *RetrospectiveUpsertOne {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.SetDocumentName(v)
	})
}

// UpdateDocumentName sets the "document_name" field to the value that was provided on create.
func (u *RetrospectiveUpsertOne) UpdateDocumentName() *RetrospectiveUpsertOne {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.UpdateDocumentName()
	})
}

// SetState sets the "state" field.
func (u *RetrospectiveUpsertOne) SetState(v retrospective.State) *RetrospectiveUpsertOne {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.SetState(v)
	})
}

// UpdateState sets the "state" field to the value that was provided on create.
func (u *RetrospectiveUpsertOne) UpdateState() *RetrospectiveUpsertOne {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.UpdateState()
	})
}

// Exec executes the query.
func (u *RetrospectiveUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for RetrospectiveCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *RetrospectiveUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *RetrospectiveUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: RetrospectiveUpsertOne.ID is not supported by MySQL driver. Use RetrospectiveUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *RetrospectiveUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// RetrospectiveCreateBulk is the builder for creating many Retrospective entities in bulk.
type RetrospectiveCreateBulk struct {
	config
	err      error
	builders []*RetrospectiveCreate
	conflict []sql.ConflictOption
}

// Save creates the Retrospective entities in the database.
func (rcb *RetrospectiveCreateBulk) Save(ctx context.Context) ([]*Retrospective, error) {
	if rcb.err != nil {
		return nil, rcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(rcb.builders))
	nodes := make([]*Retrospective, len(rcb.builders))
	mutators := make([]Mutator, len(rcb.builders))
	for i := range rcb.builders {
		func(i int, root context.Context) {
			builder := rcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*RetrospectiveMutation)
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
					_, err = mutators[i+1].Mutate(root, rcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = rcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, rcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, rcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (rcb *RetrospectiveCreateBulk) SaveX(ctx context.Context) []*Retrospective {
	v, err := rcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rcb *RetrospectiveCreateBulk) Exec(ctx context.Context) error {
	_, err := rcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rcb *RetrospectiveCreateBulk) ExecX(ctx context.Context) {
	if err := rcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Retrospective.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.RetrospectiveUpsert) {
//			SetDocumentName(v+v).
//		}).
//		Exec(ctx)
func (rcb *RetrospectiveCreateBulk) OnConflict(opts ...sql.ConflictOption) *RetrospectiveUpsertBulk {
	rcb.conflict = opts
	return &RetrospectiveUpsertBulk{
		create: rcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Retrospective.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (rcb *RetrospectiveCreateBulk) OnConflictColumns(columns ...string) *RetrospectiveUpsertBulk {
	rcb.conflict = append(rcb.conflict, sql.ConflictColumns(columns...))
	return &RetrospectiveUpsertBulk{
		create: rcb,
	}
}

// RetrospectiveUpsertBulk is the builder for "upsert"-ing
// a bulk of Retrospective nodes.
type RetrospectiveUpsertBulk struct {
	create *RetrospectiveCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Retrospective.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(retrospective.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *RetrospectiveUpsertBulk) UpdateNewValues() *RetrospectiveUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(retrospective.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Retrospective.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *RetrospectiveUpsertBulk) Ignore() *RetrospectiveUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *RetrospectiveUpsertBulk) DoNothing() *RetrospectiveUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the RetrospectiveCreateBulk.OnConflict
// documentation for more info.
func (u *RetrospectiveUpsertBulk) Update(set func(*RetrospectiveUpsert)) *RetrospectiveUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&RetrospectiveUpsert{UpdateSet: update})
	}))
	return u
}

// SetDocumentName sets the "document_name" field.
func (u *RetrospectiveUpsertBulk) SetDocumentName(v string) *RetrospectiveUpsertBulk {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.SetDocumentName(v)
	})
}

// UpdateDocumentName sets the "document_name" field to the value that was provided on create.
func (u *RetrospectiveUpsertBulk) UpdateDocumentName() *RetrospectiveUpsertBulk {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.UpdateDocumentName()
	})
}

// SetState sets the "state" field.
func (u *RetrospectiveUpsertBulk) SetState(v retrospective.State) *RetrospectiveUpsertBulk {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.SetState(v)
	})
}

// UpdateState sets the "state" field to the value that was provided on create.
func (u *RetrospectiveUpsertBulk) UpdateState() *RetrospectiveUpsertBulk {
	return u.Update(func(s *RetrospectiveUpsert) {
		s.UpdateState()
	})
}

// Exec executes the query.
func (u *RetrospectiveUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the RetrospectiveCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for RetrospectiveCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *RetrospectiveUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

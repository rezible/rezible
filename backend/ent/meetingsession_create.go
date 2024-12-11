// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/meetingsession"
)

// MeetingSessionCreate is the builder for creating a MeetingSession entity.
type MeetingSessionCreate struct {
	config
	mutation *MeetingSessionMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetTitle sets the "title" field.
func (msc *MeetingSessionCreate) SetTitle(s string) *MeetingSessionCreate {
	msc.mutation.SetTitle(s)
	return msc
}

// SetStartedAt sets the "started_at" field.
func (msc *MeetingSessionCreate) SetStartedAt(t time.Time) *MeetingSessionCreate {
	msc.mutation.SetStartedAt(t)
	return msc
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (msc *MeetingSessionCreate) SetNillableStartedAt(t *time.Time) *MeetingSessionCreate {
	if t != nil {
		msc.SetStartedAt(*t)
	}
	return msc
}

// SetEndedAt sets the "ended_at" field.
func (msc *MeetingSessionCreate) SetEndedAt(t time.Time) *MeetingSessionCreate {
	msc.mutation.SetEndedAt(t)
	return msc
}

// SetNillableEndedAt sets the "ended_at" field if the given value is not nil.
func (msc *MeetingSessionCreate) SetNillableEndedAt(t *time.Time) *MeetingSessionCreate {
	if t != nil {
		msc.SetEndedAt(*t)
	}
	return msc
}

// SetDocumentName sets the "document_name" field.
func (msc *MeetingSessionCreate) SetDocumentName(s string) *MeetingSessionCreate {
	msc.mutation.SetDocumentName(s)
	return msc
}

// SetID sets the "id" field.
func (msc *MeetingSessionCreate) SetID(u uuid.UUID) *MeetingSessionCreate {
	msc.mutation.SetID(u)
	return msc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (msc *MeetingSessionCreate) SetNillableID(u *uuid.UUID) *MeetingSessionCreate {
	if u != nil {
		msc.SetID(*u)
	}
	return msc
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (msc *MeetingSessionCreate) AddIncidentIDs(ids ...uuid.UUID) *MeetingSessionCreate {
	msc.mutation.AddIncidentIDs(ids...)
	return msc
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (msc *MeetingSessionCreate) AddIncidents(i ...*Incident) *MeetingSessionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return msc.AddIncidentIDs(ids...)
}

// Mutation returns the MeetingSessionMutation object of the builder.
func (msc *MeetingSessionCreate) Mutation() *MeetingSessionMutation {
	return msc.mutation
}

// Save creates the MeetingSession in the database.
func (msc *MeetingSessionCreate) Save(ctx context.Context) (*MeetingSession, error) {
	msc.defaults()
	return withHooks(ctx, msc.sqlSave, msc.mutation, msc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (msc *MeetingSessionCreate) SaveX(ctx context.Context) *MeetingSession {
	v, err := msc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (msc *MeetingSessionCreate) Exec(ctx context.Context) error {
	_, err := msc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (msc *MeetingSessionCreate) ExecX(ctx context.Context) {
	if err := msc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (msc *MeetingSessionCreate) defaults() {
	if _, ok := msc.mutation.StartedAt(); !ok {
		v := meetingsession.DefaultStartedAt()
		msc.mutation.SetStartedAt(v)
	}
	if _, ok := msc.mutation.ID(); !ok {
		v := meetingsession.DefaultID()
		msc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (msc *MeetingSessionCreate) check() error {
	if _, ok := msc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`ent: missing required field "MeetingSession.title"`)}
	}
	if _, ok := msc.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`ent: missing required field "MeetingSession.started_at"`)}
	}
	if _, ok := msc.mutation.DocumentName(); !ok {
		return &ValidationError{Name: "document_name", err: errors.New(`ent: missing required field "MeetingSession.document_name"`)}
	}
	return nil
}

func (msc *MeetingSessionCreate) sqlSave(ctx context.Context) (*MeetingSession, error) {
	if err := msc.check(); err != nil {
		return nil, err
	}
	_node, _spec := msc.createSpec()
	if err := sqlgraph.CreateNode(ctx, msc.driver, _spec); err != nil {
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
	msc.mutation.id = &_node.ID
	msc.mutation.done = true
	return _node, nil
}

func (msc *MeetingSessionCreate) createSpec() (*MeetingSession, *sqlgraph.CreateSpec) {
	var (
		_node = &MeetingSession{config: msc.config}
		_spec = sqlgraph.NewCreateSpec(meetingsession.Table, sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = msc.conflict
	if id, ok := msc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := msc.mutation.Title(); ok {
		_spec.SetField(meetingsession.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := msc.mutation.StartedAt(); ok {
		_spec.SetField(meetingsession.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if value, ok := msc.mutation.EndedAt(); ok {
		_spec.SetField(meetingsession.FieldEndedAt, field.TypeTime, value)
		_node.EndedAt = value
	}
	if value, ok := msc.mutation.DocumentName(); ok {
		_spec.SetField(meetingsession.FieldDocumentName, field.TypeString, value)
		_node.DocumentName = value
	}
	if nodes := msc.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   meetingsession.IncidentsTable,
			Columns: meetingsession.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
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
//	client.MeetingSession.Create().
//		SetTitle(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.MeetingSessionUpsert) {
//			SetTitle(v+v).
//		}).
//		Exec(ctx)
func (msc *MeetingSessionCreate) OnConflict(opts ...sql.ConflictOption) *MeetingSessionUpsertOne {
	msc.conflict = opts
	return &MeetingSessionUpsertOne{
		create: msc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.MeetingSession.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (msc *MeetingSessionCreate) OnConflictColumns(columns ...string) *MeetingSessionUpsertOne {
	msc.conflict = append(msc.conflict, sql.ConflictColumns(columns...))
	return &MeetingSessionUpsertOne{
		create: msc,
	}
}

type (
	// MeetingSessionUpsertOne is the builder for "upsert"-ing
	//  one MeetingSession node.
	MeetingSessionUpsertOne struct {
		create *MeetingSessionCreate
	}

	// MeetingSessionUpsert is the "OnConflict" setter.
	MeetingSessionUpsert struct {
		*sql.UpdateSet
	}
)

// SetTitle sets the "title" field.
func (u *MeetingSessionUpsert) SetTitle(v string) *MeetingSessionUpsert {
	u.Set(meetingsession.FieldTitle, v)
	return u
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *MeetingSessionUpsert) UpdateTitle() *MeetingSessionUpsert {
	u.SetExcluded(meetingsession.FieldTitle)
	return u
}

// SetStartedAt sets the "started_at" field.
func (u *MeetingSessionUpsert) SetStartedAt(v time.Time) *MeetingSessionUpsert {
	u.Set(meetingsession.FieldStartedAt, v)
	return u
}

// UpdateStartedAt sets the "started_at" field to the value that was provided on create.
func (u *MeetingSessionUpsert) UpdateStartedAt() *MeetingSessionUpsert {
	u.SetExcluded(meetingsession.FieldStartedAt)
	return u
}

// SetEndedAt sets the "ended_at" field.
func (u *MeetingSessionUpsert) SetEndedAt(v time.Time) *MeetingSessionUpsert {
	u.Set(meetingsession.FieldEndedAt, v)
	return u
}

// UpdateEndedAt sets the "ended_at" field to the value that was provided on create.
func (u *MeetingSessionUpsert) UpdateEndedAt() *MeetingSessionUpsert {
	u.SetExcluded(meetingsession.FieldEndedAt)
	return u
}

// ClearEndedAt clears the value of the "ended_at" field.
func (u *MeetingSessionUpsert) ClearEndedAt() *MeetingSessionUpsert {
	u.SetNull(meetingsession.FieldEndedAt)
	return u
}

// SetDocumentName sets the "document_name" field.
func (u *MeetingSessionUpsert) SetDocumentName(v string) *MeetingSessionUpsert {
	u.Set(meetingsession.FieldDocumentName, v)
	return u
}

// UpdateDocumentName sets the "document_name" field to the value that was provided on create.
func (u *MeetingSessionUpsert) UpdateDocumentName() *MeetingSessionUpsert {
	u.SetExcluded(meetingsession.FieldDocumentName)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.MeetingSession.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(meetingsession.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *MeetingSessionUpsertOne) UpdateNewValues() *MeetingSessionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(meetingsession.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.MeetingSession.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *MeetingSessionUpsertOne) Ignore() *MeetingSessionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *MeetingSessionUpsertOne) DoNothing() *MeetingSessionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the MeetingSessionCreate.OnConflict
// documentation for more info.
func (u *MeetingSessionUpsertOne) Update(set func(*MeetingSessionUpsert)) *MeetingSessionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&MeetingSessionUpsert{UpdateSet: update})
	}))
	return u
}

// SetTitle sets the "title" field.
func (u *MeetingSessionUpsertOne) SetTitle(v string) *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetTitle(v)
	})
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *MeetingSessionUpsertOne) UpdateTitle() *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateTitle()
	})
}

// SetStartedAt sets the "started_at" field.
func (u *MeetingSessionUpsertOne) SetStartedAt(v time.Time) *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetStartedAt(v)
	})
}

// UpdateStartedAt sets the "started_at" field to the value that was provided on create.
func (u *MeetingSessionUpsertOne) UpdateStartedAt() *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateStartedAt()
	})
}

// SetEndedAt sets the "ended_at" field.
func (u *MeetingSessionUpsertOne) SetEndedAt(v time.Time) *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetEndedAt(v)
	})
}

// UpdateEndedAt sets the "ended_at" field to the value that was provided on create.
func (u *MeetingSessionUpsertOne) UpdateEndedAt() *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateEndedAt()
	})
}

// ClearEndedAt clears the value of the "ended_at" field.
func (u *MeetingSessionUpsertOne) ClearEndedAt() *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.ClearEndedAt()
	})
}

// SetDocumentName sets the "document_name" field.
func (u *MeetingSessionUpsertOne) SetDocumentName(v string) *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetDocumentName(v)
	})
}

// UpdateDocumentName sets the "document_name" field to the value that was provided on create.
func (u *MeetingSessionUpsertOne) UpdateDocumentName() *MeetingSessionUpsertOne {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateDocumentName()
	})
}

// Exec executes the query.
func (u *MeetingSessionUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for MeetingSessionCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *MeetingSessionUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *MeetingSessionUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: MeetingSessionUpsertOne.ID is not supported by MySQL driver. Use MeetingSessionUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *MeetingSessionUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// MeetingSessionCreateBulk is the builder for creating many MeetingSession entities in bulk.
type MeetingSessionCreateBulk struct {
	config
	err      error
	builders []*MeetingSessionCreate
	conflict []sql.ConflictOption
}

// Save creates the MeetingSession entities in the database.
func (mscb *MeetingSessionCreateBulk) Save(ctx context.Context) ([]*MeetingSession, error) {
	if mscb.err != nil {
		return nil, mscb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(mscb.builders))
	nodes := make([]*MeetingSession, len(mscb.builders))
	mutators := make([]Mutator, len(mscb.builders))
	for i := range mscb.builders {
		func(i int, root context.Context) {
			builder := mscb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*MeetingSessionMutation)
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
					_, err = mutators[i+1].Mutate(root, mscb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = mscb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, mscb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, mscb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (mscb *MeetingSessionCreateBulk) SaveX(ctx context.Context) []*MeetingSession {
	v, err := mscb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mscb *MeetingSessionCreateBulk) Exec(ctx context.Context) error {
	_, err := mscb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mscb *MeetingSessionCreateBulk) ExecX(ctx context.Context) {
	if err := mscb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.MeetingSession.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.MeetingSessionUpsert) {
//			SetTitle(v+v).
//		}).
//		Exec(ctx)
func (mscb *MeetingSessionCreateBulk) OnConflict(opts ...sql.ConflictOption) *MeetingSessionUpsertBulk {
	mscb.conflict = opts
	return &MeetingSessionUpsertBulk{
		create: mscb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.MeetingSession.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (mscb *MeetingSessionCreateBulk) OnConflictColumns(columns ...string) *MeetingSessionUpsertBulk {
	mscb.conflict = append(mscb.conflict, sql.ConflictColumns(columns...))
	return &MeetingSessionUpsertBulk{
		create: mscb,
	}
}

// MeetingSessionUpsertBulk is the builder for "upsert"-ing
// a bulk of MeetingSession nodes.
type MeetingSessionUpsertBulk struct {
	create *MeetingSessionCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.MeetingSession.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(meetingsession.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *MeetingSessionUpsertBulk) UpdateNewValues() *MeetingSessionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(meetingsession.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.MeetingSession.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *MeetingSessionUpsertBulk) Ignore() *MeetingSessionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *MeetingSessionUpsertBulk) DoNothing() *MeetingSessionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the MeetingSessionCreateBulk.OnConflict
// documentation for more info.
func (u *MeetingSessionUpsertBulk) Update(set func(*MeetingSessionUpsert)) *MeetingSessionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&MeetingSessionUpsert{UpdateSet: update})
	}))
	return u
}

// SetTitle sets the "title" field.
func (u *MeetingSessionUpsertBulk) SetTitle(v string) *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetTitle(v)
	})
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *MeetingSessionUpsertBulk) UpdateTitle() *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateTitle()
	})
}

// SetStartedAt sets the "started_at" field.
func (u *MeetingSessionUpsertBulk) SetStartedAt(v time.Time) *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetStartedAt(v)
	})
}

// UpdateStartedAt sets the "started_at" field to the value that was provided on create.
func (u *MeetingSessionUpsertBulk) UpdateStartedAt() *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateStartedAt()
	})
}

// SetEndedAt sets the "ended_at" field.
func (u *MeetingSessionUpsertBulk) SetEndedAt(v time.Time) *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetEndedAt(v)
	})
}

// UpdateEndedAt sets the "ended_at" field to the value that was provided on create.
func (u *MeetingSessionUpsertBulk) UpdateEndedAt() *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateEndedAt()
	})
}

// ClearEndedAt clears the value of the "ended_at" field.
func (u *MeetingSessionUpsertBulk) ClearEndedAt() *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.ClearEndedAt()
	})
}

// SetDocumentName sets the "document_name" field.
func (u *MeetingSessionUpsertBulk) SetDocumentName(v string) *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.SetDocumentName(v)
	})
}

// UpdateDocumentName sets the "document_name" field to the value that was provided on create.
func (u *MeetingSessionUpsertBulk) UpdateDocumentName() *MeetingSessionUpsertBulk {
	return u.Update(func(s *MeetingSessionUpsert) {
		s.UpdateDocumentName()
	})
}

// Exec executes the query.
func (u *MeetingSessionUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the MeetingSessionCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for MeetingSessionCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *MeetingSessionUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

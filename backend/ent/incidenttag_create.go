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
	"github.com/rezible/rezible/ent/incidentdebriefquestion"
	"github.com/rezible/rezible/ent/incidenttag"
)

// IncidentTagCreate is the builder for creating a IncidentTag entity.
type IncidentTagCreate struct {
	config
	mutation *IncidentTagMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetArchiveTime sets the "archive_time" field.
func (itc *IncidentTagCreate) SetArchiveTime(t time.Time) *IncidentTagCreate {
	itc.mutation.SetArchiveTime(t)
	return itc
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (itc *IncidentTagCreate) SetNillableArchiveTime(t *time.Time) *IncidentTagCreate {
	if t != nil {
		itc.SetArchiveTime(*t)
	}
	return itc
}

// SetKey sets the "key" field.
func (itc *IncidentTagCreate) SetKey(s string) *IncidentTagCreate {
	itc.mutation.SetKey(s)
	return itc
}

// SetValue sets the "value" field.
func (itc *IncidentTagCreate) SetValue(s string) *IncidentTagCreate {
	itc.mutation.SetValue(s)
	return itc
}

// SetID sets the "id" field.
func (itc *IncidentTagCreate) SetID(u uuid.UUID) *IncidentTagCreate {
	itc.mutation.SetID(u)
	return itc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (itc *IncidentTagCreate) SetNillableID(u *uuid.UUID) *IncidentTagCreate {
	if u != nil {
		itc.SetID(*u)
	}
	return itc
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (itc *IncidentTagCreate) AddIncidentIDs(ids ...uuid.UUID) *IncidentTagCreate {
	itc.mutation.AddIncidentIDs(ids...)
	return itc
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (itc *IncidentTagCreate) AddIncidents(i ...*Incident) *IncidentTagCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itc.AddIncidentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (itc *IncidentTagCreate) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTagCreate {
	itc.mutation.AddDebriefQuestionIDs(ids...)
	return itc
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (itc *IncidentTagCreate) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTagCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itc.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentTagMutation object of the builder.
func (itc *IncidentTagCreate) Mutation() *IncidentTagMutation {
	return itc.mutation
}

// Save creates the IncidentTag in the database.
func (itc *IncidentTagCreate) Save(ctx context.Context) (*IncidentTag, error) {
	if err := itc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, itc.sqlSave, itc.mutation, itc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (itc *IncidentTagCreate) SaveX(ctx context.Context) *IncidentTag {
	v, err := itc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (itc *IncidentTagCreate) Exec(ctx context.Context) error {
	_, err := itc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (itc *IncidentTagCreate) ExecX(ctx context.Context) {
	if err := itc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (itc *IncidentTagCreate) defaults() error {
	if _, ok := itc.mutation.ID(); !ok {
		if incidenttag.DefaultID == nil {
			return fmt.Errorf("ent: uninitialized incidenttag.DefaultID (forgotten import ent/runtime?)")
		}
		v := incidenttag.DefaultID()
		itc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (itc *IncidentTagCreate) check() error {
	if _, ok := itc.mutation.Key(); !ok {
		return &ValidationError{Name: "key", err: errors.New(`ent: missing required field "IncidentTag.key"`)}
	}
	if _, ok := itc.mutation.Value(); !ok {
		return &ValidationError{Name: "value", err: errors.New(`ent: missing required field "IncidentTag.value"`)}
	}
	return nil
}

func (itc *IncidentTagCreate) sqlSave(ctx context.Context) (*IncidentTag, error) {
	if err := itc.check(); err != nil {
		return nil, err
	}
	_node, _spec := itc.createSpec()
	if err := sqlgraph.CreateNode(ctx, itc.driver, _spec); err != nil {
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
	itc.mutation.id = &_node.ID
	itc.mutation.done = true
	return _node, nil
}

func (itc *IncidentTagCreate) createSpec() (*IncidentTag, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentTag{config: itc.config}
		_spec = sqlgraph.NewCreateSpec(incidenttag.Table, sqlgraph.NewFieldSpec(incidenttag.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = itc.conflict
	if id, ok := itc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := itc.mutation.ArchiveTime(); ok {
		_spec.SetField(incidenttag.FieldArchiveTime, field.TypeTime, value)
		_node.ArchiveTime = value
	}
	if value, ok := itc.mutation.Key(); ok {
		_spec.SetField(incidenttag.FieldKey, field.TypeString, value)
		_node.Key = value
	}
	if value, ok := itc.mutation.Value(); ok {
		_spec.SetField(incidenttag.FieldValue, field.TypeString, value)
		_node.Value = value
	}
	if nodes := itc.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidenttag.IncidentsTable,
			Columns: incidenttag.IncidentsPrimaryKey,
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
	if nodes := itc.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidenttag.DebriefQuestionsTable,
			Columns: incidenttag.DebriefQuestionsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefquestion.FieldID, field.TypeUUID),
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
//	client.IncidentTag.Create().
//		SetArchiveTime(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentTagUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (itc *IncidentTagCreate) OnConflict(opts ...sql.ConflictOption) *IncidentTagUpsertOne {
	itc.conflict = opts
	return &IncidentTagUpsertOne{
		create: itc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentTag.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (itc *IncidentTagCreate) OnConflictColumns(columns ...string) *IncidentTagUpsertOne {
	itc.conflict = append(itc.conflict, sql.ConflictColumns(columns...))
	return &IncidentTagUpsertOne{
		create: itc,
	}
}

type (
	// IncidentTagUpsertOne is the builder for "upsert"-ing
	//  one IncidentTag node.
	IncidentTagUpsertOne struct {
		create *IncidentTagCreate
	}

	// IncidentTagUpsert is the "OnConflict" setter.
	IncidentTagUpsert struct {
		*sql.UpdateSet
	}
)

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentTagUpsert) SetArchiveTime(v time.Time) *IncidentTagUpsert {
	u.Set(incidenttag.FieldArchiveTime, v)
	return u
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentTagUpsert) UpdateArchiveTime() *IncidentTagUpsert {
	u.SetExcluded(incidenttag.FieldArchiveTime)
	return u
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentTagUpsert) ClearArchiveTime() *IncidentTagUpsert {
	u.SetNull(incidenttag.FieldArchiveTime)
	return u
}

// SetKey sets the "key" field.
func (u *IncidentTagUpsert) SetKey(v string) *IncidentTagUpsert {
	u.Set(incidenttag.FieldKey, v)
	return u
}

// UpdateKey sets the "key" field to the value that was provided on create.
func (u *IncidentTagUpsert) UpdateKey() *IncidentTagUpsert {
	u.SetExcluded(incidenttag.FieldKey)
	return u
}

// SetValue sets the "value" field.
func (u *IncidentTagUpsert) SetValue(v string) *IncidentTagUpsert {
	u.Set(incidenttag.FieldValue, v)
	return u
}

// UpdateValue sets the "value" field to the value that was provided on create.
func (u *IncidentTagUpsert) UpdateValue() *IncidentTagUpsert {
	u.SetExcluded(incidenttag.FieldValue)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentTag.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidenttag.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentTagUpsertOne) UpdateNewValues() *IncidentTagUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidenttag.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentTag.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentTagUpsertOne) Ignore() *IncidentTagUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentTagUpsertOne) DoNothing() *IncidentTagUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentTagCreate.OnConflict
// documentation for more info.
func (u *IncidentTagUpsertOne) Update(set func(*IncidentTagUpsert)) *IncidentTagUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentTagUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentTagUpsertOne) SetArchiveTime(v time.Time) *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentTagUpsertOne) UpdateArchiveTime() *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentTagUpsertOne) ClearArchiveTime() *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.ClearArchiveTime()
	})
}

// SetKey sets the "key" field.
func (u *IncidentTagUpsertOne) SetKey(v string) *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.SetKey(v)
	})
}

// UpdateKey sets the "key" field to the value that was provided on create.
func (u *IncidentTagUpsertOne) UpdateKey() *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.UpdateKey()
	})
}

// SetValue sets the "value" field.
func (u *IncidentTagUpsertOne) SetValue(v string) *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.SetValue(v)
	})
}

// UpdateValue sets the "value" field to the value that was provided on create.
func (u *IncidentTagUpsertOne) UpdateValue() *IncidentTagUpsertOne {
	return u.Update(func(s *IncidentTagUpsert) {
		s.UpdateValue()
	})
}

// Exec executes the query.
func (u *IncidentTagUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentTagCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentTagUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentTagUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentTagUpsertOne.ID is not supported by MySQL driver. Use IncidentTagUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentTagUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentTagCreateBulk is the builder for creating many IncidentTag entities in bulk.
type IncidentTagCreateBulk struct {
	config
	err      error
	builders []*IncidentTagCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentTag entities in the database.
func (itcb *IncidentTagCreateBulk) Save(ctx context.Context) ([]*IncidentTag, error) {
	if itcb.err != nil {
		return nil, itcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(itcb.builders))
	nodes := make([]*IncidentTag, len(itcb.builders))
	mutators := make([]Mutator, len(itcb.builders))
	for i := range itcb.builders {
		func(i int, root context.Context) {
			builder := itcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentTagMutation)
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
					_, err = mutators[i+1].Mutate(root, itcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = itcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, itcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, itcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (itcb *IncidentTagCreateBulk) SaveX(ctx context.Context) []*IncidentTag {
	v, err := itcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (itcb *IncidentTagCreateBulk) Exec(ctx context.Context) error {
	_, err := itcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (itcb *IncidentTagCreateBulk) ExecX(ctx context.Context) {
	if err := itcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentTag.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentTagUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (itcb *IncidentTagCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentTagUpsertBulk {
	itcb.conflict = opts
	return &IncidentTagUpsertBulk{
		create: itcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentTag.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (itcb *IncidentTagCreateBulk) OnConflictColumns(columns ...string) *IncidentTagUpsertBulk {
	itcb.conflict = append(itcb.conflict, sql.ConflictColumns(columns...))
	return &IncidentTagUpsertBulk{
		create: itcb,
	}
}

// IncidentTagUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentTag nodes.
type IncidentTagUpsertBulk struct {
	create *IncidentTagCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentTag.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidenttag.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentTagUpsertBulk) UpdateNewValues() *IncidentTagUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidenttag.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentTag.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentTagUpsertBulk) Ignore() *IncidentTagUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentTagUpsertBulk) DoNothing() *IncidentTagUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentTagCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentTagUpsertBulk) Update(set func(*IncidentTagUpsert)) *IncidentTagUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentTagUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentTagUpsertBulk) SetArchiveTime(v time.Time) *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentTagUpsertBulk) UpdateArchiveTime() *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentTagUpsertBulk) ClearArchiveTime() *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.ClearArchiveTime()
	})
}

// SetKey sets the "key" field.
func (u *IncidentTagUpsertBulk) SetKey(v string) *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.SetKey(v)
	})
}

// UpdateKey sets the "key" field to the value that was provided on create.
func (u *IncidentTagUpsertBulk) UpdateKey() *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.UpdateKey()
	})
}

// SetValue sets the "value" field.
func (u *IncidentTagUpsertBulk) SetValue(v string) *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.SetValue(v)
	})
}

// UpdateValue sets the "value" field to the value that was provided on create.
func (u *IncidentTagUpsertBulk) UpdateValue() *IncidentTagUpsertBulk {
	return u.Update(func(s *IncidentTagUpsert) {
		s.UpdateValue()
	})
}

// Exec executes the query.
func (u *IncidentTagUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentTagCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentTagCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentTagUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

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
	"github.com/rezible/rezible/ent/incidentseverity"
)

// IncidentSeverityCreate is the builder for creating a IncidentSeverity entity.
type IncidentSeverityCreate struct {
	config
	mutation *IncidentSeverityMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetArchiveTime sets the "archive_time" field.
func (isc *IncidentSeverityCreate) SetArchiveTime(t time.Time) *IncidentSeverityCreate {
	isc.mutation.SetArchiveTime(t)
	return isc
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (isc *IncidentSeverityCreate) SetNillableArchiveTime(t *time.Time) *IncidentSeverityCreate {
	if t != nil {
		isc.SetArchiveTime(*t)
	}
	return isc
}

// SetProviderID sets the "provider_id" field.
func (isc *IncidentSeverityCreate) SetProviderID(s string) *IncidentSeverityCreate {
	isc.mutation.SetProviderID(s)
	return isc
}

// SetNillableProviderID sets the "provider_id" field if the given value is not nil.
func (isc *IncidentSeverityCreate) SetNillableProviderID(s *string) *IncidentSeverityCreate {
	if s != nil {
		isc.SetProviderID(*s)
	}
	return isc
}

// SetName sets the "name" field.
func (isc *IncidentSeverityCreate) SetName(s string) *IncidentSeverityCreate {
	isc.mutation.SetName(s)
	return isc
}

// SetRank sets the "rank" field.
func (isc *IncidentSeverityCreate) SetRank(i int) *IncidentSeverityCreate {
	isc.mutation.SetRank(i)
	return isc
}

// SetColor sets the "color" field.
func (isc *IncidentSeverityCreate) SetColor(s string) *IncidentSeverityCreate {
	isc.mutation.SetColor(s)
	return isc
}

// SetNillableColor sets the "color" field if the given value is not nil.
func (isc *IncidentSeverityCreate) SetNillableColor(s *string) *IncidentSeverityCreate {
	if s != nil {
		isc.SetColor(*s)
	}
	return isc
}

// SetDescription sets the "description" field.
func (isc *IncidentSeverityCreate) SetDescription(s string) *IncidentSeverityCreate {
	isc.mutation.SetDescription(s)
	return isc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (isc *IncidentSeverityCreate) SetNillableDescription(s *string) *IncidentSeverityCreate {
	if s != nil {
		isc.SetDescription(*s)
	}
	return isc
}

// SetID sets the "id" field.
func (isc *IncidentSeverityCreate) SetID(u uuid.UUID) *IncidentSeverityCreate {
	isc.mutation.SetID(u)
	return isc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (isc *IncidentSeverityCreate) SetNillableID(u *uuid.UUID) *IncidentSeverityCreate {
	if u != nil {
		isc.SetID(*u)
	}
	return isc
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (isc *IncidentSeverityCreate) AddIncidentIDs(ids ...uuid.UUID) *IncidentSeverityCreate {
	isc.mutation.AddIncidentIDs(ids...)
	return isc
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (isc *IncidentSeverityCreate) AddIncidents(i ...*Incident) *IncidentSeverityCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return isc.AddIncidentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (isc *IncidentSeverityCreate) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentSeverityCreate {
	isc.mutation.AddDebriefQuestionIDs(ids...)
	return isc
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (isc *IncidentSeverityCreate) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentSeverityCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return isc.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentSeverityMutation object of the builder.
func (isc *IncidentSeverityCreate) Mutation() *IncidentSeverityMutation {
	return isc.mutation
}

// Save creates the IncidentSeverity in the database.
func (isc *IncidentSeverityCreate) Save(ctx context.Context) (*IncidentSeverity, error) {
	if err := isc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, isc.sqlSave, isc.mutation, isc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (isc *IncidentSeverityCreate) SaveX(ctx context.Context) *IncidentSeverity {
	v, err := isc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (isc *IncidentSeverityCreate) Exec(ctx context.Context) error {
	_, err := isc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (isc *IncidentSeverityCreate) ExecX(ctx context.Context) {
	if err := isc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (isc *IncidentSeverityCreate) defaults() error {
	if _, ok := isc.mutation.ID(); !ok {
		if incidentseverity.DefaultID == nil {
			return fmt.Errorf("ent: uninitialized incidentseverity.DefaultID (forgotten import ent/runtime?)")
		}
		v := incidentseverity.DefaultID()
		isc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (isc *IncidentSeverityCreate) check() error {
	if _, ok := isc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "IncidentSeverity.name"`)}
	}
	if _, ok := isc.mutation.Rank(); !ok {
		return &ValidationError{Name: "rank", err: errors.New(`ent: missing required field "IncidentSeverity.rank"`)}
	}
	return nil
}

func (isc *IncidentSeverityCreate) sqlSave(ctx context.Context) (*IncidentSeverity, error) {
	if err := isc.check(); err != nil {
		return nil, err
	}
	_node, _spec := isc.createSpec()
	if err := sqlgraph.CreateNode(ctx, isc.driver, _spec); err != nil {
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
	isc.mutation.id = &_node.ID
	isc.mutation.done = true
	return _node, nil
}

func (isc *IncidentSeverityCreate) createSpec() (*IncidentSeverity, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentSeverity{config: isc.config}
		_spec = sqlgraph.NewCreateSpec(incidentseverity.Table, sqlgraph.NewFieldSpec(incidentseverity.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = isc.conflict
	if id, ok := isc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := isc.mutation.ArchiveTime(); ok {
		_spec.SetField(incidentseverity.FieldArchiveTime, field.TypeTime, value)
		_node.ArchiveTime = value
	}
	if value, ok := isc.mutation.ProviderID(); ok {
		_spec.SetField(incidentseverity.FieldProviderID, field.TypeString, value)
		_node.ProviderID = value
	}
	if value, ok := isc.mutation.Name(); ok {
		_spec.SetField(incidentseverity.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := isc.mutation.Rank(); ok {
		_spec.SetField(incidentseverity.FieldRank, field.TypeInt, value)
		_node.Rank = value
	}
	if value, ok := isc.mutation.Color(); ok {
		_spec.SetField(incidentseverity.FieldColor, field.TypeString, value)
		_node.Color = value
	}
	if value, ok := isc.mutation.Description(); ok {
		_spec.SetField(incidentseverity.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if nodes := isc.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentseverity.IncidentsTable,
			Columns: []string{incidentseverity.IncidentsColumn},
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
	if nodes := isc.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentseverity.DebriefQuestionsTable,
			Columns: incidentseverity.DebriefQuestionsPrimaryKey,
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
//	client.IncidentSeverity.Create().
//		SetArchiveTime(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentSeverityUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (isc *IncidentSeverityCreate) OnConflict(opts ...sql.ConflictOption) *IncidentSeverityUpsertOne {
	isc.conflict = opts
	return &IncidentSeverityUpsertOne{
		create: isc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentSeverity.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (isc *IncidentSeverityCreate) OnConflictColumns(columns ...string) *IncidentSeverityUpsertOne {
	isc.conflict = append(isc.conflict, sql.ConflictColumns(columns...))
	return &IncidentSeverityUpsertOne{
		create: isc,
	}
}

type (
	// IncidentSeverityUpsertOne is the builder for "upsert"-ing
	//  one IncidentSeverity node.
	IncidentSeverityUpsertOne struct {
		create *IncidentSeverityCreate
	}

	// IncidentSeverityUpsert is the "OnConflict" setter.
	IncidentSeverityUpsert struct {
		*sql.UpdateSet
	}
)

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentSeverityUpsert) SetArchiveTime(v time.Time) *IncidentSeverityUpsert {
	u.Set(incidentseverity.FieldArchiveTime, v)
	return u
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentSeverityUpsert) UpdateArchiveTime() *IncidentSeverityUpsert {
	u.SetExcluded(incidentseverity.FieldArchiveTime)
	return u
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentSeverityUpsert) ClearArchiveTime() *IncidentSeverityUpsert {
	u.SetNull(incidentseverity.FieldArchiveTime)
	return u
}

// SetProviderID sets the "provider_id" field.
func (u *IncidentSeverityUpsert) SetProviderID(v string) *IncidentSeverityUpsert {
	u.Set(incidentseverity.FieldProviderID, v)
	return u
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *IncidentSeverityUpsert) UpdateProviderID() *IncidentSeverityUpsert {
	u.SetExcluded(incidentseverity.FieldProviderID)
	return u
}

// ClearProviderID clears the value of the "provider_id" field.
func (u *IncidentSeverityUpsert) ClearProviderID() *IncidentSeverityUpsert {
	u.SetNull(incidentseverity.FieldProviderID)
	return u
}

// SetName sets the "name" field.
func (u *IncidentSeverityUpsert) SetName(v string) *IncidentSeverityUpsert {
	u.Set(incidentseverity.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *IncidentSeverityUpsert) UpdateName() *IncidentSeverityUpsert {
	u.SetExcluded(incidentseverity.FieldName)
	return u
}

// SetRank sets the "rank" field.
func (u *IncidentSeverityUpsert) SetRank(v int) *IncidentSeverityUpsert {
	u.Set(incidentseverity.FieldRank, v)
	return u
}

// UpdateRank sets the "rank" field to the value that was provided on create.
func (u *IncidentSeverityUpsert) UpdateRank() *IncidentSeverityUpsert {
	u.SetExcluded(incidentseverity.FieldRank)
	return u
}

// AddRank adds v to the "rank" field.
func (u *IncidentSeverityUpsert) AddRank(v int) *IncidentSeverityUpsert {
	u.Add(incidentseverity.FieldRank, v)
	return u
}

// SetColor sets the "color" field.
func (u *IncidentSeverityUpsert) SetColor(v string) *IncidentSeverityUpsert {
	u.Set(incidentseverity.FieldColor, v)
	return u
}

// UpdateColor sets the "color" field to the value that was provided on create.
func (u *IncidentSeverityUpsert) UpdateColor() *IncidentSeverityUpsert {
	u.SetExcluded(incidentseverity.FieldColor)
	return u
}

// ClearColor clears the value of the "color" field.
func (u *IncidentSeverityUpsert) ClearColor() *IncidentSeverityUpsert {
	u.SetNull(incidentseverity.FieldColor)
	return u
}

// SetDescription sets the "description" field.
func (u *IncidentSeverityUpsert) SetDescription(v string) *IncidentSeverityUpsert {
	u.Set(incidentseverity.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *IncidentSeverityUpsert) UpdateDescription() *IncidentSeverityUpsert {
	u.SetExcluded(incidentseverity.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *IncidentSeverityUpsert) ClearDescription() *IncidentSeverityUpsert {
	u.SetNull(incidentseverity.FieldDescription)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentSeverity.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentseverity.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentSeverityUpsertOne) UpdateNewValues() *IncidentSeverityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidentseverity.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentSeverity.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentSeverityUpsertOne) Ignore() *IncidentSeverityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentSeverityUpsertOne) DoNothing() *IncidentSeverityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentSeverityCreate.OnConflict
// documentation for more info.
func (u *IncidentSeverityUpsertOne) Update(set func(*IncidentSeverityUpsert)) *IncidentSeverityUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentSeverityUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentSeverityUpsertOne) SetArchiveTime(v time.Time) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentSeverityUpsertOne) UpdateArchiveTime() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentSeverityUpsertOne) ClearArchiveTime() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearArchiveTime()
	})
}

// SetProviderID sets the "provider_id" field.
func (u *IncidentSeverityUpsertOne) SetProviderID(v string) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetProviderID(v)
	})
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *IncidentSeverityUpsertOne) UpdateProviderID() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateProviderID()
	})
}

// ClearProviderID clears the value of the "provider_id" field.
func (u *IncidentSeverityUpsertOne) ClearProviderID() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearProviderID()
	})
}

// SetName sets the "name" field.
func (u *IncidentSeverityUpsertOne) SetName(v string) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *IncidentSeverityUpsertOne) UpdateName() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateName()
	})
}

// SetRank sets the "rank" field.
func (u *IncidentSeverityUpsertOne) SetRank(v int) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetRank(v)
	})
}

// AddRank adds v to the "rank" field.
func (u *IncidentSeverityUpsertOne) AddRank(v int) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.AddRank(v)
	})
}

// UpdateRank sets the "rank" field to the value that was provided on create.
func (u *IncidentSeverityUpsertOne) UpdateRank() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateRank()
	})
}

// SetColor sets the "color" field.
func (u *IncidentSeverityUpsertOne) SetColor(v string) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetColor(v)
	})
}

// UpdateColor sets the "color" field to the value that was provided on create.
func (u *IncidentSeverityUpsertOne) UpdateColor() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateColor()
	})
}

// ClearColor clears the value of the "color" field.
func (u *IncidentSeverityUpsertOne) ClearColor() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearColor()
	})
}

// SetDescription sets the "description" field.
func (u *IncidentSeverityUpsertOne) SetDescription(v string) *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *IncidentSeverityUpsertOne) UpdateDescription() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *IncidentSeverityUpsertOne) ClearDescription() *IncidentSeverityUpsertOne {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearDescription()
	})
}

// Exec executes the query.
func (u *IncidentSeverityUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentSeverityCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentSeverityUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentSeverityUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentSeverityUpsertOne.ID is not supported by MySQL driver. Use IncidentSeverityUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentSeverityUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentSeverityCreateBulk is the builder for creating many IncidentSeverity entities in bulk.
type IncidentSeverityCreateBulk struct {
	config
	err      error
	builders []*IncidentSeverityCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentSeverity entities in the database.
func (iscb *IncidentSeverityCreateBulk) Save(ctx context.Context) ([]*IncidentSeverity, error) {
	if iscb.err != nil {
		return nil, iscb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(iscb.builders))
	nodes := make([]*IncidentSeverity, len(iscb.builders))
	mutators := make([]Mutator, len(iscb.builders))
	for i := range iscb.builders {
		func(i int, root context.Context) {
			builder := iscb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentSeverityMutation)
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
					_, err = mutators[i+1].Mutate(root, iscb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = iscb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, iscb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, iscb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (iscb *IncidentSeverityCreateBulk) SaveX(ctx context.Context) []*IncidentSeverity {
	v, err := iscb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (iscb *IncidentSeverityCreateBulk) Exec(ctx context.Context) error {
	_, err := iscb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iscb *IncidentSeverityCreateBulk) ExecX(ctx context.Context) {
	if err := iscb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentSeverity.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentSeverityUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (iscb *IncidentSeverityCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentSeverityUpsertBulk {
	iscb.conflict = opts
	return &IncidentSeverityUpsertBulk{
		create: iscb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentSeverity.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (iscb *IncidentSeverityCreateBulk) OnConflictColumns(columns ...string) *IncidentSeverityUpsertBulk {
	iscb.conflict = append(iscb.conflict, sql.ConflictColumns(columns...))
	return &IncidentSeverityUpsertBulk{
		create: iscb,
	}
}

// IncidentSeverityUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentSeverity nodes.
type IncidentSeverityUpsertBulk struct {
	create *IncidentSeverityCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentSeverity.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentseverity.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentSeverityUpsertBulk) UpdateNewValues() *IncidentSeverityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidentseverity.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentSeverity.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentSeverityUpsertBulk) Ignore() *IncidentSeverityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentSeverityUpsertBulk) DoNothing() *IncidentSeverityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentSeverityCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentSeverityUpsertBulk) Update(set func(*IncidentSeverityUpsert)) *IncidentSeverityUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentSeverityUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentSeverityUpsertBulk) SetArchiveTime(v time.Time) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentSeverityUpsertBulk) UpdateArchiveTime() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentSeverityUpsertBulk) ClearArchiveTime() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearArchiveTime()
	})
}

// SetProviderID sets the "provider_id" field.
func (u *IncidentSeverityUpsertBulk) SetProviderID(v string) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetProviderID(v)
	})
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *IncidentSeverityUpsertBulk) UpdateProviderID() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateProviderID()
	})
}

// ClearProviderID clears the value of the "provider_id" field.
func (u *IncidentSeverityUpsertBulk) ClearProviderID() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearProviderID()
	})
}

// SetName sets the "name" field.
func (u *IncidentSeverityUpsertBulk) SetName(v string) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *IncidentSeverityUpsertBulk) UpdateName() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateName()
	})
}

// SetRank sets the "rank" field.
func (u *IncidentSeverityUpsertBulk) SetRank(v int) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetRank(v)
	})
}

// AddRank adds v to the "rank" field.
func (u *IncidentSeverityUpsertBulk) AddRank(v int) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.AddRank(v)
	})
}

// UpdateRank sets the "rank" field to the value that was provided on create.
func (u *IncidentSeverityUpsertBulk) UpdateRank() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateRank()
	})
}

// SetColor sets the "color" field.
func (u *IncidentSeverityUpsertBulk) SetColor(v string) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetColor(v)
	})
}

// UpdateColor sets the "color" field to the value that was provided on create.
func (u *IncidentSeverityUpsertBulk) UpdateColor() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateColor()
	})
}

// ClearColor clears the value of the "color" field.
func (u *IncidentSeverityUpsertBulk) ClearColor() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearColor()
	})
}

// SetDescription sets the "description" field.
func (u *IncidentSeverityUpsertBulk) SetDescription(v string) *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *IncidentSeverityUpsertBulk) UpdateDescription() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *IncidentSeverityUpsertBulk) ClearDescription() *IncidentSeverityUpsertBulk {
	return u.Update(func(s *IncidentSeverityUpsert) {
		s.ClearDescription()
	})
}

// Exec executes the query.
func (u *IncidentSeverityUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentSeverityCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentSeverityCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentSeverityUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

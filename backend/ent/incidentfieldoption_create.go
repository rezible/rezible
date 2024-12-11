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
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentfield"
	"github.com/twohundreds/rezible/ent/incidentfieldoption"
)

// IncidentFieldOptionCreate is the builder for creating a IncidentFieldOption entity.
type IncidentFieldOptionCreate struct {
	config
	mutation *IncidentFieldOptionMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetArchiveTime sets the "archive_time" field.
func (ifoc *IncidentFieldOptionCreate) SetArchiveTime(t time.Time) *IncidentFieldOptionCreate {
	ifoc.mutation.SetArchiveTime(t)
	return ifoc
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (ifoc *IncidentFieldOptionCreate) SetNillableArchiveTime(t *time.Time) *IncidentFieldOptionCreate {
	if t != nil {
		ifoc.SetArchiveTime(*t)
	}
	return ifoc
}

// SetIncidentFieldID sets the "incident_field_id" field.
func (ifoc *IncidentFieldOptionCreate) SetIncidentFieldID(u uuid.UUID) *IncidentFieldOptionCreate {
	ifoc.mutation.SetIncidentFieldID(u)
	return ifoc
}

// SetType sets the "type" field.
func (ifoc *IncidentFieldOptionCreate) SetType(i incidentfieldoption.Type) *IncidentFieldOptionCreate {
	ifoc.mutation.SetType(i)
	return ifoc
}

// SetValue sets the "value" field.
func (ifoc *IncidentFieldOptionCreate) SetValue(s string) *IncidentFieldOptionCreate {
	ifoc.mutation.SetValue(s)
	return ifoc
}

// SetID sets the "id" field.
func (ifoc *IncidentFieldOptionCreate) SetID(u uuid.UUID) *IncidentFieldOptionCreate {
	ifoc.mutation.SetID(u)
	return ifoc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ifoc *IncidentFieldOptionCreate) SetNillableID(u *uuid.UUID) *IncidentFieldOptionCreate {
	if u != nil {
		ifoc.SetID(*u)
	}
	return ifoc
}

// SetIncidentField sets the "incident_field" edge to the IncidentField entity.
func (ifoc *IncidentFieldOptionCreate) SetIncidentField(i *IncidentField) *IncidentFieldOptionCreate {
	return ifoc.SetIncidentFieldID(i.ID)
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (ifoc *IncidentFieldOptionCreate) AddIncidentIDs(ids ...uuid.UUID) *IncidentFieldOptionCreate {
	ifoc.mutation.AddIncidentIDs(ids...)
	return ifoc
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (ifoc *IncidentFieldOptionCreate) AddIncidents(i ...*Incident) *IncidentFieldOptionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ifoc.AddIncidentIDs(ids...)
}

// Mutation returns the IncidentFieldOptionMutation object of the builder.
func (ifoc *IncidentFieldOptionCreate) Mutation() *IncidentFieldOptionMutation {
	return ifoc.mutation
}

// Save creates the IncidentFieldOption in the database.
func (ifoc *IncidentFieldOptionCreate) Save(ctx context.Context) (*IncidentFieldOption, error) {
	if err := ifoc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, ifoc.sqlSave, ifoc.mutation, ifoc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ifoc *IncidentFieldOptionCreate) SaveX(ctx context.Context) *IncidentFieldOption {
	v, err := ifoc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ifoc *IncidentFieldOptionCreate) Exec(ctx context.Context) error {
	_, err := ifoc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ifoc *IncidentFieldOptionCreate) ExecX(ctx context.Context) {
	if err := ifoc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ifoc *IncidentFieldOptionCreate) defaults() error {
	if _, ok := ifoc.mutation.ID(); !ok {
		if incidentfieldoption.DefaultID == nil {
			return fmt.Errorf("ent: uninitialized incidentfieldoption.DefaultID (forgotten import ent/runtime?)")
		}
		v := incidentfieldoption.DefaultID()
		ifoc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (ifoc *IncidentFieldOptionCreate) check() error {
	if _, ok := ifoc.mutation.IncidentFieldID(); !ok {
		return &ValidationError{Name: "incident_field_id", err: errors.New(`ent: missing required field "IncidentFieldOption.incident_field_id"`)}
	}
	if _, ok := ifoc.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "IncidentFieldOption.type"`)}
	}
	if v, ok := ifoc.mutation.GetType(); ok {
		if err := incidentfieldoption.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "IncidentFieldOption.type": %w`, err)}
		}
	}
	if _, ok := ifoc.mutation.Value(); !ok {
		return &ValidationError{Name: "value", err: errors.New(`ent: missing required field "IncidentFieldOption.value"`)}
	}
	if len(ifoc.mutation.IncidentFieldIDs()) == 0 {
		return &ValidationError{Name: "incident_field", err: errors.New(`ent: missing required edge "IncidentFieldOption.incident_field"`)}
	}
	return nil
}

func (ifoc *IncidentFieldOptionCreate) sqlSave(ctx context.Context) (*IncidentFieldOption, error) {
	if err := ifoc.check(); err != nil {
		return nil, err
	}
	_node, _spec := ifoc.createSpec()
	if err := sqlgraph.CreateNode(ctx, ifoc.driver, _spec); err != nil {
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
	ifoc.mutation.id = &_node.ID
	ifoc.mutation.done = true
	return _node, nil
}

func (ifoc *IncidentFieldOptionCreate) createSpec() (*IncidentFieldOption, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentFieldOption{config: ifoc.config}
		_spec = sqlgraph.NewCreateSpec(incidentfieldoption.Table, sqlgraph.NewFieldSpec(incidentfieldoption.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = ifoc.conflict
	if id, ok := ifoc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := ifoc.mutation.ArchiveTime(); ok {
		_spec.SetField(incidentfieldoption.FieldArchiveTime, field.TypeTime, value)
		_node.ArchiveTime = value
	}
	if value, ok := ifoc.mutation.GetType(); ok {
		_spec.SetField(incidentfieldoption.FieldType, field.TypeEnum, value)
		_node.Type = value
	}
	if value, ok := ifoc.mutation.Value(); ok {
		_spec.SetField(incidentfieldoption.FieldValue, field.TypeString, value)
		_node.Value = value
	}
	if nodes := ifoc.mutation.IncidentFieldIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentfieldoption.IncidentFieldTable,
			Columns: []string{incidentfieldoption.IncidentFieldColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentfield.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.IncidentFieldID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ifoc.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentfieldoption.IncidentsTable,
			Columns: incidentfieldoption.IncidentsPrimaryKey,
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
//	client.IncidentFieldOption.Create().
//		SetArchiveTime(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentFieldOptionUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (ifoc *IncidentFieldOptionCreate) OnConflict(opts ...sql.ConflictOption) *IncidentFieldOptionUpsertOne {
	ifoc.conflict = opts
	return &IncidentFieldOptionUpsertOne{
		create: ifoc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentFieldOption.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ifoc *IncidentFieldOptionCreate) OnConflictColumns(columns ...string) *IncidentFieldOptionUpsertOne {
	ifoc.conflict = append(ifoc.conflict, sql.ConflictColumns(columns...))
	return &IncidentFieldOptionUpsertOne{
		create: ifoc,
	}
}

type (
	// IncidentFieldOptionUpsertOne is the builder for "upsert"-ing
	//  one IncidentFieldOption node.
	IncidentFieldOptionUpsertOne struct {
		create *IncidentFieldOptionCreate
	}

	// IncidentFieldOptionUpsert is the "OnConflict" setter.
	IncidentFieldOptionUpsert struct {
		*sql.UpdateSet
	}
)

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentFieldOptionUpsert) SetArchiveTime(v time.Time) *IncidentFieldOptionUpsert {
	u.Set(incidentfieldoption.FieldArchiveTime, v)
	return u
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsert) UpdateArchiveTime() *IncidentFieldOptionUpsert {
	u.SetExcluded(incidentfieldoption.FieldArchiveTime)
	return u
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentFieldOptionUpsert) ClearArchiveTime() *IncidentFieldOptionUpsert {
	u.SetNull(incidentfieldoption.FieldArchiveTime)
	return u
}

// SetIncidentFieldID sets the "incident_field_id" field.
func (u *IncidentFieldOptionUpsert) SetIncidentFieldID(v uuid.UUID) *IncidentFieldOptionUpsert {
	u.Set(incidentfieldoption.FieldIncidentFieldID, v)
	return u
}

// UpdateIncidentFieldID sets the "incident_field_id" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsert) UpdateIncidentFieldID() *IncidentFieldOptionUpsert {
	u.SetExcluded(incidentfieldoption.FieldIncidentFieldID)
	return u
}

// SetType sets the "type" field.
func (u *IncidentFieldOptionUpsert) SetType(v incidentfieldoption.Type) *IncidentFieldOptionUpsert {
	u.Set(incidentfieldoption.FieldType, v)
	return u
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsert) UpdateType() *IncidentFieldOptionUpsert {
	u.SetExcluded(incidentfieldoption.FieldType)
	return u
}

// SetValue sets the "value" field.
func (u *IncidentFieldOptionUpsert) SetValue(v string) *IncidentFieldOptionUpsert {
	u.Set(incidentfieldoption.FieldValue, v)
	return u
}

// UpdateValue sets the "value" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsert) UpdateValue() *IncidentFieldOptionUpsert {
	u.SetExcluded(incidentfieldoption.FieldValue)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentFieldOption.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentfieldoption.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentFieldOptionUpsertOne) UpdateNewValues() *IncidentFieldOptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidentfieldoption.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentFieldOption.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentFieldOptionUpsertOne) Ignore() *IncidentFieldOptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentFieldOptionUpsertOne) DoNothing() *IncidentFieldOptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentFieldOptionCreate.OnConflict
// documentation for more info.
func (u *IncidentFieldOptionUpsertOne) Update(set func(*IncidentFieldOptionUpsert)) *IncidentFieldOptionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentFieldOptionUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentFieldOptionUpsertOne) SetArchiveTime(v time.Time) *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertOne) UpdateArchiveTime() *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentFieldOptionUpsertOne) ClearArchiveTime() *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.ClearArchiveTime()
	})
}

// SetIncidentFieldID sets the "incident_field_id" field.
func (u *IncidentFieldOptionUpsertOne) SetIncidentFieldID(v uuid.UUID) *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetIncidentFieldID(v)
	})
}

// UpdateIncidentFieldID sets the "incident_field_id" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertOne) UpdateIncidentFieldID() *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateIncidentFieldID()
	})
}

// SetType sets the "type" field.
func (u *IncidentFieldOptionUpsertOne) SetType(v incidentfieldoption.Type) *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertOne) UpdateType() *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateType()
	})
}

// SetValue sets the "value" field.
func (u *IncidentFieldOptionUpsertOne) SetValue(v string) *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetValue(v)
	})
}

// UpdateValue sets the "value" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertOne) UpdateValue() *IncidentFieldOptionUpsertOne {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateValue()
	})
}

// Exec executes the query.
func (u *IncidentFieldOptionUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentFieldOptionCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentFieldOptionUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentFieldOptionUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentFieldOptionUpsertOne.ID is not supported by MySQL driver. Use IncidentFieldOptionUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentFieldOptionUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentFieldOptionCreateBulk is the builder for creating many IncidentFieldOption entities in bulk.
type IncidentFieldOptionCreateBulk struct {
	config
	err      error
	builders []*IncidentFieldOptionCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentFieldOption entities in the database.
func (ifocb *IncidentFieldOptionCreateBulk) Save(ctx context.Context) ([]*IncidentFieldOption, error) {
	if ifocb.err != nil {
		return nil, ifocb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ifocb.builders))
	nodes := make([]*IncidentFieldOption, len(ifocb.builders))
	mutators := make([]Mutator, len(ifocb.builders))
	for i := range ifocb.builders {
		func(i int, root context.Context) {
			builder := ifocb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentFieldOptionMutation)
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
					_, err = mutators[i+1].Mutate(root, ifocb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ifocb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ifocb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ifocb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ifocb *IncidentFieldOptionCreateBulk) SaveX(ctx context.Context) []*IncidentFieldOption {
	v, err := ifocb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ifocb *IncidentFieldOptionCreateBulk) Exec(ctx context.Context) error {
	_, err := ifocb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ifocb *IncidentFieldOptionCreateBulk) ExecX(ctx context.Context) {
	if err := ifocb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentFieldOption.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentFieldOptionUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (ifocb *IncidentFieldOptionCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentFieldOptionUpsertBulk {
	ifocb.conflict = opts
	return &IncidentFieldOptionUpsertBulk{
		create: ifocb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentFieldOption.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ifocb *IncidentFieldOptionCreateBulk) OnConflictColumns(columns ...string) *IncidentFieldOptionUpsertBulk {
	ifocb.conflict = append(ifocb.conflict, sql.ConflictColumns(columns...))
	return &IncidentFieldOptionUpsertBulk{
		create: ifocb,
	}
}

// IncidentFieldOptionUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentFieldOption nodes.
type IncidentFieldOptionUpsertBulk struct {
	create *IncidentFieldOptionCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentFieldOption.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentfieldoption.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentFieldOptionUpsertBulk) UpdateNewValues() *IncidentFieldOptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidentfieldoption.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentFieldOption.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentFieldOptionUpsertBulk) Ignore() *IncidentFieldOptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentFieldOptionUpsertBulk) DoNothing() *IncidentFieldOptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentFieldOptionCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentFieldOptionUpsertBulk) Update(set func(*IncidentFieldOptionUpsert)) *IncidentFieldOptionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentFieldOptionUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentFieldOptionUpsertBulk) SetArchiveTime(v time.Time) *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertBulk) UpdateArchiveTime() *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentFieldOptionUpsertBulk) ClearArchiveTime() *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.ClearArchiveTime()
	})
}

// SetIncidentFieldID sets the "incident_field_id" field.
func (u *IncidentFieldOptionUpsertBulk) SetIncidentFieldID(v uuid.UUID) *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetIncidentFieldID(v)
	})
}

// UpdateIncidentFieldID sets the "incident_field_id" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertBulk) UpdateIncidentFieldID() *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateIncidentFieldID()
	})
}

// SetType sets the "type" field.
func (u *IncidentFieldOptionUpsertBulk) SetType(v incidentfieldoption.Type) *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertBulk) UpdateType() *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateType()
	})
}

// SetValue sets the "value" field.
func (u *IncidentFieldOptionUpsertBulk) SetValue(v string) *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.SetValue(v)
	})
}

// UpdateValue sets the "value" field to the value that was provided on create.
func (u *IncidentFieldOptionUpsertBulk) UpdateValue() *IncidentFieldOptionUpsertBulk {
	return u.Update(func(s *IncidentFieldOptionUpsert) {
		s.UpdateValue()
	})
}

// Exec executes the query.
func (u *IncidentFieldOptionUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentFieldOptionCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentFieldOptionCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentFieldOptionUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
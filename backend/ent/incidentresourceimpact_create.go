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
	"github.com/twohundreds/rezible/ent/functionality"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentlink"
	"github.com/twohundreds/rezible/ent/incidentresourceimpact"
	"github.com/twohundreds/rezible/ent/service"
)

// IncidentResourceImpactCreate is the builder for creating a IncidentResourceImpact entity.
type IncidentResourceImpactCreate struct {
	config
	mutation *IncidentResourceImpactMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetIncidentID sets the "incident_id" field.
func (iric *IncidentResourceImpactCreate) SetIncidentID(u uuid.UUID) *IncidentResourceImpactCreate {
	iric.mutation.SetIncidentID(u)
	return iric
}

// SetServiceID sets the "service_id" field.
func (iric *IncidentResourceImpactCreate) SetServiceID(u uuid.UUID) *IncidentResourceImpactCreate {
	iric.mutation.SetServiceID(u)
	return iric
}

// SetNillableServiceID sets the "service_id" field if the given value is not nil.
func (iric *IncidentResourceImpactCreate) SetNillableServiceID(u *uuid.UUID) *IncidentResourceImpactCreate {
	if u != nil {
		iric.SetServiceID(*u)
	}
	return iric
}

// SetFunctionalityID sets the "functionality_id" field.
func (iric *IncidentResourceImpactCreate) SetFunctionalityID(u uuid.UUID) *IncidentResourceImpactCreate {
	iric.mutation.SetFunctionalityID(u)
	return iric
}

// SetNillableFunctionalityID sets the "functionality_id" field if the given value is not nil.
func (iric *IncidentResourceImpactCreate) SetNillableFunctionalityID(u *uuid.UUID) *IncidentResourceImpactCreate {
	if u != nil {
		iric.SetFunctionalityID(*u)
	}
	return iric
}

// SetID sets the "id" field.
func (iric *IncidentResourceImpactCreate) SetID(u uuid.UUID) *IncidentResourceImpactCreate {
	iric.mutation.SetID(u)
	return iric
}

// SetNillableID sets the "id" field if the given value is not nil.
func (iric *IncidentResourceImpactCreate) SetNillableID(u *uuid.UUID) *IncidentResourceImpactCreate {
	if u != nil {
		iric.SetID(*u)
	}
	return iric
}

// SetIncident sets the "incident" edge to the Incident entity.
func (iric *IncidentResourceImpactCreate) SetIncident(i *Incident) *IncidentResourceImpactCreate {
	return iric.SetIncidentID(i.ID)
}

// SetService sets the "service" edge to the Service entity.
func (iric *IncidentResourceImpactCreate) SetService(s *Service) *IncidentResourceImpactCreate {
	return iric.SetServiceID(s.ID)
}

// SetFunctionality sets the "functionality" edge to the Functionality entity.
func (iric *IncidentResourceImpactCreate) SetFunctionality(f *Functionality) *IncidentResourceImpactCreate {
	return iric.SetFunctionalityID(f.ID)
}

// AddResultingIncidentIDs adds the "resulting_incidents" edge to the IncidentLink entity by IDs.
func (iric *IncidentResourceImpactCreate) AddResultingIncidentIDs(ids ...int) *IncidentResourceImpactCreate {
	iric.mutation.AddResultingIncidentIDs(ids...)
	return iric
}

// AddResultingIncidents adds the "resulting_incidents" edges to the IncidentLink entity.
func (iric *IncidentResourceImpactCreate) AddResultingIncidents(i ...*IncidentLink) *IncidentResourceImpactCreate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iric.AddResultingIncidentIDs(ids...)
}

// Mutation returns the IncidentResourceImpactMutation object of the builder.
func (iric *IncidentResourceImpactCreate) Mutation() *IncidentResourceImpactMutation {
	return iric.mutation
}

// Save creates the IncidentResourceImpact in the database.
func (iric *IncidentResourceImpactCreate) Save(ctx context.Context) (*IncidentResourceImpact, error) {
	iric.defaults()
	return withHooks(ctx, iric.sqlSave, iric.mutation, iric.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (iric *IncidentResourceImpactCreate) SaveX(ctx context.Context) *IncidentResourceImpact {
	v, err := iric.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (iric *IncidentResourceImpactCreate) Exec(ctx context.Context) error {
	_, err := iric.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iric *IncidentResourceImpactCreate) ExecX(ctx context.Context) {
	if err := iric.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (iric *IncidentResourceImpactCreate) defaults() {
	if _, ok := iric.mutation.ID(); !ok {
		v := incidentresourceimpact.DefaultID()
		iric.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iric *IncidentResourceImpactCreate) check() error {
	if _, ok := iric.mutation.IncidentID(); !ok {
		return &ValidationError{Name: "incident_id", err: errors.New(`ent: missing required field "IncidentResourceImpact.incident_id"`)}
	}
	if len(iric.mutation.IncidentIDs()) == 0 {
		return &ValidationError{Name: "incident", err: errors.New(`ent: missing required edge "IncidentResourceImpact.incident"`)}
	}
	return nil
}

func (iric *IncidentResourceImpactCreate) sqlSave(ctx context.Context) (*IncidentResourceImpact, error) {
	if err := iric.check(); err != nil {
		return nil, err
	}
	_node, _spec := iric.createSpec()
	if err := sqlgraph.CreateNode(ctx, iric.driver, _spec); err != nil {
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
	iric.mutation.id = &_node.ID
	iric.mutation.done = true
	return _node, nil
}

func (iric *IncidentResourceImpactCreate) createSpec() (*IncidentResourceImpact, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentResourceImpact{config: iric.config}
		_spec = sqlgraph.NewCreateSpec(incidentresourceimpact.Table, sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = iric.conflict
	if id, ok := iric.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if nodes := iric.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.IncidentTable,
			Columns: []string{incidentresourceimpact.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.IncidentID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := iric.mutation.ServiceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.ServiceTable,
			Columns: []string{incidentresourceimpact.ServiceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ServiceID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := iric.mutation.FunctionalityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.FunctionalityTable,
			Columns: []string{incidentresourceimpact.FunctionalityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.FunctionalityID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := iric.mutation.ResultingIncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
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
//	client.IncidentResourceImpact.Create().
//		SetIncidentID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentResourceImpactUpsert) {
//			SetIncidentID(v+v).
//		}).
//		Exec(ctx)
func (iric *IncidentResourceImpactCreate) OnConflict(opts ...sql.ConflictOption) *IncidentResourceImpactUpsertOne {
	iric.conflict = opts
	return &IncidentResourceImpactUpsertOne{
		create: iric,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentResourceImpact.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (iric *IncidentResourceImpactCreate) OnConflictColumns(columns ...string) *IncidentResourceImpactUpsertOne {
	iric.conflict = append(iric.conflict, sql.ConflictColumns(columns...))
	return &IncidentResourceImpactUpsertOne{
		create: iric,
	}
}

type (
	// IncidentResourceImpactUpsertOne is the builder for "upsert"-ing
	//  one IncidentResourceImpact node.
	IncidentResourceImpactUpsertOne struct {
		create *IncidentResourceImpactCreate
	}

	// IncidentResourceImpactUpsert is the "OnConflict" setter.
	IncidentResourceImpactUpsert struct {
		*sql.UpdateSet
	}
)

// SetIncidentID sets the "incident_id" field.
func (u *IncidentResourceImpactUpsert) SetIncidentID(v uuid.UUID) *IncidentResourceImpactUpsert {
	u.Set(incidentresourceimpact.FieldIncidentID, v)
	return u
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsert) UpdateIncidentID() *IncidentResourceImpactUpsert {
	u.SetExcluded(incidentresourceimpact.FieldIncidentID)
	return u
}

// SetServiceID sets the "service_id" field.
func (u *IncidentResourceImpactUpsert) SetServiceID(v uuid.UUID) *IncidentResourceImpactUpsert {
	u.Set(incidentresourceimpact.FieldServiceID, v)
	return u
}

// UpdateServiceID sets the "service_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsert) UpdateServiceID() *IncidentResourceImpactUpsert {
	u.SetExcluded(incidentresourceimpact.FieldServiceID)
	return u
}

// ClearServiceID clears the value of the "service_id" field.
func (u *IncidentResourceImpactUpsert) ClearServiceID() *IncidentResourceImpactUpsert {
	u.SetNull(incidentresourceimpact.FieldServiceID)
	return u
}

// SetFunctionalityID sets the "functionality_id" field.
func (u *IncidentResourceImpactUpsert) SetFunctionalityID(v uuid.UUID) *IncidentResourceImpactUpsert {
	u.Set(incidentresourceimpact.FieldFunctionalityID, v)
	return u
}

// UpdateFunctionalityID sets the "functionality_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsert) UpdateFunctionalityID() *IncidentResourceImpactUpsert {
	u.SetExcluded(incidentresourceimpact.FieldFunctionalityID)
	return u
}

// ClearFunctionalityID clears the value of the "functionality_id" field.
func (u *IncidentResourceImpactUpsert) ClearFunctionalityID() *IncidentResourceImpactUpsert {
	u.SetNull(incidentresourceimpact.FieldFunctionalityID)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentResourceImpact.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentresourceimpact.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentResourceImpactUpsertOne) UpdateNewValues() *IncidentResourceImpactUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidentresourceimpact.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentResourceImpact.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentResourceImpactUpsertOne) Ignore() *IncidentResourceImpactUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentResourceImpactUpsertOne) DoNothing() *IncidentResourceImpactUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentResourceImpactCreate.OnConflict
// documentation for more info.
func (u *IncidentResourceImpactUpsertOne) Update(set func(*IncidentResourceImpactUpsert)) *IncidentResourceImpactUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentResourceImpactUpsert{UpdateSet: update})
	}))
	return u
}

// SetIncidentID sets the "incident_id" field.
func (u *IncidentResourceImpactUpsertOne) SetIncidentID(v uuid.UUID) *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.SetIncidentID(v)
	})
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsertOne) UpdateIncidentID() *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.UpdateIncidentID()
	})
}

// SetServiceID sets the "service_id" field.
func (u *IncidentResourceImpactUpsertOne) SetServiceID(v uuid.UUID) *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.SetServiceID(v)
	})
}

// UpdateServiceID sets the "service_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsertOne) UpdateServiceID() *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.UpdateServiceID()
	})
}

// ClearServiceID clears the value of the "service_id" field.
func (u *IncidentResourceImpactUpsertOne) ClearServiceID() *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.ClearServiceID()
	})
}

// SetFunctionalityID sets the "functionality_id" field.
func (u *IncidentResourceImpactUpsertOne) SetFunctionalityID(v uuid.UUID) *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.SetFunctionalityID(v)
	})
}

// UpdateFunctionalityID sets the "functionality_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsertOne) UpdateFunctionalityID() *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.UpdateFunctionalityID()
	})
}

// ClearFunctionalityID clears the value of the "functionality_id" field.
func (u *IncidentResourceImpactUpsertOne) ClearFunctionalityID() *IncidentResourceImpactUpsertOne {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.ClearFunctionalityID()
	})
}

// Exec executes the query.
func (u *IncidentResourceImpactUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentResourceImpactCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentResourceImpactUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentResourceImpactUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentResourceImpactUpsertOne.ID is not supported by MySQL driver. Use IncidentResourceImpactUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentResourceImpactUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentResourceImpactCreateBulk is the builder for creating many IncidentResourceImpact entities in bulk.
type IncidentResourceImpactCreateBulk struct {
	config
	err      error
	builders []*IncidentResourceImpactCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentResourceImpact entities in the database.
func (iricb *IncidentResourceImpactCreateBulk) Save(ctx context.Context) ([]*IncidentResourceImpact, error) {
	if iricb.err != nil {
		return nil, iricb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(iricb.builders))
	nodes := make([]*IncidentResourceImpact, len(iricb.builders))
	mutators := make([]Mutator, len(iricb.builders))
	for i := range iricb.builders {
		func(i int, root context.Context) {
			builder := iricb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentResourceImpactMutation)
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
					_, err = mutators[i+1].Mutate(root, iricb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = iricb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, iricb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, iricb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (iricb *IncidentResourceImpactCreateBulk) SaveX(ctx context.Context) []*IncidentResourceImpact {
	v, err := iricb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (iricb *IncidentResourceImpactCreateBulk) Exec(ctx context.Context) error {
	_, err := iricb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iricb *IncidentResourceImpactCreateBulk) ExecX(ctx context.Context) {
	if err := iricb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentResourceImpact.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentResourceImpactUpsert) {
//			SetIncidentID(v+v).
//		}).
//		Exec(ctx)
func (iricb *IncidentResourceImpactCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentResourceImpactUpsertBulk {
	iricb.conflict = opts
	return &IncidentResourceImpactUpsertBulk{
		create: iricb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentResourceImpact.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (iricb *IncidentResourceImpactCreateBulk) OnConflictColumns(columns ...string) *IncidentResourceImpactUpsertBulk {
	iricb.conflict = append(iricb.conflict, sql.ConflictColumns(columns...))
	return &IncidentResourceImpactUpsertBulk{
		create: iricb,
	}
}

// IncidentResourceImpactUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentResourceImpact nodes.
type IncidentResourceImpactUpsertBulk struct {
	create *IncidentResourceImpactCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentResourceImpact.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentresourceimpact.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentResourceImpactUpsertBulk) UpdateNewValues() *IncidentResourceImpactUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidentresourceimpact.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentResourceImpact.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentResourceImpactUpsertBulk) Ignore() *IncidentResourceImpactUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentResourceImpactUpsertBulk) DoNothing() *IncidentResourceImpactUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentResourceImpactCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentResourceImpactUpsertBulk) Update(set func(*IncidentResourceImpactUpsert)) *IncidentResourceImpactUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentResourceImpactUpsert{UpdateSet: update})
	}))
	return u
}

// SetIncidentID sets the "incident_id" field.
func (u *IncidentResourceImpactUpsertBulk) SetIncidentID(v uuid.UUID) *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.SetIncidentID(v)
	})
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsertBulk) UpdateIncidentID() *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.UpdateIncidentID()
	})
}

// SetServiceID sets the "service_id" field.
func (u *IncidentResourceImpactUpsertBulk) SetServiceID(v uuid.UUID) *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.SetServiceID(v)
	})
}

// UpdateServiceID sets the "service_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsertBulk) UpdateServiceID() *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.UpdateServiceID()
	})
}

// ClearServiceID clears the value of the "service_id" field.
func (u *IncidentResourceImpactUpsertBulk) ClearServiceID() *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.ClearServiceID()
	})
}

// SetFunctionalityID sets the "functionality_id" field.
func (u *IncidentResourceImpactUpsertBulk) SetFunctionalityID(v uuid.UUID) *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.SetFunctionalityID(v)
	})
}

// UpdateFunctionalityID sets the "functionality_id" field to the value that was provided on create.
func (u *IncidentResourceImpactUpsertBulk) UpdateFunctionalityID() *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.UpdateFunctionalityID()
	})
}

// ClearFunctionalityID clears the value of the "functionality_id" field.
func (u *IncidentResourceImpactUpsertBulk) ClearFunctionalityID() *IncidentResourceImpactUpsertBulk {
	return u.Update(func(s *IncidentResourceImpactUpsert) {
		s.ClearFunctionalityID()
	})
}

// Exec executes the query.
func (u *IncidentResourceImpactUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentResourceImpactCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentResourceImpactCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentResourceImpactUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

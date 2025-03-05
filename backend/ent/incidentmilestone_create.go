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
	"github.com/rezible/rezible/ent/incidentmilestone"
)

// IncidentMilestoneCreate is the builder for creating a IncidentMilestone entity.
type IncidentMilestoneCreate struct {
	config
	mutation *IncidentMilestoneMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetIncidentID sets the "incident_id" field.
func (imc *IncidentMilestoneCreate) SetIncidentID(u uuid.UUID) *IncidentMilestoneCreate {
	imc.mutation.SetIncidentID(u)
	return imc
}

// SetKind sets the "kind" field.
func (imc *IncidentMilestoneCreate) SetKind(i incidentmilestone.Kind) *IncidentMilestoneCreate {
	imc.mutation.SetKind(i)
	return imc
}

// SetDescription sets the "description" field.
func (imc *IncidentMilestoneCreate) SetDescription(s string) *IncidentMilestoneCreate {
	imc.mutation.SetDescription(s)
	return imc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (imc *IncidentMilestoneCreate) SetNillableDescription(s *string) *IncidentMilestoneCreate {
	if s != nil {
		imc.SetDescription(*s)
	}
	return imc
}

// SetTime sets the "time" field.
func (imc *IncidentMilestoneCreate) SetTime(t time.Time) *IncidentMilestoneCreate {
	imc.mutation.SetTime(t)
	return imc
}

// SetID sets the "id" field.
func (imc *IncidentMilestoneCreate) SetID(u uuid.UUID) *IncidentMilestoneCreate {
	imc.mutation.SetID(u)
	return imc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (imc *IncidentMilestoneCreate) SetNillableID(u *uuid.UUID) *IncidentMilestoneCreate {
	if u != nil {
		imc.SetID(*u)
	}
	return imc
}

// SetIncident sets the "incident" edge to the Incident entity.
func (imc *IncidentMilestoneCreate) SetIncident(i *Incident) *IncidentMilestoneCreate {
	return imc.SetIncidentID(i.ID)
}

// Mutation returns the IncidentMilestoneMutation object of the builder.
func (imc *IncidentMilestoneCreate) Mutation() *IncidentMilestoneMutation {
	return imc.mutation
}

// Save creates the IncidentMilestone in the database.
func (imc *IncidentMilestoneCreate) Save(ctx context.Context) (*IncidentMilestone, error) {
	imc.defaults()
	return withHooks(ctx, imc.sqlSave, imc.mutation, imc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (imc *IncidentMilestoneCreate) SaveX(ctx context.Context) *IncidentMilestone {
	v, err := imc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (imc *IncidentMilestoneCreate) Exec(ctx context.Context) error {
	_, err := imc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (imc *IncidentMilestoneCreate) ExecX(ctx context.Context) {
	if err := imc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (imc *IncidentMilestoneCreate) defaults() {
	if _, ok := imc.mutation.ID(); !ok {
		v := incidentmilestone.DefaultID()
		imc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (imc *IncidentMilestoneCreate) check() error {
	if _, ok := imc.mutation.IncidentID(); !ok {
		return &ValidationError{Name: "incident_id", err: errors.New(`ent: missing required field "IncidentMilestone.incident_id"`)}
	}
	if _, ok := imc.mutation.Kind(); !ok {
		return &ValidationError{Name: "kind", err: errors.New(`ent: missing required field "IncidentMilestone.kind"`)}
	}
	if v, ok := imc.mutation.Kind(); ok {
		if err := incidentmilestone.KindValidator(v); err != nil {
			return &ValidationError{Name: "kind", err: fmt.Errorf(`ent: validator failed for field "IncidentMilestone.kind": %w`, err)}
		}
	}
	if _, ok := imc.mutation.Time(); !ok {
		return &ValidationError{Name: "time", err: errors.New(`ent: missing required field "IncidentMilestone.time"`)}
	}
	if len(imc.mutation.IncidentIDs()) == 0 {
		return &ValidationError{Name: "incident", err: errors.New(`ent: missing required edge "IncidentMilestone.incident"`)}
	}
	return nil
}

func (imc *IncidentMilestoneCreate) sqlSave(ctx context.Context) (*IncidentMilestone, error) {
	if err := imc.check(); err != nil {
		return nil, err
	}
	_node, _spec := imc.createSpec()
	if err := sqlgraph.CreateNode(ctx, imc.driver, _spec); err != nil {
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
	imc.mutation.id = &_node.ID
	imc.mutation.done = true
	return _node, nil
}

func (imc *IncidentMilestoneCreate) createSpec() (*IncidentMilestone, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentMilestone{config: imc.config}
		_spec = sqlgraph.NewCreateSpec(incidentmilestone.Table, sqlgraph.NewFieldSpec(incidentmilestone.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = imc.conflict
	if id, ok := imc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := imc.mutation.Kind(); ok {
		_spec.SetField(incidentmilestone.FieldKind, field.TypeEnum, value)
		_node.Kind = value
	}
	if value, ok := imc.mutation.Description(); ok {
		_spec.SetField(incidentmilestone.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := imc.mutation.Time(); ok {
		_spec.SetField(incidentmilestone.FieldTime, field.TypeTime, value)
		_node.Time = value
	}
	if nodes := imc.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentmilestone.IncidentTable,
			Columns: []string{incidentmilestone.IncidentColumn},
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
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentMilestone.Create().
//		SetIncidentID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentMilestoneUpsert) {
//			SetIncidentID(v+v).
//		}).
//		Exec(ctx)
func (imc *IncidentMilestoneCreate) OnConflict(opts ...sql.ConflictOption) *IncidentMilestoneUpsertOne {
	imc.conflict = opts
	return &IncidentMilestoneUpsertOne{
		create: imc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentMilestone.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (imc *IncidentMilestoneCreate) OnConflictColumns(columns ...string) *IncidentMilestoneUpsertOne {
	imc.conflict = append(imc.conflict, sql.ConflictColumns(columns...))
	return &IncidentMilestoneUpsertOne{
		create: imc,
	}
}

type (
	// IncidentMilestoneUpsertOne is the builder for "upsert"-ing
	//  one IncidentMilestone node.
	IncidentMilestoneUpsertOne struct {
		create *IncidentMilestoneCreate
	}

	// IncidentMilestoneUpsert is the "OnConflict" setter.
	IncidentMilestoneUpsert struct {
		*sql.UpdateSet
	}
)

// SetIncidentID sets the "incident_id" field.
func (u *IncidentMilestoneUpsert) SetIncidentID(v uuid.UUID) *IncidentMilestoneUpsert {
	u.Set(incidentmilestone.FieldIncidentID, v)
	return u
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *IncidentMilestoneUpsert) UpdateIncidentID() *IncidentMilestoneUpsert {
	u.SetExcluded(incidentmilestone.FieldIncidentID)
	return u
}

// SetKind sets the "kind" field.
func (u *IncidentMilestoneUpsert) SetKind(v incidentmilestone.Kind) *IncidentMilestoneUpsert {
	u.Set(incidentmilestone.FieldKind, v)
	return u
}

// UpdateKind sets the "kind" field to the value that was provided on create.
func (u *IncidentMilestoneUpsert) UpdateKind() *IncidentMilestoneUpsert {
	u.SetExcluded(incidentmilestone.FieldKind)
	return u
}

// SetDescription sets the "description" field.
func (u *IncidentMilestoneUpsert) SetDescription(v string) *IncidentMilestoneUpsert {
	u.Set(incidentmilestone.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *IncidentMilestoneUpsert) UpdateDescription() *IncidentMilestoneUpsert {
	u.SetExcluded(incidentmilestone.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *IncidentMilestoneUpsert) ClearDescription() *IncidentMilestoneUpsert {
	u.SetNull(incidentmilestone.FieldDescription)
	return u
}

// SetTime sets the "time" field.
func (u *IncidentMilestoneUpsert) SetTime(v time.Time) *IncidentMilestoneUpsert {
	u.Set(incidentmilestone.FieldTime, v)
	return u
}

// UpdateTime sets the "time" field to the value that was provided on create.
func (u *IncidentMilestoneUpsert) UpdateTime() *IncidentMilestoneUpsert {
	u.SetExcluded(incidentmilestone.FieldTime)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentMilestone.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentmilestone.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentMilestoneUpsertOne) UpdateNewValues() *IncidentMilestoneUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidentmilestone.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentMilestone.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentMilestoneUpsertOne) Ignore() *IncidentMilestoneUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentMilestoneUpsertOne) DoNothing() *IncidentMilestoneUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentMilestoneCreate.OnConflict
// documentation for more info.
func (u *IncidentMilestoneUpsertOne) Update(set func(*IncidentMilestoneUpsert)) *IncidentMilestoneUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentMilestoneUpsert{UpdateSet: update})
	}))
	return u
}

// SetIncidentID sets the "incident_id" field.
func (u *IncidentMilestoneUpsertOne) SetIncidentID(v uuid.UUID) *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetIncidentID(v)
	})
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertOne) UpdateIncidentID() *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateIncidentID()
	})
}

// SetKind sets the "kind" field.
func (u *IncidentMilestoneUpsertOne) SetKind(v incidentmilestone.Kind) *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetKind(v)
	})
}

// UpdateKind sets the "kind" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertOne) UpdateKind() *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateKind()
	})
}

// SetDescription sets the "description" field.
func (u *IncidentMilestoneUpsertOne) SetDescription(v string) *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertOne) UpdateDescription() *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *IncidentMilestoneUpsertOne) ClearDescription() *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.ClearDescription()
	})
}

// SetTime sets the "time" field.
func (u *IncidentMilestoneUpsertOne) SetTime(v time.Time) *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetTime(v)
	})
}

// UpdateTime sets the "time" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertOne) UpdateTime() *IncidentMilestoneUpsertOne {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateTime()
	})
}

// Exec executes the query.
func (u *IncidentMilestoneUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentMilestoneCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentMilestoneUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentMilestoneUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentMilestoneUpsertOne.ID is not supported by MySQL driver. Use IncidentMilestoneUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentMilestoneUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentMilestoneCreateBulk is the builder for creating many IncidentMilestone entities in bulk.
type IncidentMilestoneCreateBulk struct {
	config
	err      error
	builders []*IncidentMilestoneCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentMilestone entities in the database.
func (imcb *IncidentMilestoneCreateBulk) Save(ctx context.Context) ([]*IncidentMilestone, error) {
	if imcb.err != nil {
		return nil, imcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(imcb.builders))
	nodes := make([]*IncidentMilestone, len(imcb.builders))
	mutators := make([]Mutator, len(imcb.builders))
	for i := range imcb.builders {
		func(i int, root context.Context) {
			builder := imcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentMilestoneMutation)
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
					_, err = mutators[i+1].Mutate(root, imcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = imcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, imcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, imcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (imcb *IncidentMilestoneCreateBulk) SaveX(ctx context.Context) []*IncidentMilestone {
	v, err := imcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (imcb *IncidentMilestoneCreateBulk) Exec(ctx context.Context) error {
	_, err := imcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (imcb *IncidentMilestoneCreateBulk) ExecX(ctx context.Context) {
	if err := imcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentMilestone.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentMilestoneUpsert) {
//			SetIncidentID(v+v).
//		}).
//		Exec(ctx)
func (imcb *IncidentMilestoneCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentMilestoneUpsertBulk {
	imcb.conflict = opts
	return &IncidentMilestoneUpsertBulk{
		create: imcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentMilestone.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (imcb *IncidentMilestoneCreateBulk) OnConflictColumns(columns ...string) *IncidentMilestoneUpsertBulk {
	imcb.conflict = append(imcb.conflict, sql.ConflictColumns(columns...))
	return &IncidentMilestoneUpsertBulk{
		create: imcb,
	}
}

// IncidentMilestoneUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentMilestone nodes.
type IncidentMilestoneUpsertBulk struct {
	create *IncidentMilestoneCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentMilestone.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentmilestone.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentMilestoneUpsertBulk) UpdateNewValues() *IncidentMilestoneUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidentmilestone.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentMilestone.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentMilestoneUpsertBulk) Ignore() *IncidentMilestoneUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentMilestoneUpsertBulk) DoNothing() *IncidentMilestoneUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentMilestoneCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentMilestoneUpsertBulk) Update(set func(*IncidentMilestoneUpsert)) *IncidentMilestoneUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentMilestoneUpsert{UpdateSet: update})
	}))
	return u
}

// SetIncidentID sets the "incident_id" field.
func (u *IncidentMilestoneUpsertBulk) SetIncidentID(v uuid.UUID) *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetIncidentID(v)
	})
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertBulk) UpdateIncidentID() *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateIncidentID()
	})
}

// SetKind sets the "kind" field.
func (u *IncidentMilestoneUpsertBulk) SetKind(v incidentmilestone.Kind) *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetKind(v)
	})
}

// UpdateKind sets the "kind" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertBulk) UpdateKind() *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateKind()
	})
}

// SetDescription sets the "description" field.
func (u *IncidentMilestoneUpsertBulk) SetDescription(v string) *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertBulk) UpdateDescription() *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *IncidentMilestoneUpsertBulk) ClearDescription() *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.ClearDescription()
	})
}

// SetTime sets the "time" field.
func (u *IncidentMilestoneUpsertBulk) SetTime(v time.Time) *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.SetTime(v)
	})
}

// UpdateTime sets the "time" field to the value that was provided on create.
func (u *IncidentMilestoneUpsertBulk) UpdateTime() *IncidentMilestoneUpsertBulk {
	return u.Update(func(s *IncidentMilestoneUpsert) {
		s.UpdateTime()
	})
}

// Exec executes the query.
func (u *IncidentMilestoneUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentMilestoneCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentMilestoneCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentMilestoneUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

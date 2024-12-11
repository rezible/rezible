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
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rezible/rezible/ent/oncallusershiftannotation"
	"github.com/rezible/rezible/ent/oncallusershiftcover"
	"github.com/rezible/rezible/ent/oncallusershifthandover"
	"github.com/rezible/rezible/ent/user"
)

// OncallUserShiftCreate is the builder for creating a OncallUserShift entity.
type OncallUserShiftCreate struct {
	config
	mutation *OncallUserShiftMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetUserID sets the "user_id" field.
func (ousc *OncallUserShiftCreate) SetUserID(u uuid.UUID) *OncallUserShiftCreate {
	ousc.mutation.SetUserID(u)
	return ousc
}

// SetRosterID sets the "roster_id" field.
func (ousc *OncallUserShiftCreate) SetRosterID(u uuid.UUID) *OncallUserShiftCreate {
	ousc.mutation.SetRosterID(u)
	return ousc
}

// SetStartAt sets the "start_at" field.
func (ousc *OncallUserShiftCreate) SetStartAt(t time.Time) *OncallUserShiftCreate {
	ousc.mutation.SetStartAt(t)
	return ousc
}

// SetEndAt sets the "end_at" field.
func (ousc *OncallUserShiftCreate) SetEndAt(t time.Time) *OncallUserShiftCreate {
	ousc.mutation.SetEndAt(t)
	return ousc
}

// SetProviderID sets the "provider_id" field.
func (ousc *OncallUserShiftCreate) SetProviderID(s string) *OncallUserShiftCreate {
	ousc.mutation.SetProviderID(s)
	return ousc
}

// SetNillableProviderID sets the "provider_id" field if the given value is not nil.
func (ousc *OncallUserShiftCreate) SetNillableProviderID(s *string) *OncallUserShiftCreate {
	if s != nil {
		ousc.SetProviderID(*s)
	}
	return ousc
}

// SetID sets the "id" field.
func (ousc *OncallUserShiftCreate) SetID(u uuid.UUID) *OncallUserShiftCreate {
	ousc.mutation.SetID(u)
	return ousc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ousc *OncallUserShiftCreate) SetNillableID(u *uuid.UUID) *OncallUserShiftCreate {
	if u != nil {
		ousc.SetID(*u)
	}
	return ousc
}

// SetUser sets the "user" edge to the User entity.
func (ousc *OncallUserShiftCreate) SetUser(u *User) *OncallUserShiftCreate {
	return ousc.SetUserID(u.ID)
}

// SetRoster sets the "roster" edge to the OncallRoster entity.
func (ousc *OncallUserShiftCreate) SetRoster(o *OncallRoster) *OncallUserShiftCreate {
	return ousc.SetRosterID(o.ID)
}

// AddCoverIDs adds the "covers" edge to the OncallUserShiftCover entity by IDs.
func (ousc *OncallUserShiftCreate) AddCoverIDs(ids ...uuid.UUID) *OncallUserShiftCreate {
	ousc.mutation.AddCoverIDs(ids...)
	return ousc
}

// AddCovers adds the "covers" edges to the OncallUserShiftCover entity.
func (ousc *OncallUserShiftCreate) AddCovers(o ...*OncallUserShiftCover) *OncallUserShiftCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return ousc.AddCoverIDs(ids...)
}

// AddAnnotationIDs adds the "annotations" edge to the OncallUserShiftAnnotation entity by IDs.
func (ousc *OncallUserShiftCreate) AddAnnotationIDs(ids ...uuid.UUID) *OncallUserShiftCreate {
	ousc.mutation.AddAnnotationIDs(ids...)
	return ousc
}

// AddAnnotations adds the "annotations" edges to the OncallUserShiftAnnotation entity.
func (ousc *OncallUserShiftCreate) AddAnnotations(o ...*OncallUserShiftAnnotation) *OncallUserShiftCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return ousc.AddAnnotationIDs(ids...)
}

// SetHandoverID sets the "handover" edge to the OncallUserShiftHandover entity by ID.
func (ousc *OncallUserShiftCreate) SetHandoverID(id uuid.UUID) *OncallUserShiftCreate {
	ousc.mutation.SetHandoverID(id)
	return ousc
}

// SetNillableHandoverID sets the "handover" edge to the OncallUserShiftHandover entity by ID if the given value is not nil.
func (ousc *OncallUserShiftCreate) SetNillableHandoverID(id *uuid.UUID) *OncallUserShiftCreate {
	if id != nil {
		ousc = ousc.SetHandoverID(*id)
	}
	return ousc
}

// SetHandover sets the "handover" edge to the OncallUserShiftHandover entity.
func (ousc *OncallUserShiftCreate) SetHandover(o *OncallUserShiftHandover) *OncallUserShiftCreate {
	return ousc.SetHandoverID(o.ID)
}

// Mutation returns the OncallUserShiftMutation object of the builder.
func (ousc *OncallUserShiftCreate) Mutation() *OncallUserShiftMutation {
	return ousc.mutation
}

// Save creates the OncallUserShift in the database.
func (ousc *OncallUserShiftCreate) Save(ctx context.Context) (*OncallUserShift, error) {
	ousc.defaults()
	return withHooks(ctx, ousc.sqlSave, ousc.mutation, ousc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ousc *OncallUserShiftCreate) SaveX(ctx context.Context) *OncallUserShift {
	v, err := ousc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ousc *OncallUserShiftCreate) Exec(ctx context.Context) error {
	_, err := ousc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ousc *OncallUserShiftCreate) ExecX(ctx context.Context) {
	if err := ousc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ousc *OncallUserShiftCreate) defaults() {
	if _, ok := ousc.mutation.ID(); !ok {
		v := oncallusershift.DefaultID()
		ousc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ousc *OncallUserShiftCreate) check() error {
	if _, ok := ousc.mutation.UserID(); !ok {
		return &ValidationError{Name: "user_id", err: errors.New(`ent: missing required field "OncallUserShift.user_id"`)}
	}
	if _, ok := ousc.mutation.RosterID(); !ok {
		return &ValidationError{Name: "roster_id", err: errors.New(`ent: missing required field "OncallUserShift.roster_id"`)}
	}
	if _, ok := ousc.mutation.StartAt(); !ok {
		return &ValidationError{Name: "start_at", err: errors.New(`ent: missing required field "OncallUserShift.start_at"`)}
	}
	if _, ok := ousc.mutation.EndAt(); !ok {
		return &ValidationError{Name: "end_at", err: errors.New(`ent: missing required field "OncallUserShift.end_at"`)}
	}
	if len(ousc.mutation.UserIDs()) == 0 {
		return &ValidationError{Name: "user", err: errors.New(`ent: missing required edge "OncallUserShift.user"`)}
	}
	if len(ousc.mutation.RosterIDs()) == 0 {
		return &ValidationError{Name: "roster", err: errors.New(`ent: missing required edge "OncallUserShift.roster"`)}
	}
	return nil
}

func (ousc *OncallUserShiftCreate) sqlSave(ctx context.Context) (*OncallUserShift, error) {
	if err := ousc.check(); err != nil {
		return nil, err
	}
	_node, _spec := ousc.createSpec()
	if err := sqlgraph.CreateNode(ctx, ousc.driver, _spec); err != nil {
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
	ousc.mutation.id = &_node.ID
	ousc.mutation.done = true
	return _node, nil
}

func (ousc *OncallUserShiftCreate) createSpec() (*OncallUserShift, *sqlgraph.CreateSpec) {
	var (
		_node = &OncallUserShift{config: ousc.config}
		_spec = sqlgraph.NewCreateSpec(oncallusershift.Table, sqlgraph.NewFieldSpec(oncallusershift.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = ousc.conflict
	if id, ok := ousc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := ousc.mutation.StartAt(); ok {
		_spec.SetField(oncallusershift.FieldStartAt, field.TypeTime, value)
		_node.StartAt = value
	}
	if value, ok := ousc.mutation.EndAt(); ok {
		_spec.SetField(oncallusershift.FieldEndAt, field.TypeTime, value)
		_node.EndAt = value
	}
	if value, ok := ousc.mutation.ProviderID(); ok {
		_spec.SetField(oncallusershift.FieldProviderID, field.TypeString, value)
		_node.ProviderID = value
	}
	if nodes := ousc.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   oncallusershift.UserTable,
			Columns: []string{oncallusershift.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.UserID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ousc.mutation.RosterIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   oncallusershift.RosterTable,
			Columns: []string{oncallusershift.RosterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallroster.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.RosterID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ousc.mutation.CoversIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallusershift.CoversTable,
			Columns: []string{oncallusershift.CoversColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallusershiftcover.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ousc.mutation.AnnotationsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallusershift.AnnotationsTable,
			Columns: []string{oncallusershift.AnnotationsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallusershiftannotation.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ousc.mutation.HandoverIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   oncallusershift.HandoverTable,
			Columns: []string{oncallusershift.HandoverColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallusershifthandover.FieldID, field.TypeUUID),
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
//	client.OncallUserShift.Create().
//		SetUserID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.OncallUserShiftUpsert) {
//			SetUserID(v+v).
//		}).
//		Exec(ctx)
func (ousc *OncallUserShiftCreate) OnConflict(opts ...sql.ConflictOption) *OncallUserShiftUpsertOne {
	ousc.conflict = opts
	return &OncallUserShiftUpsertOne{
		create: ousc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.OncallUserShift.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ousc *OncallUserShiftCreate) OnConflictColumns(columns ...string) *OncallUserShiftUpsertOne {
	ousc.conflict = append(ousc.conflict, sql.ConflictColumns(columns...))
	return &OncallUserShiftUpsertOne{
		create: ousc,
	}
}

type (
	// OncallUserShiftUpsertOne is the builder for "upsert"-ing
	//  one OncallUserShift node.
	OncallUserShiftUpsertOne struct {
		create *OncallUserShiftCreate
	}

	// OncallUserShiftUpsert is the "OnConflict" setter.
	OncallUserShiftUpsert struct {
		*sql.UpdateSet
	}
)

// SetUserID sets the "user_id" field.
func (u *OncallUserShiftUpsert) SetUserID(v uuid.UUID) *OncallUserShiftUpsert {
	u.Set(oncallusershift.FieldUserID, v)
	return u
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsert) UpdateUserID() *OncallUserShiftUpsert {
	u.SetExcluded(oncallusershift.FieldUserID)
	return u
}

// SetRosterID sets the "roster_id" field.
func (u *OncallUserShiftUpsert) SetRosterID(v uuid.UUID) *OncallUserShiftUpsert {
	u.Set(oncallusershift.FieldRosterID, v)
	return u
}

// UpdateRosterID sets the "roster_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsert) UpdateRosterID() *OncallUserShiftUpsert {
	u.SetExcluded(oncallusershift.FieldRosterID)
	return u
}

// SetStartAt sets the "start_at" field.
func (u *OncallUserShiftUpsert) SetStartAt(v time.Time) *OncallUserShiftUpsert {
	u.Set(oncallusershift.FieldStartAt, v)
	return u
}

// UpdateStartAt sets the "start_at" field to the value that was provided on create.
func (u *OncallUserShiftUpsert) UpdateStartAt() *OncallUserShiftUpsert {
	u.SetExcluded(oncallusershift.FieldStartAt)
	return u
}

// SetEndAt sets the "end_at" field.
func (u *OncallUserShiftUpsert) SetEndAt(v time.Time) *OncallUserShiftUpsert {
	u.Set(oncallusershift.FieldEndAt, v)
	return u
}

// UpdateEndAt sets the "end_at" field to the value that was provided on create.
func (u *OncallUserShiftUpsert) UpdateEndAt() *OncallUserShiftUpsert {
	u.SetExcluded(oncallusershift.FieldEndAt)
	return u
}

// SetProviderID sets the "provider_id" field.
func (u *OncallUserShiftUpsert) SetProviderID(v string) *OncallUserShiftUpsert {
	u.Set(oncallusershift.FieldProviderID, v)
	return u
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsert) UpdateProviderID() *OncallUserShiftUpsert {
	u.SetExcluded(oncallusershift.FieldProviderID)
	return u
}

// ClearProviderID clears the value of the "provider_id" field.
func (u *OncallUserShiftUpsert) ClearProviderID() *OncallUserShiftUpsert {
	u.SetNull(oncallusershift.FieldProviderID)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.OncallUserShift.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(oncallusershift.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *OncallUserShiftUpsertOne) UpdateNewValues() *OncallUserShiftUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(oncallusershift.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.OncallUserShift.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *OncallUserShiftUpsertOne) Ignore() *OncallUserShiftUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *OncallUserShiftUpsertOne) DoNothing() *OncallUserShiftUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the OncallUserShiftCreate.OnConflict
// documentation for more info.
func (u *OncallUserShiftUpsertOne) Update(set func(*OncallUserShiftUpsert)) *OncallUserShiftUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&OncallUserShiftUpsert{UpdateSet: update})
	}))
	return u
}

// SetUserID sets the "user_id" field.
func (u *OncallUserShiftUpsertOne) SetUserID(v uuid.UUID) *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetUserID(v)
	})
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsertOne) UpdateUserID() *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateUserID()
	})
}

// SetRosterID sets the "roster_id" field.
func (u *OncallUserShiftUpsertOne) SetRosterID(v uuid.UUID) *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetRosterID(v)
	})
}

// UpdateRosterID sets the "roster_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsertOne) UpdateRosterID() *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateRosterID()
	})
}

// SetStartAt sets the "start_at" field.
func (u *OncallUserShiftUpsertOne) SetStartAt(v time.Time) *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetStartAt(v)
	})
}

// UpdateStartAt sets the "start_at" field to the value that was provided on create.
func (u *OncallUserShiftUpsertOne) UpdateStartAt() *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateStartAt()
	})
}

// SetEndAt sets the "end_at" field.
func (u *OncallUserShiftUpsertOne) SetEndAt(v time.Time) *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetEndAt(v)
	})
}

// UpdateEndAt sets the "end_at" field to the value that was provided on create.
func (u *OncallUserShiftUpsertOne) UpdateEndAt() *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateEndAt()
	})
}

// SetProviderID sets the "provider_id" field.
func (u *OncallUserShiftUpsertOne) SetProviderID(v string) *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetProviderID(v)
	})
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsertOne) UpdateProviderID() *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateProviderID()
	})
}

// ClearProviderID clears the value of the "provider_id" field.
func (u *OncallUserShiftUpsertOne) ClearProviderID() *OncallUserShiftUpsertOne {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.ClearProviderID()
	})
}

// Exec executes the query.
func (u *OncallUserShiftUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for OncallUserShiftCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *OncallUserShiftUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *OncallUserShiftUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: OncallUserShiftUpsertOne.ID is not supported by MySQL driver. Use OncallUserShiftUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *OncallUserShiftUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// OncallUserShiftCreateBulk is the builder for creating many OncallUserShift entities in bulk.
type OncallUserShiftCreateBulk struct {
	config
	err      error
	builders []*OncallUserShiftCreate
	conflict []sql.ConflictOption
}

// Save creates the OncallUserShift entities in the database.
func (ouscb *OncallUserShiftCreateBulk) Save(ctx context.Context) ([]*OncallUserShift, error) {
	if ouscb.err != nil {
		return nil, ouscb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ouscb.builders))
	nodes := make([]*OncallUserShift, len(ouscb.builders))
	mutators := make([]Mutator, len(ouscb.builders))
	for i := range ouscb.builders {
		func(i int, root context.Context) {
			builder := ouscb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*OncallUserShiftMutation)
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
					_, err = mutators[i+1].Mutate(root, ouscb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ouscb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ouscb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ouscb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ouscb *OncallUserShiftCreateBulk) SaveX(ctx context.Context) []*OncallUserShift {
	v, err := ouscb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ouscb *OncallUserShiftCreateBulk) Exec(ctx context.Context) error {
	_, err := ouscb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ouscb *OncallUserShiftCreateBulk) ExecX(ctx context.Context) {
	if err := ouscb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.OncallUserShift.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.OncallUserShiftUpsert) {
//			SetUserID(v+v).
//		}).
//		Exec(ctx)
func (ouscb *OncallUserShiftCreateBulk) OnConflict(opts ...sql.ConflictOption) *OncallUserShiftUpsertBulk {
	ouscb.conflict = opts
	return &OncallUserShiftUpsertBulk{
		create: ouscb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.OncallUserShift.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ouscb *OncallUserShiftCreateBulk) OnConflictColumns(columns ...string) *OncallUserShiftUpsertBulk {
	ouscb.conflict = append(ouscb.conflict, sql.ConflictColumns(columns...))
	return &OncallUserShiftUpsertBulk{
		create: ouscb,
	}
}

// OncallUserShiftUpsertBulk is the builder for "upsert"-ing
// a bulk of OncallUserShift nodes.
type OncallUserShiftUpsertBulk struct {
	create *OncallUserShiftCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.OncallUserShift.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(oncallusershift.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *OncallUserShiftUpsertBulk) UpdateNewValues() *OncallUserShiftUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(oncallusershift.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.OncallUserShift.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *OncallUserShiftUpsertBulk) Ignore() *OncallUserShiftUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *OncallUserShiftUpsertBulk) DoNothing() *OncallUserShiftUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the OncallUserShiftCreateBulk.OnConflict
// documentation for more info.
func (u *OncallUserShiftUpsertBulk) Update(set func(*OncallUserShiftUpsert)) *OncallUserShiftUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&OncallUserShiftUpsert{UpdateSet: update})
	}))
	return u
}

// SetUserID sets the "user_id" field.
func (u *OncallUserShiftUpsertBulk) SetUserID(v uuid.UUID) *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetUserID(v)
	})
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsertBulk) UpdateUserID() *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateUserID()
	})
}

// SetRosterID sets the "roster_id" field.
func (u *OncallUserShiftUpsertBulk) SetRosterID(v uuid.UUID) *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetRosterID(v)
	})
}

// UpdateRosterID sets the "roster_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsertBulk) UpdateRosterID() *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateRosterID()
	})
}

// SetStartAt sets the "start_at" field.
func (u *OncallUserShiftUpsertBulk) SetStartAt(v time.Time) *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetStartAt(v)
	})
}

// UpdateStartAt sets the "start_at" field to the value that was provided on create.
func (u *OncallUserShiftUpsertBulk) UpdateStartAt() *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateStartAt()
	})
}

// SetEndAt sets the "end_at" field.
func (u *OncallUserShiftUpsertBulk) SetEndAt(v time.Time) *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetEndAt(v)
	})
}

// UpdateEndAt sets the "end_at" field to the value that was provided on create.
func (u *OncallUserShiftUpsertBulk) UpdateEndAt() *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateEndAt()
	})
}

// SetProviderID sets the "provider_id" field.
func (u *OncallUserShiftUpsertBulk) SetProviderID(v string) *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.SetProviderID(v)
	})
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *OncallUserShiftUpsertBulk) UpdateProviderID() *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.UpdateProviderID()
	})
}

// ClearProviderID clears the value of the "provider_id" field.
func (u *OncallUserShiftUpsertBulk) ClearProviderID() *OncallUserShiftUpsertBulk {
	return u.Update(func(s *OncallUserShiftUpsert) {
		s.ClearProviderID()
	})
}

// Exec executes the query.
func (u *OncallUserShiftUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the OncallUserShiftCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for OncallUserShiftCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *OncallUserShiftUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

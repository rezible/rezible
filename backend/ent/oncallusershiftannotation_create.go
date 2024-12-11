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
	"github.com/twohundreds/rezible/ent/oncallusershift"
	"github.com/twohundreds/rezible/ent/oncallusershiftannotation"
)

// OncallUserShiftAnnotationCreate is the builder for creating a OncallUserShiftAnnotation entity.
type OncallUserShiftAnnotationCreate struct {
	config
	mutation *OncallUserShiftAnnotationMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetShiftID sets the "shift_id" field.
func (ousac *OncallUserShiftAnnotationCreate) SetShiftID(u uuid.UUID) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetShiftID(u)
	return ousac
}

// SetEventID sets the "event_id" field.
func (ousac *OncallUserShiftAnnotationCreate) SetEventID(s string) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetEventID(s)
	return ousac
}

// SetEventKind sets the "event_kind" field.
func (ousac *OncallUserShiftAnnotationCreate) SetEventKind(ok oncallusershiftannotation.EventKind) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetEventKind(ok)
	return ousac
}

// SetTitle sets the "title" field.
func (ousac *OncallUserShiftAnnotationCreate) SetTitle(s string) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetTitle(s)
	return ousac
}

// SetOccurredAt sets the "occurred_at" field.
func (ousac *OncallUserShiftAnnotationCreate) SetOccurredAt(t time.Time) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetOccurredAt(t)
	return ousac
}

// SetMinutesOccupied sets the "minutes_occupied" field.
func (ousac *OncallUserShiftAnnotationCreate) SetMinutesOccupied(i int) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetMinutesOccupied(i)
	return ousac
}

// SetNotes sets the "notes" field.
func (ousac *OncallUserShiftAnnotationCreate) SetNotes(s string) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetNotes(s)
	return ousac
}

// SetPinned sets the "pinned" field.
func (ousac *OncallUserShiftAnnotationCreate) SetPinned(b bool) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetPinned(b)
	return ousac
}

// SetID sets the "id" field.
func (ousac *OncallUserShiftAnnotationCreate) SetID(u uuid.UUID) *OncallUserShiftAnnotationCreate {
	ousac.mutation.SetID(u)
	return ousac
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ousac *OncallUserShiftAnnotationCreate) SetNillableID(u *uuid.UUID) *OncallUserShiftAnnotationCreate {
	if u != nil {
		ousac.SetID(*u)
	}
	return ousac
}

// SetShift sets the "shift" edge to the OncallUserShift entity.
func (ousac *OncallUserShiftAnnotationCreate) SetShift(o *OncallUserShift) *OncallUserShiftAnnotationCreate {
	return ousac.SetShiftID(o.ID)
}

// Mutation returns the OncallUserShiftAnnotationMutation object of the builder.
func (ousac *OncallUserShiftAnnotationCreate) Mutation() *OncallUserShiftAnnotationMutation {
	return ousac.mutation
}

// Save creates the OncallUserShiftAnnotation in the database.
func (ousac *OncallUserShiftAnnotationCreate) Save(ctx context.Context) (*OncallUserShiftAnnotation, error) {
	ousac.defaults()
	return withHooks(ctx, ousac.sqlSave, ousac.mutation, ousac.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ousac *OncallUserShiftAnnotationCreate) SaveX(ctx context.Context) *OncallUserShiftAnnotation {
	v, err := ousac.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ousac *OncallUserShiftAnnotationCreate) Exec(ctx context.Context) error {
	_, err := ousac.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ousac *OncallUserShiftAnnotationCreate) ExecX(ctx context.Context) {
	if err := ousac.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ousac *OncallUserShiftAnnotationCreate) defaults() {
	if _, ok := ousac.mutation.ID(); !ok {
		v := oncallusershiftannotation.DefaultID()
		ousac.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ousac *OncallUserShiftAnnotationCreate) check() error {
	if _, ok := ousac.mutation.ShiftID(); !ok {
		return &ValidationError{Name: "shift_id", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.shift_id"`)}
	}
	if _, ok := ousac.mutation.EventID(); !ok {
		return &ValidationError{Name: "event_id", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.event_id"`)}
	}
	if _, ok := ousac.mutation.EventKind(); !ok {
		return &ValidationError{Name: "event_kind", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.event_kind"`)}
	}
	if v, ok := ousac.mutation.EventKind(); ok {
		if err := oncallusershiftannotation.EventKindValidator(v); err != nil {
			return &ValidationError{Name: "event_kind", err: fmt.Errorf(`ent: validator failed for field "OncallUserShiftAnnotation.event_kind": %w`, err)}
		}
	}
	if _, ok := ousac.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.title"`)}
	}
	if _, ok := ousac.mutation.OccurredAt(); !ok {
		return &ValidationError{Name: "occurred_at", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.occurred_at"`)}
	}
	if _, ok := ousac.mutation.MinutesOccupied(); !ok {
		return &ValidationError{Name: "minutes_occupied", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.minutes_occupied"`)}
	}
	if _, ok := ousac.mutation.Notes(); !ok {
		return &ValidationError{Name: "notes", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.notes"`)}
	}
	if _, ok := ousac.mutation.Pinned(); !ok {
		return &ValidationError{Name: "pinned", err: errors.New(`ent: missing required field "OncallUserShiftAnnotation.pinned"`)}
	}
	if len(ousac.mutation.ShiftIDs()) == 0 {
		return &ValidationError{Name: "shift", err: errors.New(`ent: missing required edge "OncallUserShiftAnnotation.shift"`)}
	}
	return nil
}

func (ousac *OncallUserShiftAnnotationCreate) sqlSave(ctx context.Context) (*OncallUserShiftAnnotation, error) {
	if err := ousac.check(); err != nil {
		return nil, err
	}
	_node, _spec := ousac.createSpec()
	if err := sqlgraph.CreateNode(ctx, ousac.driver, _spec); err != nil {
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
	ousac.mutation.id = &_node.ID
	ousac.mutation.done = true
	return _node, nil
}

func (ousac *OncallUserShiftAnnotationCreate) createSpec() (*OncallUserShiftAnnotation, *sqlgraph.CreateSpec) {
	var (
		_node = &OncallUserShiftAnnotation{config: ousac.config}
		_spec = sqlgraph.NewCreateSpec(oncallusershiftannotation.Table, sqlgraph.NewFieldSpec(oncallusershiftannotation.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = ousac.conflict
	if id, ok := ousac.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := ousac.mutation.EventID(); ok {
		_spec.SetField(oncallusershiftannotation.FieldEventID, field.TypeString, value)
		_node.EventID = value
	}
	if value, ok := ousac.mutation.EventKind(); ok {
		_spec.SetField(oncallusershiftannotation.FieldEventKind, field.TypeEnum, value)
		_node.EventKind = value
	}
	if value, ok := ousac.mutation.Title(); ok {
		_spec.SetField(oncallusershiftannotation.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := ousac.mutation.OccurredAt(); ok {
		_spec.SetField(oncallusershiftannotation.FieldOccurredAt, field.TypeTime, value)
		_node.OccurredAt = value
	}
	if value, ok := ousac.mutation.MinutesOccupied(); ok {
		_spec.SetField(oncallusershiftannotation.FieldMinutesOccupied, field.TypeInt, value)
		_node.MinutesOccupied = value
	}
	if value, ok := ousac.mutation.Notes(); ok {
		_spec.SetField(oncallusershiftannotation.FieldNotes, field.TypeString, value)
		_node.Notes = value
	}
	if value, ok := ousac.mutation.Pinned(); ok {
		_spec.SetField(oncallusershiftannotation.FieldPinned, field.TypeBool, value)
		_node.Pinned = value
	}
	if nodes := ousac.mutation.ShiftIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   oncallusershiftannotation.ShiftTable,
			Columns: []string{oncallusershiftannotation.ShiftColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallusershift.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ShiftID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.OncallUserShiftAnnotation.Create().
//		SetShiftID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.OncallUserShiftAnnotationUpsert) {
//			SetShiftID(v+v).
//		}).
//		Exec(ctx)
func (ousac *OncallUserShiftAnnotationCreate) OnConflict(opts ...sql.ConflictOption) *OncallUserShiftAnnotationUpsertOne {
	ousac.conflict = opts
	return &OncallUserShiftAnnotationUpsertOne{
		create: ousac,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.OncallUserShiftAnnotation.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ousac *OncallUserShiftAnnotationCreate) OnConflictColumns(columns ...string) *OncallUserShiftAnnotationUpsertOne {
	ousac.conflict = append(ousac.conflict, sql.ConflictColumns(columns...))
	return &OncallUserShiftAnnotationUpsertOne{
		create: ousac,
	}
}

type (
	// OncallUserShiftAnnotationUpsertOne is the builder for "upsert"-ing
	//  one OncallUserShiftAnnotation node.
	OncallUserShiftAnnotationUpsertOne struct {
		create *OncallUserShiftAnnotationCreate
	}

	// OncallUserShiftAnnotationUpsert is the "OnConflict" setter.
	OncallUserShiftAnnotationUpsert struct {
		*sql.UpdateSet
	}
)

// SetShiftID sets the "shift_id" field.
func (u *OncallUserShiftAnnotationUpsert) SetShiftID(v uuid.UUID) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldShiftID, v)
	return u
}

// UpdateShiftID sets the "shift_id" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateShiftID() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldShiftID)
	return u
}

// SetEventID sets the "event_id" field.
func (u *OncallUserShiftAnnotationUpsert) SetEventID(v string) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldEventID, v)
	return u
}

// UpdateEventID sets the "event_id" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateEventID() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldEventID)
	return u
}

// SetEventKind sets the "event_kind" field.
func (u *OncallUserShiftAnnotationUpsert) SetEventKind(v oncallusershiftannotation.EventKind) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldEventKind, v)
	return u
}

// UpdateEventKind sets the "event_kind" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateEventKind() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldEventKind)
	return u
}

// SetTitle sets the "title" field.
func (u *OncallUserShiftAnnotationUpsert) SetTitle(v string) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldTitle, v)
	return u
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateTitle() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldTitle)
	return u
}

// SetOccurredAt sets the "occurred_at" field.
func (u *OncallUserShiftAnnotationUpsert) SetOccurredAt(v time.Time) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldOccurredAt, v)
	return u
}

// UpdateOccurredAt sets the "occurred_at" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateOccurredAt() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldOccurredAt)
	return u
}

// SetMinutesOccupied sets the "minutes_occupied" field.
func (u *OncallUserShiftAnnotationUpsert) SetMinutesOccupied(v int) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldMinutesOccupied, v)
	return u
}

// UpdateMinutesOccupied sets the "minutes_occupied" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateMinutesOccupied() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldMinutesOccupied)
	return u
}

// AddMinutesOccupied adds v to the "minutes_occupied" field.
func (u *OncallUserShiftAnnotationUpsert) AddMinutesOccupied(v int) *OncallUserShiftAnnotationUpsert {
	u.Add(oncallusershiftannotation.FieldMinutesOccupied, v)
	return u
}

// SetNotes sets the "notes" field.
func (u *OncallUserShiftAnnotationUpsert) SetNotes(v string) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldNotes, v)
	return u
}

// UpdateNotes sets the "notes" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdateNotes() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldNotes)
	return u
}

// SetPinned sets the "pinned" field.
func (u *OncallUserShiftAnnotationUpsert) SetPinned(v bool) *OncallUserShiftAnnotationUpsert {
	u.Set(oncallusershiftannotation.FieldPinned, v)
	return u
}

// UpdatePinned sets the "pinned" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsert) UpdatePinned() *OncallUserShiftAnnotationUpsert {
	u.SetExcluded(oncallusershiftannotation.FieldPinned)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.OncallUserShiftAnnotation.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(oncallusershiftannotation.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *OncallUserShiftAnnotationUpsertOne) UpdateNewValues() *OncallUserShiftAnnotationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(oncallusershiftannotation.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.OncallUserShiftAnnotation.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *OncallUserShiftAnnotationUpsertOne) Ignore() *OncallUserShiftAnnotationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *OncallUserShiftAnnotationUpsertOne) DoNothing() *OncallUserShiftAnnotationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the OncallUserShiftAnnotationCreate.OnConflict
// documentation for more info.
func (u *OncallUserShiftAnnotationUpsertOne) Update(set func(*OncallUserShiftAnnotationUpsert)) *OncallUserShiftAnnotationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&OncallUserShiftAnnotationUpsert{UpdateSet: update})
	}))
	return u
}

// SetShiftID sets the "shift_id" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetShiftID(v uuid.UUID) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetShiftID(v)
	})
}

// UpdateShiftID sets the "shift_id" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateShiftID() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateShiftID()
	})
}

// SetEventID sets the "event_id" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetEventID(v string) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetEventID(v)
	})
}

// UpdateEventID sets the "event_id" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateEventID() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateEventID()
	})
}

// SetEventKind sets the "event_kind" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetEventKind(v oncallusershiftannotation.EventKind) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetEventKind(v)
	})
}

// UpdateEventKind sets the "event_kind" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateEventKind() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateEventKind()
	})
}

// SetTitle sets the "title" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetTitle(v string) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetTitle(v)
	})
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateTitle() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateTitle()
	})
}

// SetOccurredAt sets the "occurred_at" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetOccurredAt(v time.Time) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetOccurredAt(v)
	})
}

// UpdateOccurredAt sets the "occurred_at" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateOccurredAt() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateOccurredAt()
	})
}

// SetMinutesOccupied sets the "minutes_occupied" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetMinutesOccupied(v int) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetMinutesOccupied(v)
	})
}

// AddMinutesOccupied adds v to the "minutes_occupied" field.
func (u *OncallUserShiftAnnotationUpsertOne) AddMinutesOccupied(v int) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.AddMinutesOccupied(v)
	})
}

// UpdateMinutesOccupied sets the "minutes_occupied" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateMinutesOccupied() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateMinutesOccupied()
	})
}

// SetNotes sets the "notes" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetNotes(v string) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetNotes(v)
	})
}

// UpdateNotes sets the "notes" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdateNotes() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateNotes()
	})
}

// SetPinned sets the "pinned" field.
func (u *OncallUserShiftAnnotationUpsertOne) SetPinned(v bool) *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetPinned(v)
	})
}

// UpdatePinned sets the "pinned" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertOne) UpdatePinned() *OncallUserShiftAnnotationUpsertOne {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdatePinned()
	})
}

// Exec executes the query.
func (u *OncallUserShiftAnnotationUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for OncallUserShiftAnnotationCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *OncallUserShiftAnnotationUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *OncallUserShiftAnnotationUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: OncallUserShiftAnnotationUpsertOne.ID is not supported by MySQL driver. Use OncallUserShiftAnnotationUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *OncallUserShiftAnnotationUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// OncallUserShiftAnnotationCreateBulk is the builder for creating many OncallUserShiftAnnotation entities in bulk.
type OncallUserShiftAnnotationCreateBulk struct {
	config
	err      error
	builders []*OncallUserShiftAnnotationCreate
	conflict []sql.ConflictOption
}

// Save creates the OncallUserShiftAnnotation entities in the database.
func (ousacb *OncallUserShiftAnnotationCreateBulk) Save(ctx context.Context) ([]*OncallUserShiftAnnotation, error) {
	if ousacb.err != nil {
		return nil, ousacb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ousacb.builders))
	nodes := make([]*OncallUserShiftAnnotation, len(ousacb.builders))
	mutators := make([]Mutator, len(ousacb.builders))
	for i := range ousacb.builders {
		func(i int, root context.Context) {
			builder := ousacb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*OncallUserShiftAnnotationMutation)
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
					_, err = mutators[i+1].Mutate(root, ousacb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ousacb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ousacb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ousacb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ousacb *OncallUserShiftAnnotationCreateBulk) SaveX(ctx context.Context) []*OncallUserShiftAnnotation {
	v, err := ousacb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ousacb *OncallUserShiftAnnotationCreateBulk) Exec(ctx context.Context) error {
	_, err := ousacb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ousacb *OncallUserShiftAnnotationCreateBulk) ExecX(ctx context.Context) {
	if err := ousacb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.OncallUserShiftAnnotation.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.OncallUserShiftAnnotationUpsert) {
//			SetShiftID(v+v).
//		}).
//		Exec(ctx)
func (ousacb *OncallUserShiftAnnotationCreateBulk) OnConflict(opts ...sql.ConflictOption) *OncallUserShiftAnnotationUpsertBulk {
	ousacb.conflict = opts
	return &OncallUserShiftAnnotationUpsertBulk{
		create: ousacb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.OncallUserShiftAnnotation.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ousacb *OncallUserShiftAnnotationCreateBulk) OnConflictColumns(columns ...string) *OncallUserShiftAnnotationUpsertBulk {
	ousacb.conflict = append(ousacb.conflict, sql.ConflictColumns(columns...))
	return &OncallUserShiftAnnotationUpsertBulk{
		create: ousacb,
	}
}

// OncallUserShiftAnnotationUpsertBulk is the builder for "upsert"-ing
// a bulk of OncallUserShiftAnnotation nodes.
type OncallUserShiftAnnotationUpsertBulk struct {
	create *OncallUserShiftAnnotationCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.OncallUserShiftAnnotation.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(oncallusershiftannotation.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateNewValues() *OncallUserShiftAnnotationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(oncallusershiftannotation.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.OncallUserShiftAnnotation.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *OncallUserShiftAnnotationUpsertBulk) Ignore() *OncallUserShiftAnnotationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *OncallUserShiftAnnotationUpsertBulk) DoNothing() *OncallUserShiftAnnotationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the OncallUserShiftAnnotationCreateBulk.OnConflict
// documentation for more info.
func (u *OncallUserShiftAnnotationUpsertBulk) Update(set func(*OncallUserShiftAnnotationUpsert)) *OncallUserShiftAnnotationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&OncallUserShiftAnnotationUpsert{UpdateSet: update})
	}))
	return u
}

// SetShiftID sets the "shift_id" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetShiftID(v uuid.UUID) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetShiftID(v)
	})
}

// UpdateShiftID sets the "shift_id" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateShiftID() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateShiftID()
	})
}

// SetEventID sets the "event_id" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetEventID(v string) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetEventID(v)
	})
}

// UpdateEventID sets the "event_id" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateEventID() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateEventID()
	})
}

// SetEventKind sets the "event_kind" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetEventKind(v oncallusershiftannotation.EventKind) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetEventKind(v)
	})
}

// UpdateEventKind sets the "event_kind" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateEventKind() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateEventKind()
	})
}

// SetTitle sets the "title" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetTitle(v string) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetTitle(v)
	})
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateTitle() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateTitle()
	})
}

// SetOccurredAt sets the "occurred_at" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetOccurredAt(v time.Time) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetOccurredAt(v)
	})
}

// UpdateOccurredAt sets the "occurred_at" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateOccurredAt() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateOccurredAt()
	})
}

// SetMinutesOccupied sets the "minutes_occupied" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetMinutesOccupied(v int) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetMinutesOccupied(v)
	})
}

// AddMinutesOccupied adds v to the "minutes_occupied" field.
func (u *OncallUserShiftAnnotationUpsertBulk) AddMinutesOccupied(v int) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.AddMinutesOccupied(v)
	})
}

// UpdateMinutesOccupied sets the "minutes_occupied" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateMinutesOccupied() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateMinutesOccupied()
	})
}

// SetNotes sets the "notes" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetNotes(v string) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetNotes(v)
	})
}

// UpdateNotes sets the "notes" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdateNotes() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdateNotes()
	})
}

// SetPinned sets the "pinned" field.
func (u *OncallUserShiftAnnotationUpsertBulk) SetPinned(v bool) *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.SetPinned(v)
	})
}

// UpdatePinned sets the "pinned" field to the value that was provided on create.
func (u *OncallUserShiftAnnotationUpsertBulk) UpdatePinned() *OncallUserShiftAnnotationUpsertBulk {
	return u.Update(func(s *OncallUserShiftAnnotationUpsert) {
		s.UpdatePinned()
	})
}

// Exec executes the query.
func (u *OncallUserShiftAnnotationUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the OncallUserShiftAnnotationCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for OncallUserShiftAnnotationCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *OncallUserShiftAnnotationUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
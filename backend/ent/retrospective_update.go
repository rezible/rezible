// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivediscussion"
)

// RetrospectiveUpdate is the builder for updating Retrospective entities.
type RetrospectiveUpdate struct {
	config
	hooks     []Hook
	mutation  *RetrospectiveMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the RetrospectiveUpdate builder.
func (ru *RetrospectiveUpdate) Where(ps ...predicate.Retrospective) *RetrospectiveUpdate {
	ru.mutation.Where(ps...)
	return ru
}

// SetDocumentName sets the "document_name" field.
func (ru *RetrospectiveUpdate) SetDocumentName(s string) *RetrospectiveUpdate {
	ru.mutation.SetDocumentName(s)
	return ru
}

// SetNillableDocumentName sets the "document_name" field if the given value is not nil.
func (ru *RetrospectiveUpdate) SetNillableDocumentName(s *string) *RetrospectiveUpdate {
	if s != nil {
		ru.SetDocumentName(*s)
	}
	return ru
}

// SetType sets the "type" field.
func (ru *RetrospectiveUpdate) SetType(r retrospective.Type) *RetrospectiveUpdate {
	ru.mutation.SetType(r)
	return ru
}

// SetNillableType sets the "type" field if the given value is not nil.
func (ru *RetrospectiveUpdate) SetNillableType(r *retrospective.Type) *RetrospectiveUpdate {
	if r != nil {
		ru.SetType(*r)
	}
	return ru
}

// SetState sets the "state" field.
func (ru *RetrospectiveUpdate) SetState(r retrospective.State) *RetrospectiveUpdate {
	ru.mutation.SetState(r)
	return ru
}

// SetNillableState sets the "state" field if the given value is not nil.
func (ru *RetrospectiveUpdate) SetNillableState(r *retrospective.State) *RetrospectiveUpdate {
	if r != nil {
		ru.SetState(*r)
	}
	return ru
}

// SetIncidentID sets the "incident" edge to the Incident entity by ID.
func (ru *RetrospectiveUpdate) SetIncidentID(id uuid.UUID) *RetrospectiveUpdate {
	ru.mutation.SetIncidentID(id)
	return ru
}

// SetNillableIncidentID sets the "incident" edge to the Incident entity by ID if the given value is not nil.
func (ru *RetrospectiveUpdate) SetNillableIncidentID(id *uuid.UUID) *RetrospectiveUpdate {
	if id != nil {
		ru = ru.SetIncidentID(*id)
	}
	return ru
}

// SetIncident sets the "incident" edge to the Incident entity.
func (ru *RetrospectiveUpdate) SetIncident(i *Incident) *RetrospectiveUpdate {
	return ru.SetIncidentID(i.ID)
}

// AddDiscussionIDs adds the "discussions" edge to the RetrospectiveDiscussion entity by IDs.
func (ru *RetrospectiveUpdate) AddDiscussionIDs(ids ...uuid.UUID) *RetrospectiveUpdate {
	ru.mutation.AddDiscussionIDs(ids...)
	return ru
}

// AddDiscussions adds the "discussions" edges to the RetrospectiveDiscussion entity.
func (ru *RetrospectiveUpdate) AddDiscussions(r ...*RetrospectiveDiscussion) *RetrospectiveUpdate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return ru.AddDiscussionIDs(ids...)
}

// Mutation returns the RetrospectiveMutation object of the builder.
func (ru *RetrospectiveUpdate) Mutation() *RetrospectiveMutation {
	return ru.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (ru *RetrospectiveUpdate) ClearIncident() *RetrospectiveUpdate {
	ru.mutation.ClearIncident()
	return ru
}

// ClearDiscussions clears all "discussions" edges to the RetrospectiveDiscussion entity.
func (ru *RetrospectiveUpdate) ClearDiscussions() *RetrospectiveUpdate {
	ru.mutation.ClearDiscussions()
	return ru
}

// RemoveDiscussionIDs removes the "discussions" edge to RetrospectiveDiscussion entities by IDs.
func (ru *RetrospectiveUpdate) RemoveDiscussionIDs(ids ...uuid.UUID) *RetrospectiveUpdate {
	ru.mutation.RemoveDiscussionIDs(ids...)
	return ru
}

// RemoveDiscussions removes "discussions" edges to RetrospectiveDiscussion entities.
func (ru *RetrospectiveUpdate) RemoveDiscussions(r ...*RetrospectiveDiscussion) *RetrospectiveUpdate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return ru.RemoveDiscussionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (ru *RetrospectiveUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, ru.sqlSave, ru.mutation, ru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ru *RetrospectiveUpdate) SaveX(ctx context.Context) int {
	affected, err := ru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (ru *RetrospectiveUpdate) Exec(ctx context.Context) error {
	_, err := ru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ru *RetrospectiveUpdate) ExecX(ctx context.Context) {
	if err := ru.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ru *RetrospectiveUpdate) check() error {
	if v, ok := ru.mutation.GetType(); ok {
		if err := retrospective.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Retrospective.type": %w`, err)}
		}
	}
	if v, ok := ru.mutation.State(); ok {
		if err := retrospective.StateValidator(v); err != nil {
			return &ValidationError{Name: "state", err: fmt.Errorf(`ent: validator failed for field "Retrospective.state": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (ru *RetrospectiveUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *RetrospectiveUpdate {
	ru.modifiers = append(ru.modifiers, modifiers...)
	return ru
}

func (ru *RetrospectiveUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := ru.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(retrospective.Table, retrospective.Columns, sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID))
	if ps := ru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ru.mutation.DocumentName(); ok {
		_spec.SetField(retrospective.FieldDocumentName, field.TypeString, value)
	}
	if value, ok := ru.mutation.GetType(); ok {
		_spec.SetField(retrospective.FieldType, field.TypeEnum, value)
	}
	if value, ok := ru.mutation.State(); ok {
		_spec.SetField(retrospective.FieldState, field.TypeEnum, value)
	}
	if ru.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   retrospective.IncidentTable,
			Columns: []string{retrospective.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ru.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if ru.mutation.DiscussionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ru.mutation.RemovedDiscussionsIDs(); len(nodes) > 0 && !ru.mutation.DiscussionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ru.mutation.DiscussionsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(ru.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, ru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{retrospective.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	ru.mutation.done = true
	return n, nil
}

// RetrospectiveUpdateOne is the builder for updating a single Retrospective entity.
type RetrospectiveUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *RetrospectiveMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetDocumentName sets the "document_name" field.
func (ruo *RetrospectiveUpdateOne) SetDocumentName(s string) *RetrospectiveUpdateOne {
	ruo.mutation.SetDocumentName(s)
	return ruo
}

// SetNillableDocumentName sets the "document_name" field if the given value is not nil.
func (ruo *RetrospectiveUpdateOne) SetNillableDocumentName(s *string) *RetrospectiveUpdateOne {
	if s != nil {
		ruo.SetDocumentName(*s)
	}
	return ruo
}

// SetType sets the "type" field.
func (ruo *RetrospectiveUpdateOne) SetType(r retrospective.Type) *RetrospectiveUpdateOne {
	ruo.mutation.SetType(r)
	return ruo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (ruo *RetrospectiveUpdateOne) SetNillableType(r *retrospective.Type) *RetrospectiveUpdateOne {
	if r != nil {
		ruo.SetType(*r)
	}
	return ruo
}

// SetState sets the "state" field.
func (ruo *RetrospectiveUpdateOne) SetState(r retrospective.State) *RetrospectiveUpdateOne {
	ruo.mutation.SetState(r)
	return ruo
}

// SetNillableState sets the "state" field if the given value is not nil.
func (ruo *RetrospectiveUpdateOne) SetNillableState(r *retrospective.State) *RetrospectiveUpdateOne {
	if r != nil {
		ruo.SetState(*r)
	}
	return ruo
}

// SetIncidentID sets the "incident" edge to the Incident entity by ID.
func (ruo *RetrospectiveUpdateOne) SetIncidentID(id uuid.UUID) *RetrospectiveUpdateOne {
	ruo.mutation.SetIncidentID(id)
	return ruo
}

// SetNillableIncidentID sets the "incident" edge to the Incident entity by ID if the given value is not nil.
func (ruo *RetrospectiveUpdateOne) SetNillableIncidentID(id *uuid.UUID) *RetrospectiveUpdateOne {
	if id != nil {
		ruo = ruo.SetIncidentID(*id)
	}
	return ruo
}

// SetIncident sets the "incident" edge to the Incident entity.
func (ruo *RetrospectiveUpdateOne) SetIncident(i *Incident) *RetrospectiveUpdateOne {
	return ruo.SetIncidentID(i.ID)
}

// AddDiscussionIDs adds the "discussions" edge to the RetrospectiveDiscussion entity by IDs.
func (ruo *RetrospectiveUpdateOne) AddDiscussionIDs(ids ...uuid.UUID) *RetrospectiveUpdateOne {
	ruo.mutation.AddDiscussionIDs(ids...)
	return ruo
}

// AddDiscussions adds the "discussions" edges to the RetrospectiveDiscussion entity.
func (ruo *RetrospectiveUpdateOne) AddDiscussions(r ...*RetrospectiveDiscussion) *RetrospectiveUpdateOne {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return ruo.AddDiscussionIDs(ids...)
}

// Mutation returns the RetrospectiveMutation object of the builder.
func (ruo *RetrospectiveUpdateOne) Mutation() *RetrospectiveMutation {
	return ruo.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (ruo *RetrospectiveUpdateOne) ClearIncident() *RetrospectiveUpdateOne {
	ruo.mutation.ClearIncident()
	return ruo
}

// ClearDiscussions clears all "discussions" edges to the RetrospectiveDiscussion entity.
func (ruo *RetrospectiveUpdateOne) ClearDiscussions() *RetrospectiveUpdateOne {
	ruo.mutation.ClearDiscussions()
	return ruo
}

// RemoveDiscussionIDs removes the "discussions" edge to RetrospectiveDiscussion entities by IDs.
func (ruo *RetrospectiveUpdateOne) RemoveDiscussionIDs(ids ...uuid.UUID) *RetrospectiveUpdateOne {
	ruo.mutation.RemoveDiscussionIDs(ids...)
	return ruo
}

// RemoveDiscussions removes "discussions" edges to RetrospectiveDiscussion entities.
func (ruo *RetrospectiveUpdateOne) RemoveDiscussions(r ...*RetrospectiveDiscussion) *RetrospectiveUpdateOne {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return ruo.RemoveDiscussionIDs(ids...)
}

// Where appends a list predicates to the RetrospectiveUpdate builder.
func (ruo *RetrospectiveUpdateOne) Where(ps ...predicate.Retrospective) *RetrospectiveUpdateOne {
	ruo.mutation.Where(ps...)
	return ruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (ruo *RetrospectiveUpdateOne) Select(field string, fields ...string) *RetrospectiveUpdateOne {
	ruo.fields = append([]string{field}, fields...)
	return ruo
}

// Save executes the query and returns the updated Retrospective entity.
func (ruo *RetrospectiveUpdateOne) Save(ctx context.Context) (*Retrospective, error) {
	return withHooks(ctx, ruo.sqlSave, ruo.mutation, ruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ruo *RetrospectiveUpdateOne) SaveX(ctx context.Context) *Retrospective {
	node, err := ruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (ruo *RetrospectiveUpdateOne) Exec(ctx context.Context) error {
	_, err := ruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ruo *RetrospectiveUpdateOne) ExecX(ctx context.Context) {
	if err := ruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ruo *RetrospectiveUpdateOne) check() error {
	if v, ok := ruo.mutation.GetType(); ok {
		if err := retrospective.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Retrospective.type": %w`, err)}
		}
	}
	if v, ok := ruo.mutation.State(); ok {
		if err := retrospective.StateValidator(v); err != nil {
			return &ValidationError{Name: "state", err: fmt.Errorf(`ent: validator failed for field "Retrospective.state": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (ruo *RetrospectiveUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *RetrospectiveUpdateOne {
	ruo.modifiers = append(ruo.modifiers, modifiers...)
	return ruo
}

func (ruo *RetrospectiveUpdateOne) sqlSave(ctx context.Context) (_node *Retrospective, err error) {
	if err := ruo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(retrospective.Table, retrospective.Columns, sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID))
	id, ok := ruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Retrospective.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := ruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, retrospective.FieldID)
		for _, f := range fields {
			if !retrospective.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != retrospective.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := ruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ruo.mutation.DocumentName(); ok {
		_spec.SetField(retrospective.FieldDocumentName, field.TypeString, value)
	}
	if value, ok := ruo.mutation.GetType(); ok {
		_spec.SetField(retrospective.FieldType, field.TypeEnum, value)
	}
	if value, ok := ruo.mutation.State(); ok {
		_spec.SetField(retrospective.FieldState, field.TypeEnum, value)
	}
	if ruo.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   retrospective.IncidentTable,
			Columns: []string{retrospective.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ruo.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if ruo.mutation.DiscussionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ruo.mutation.RemovedDiscussionsIDs(); len(nodes) > 0 && !ruo.mutation.DiscussionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ruo.mutation.DiscussionsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(ruo.modifiers...)
	_node = &Retrospective{config: ruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, ruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{retrospective.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	ruo.mutation.done = true
	return _node, nil
}

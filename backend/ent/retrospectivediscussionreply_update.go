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
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospectivediscussion"
	"github.com/rezible/rezible/ent/retrospectivediscussionreply"
)

// RetrospectiveDiscussionReplyUpdate is the builder for updating RetrospectiveDiscussionReply entities.
type RetrospectiveDiscussionReplyUpdate struct {
	config
	hooks     []Hook
	mutation  *RetrospectiveDiscussionReplyMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the RetrospectiveDiscussionReplyUpdate builder.
func (rdru *RetrospectiveDiscussionReplyUpdate) Where(ps ...predicate.RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.Where(ps...)
	return rdru
}

// SetContent sets the "content" field.
func (rdru *RetrospectiveDiscussionReplyUpdate) SetContent(b []byte) *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.SetContent(b)
	return rdru
}

// SetDiscussionID sets the "discussion" edge to the RetrospectiveDiscussion entity by ID.
func (rdru *RetrospectiveDiscussionReplyUpdate) SetDiscussionID(id uuid.UUID) *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.SetDiscussionID(id)
	return rdru
}

// SetDiscussion sets the "discussion" edge to the RetrospectiveDiscussion entity.
func (rdru *RetrospectiveDiscussionReplyUpdate) SetDiscussion(r *RetrospectiveDiscussion) *RetrospectiveDiscussionReplyUpdate {
	return rdru.SetDiscussionID(r.ID)
}

// SetParentReplyID sets the "parent_reply" edge to the RetrospectiveDiscussionReply entity by ID.
func (rdru *RetrospectiveDiscussionReplyUpdate) SetParentReplyID(id uuid.UUID) *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.SetParentReplyID(id)
	return rdru
}

// SetNillableParentReplyID sets the "parent_reply" edge to the RetrospectiveDiscussionReply entity by ID if the given value is not nil.
func (rdru *RetrospectiveDiscussionReplyUpdate) SetNillableParentReplyID(id *uuid.UUID) *RetrospectiveDiscussionReplyUpdate {
	if id != nil {
		rdru = rdru.SetParentReplyID(*id)
	}
	return rdru
}

// SetParentReply sets the "parent_reply" edge to the RetrospectiveDiscussionReply entity.
func (rdru *RetrospectiveDiscussionReplyUpdate) SetParentReply(r *RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdate {
	return rdru.SetParentReplyID(r.ID)
}

// AddReplyIDs adds the "replies" edge to the RetrospectiveDiscussionReply entity by IDs.
func (rdru *RetrospectiveDiscussionReplyUpdate) AddReplyIDs(ids ...uuid.UUID) *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.AddReplyIDs(ids...)
	return rdru
}

// AddReplies adds the "replies" edges to the RetrospectiveDiscussionReply entity.
func (rdru *RetrospectiveDiscussionReplyUpdate) AddReplies(r ...*RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return rdru.AddReplyIDs(ids...)
}

// Mutation returns the RetrospectiveDiscussionReplyMutation object of the builder.
func (rdru *RetrospectiveDiscussionReplyUpdate) Mutation() *RetrospectiveDiscussionReplyMutation {
	return rdru.mutation
}

// ClearDiscussion clears the "discussion" edge to the RetrospectiveDiscussion entity.
func (rdru *RetrospectiveDiscussionReplyUpdate) ClearDiscussion() *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.ClearDiscussion()
	return rdru
}

// ClearParentReply clears the "parent_reply" edge to the RetrospectiveDiscussionReply entity.
func (rdru *RetrospectiveDiscussionReplyUpdate) ClearParentReply() *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.ClearParentReply()
	return rdru
}

// ClearReplies clears all "replies" edges to the RetrospectiveDiscussionReply entity.
func (rdru *RetrospectiveDiscussionReplyUpdate) ClearReplies() *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.ClearReplies()
	return rdru
}

// RemoveReplyIDs removes the "replies" edge to RetrospectiveDiscussionReply entities by IDs.
func (rdru *RetrospectiveDiscussionReplyUpdate) RemoveReplyIDs(ids ...uuid.UUID) *RetrospectiveDiscussionReplyUpdate {
	rdru.mutation.RemoveReplyIDs(ids...)
	return rdru
}

// RemoveReplies removes "replies" edges to RetrospectiveDiscussionReply entities.
func (rdru *RetrospectiveDiscussionReplyUpdate) RemoveReplies(r ...*RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return rdru.RemoveReplyIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (rdru *RetrospectiveDiscussionReplyUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, rdru.sqlSave, rdru.mutation, rdru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (rdru *RetrospectiveDiscussionReplyUpdate) SaveX(ctx context.Context) int {
	affected, err := rdru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (rdru *RetrospectiveDiscussionReplyUpdate) Exec(ctx context.Context) error {
	_, err := rdru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rdru *RetrospectiveDiscussionReplyUpdate) ExecX(ctx context.Context) {
	if err := rdru.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (rdru *RetrospectiveDiscussionReplyUpdate) check() error {
	if rdru.mutation.DiscussionCleared() && len(rdru.mutation.DiscussionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "RetrospectiveDiscussionReply.discussion"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (rdru *RetrospectiveDiscussionReplyUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *RetrospectiveDiscussionReplyUpdate {
	rdru.modifiers = append(rdru.modifiers, modifiers...)
	return rdru
}

func (rdru *RetrospectiveDiscussionReplyUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := rdru.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(retrospectivediscussionreply.Table, retrospectivediscussionreply.Columns, sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID))
	if ps := rdru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := rdru.mutation.Content(); ok {
		_spec.SetField(retrospectivediscussionreply.FieldContent, field.TypeBytes, value)
	}
	if rdru.mutation.DiscussionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   retrospectivediscussionreply.DiscussionTable,
			Columns: []string{retrospectivediscussionreply.DiscussionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussion.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdru.mutation.DiscussionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   retrospectivediscussionreply.DiscussionTable,
			Columns: []string{retrospectivediscussionreply.DiscussionColumn},
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
	if rdru.mutation.ParentReplyCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   retrospectivediscussionreply.ParentReplyTable,
			Columns: []string{retrospectivediscussionreply.ParentReplyColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdru.mutation.ParentReplyIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   retrospectivediscussionreply.ParentReplyTable,
			Columns: []string{retrospectivediscussionreply.ParentReplyColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if rdru.mutation.RepliesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   retrospectivediscussionreply.RepliesTable,
			Columns: []string{retrospectivediscussionreply.RepliesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdru.mutation.RemovedRepliesIDs(); len(nodes) > 0 && !rdru.mutation.RepliesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   retrospectivediscussionreply.RepliesTable,
			Columns: []string{retrospectivediscussionreply.RepliesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdru.mutation.RepliesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   retrospectivediscussionreply.RepliesTable,
			Columns: []string{retrospectivediscussionreply.RepliesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(rdru.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, rdru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{retrospectivediscussionreply.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	rdru.mutation.done = true
	return n, nil
}

// RetrospectiveDiscussionReplyUpdateOne is the builder for updating a single RetrospectiveDiscussionReply entity.
type RetrospectiveDiscussionReplyUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *RetrospectiveDiscussionReplyMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetContent sets the "content" field.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SetContent(b []byte) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.SetContent(b)
	return rdruo
}

// SetDiscussionID sets the "discussion" edge to the RetrospectiveDiscussion entity by ID.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SetDiscussionID(id uuid.UUID) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.SetDiscussionID(id)
	return rdruo
}

// SetDiscussion sets the "discussion" edge to the RetrospectiveDiscussion entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SetDiscussion(r *RetrospectiveDiscussion) *RetrospectiveDiscussionReplyUpdateOne {
	return rdruo.SetDiscussionID(r.ID)
}

// SetParentReplyID sets the "parent_reply" edge to the RetrospectiveDiscussionReply entity by ID.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SetParentReplyID(id uuid.UUID) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.SetParentReplyID(id)
	return rdruo
}

// SetNillableParentReplyID sets the "parent_reply" edge to the RetrospectiveDiscussionReply entity by ID if the given value is not nil.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SetNillableParentReplyID(id *uuid.UUID) *RetrospectiveDiscussionReplyUpdateOne {
	if id != nil {
		rdruo = rdruo.SetParentReplyID(*id)
	}
	return rdruo
}

// SetParentReply sets the "parent_reply" edge to the RetrospectiveDiscussionReply entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SetParentReply(r *RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdateOne {
	return rdruo.SetParentReplyID(r.ID)
}

// AddReplyIDs adds the "replies" edge to the RetrospectiveDiscussionReply entity by IDs.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) AddReplyIDs(ids ...uuid.UUID) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.AddReplyIDs(ids...)
	return rdruo
}

// AddReplies adds the "replies" edges to the RetrospectiveDiscussionReply entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) AddReplies(r ...*RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdateOne {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return rdruo.AddReplyIDs(ids...)
}

// Mutation returns the RetrospectiveDiscussionReplyMutation object of the builder.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) Mutation() *RetrospectiveDiscussionReplyMutation {
	return rdruo.mutation
}

// ClearDiscussion clears the "discussion" edge to the RetrospectiveDiscussion entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) ClearDiscussion() *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.ClearDiscussion()
	return rdruo
}

// ClearParentReply clears the "parent_reply" edge to the RetrospectiveDiscussionReply entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) ClearParentReply() *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.ClearParentReply()
	return rdruo
}

// ClearReplies clears all "replies" edges to the RetrospectiveDiscussionReply entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) ClearReplies() *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.ClearReplies()
	return rdruo
}

// RemoveReplyIDs removes the "replies" edge to RetrospectiveDiscussionReply entities by IDs.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) RemoveReplyIDs(ids ...uuid.UUID) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.RemoveReplyIDs(ids...)
	return rdruo
}

// RemoveReplies removes "replies" edges to RetrospectiveDiscussionReply entities.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) RemoveReplies(r ...*RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdateOne {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return rdruo.RemoveReplyIDs(ids...)
}

// Where appends a list predicates to the RetrospectiveDiscussionReplyUpdate builder.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) Where(ps ...predicate.RetrospectiveDiscussionReply) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.mutation.Where(ps...)
	return rdruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) Select(field string, fields ...string) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.fields = append([]string{field}, fields...)
	return rdruo
}

// Save executes the query and returns the updated RetrospectiveDiscussionReply entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) Save(ctx context.Context) (*RetrospectiveDiscussionReply, error) {
	return withHooks(ctx, rdruo.sqlSave, rdruo.mutation, rdruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) SaveX(ctx context.Context) *RetrospectiveDiscussionReply {
	node, err := rdruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) Exec(ctx context.Context) error {
	_, err := rdruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) ExecX(ctx context.Context) {
	if err := rdruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) check() error {
	if rdruo.mutation.DiscussionCleared() && len(rdruo.mutation.DiscussionIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "RetrospectiveDiscussionReply.discussion"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (rdruo *RetrospectiveDiscussionReplyUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *RetrospectiveDiscussionReplyUpdateOne {
	rdruo.modifiers = append(rdruo.modifiers, modifiers...)
	return rdruo
}

func (rdruo *RetrospectiveDiscussionReplyUpdateOne) sqlSave(ctx context.Context) (_node *RetrospectiveDiscussionReply, err error) {
	if err := rdruo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(retrospectivediscussionreply.Table, retrospectivediscussionreply.Columns, sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID))
	id, ok := rdruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "RetrospectiveDiscussionReply.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := rdruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, retrospectivediscussionreply.FieldID)
		for _, f := range fields {
			if !retrospectivediscussionreply.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != retrospectivediscussionreply.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := rdruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := rdruo.mutation.Content(); ok {
		_spec.SetField(retrospectivediscussionreply.FieldContent, field.TypeBytes, value)
	}
	if rdruo.mutation.DiscussionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   retrospectivediscussionreply.DiscussionTable,
			Columns: []string{retrospectivediscussionreply.DiscussionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussion.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdruo.mutation.DiscussionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   retrospectivediscussionreply.DiscussionTable,
			Columns: []string{retrospectivediscussionreply.DiscussionColumn},
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
	if rdruo.mutation.ParentReplyCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   retrospectivediscussionreply.ParentReplyTable,
			Columns: []string{retrospectivediscussionreply.ParentReplyColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdruo.mutation.ParentReplyIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   retrospectivediscussionreply.ParentReplyTable,
			Columns: []string{retrospectivediscussionreply.ParentReplyColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if rdruo.mutation.RepliesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   retrospectivediscussionreply.RepliesTable,
			Columns: []string{retrospectivediscussionreply.RepliesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdruo.mutation.RemovedRepliesIDs(); len(nodes) > 0 && !rdruo.mutation.RepliesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   retrospectivediscussionreply.RepliesTable,
			Columns: []string{retrospectivediscussionreply.RepliesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rdruo.mutation.RepliesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   retrospectivediscussionreply.RepliesTable,
			Columns: []string{retrospectivediscussionreply.RepliesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivediscussionreply.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(rdruo.modifiers...)
	_node = &RetrospectiveDiscussionReply{config: rdruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, rdruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{retrospectivediscussionreply.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	rdruo.mutation.done = true
	return _node, nil
}

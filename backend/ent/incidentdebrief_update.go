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
	"github.com/rezible/rezible/ent/incidentdebrief"
	"github.com/rezible/rezible/ent/incidentdebriefmessage"
	"github.com/rezible/rezible/ent/incidentdebriefsuggestion"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/user"
)

// IncidentDebriefUpdate is the builder for updating IncidentDebrief entities.
type IncidentDebriefUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentDebriefMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentDebriefUpdate builder.
func (idu *IncidentDebriefUpdate) Where(ps ...predicate.IncidentDebrief) *IncidentDebriefUpdate {
	idu.mutation.Where(ps...)
	return idu
}

// SetIncidentID sets the "incident_id" field.
func (idu *IncidentDebriefUpdate) SetIncidentID(u uuid.UUID) *IncidentDebriefUpdate {
	idu.mutation.SetIncidentID(u)
	return idu
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (idu *IncidentDebriefUpdate) SetNillableIncidentID(u *uuid.UUID) *IncidentDebriefUpdate {
	if u != nil {
		idu.SetIncidentID(*u)
	}
	return idu
}

// SetUserID sets the "user_id" field.
func (idu *IncidentDebriefUpdate) SetUserID(u uuid.UUID) *IncidentDebriefUpdate {
	idu.mutation.SetUserID(u)
	return idu
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (idu *IncidentDebriefUpdate) SetNillableUserID(u *uuid.UUID) *IncidentDebriefUpdate {
	if u != nil {
		idu.SetUserID(*u)
	}
	return idu
}

// SetRequired sets the "required" field.
func (idu *IncidentDebriefUpdate) SetRequired(b bool) *IncidentDebriefUpdate {
	idu.mutation.SetRequired(b)
	return idu
}

// SetNillableRequired sets the "required" field if the given value is not nil.
func (idu *IncidentDebriefUpdate) SetNillableRequired(b *bool) *IncidentDebriefUpdate {
	if b != nil {
		idu.SetRequired(*b)
	}
	return idu
}

// SetStarted sets the "started" field.
func (idu *IncidentDebriefUpdate) SetStarted(b bool) *IncidentDebriefUpdate {
	idu.mutation.SetStarted(b)
	return idu
}

// SetNillableStarted sets the "started" field if the given value is not nil.
func (idu *IncidentDebriefUpdate) SetNillableStarted(b *bool) *IncidentDebriefUpdate {
	if b != nil {
		idu.SetStarted(*b)
	}
	return idu
}

// SetIncident sets the "incident" edge to the Incident entity.
func (idu *IncidentDebriefUpdate) SetIncident(i *Incident) *IncidentDebriefUpdate {
	return idu.SetIncidentID(i.ID)
}

// SetUser sets the "user" edge to the User entity.
func (idu *IncidentDebriefUpdate) SetUser(u *User) *IncidentDebriefUpdate {
	return idu.SetUserID(u.ID)
}

// AddMessageIDs adds the "messages" edge to the IncidentDebriefMessage entity by IDs.
func (idu *IncidentDebriefUpdate) AddMessageIDs(ids ...uuid.UUID) *IncidentDebriefUpdate {
	idu.mutation.AddMessageIDs(ids...)
	return idu
}

// AddMessages adds the "messages" edges to the IncidentDebriefMessage entity.
func (idu *IncidentDebriefUpdate) AddMessages(i ...*IncidentDebriefMessage) *IncidentDebriefUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idu.AddMessageIDs(ids...)
}

// AddSuggestionIDs adds the "suggestions" edge to the IncidentDebriefSuggestion entity by IDs.
func (idu *IncidentDebriefUpdate) AddSuggestionIDs(ids ...uuid.UUID) *IncidentDebriefUpdate {
	idu.mutation.AddSuggestionIDs(ids...)
	return idu
}

// AddSuggestions adds the "suggestions" edges to the IncidentDebriefSuggestion entity.
func (idu *IncidentDebriefUpdate) AddSuggestions(i ...*IncidentDebriefSuggestion) *IncidentDebriefUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idu.AddSuggestionIDs(ids...)
}

// Mutation returns the IncidentDebriefMutation object of the builder.
func (idu *IncidentDebriefUpdate) Mutation() *IncidentDebriefMutation {
	return idu.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (idu *IncidentDebriefUpdate) ClearIncident() *IncidentDebriefUpdate {
	idu.mutation.ClearIncident()
	return idu
}

// ClearUser clears the "user" edge to the User entity.
func (idu *IncidentDebriefUpdate) ClearUser() *IncidentDebriefUpdate {
	idu.mutation.ClearUser()
	return idu
}

// ClearMessages clears all "messages" edges to the IncidentDebriefMessage entity.
func (idu *IncidentDebriefUpdate) ClearMessages() *IncidentDebriefUpdate {
	idu.mutation.ClearMessages()
	return idu
}

// RemoveMessageIDs removes the "messages" edge to IncidentDebriefMessage entities by IDs.
func (idu *IncidentDebriefUpdate) RemoveMessageIDs(ids ...uuid.UUID) *IncidentDebriefUpdate {
	idu.mutation.RemoveMessageIDs(ids...)
	return idu
}

// RemoveMessages removes "messages" edges to IncidentDebriefMessage entities.
func (idu *IncidentDebriefUpdate) RemoveMessages(i ...*IncidentDebriefMessage) *IncidentDebriefUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idu.RemoveMessageIDs(ids...)
}

// ClearSuggestions clears all "suggestions" edges to the IncidentDebriefSuggestion entity.
func (idu *IncidentDebriefUpdate) ClearSuggestions() *IncidentDebriefUpdate {
	idu.mutation.ClearSuggestions()
	return idu
}

// RemoveSuggestionIDs removes the "suggestions" edge to IncidentDebriefSuggestion entities by IDs.
func (idu *IncidentDebriefUpdate) RemoveSuggestionIDs(ids ...uuid.UUID) *IncidentDebriefUpdate {
	idu.mutation.RemoveSuggestionIDs(ids...)
	return idu
}

// RemoveSuggestions removes "suggestions" edges to IncidentDebriefSuggestion entities.
func (idu *IncidentDebriefUpdate) RemoveSuggestions(i ...*IncidentDebriefSuggestion) *IncidentDebriefUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idu.RemoveSuggestionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (idu *IncidentDebriefUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, idu.sqlSave, idu.mutation, idu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (idu *IncidentDebriefUpdate) SaveX(ctx context.Context) int {
	affected, err := idu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (idu *IncidentDebriefUpdate) Exec(ctx context.Context) error {
	_, err := idu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (idu *IncidentDebriefUpdate) ExecX(ctx context.Context) {
	if err := idu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (idu *IncidentDebriefUpdate) check() error {
	if idu.mutation.IncidentCleared() && len(idu.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentDebrief.incident"`)
	}
	if idu.mutation.UserCleared() && len(idu.mutation.UserIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentDebrief.user"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (idu *IncidentDebriefUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentDebriefUpdate {
	idu.modifiers = append(idu.modifiers, modifiers...)
	return idu
}

func (idu *IncidentDebriefUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := idu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentdebrief.Table, incidentdebrief.Columns, sqlgraph.NewFieldSpec(incidentdebrief.FieldID, field.TypeUUID))
	if ps := idu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := idu.mutation.Required(); ok {
		_spec.SetField(incidentdebrief.FieldRequired, field.TypeBool, value)
	}
	if value, ok := idu.mutation.Started(); ok {
		_spec.SetField(incidentdebrief.FieldStarted, field.TypeBool, value)
	}
	if idu.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.IncidentTable,
			Columns: []string{incidentdebrief.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := idu.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.IncidentTable,
			Columns: []string{incidentdebrief.IncidentColumn},
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
	if idu.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.UserTable,
			Columns: []string{incidentdebrief.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := idu.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.UserTable,
			Columns: []string{incidentdebrief.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if idu.mutation.MessagesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.MessagesTable,
			Columns: []string{incidentdebrief.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := idu.mutation.RemovedMessagesIDs(); len(nodes) > 0 && !idu.mutation.MessagesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.MessagesTable,
			Columns: []string{incidentdebrief.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := idu.mutation.MessagesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.MessagesTable,
			Columns: []string{incidentdebrief.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if idu.mutation.SuggestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.SuggestionsTable,
			Columns: []string{incidentdebrief.SuggestionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := idu.mutation.RemovedSuggestionsIDs(); len(nodes) > 0 && !idu.mutation.SuggestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.SuggestionsTable,
			Columns: []string{incidentdebrief.SuggestionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := idu.mutation.SuggestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.SuggestionsTable,
			Columns: []string{incidentdebrief.SuggestionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(idu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, idu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentdebrief.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	idu.mutation.done = true
	return n, nil
}

// IncidentDebriefUpdateOne is the builder for updating a single IncidentDebrief entity.
type IncidentDebriefUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentDebriefMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetIncidentID sets the "incident_id" field.
func (iduo *IncidentDebriefUpdateOne) SetIncidentID(u uuid.UUID) *IncidentDebriefUpdateOne {
	iduo.mutation.SetIncidentID(u)
	return iduo
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (iduo *IncidentDebriefUpdateOne) SetNillableIncidentID(u *uuid.UUID) *IncidentDebriefUpdateOne {
	if u != nil {
		iduo.SetIncidentID(*u)
	}
	return iduo
}

// SetUserID sets the "user_id" field.
func (iduo *IncidentDebriefUpdateOne) SetUserID(u uuid.UUID) *IncidentDebriefUpdateOne {
	iduo.mutation.SetUserID(u)
	return iduo
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (iduo *IncidentDebriefUpdateOne) SetNillableUserID(u *uuid.UUID) *IncidentDebriefUpdateOne {
	if u != nil {
		iduo.SetUserID(*u)
	}
	return iduo
}

// SetRequired sets the "required" field.
func (iduo *IncidentDebriefUpdateOne) SetRequired(b bool) *IncidentDebriefUpdateOne {
	iduo.mutation.SetRequired(b)
	return iduo
}

// SetNillableRequired sets the "required" field if the given value is not nil.
func (iduo *IncidentDebriefUpdateOne) SetNillableRequired(b *bool) *IncidentDebriefUpdateOne {
	if b != nil {
		iduo.SetRequired(*b)
	}
	return iduo
}

// SetStarted sets the "started" field.
func (iduo *IncidentDebriefUpdateOne) SetStarted(b bool) *IncidentDebriefUpdateOne {
	iduo.mutation.SetStarted(b)
	return iduo
}

// SetNillableStarted sets the "started" field if the given value is not nil.
func (iduo *IncidentDebriefUpdateOne) SetNillableStarted(b *bool) *IncidentDebriefUpdateOne {
	if b != nil {
		iduo.SetStarted(*b)
	}
	return iduo
}

// SetIncident sets the "incident" edge to the Incident entity.
func (iduo *IncidentDebriefUpdateOne) SetIncident(i *Incident) *IncidentDebriefUpdateOne {
	return iduo.SetIncidentID(i.ID)
}

// SetUser sets the "user" edge to the User entity.
func (iduo *IncidentDebriefUpdateOne) SetUser(u *User) *IncidentDebriefUpdateOne {
	return iduo.SetUserID(u.ID)
}

// AddMessageIDs adds the "messages" edge to the IncidentDebriefMessage entity by IDs.
func (iduo *IncidentDebriefUpdateOne) AddMessageIDs(ids ...uuid.UUID) *IncidentDebriefUpdateOne {
	iduo.mutation.AddMessageIDs(ids...)
	return iduo
}

// AddMessages adds the "messages" edges to the IncidentDebriefMessage entity.
func (iduo *IncidentDebriefUpdateOne) AddMessages(i ...*IncidentDebriefMessage) *IncidentDebriefUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iduo.AddMessageIDs(ids...)
}

// AddSuggestionIDs adds the "suggestions" edge to the IncidentDebriefSuggestion entity by IDs.
func (iduo *IncidentDebriefUpdateOne) AddSuggestionIDs(ids ...uuid.UUID) *IncidentDebriefUpdateOne {
	iduo.mutation.AddSuggestionIDs(ids...)
	return iduo
}

// AddSuggestions adds the "suggestions" edges to the IncidentDebriefSuggestion entity.
func (iduo *IncidentDebriefUpdateOne) AddSuggestions(i ...*IncidentDebriefSuggestion) *IncidentDebriefUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iduo.AddSuggestionIDs(ids...)
}

// Mutation returns the IncidentDebriefMutation object of the builder.
func (iduo *IncidentDebriefUpdateOne) Mutation() *IncidentDebriefMutation {
	return iduo.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (iduo *IncidentDebriefUpdateOne) ClearIncident() *IncidentDebriefUpdateOne {
	iduo.mutation.ClearIncident()
	return iduo
}

// ClearUser clears the "user" edge to the User entity.
func (iduo *IncidentDebriefUpdateOne) ClearUser() *IncidentDebriefUpdateOne {
	iduo.mutation.ClearUser()
	return iduo
}

// ClearMessages clears all "messages" edges to the IncidentDebriefMessage entity.
func (iduo *IncidentDebriefUpdateOne) ClearMessages() *IncidentDebriefUpdateOne {
	iduo.mutation.ClearMessages()
	return iduo
}

// RemoveMessageIDs removes the "messages" edge to IncidentDebriefMessage entities by IDs.
func (iduo *IncidentDebriefUpdateOne) RemoveMessageIDs(ids ...uuid.UUID) *IncidentDebriefUpdateOne {
	iduo.mutation.RemoveMessageIDs(ids...)
	return iduo
}

// RemoveMessages removes "messages" edges to IncidentDebriefMessage entities.
func (iduo *IncidentDebriefUpdateOne) RemoveMessages(i ...*IncidentDebriefMessage) *IncidentDebriefUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iduo.RemoveMessageIDs(ids...)
}

// ClearSuggestions clears all "suggestions" edges to the IncidentDebriefSuggestion entity.
func (iduo *IncidentDebriefUpdateOne) ClearSuggestions() *IncidentDebriefUpdateOne {
	iduo.mutation.ClearSuggestions()
	return iduo
}

// RemoveSuggestionIDs removes the "suggestions" edge to IncidentDebriefSuggestion entities by IDs.
func (iduo *IncidentDebriefUpdateOne) RemoveSuggestionIDs(ids ...uuid.UUID) *IncidentDebriefUpdateOne {
	iduo.mutation.RemoveSuggestionIDs(ids...)
	return iduo
}

// RemoveSuggestions removes "suggestions" edges to IncidentDebriefSuggestion entities.
func (iduo *IncidentDebriefUpdateOne) RemoveSuggestions(i ...*IncidentDebriefSuggestion) *IncidentDebriefUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iduo.RemoveSuggestionIDs(ids...)
}

// Where appends a list predicates to the IncidentDebriefUpdate builder.
func (iduo *IncidentDebriefUpdateOne) Where(ps ...predicate.IncidentDebrief) *IncidentDebriefUpdateOne {
	iduo.mutation.Where(ps...)
	return iduo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (iduo *IncidentDebriefUpdateOne) Select(field string, fields ...string) *IncidentDebriefUpdateOne {
	iduo.fields = append([]string{field}, fields...)
	return iduo
}

// Save executes the query and returns the updated IncidentDebrief entity.
func (iduo *IncidentDebriefUpdateOne) Save(ctx context.Context) (*IncidentDebrief, error) {
	return withHooks(ctx, iduo.sqlSave, iduo.mutation, iduo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iduo *IncidentDebriefUpdateOne) SaveX(ctx context.Context) *IncidentDebrief {
	node, err := iduo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (iduo *IncidentDebriefUpdateOne) Exec(ctx context.Context) error {
	_, err := iduo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iduo *IncidentDebriefUpdateOne) ExecX(ctx context.Context) {
	if err := iduo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iduo *IncidentDebriefUpdateOne) check() error {
	if iduo.mutation.IncidentCleared() && len(iduo.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentDebrief.incident"`)
	}
	if iduo.mutation.UserCleared() && len(iduo.mutation.UserIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentDebrief.user"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iduo *IncidentDebriefUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentDebriefUpdateOne {
	iduo.modifiers = append(iduo.modifiers, modifiers...)
	return iduo
}

func (iduo *IncidentDebriefUpdateOne) sqlSave(ctx context.Context) (_node *IncidentDebrief, err error) {
	if err := iduo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentdebrief.Table, incidentdebrief.Columns, sqlgraph.NewFieldSpec(incidentdebrief.FieldID, field.TypeUUID))
	id, ok := iduo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentDebrief.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := iduo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentdebrief.FieldID)
		for _, f := range fields {
			if !incidentdebrief.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidentdebrief.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := iduo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iduo.mutation.Required(); ok {
		_spec.SetField(incidentdebrief.FieldRequired, field.TypeBool, value)
	}
	if value, ok := iduo.mutation.Started(); ok {
		_spec.SetField(incidentdebrief.FieldStarted, field.TypeBool, value)
	}
	if iduo.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.IncidentTable,
			Columns: []string{incidentdebrief.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iduo.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.IncidentTable,
			Columns: []string{incidentdebrief.IncidentColumn},
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
	if iduo.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.UserTable,
			Columns: []string{incidentdebrief.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iduo.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentdebrief.UserTable,
			Columns: []string{incidentdebrief.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if iduo.mutation.MessagesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.MessagesTable,
			Columns: []string{incidentdebrief.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iduo.mutation.RemovedMessagesIDs(); len(nodes) > 0 && !iduo.mutation.MessagesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.MessagesTable,
			Columns: []string{incidentdebrief.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iduo.mutation.MessagesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.MessagesTable,
			Columns: []string{incidentdebrief.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if iduo.mutation.SuggestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.SuggestionsTable,
			Columns: []string{incidentdebrief.SuggestionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iduo.mutation.RemovedSuggestionsIDs(); len(nodes) > 0 && !iduo.mutation.SuggestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.SuggestionsTable,
			Columns: []string{incidentdebrief.SuggestionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iduo.mutation.SuggestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentdebrief.SuggestionsTable,
			Columns: []string{incidentdebrief.SuggestionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefsuggestion.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(iduo.modifiers...)
	_node = &IncidentDebrief{config: iduo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, iduo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentdebrief.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	iduo.mutation.done = true
	return _node, nil
}

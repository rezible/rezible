// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentdebriefquestion"
	"github.com/rezible/rezible/ent/incidenttag"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentTagUpdate is the builder for updating IncidentTag entities.
type IncidentTagUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentTagMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentTagUpdate builder.
func (itu *IncidentTagUpdate) Where(ps ...predicate.IncidentTag) *IncidentTagUpdate {
	itu.mutation.Where(ps...)
	return itu
}

// SetArchiveTime sets the "archive_time" field.
func (itu *IncidentTagUpdate) SetArchiveTime(t time.Time) *IncidentTagUpdate {
	itu.mutation.SetArchiveTime(t)
	return itu
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (itu *IncidentTagUpdate) SetNillableArchiveTime(t *time.Time) *IncidentTagUpdate {
	if t != nil {
		itu.SetArchiveTime(*t)
	}
	return itu
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (itu *IncidentTagUpdate) ClearArchiveTime() *IncidentTagUpdate {
	itu.mutation.ClearArchiveTime()
	return itu
}

// SetKey sets the "key" field.
func (itu *IncidentTagUpdate) SetKey(s string) *IncidentTagUpdate {
	itu.mutation.SetKey(s)
	return itu
}

// SetNillableKey sets the "key" field if the given value is not nil.
func (itu *IncidentTagUpdate) SetNillableKey(s *string) *IncidentTagUpdate {
	if s != nil {
		itu.SetKey(*s)
	}
	return itu
}

// SetValue sets the "value" field.
func (itu *IncidentTagUpdate) SetValue(s string) *IncidentTagUpdate {
	itu.mutation.SetValue(s)
	return itu
}

// SetNillableValue sets the "value" field if the given value is not nil.
func (itu *IncidentTagUpdate) SetNillableValue(s *string) *IncidentTagUpdate {
	if s != nil {
		itu.SetValue(*s)
	}
	return itu
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (itu *IncidentTagUpdate) AddIncidentIDs(ids ...uuid.UUID) *IncidentTagUpdate {
	itu.mutation.AddIncidentIDs(ids...)
	return itu
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (itu *IncidentTagUpdate) AddIncidents(i ...*Incident) *IncidentTagUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.AddIncidentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (itu *IncidentTagUpdate) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTagUpdate {
	itu.mutation.AddDebriefQuestionIDs(ids...)
	return itu
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (itu *IncidentTagUpdate) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTagUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentTagMutation object of the builder.
func (itu *IncidentTagUpdate) Mutation() *IncidentTagMutation {
	return itu.mutation
}

// ClearIncidents clears all "incidents" edges to the Incident entity.
func (itu *IncidentTagUpdate) ClearIncidents() *IncidentTagUpdate {
	itu.mutation.ClearIncidents()
	return itu
}

// RemoveIncidentIDs removes the "incidents" edge to Incident entities by IDs.
func (itu *IncidentTagUpdate) RemoveIncidentIDs(ids ...uuid.UUID) *IncidentTagUpdate {
	itu.mutation.RemoveIncidentIDs(ids...)
	return itu
}

// RemoveIncidents removes "incidents" edges to Incident entities.
func (itu *IncidentTagUpdate) RemoveIncidents(i ...*Incident) *IncidentTagUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.RemoveIncidentIDs(ids...)
}

// ClearDebriefQuestions clears all "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (itu *IncidentTagUpdate) ClearDebriefQuestions() *IncidentTagUpdate {
	itu.mutation.ClearDebriefQuestions()
	return itu
}

// RemoveDebriefQuestionIDs removes the "debrief_questions" edge to IncidentDebriefQuestion entities by IDs.
func (itu *IncidentTagUpdate) RemoveDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTagUpdate {
	itu.mutation.RemoveDebriefQuestionIDs(ids...)
	return itu
}

// RemoveDebriefQuestions removes "debrief_questions" edges to IncidentDebriefQuestion entities.
func (itu *IncidentTagUpdate) RemoveDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTagUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.RemoveDebriefQuestionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (itu *IncidentTagUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, itu.sqlSave, itu.mutation, itu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (itu *IncidentTagUpdate) SaveX(ctx context.Context) int {
	affected, err := itu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (itu *IncidentTagUpdate) Exec(ctx context.Context) error {
	_, err := itu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (itu *IncidentTagUpdate) ExecX(ctx context.Context) {
	if err := itu.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (itu *IncidentTagUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentTagUpdate {
	itu.modifiers = append(itu.modifiers, modifiers...)
	return itu
}

func (itu *IncidentTagUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(incidenttag.Table, incidenttag.Columns, sqlgraph.NewFieldSpec(incidenttag.FieldID, field.TypeUUID))
	if ps := itu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := itu.mutation.ArchiveTime(); ok {
		_spec.SetField(incidenttag.FieldArchiveTime, field.TypeTime, value)
	}
	if itu.mutation.ArchiveTimeCleared() {
		_spec.ClearField(incidenttag.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := itu.mutation.Key(); ok {
		_spec.SetField(incidenttag.FieldKey, field.TypeString, value)
	}
	if value, ok := itu.mutation.Value(); ok {
		_spec.SetField(incidenttag.FieldValue, field.TypeString, value)
	}
	if itu.mutation.IncidentsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := itu.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !itu.mutation.IncidentsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := itu.mutation.IncidentsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if itu.mutation.DebriefQuestionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := itu.mutation.RemovedDebriefQuestionsIDs(); len(nodes) > 0 && !itu.mutation.DebriefQuestionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := itu.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(itu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, itu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidenttag.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	itu.mutation.done = true
	return n, nil
}

// IncidentTagUpdateOne is the builder for updating a single IncidentTag entity.
type IncidentTagUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentTagMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetArchiveTime sets the "archive_time" field.
func (ituo *IncidentTagUpdateOne) SetArchiveTime(t time.Time) *IncidentTagUpdateOne {
	ituo.mutation.SetArchiveTime(t)
	return ituo
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (ituo *IncidentTagUpdateOne) SetNillableArchiveTime(t *time.Time) *IncidentTagUpdateOne {
	if t != nil {
		ituo.SetArchiveTime(*t)
	}
	return ituo
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (ituo *IncidentTagUpdateOne) ClearArchiveTime() *IncidentTagUpdateOne {
	ituo.mutation.ClearArchiveTime()
	return ituo
}

// SetKey sets the "key" field.
func (ituo *IncidentTagUpdateOne) SetKey(s string) *IncidentTagUpdateOne {
	ituo.mutation.SetKey(s)
	return ituo
}

// SetNillableKey sets the "key" field if the given value is not nil.
func (ituo *IncidentTagUpdateOne) SetNillableKey(s *string) *IncidentTagUpdateOne {
	if s != nil {
		ituo.SetKey(*s)
	}
	return ituo
}

// SetValue sets the "value" field.
func (ituo *IncidentTagUpdateOne) SetValue(s string) *IncidentTagUpdateOne {
	ituo.mutation.SetValue(s)
	return ituo
}

// SetNillableValue sets the "value" field if the given value is not nil.
func (ituo *IncidentTagUpdateOne) SetNillableValue(s *string) *IncidentTagUpdateOne {
	if s != nil {
		ituo.SetValue(*s)
	}
	return ituo
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (ituo *IncidentTagUpdateOne) AddIncidentIDs(ids ...uuid.UUID) *IncidentTagUpdateOne {
	ituo.mutation.AddIncidentIDs(ids...)
	return ituo
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (ituo *IncidentTagUpdateOne) AddIncidents(i ...*Incident) *IncidentTagUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.AddIncidentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (ituo *IncidentTagUpdateOne) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTagUpdateOne {
	ituo.mutation.AddDebriefQuestionIDs(ids...)
	return ituo
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (ituo *IncidentTagUpdateOne) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTagUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentTagMutation object of the builder.
func (ituo *IncidentTagUpdateOne) Mutation() *IncidentTagMutation {
	return ituo.mutation
}

// ClearIncidents clears all "incidents" edges to the Incident entity.
func (ituo *IncidentTagUpdateOne) ClearIncidents() *IncidentTagUpdateOne {
	ituo.mutation.ClearIncidents()
	return ituo
}

// RemoveIncidentIDs removes the "incidents" edge to Incident entities by IDs.
func (ituo *IncidentTagUpdateOne) RemoveIncidentIDs(ids ...uuid.UUID) *IncidentTagUpdateOne {
	ituo.mutation.RemoveIncidentIDs(ids...)
	return ituo
}

// RemoveIncidents removes "incidents" edges to Incident entities.
func (ituo *IncidentTagUpdateOne) RemoveIncidents(i ...*Incident) *IncidentTagUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.RemoveIncidentIDs(ids...)
}

// ClearDebriefQuestions clears all "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (ituo *IncidentTagUpdateOne) ClearDebriefQuestions() *IncidentTagUpdateOne {
	ituo.mutation.ClearDebriefQuestions()
	return ituo
}

// RemoveDebriefQuestionIDs removes the "debrief_questions" edge to IncidentDebriefQuestion entities by IDs.
func (ituo *IncidentTagUpdateOne) RemoveDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTagUpdateOne {
	ituo.mutation.RemoveDebriefQuestionIDs(ids...)
	return ituo
}

// RemoveDebriefQuestions removes "debrief_questions" edges to IncidentDebriefQuestion entities.
func (ituo *IncidentTagUpdateOne) RemoveDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTagUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.RemoveDebriefQuestionIDs(ids...)
}

// Where appends a list predicates to the IncidentTagUpdate builder.
func (ituo *IncidentTagUpdateOne) Where(ps ...predicate.IncidentTag) *IncidentTagUpdateOne {
	ituo.mutation.Where(ps...)
	return ituo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (ituo *IncidentTagUpdateOne) Select(field string, fields ...string) *IncidentTagUpdateOne {
	ituo.fields = append([]string{field}, fields...)
	return ituo
}

// Save executes the query and returns the updated IncidentTag entity.
func (ituo *IncidentTagUpdateOne) Save(ctx context.Context) (*IncidentTag, error) {
	return withHooks(ctx, ituo.sqlSave, ituo.mutation, ituo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ituo *IncidentTagUpdateOne) SaveX(ctx context.Context) *IncidentTag {
	node, err := ituo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (ituo *IncidentTagUpdateOne) Exec(ctx context.Context) error {
	_, err := ituo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ituo *IncidentTagUpdateOne) ExecX(ctx context.Context) {
	if err := ituo.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (ituo *IncidentTagUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentTagUpdateOne {
	ituo.modifiers = append(ituo.modifiers, modifiers...)
	return ituo
}

func (ituo *IncidentTagUpdateOne) sqlSave(ctx context.Context) (_node *IncidentTag, err error) {
	_spec := sqlgraph.NewUpdateSpec(incidenttag.Table, incidenttag.Columns, sqlgraph.NewFieldSpec(incidenttag.FieldID, field.TypeUUID))
	id, ok := ituo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentTag.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := ituo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidenttag.FieldID)
		for _, f := range fields {
			if !incidenttag.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidenttag.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := ituo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ituo.mutation.ArchiveTime(); ok {
		_spec.SetField(incidenttag.FieldArchiveTime, field.TypeTime, value)
	}
	if ituo.mutation.ArchiveTimeCleared() {
		_spec.ClearField(incidenttag.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := ituo.mutation.Key(); ok {
		_spec.SetField(incidenttag.FieldKey, field.TypeString, value)
	}
	if value, ok := ituo.mutation.Value(); ok {
		_spec.SetField(incidenttag.FieldValue, field.TypeString, value)
	}
	if ituo.mutation.IncidentsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ituo.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !ituo.mutation.IncidentsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ituo.mutation.IncidentsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if ituo.mutation.DebriefQuestionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ituo.mutation.RemovedDebriefQuestionsIDs(); len(nodes) > 0 && !ituo.mutation.DebriefQuestionsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ituo.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(ituo.modifiers...)
	_node = &IncidentTag{config: ituo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, ituo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidenttag.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	ituo.mutation.done = true
	return _node, nil
}

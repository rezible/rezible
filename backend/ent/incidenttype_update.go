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
	"github.com/rezible/rezible/ent/incidenttype"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentTypeUpdate is the builder for updating IncidentType entities.
type IncidentTypeUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentTypeMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentTypeUpdate builder.
func (itu *IncidentTypeUpdate) Where(ps ...predicate.IncidentType) *IncidentTypeUpdate {
	itu.mutation.Where(ps...)
	return itu
}

// SetArchiveTime sets the "archive_time" field.
func (itu *IncidentTypeUpdate) SetArchiveTime(t time.Time) *IncidentTypeUpdate {
	itu.mutation.SetArchiveTime(t)
	return itu
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (itu *IncidentTypeUpdate) SetNillableArchiveTime(t *time.Time) *IncidentTypeUpdate {
	if t != nil {
		itu.SetArchiveTime(*t)
	}
	return itu
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (itu *IncidentTypeUpdate) ClearArchiveTime() *IncidentTypeUpdate {
	itu.mutation.ClearArchiveTime()
	return itu
}

// SetName sets the "name" field.
func (itu *IncidentTypeUpdate) SetName(s string) *IncidentTypeUpdate {
	itu.mutation.SetName(s)
	return itu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (itu *IncidentTypeUpdate) SetNillableName(s *string) *IncidentTypeUpdate {
	if s != nil {
		itu.SetName(*s)
	}
	return itu
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (itu *IncidentTypeUpdate) AddIncidentIDs(ids ...uuid.UUID) *IncidentTypeUpdate {
	itu.mutation.AddIncidentIDs(ids...)
	return itu
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (itu *IncidentTypeUpdate) AddIncidents(i ...*Incident) *IncidentTypeUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.AddIncidentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (itu *IncidentTypeUpdate) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTypeUpdate {
	itu.mutation.AddDebriefQuestionIDs(ids...)
	return itu
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (itu *IncidentTypeUpdate) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTypeUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentTypeMutation object of the builder.
func (itu *IncidentTypeUpdate) Mutation() *IncidentTypeMutation {
	return itu.mutation
}

// ClearIncidents clears all "incidents" edges to the Incident entity.
func (itu *IncidentTypeUpdate) ClearIncidents() *IncidentTypeUpdate {
	itu.mutation.ClearIncidents()
	return itu
}

// RemoveIncidentIDs removes the "incidents" edge to Incident entities by IDs.
func (itu *IncidentTypeUpdate) RemoveIncidentIDs(ids ...uuid.UUID) *IncidentTypeUpdate {
	itu.mutation.RemoveIncidentIDs(ids...)
	return itu
}

// RemoveIncidents removes "incidents" edges to Incident entities.
func (itu *IncidentTypeUpdate) RemoveIncidents(i ...*Incident) *IncidentTypeUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.RemoveIncidentIDs(ids...)
}

// ClearDebriefQuestions clears all "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (itu *IncidentTypeUpdate) ClearDebriefQuestions() *IncidentTypeUpdate {
	itu.mutation.ClearDebriefQuestions()
	return itu
}

// RemoveDebriefQuestionIDs removes the "debrief_questions" edge to IncidentDebriefQuestion entities by IDs.
func (itu *IncidentTypeUpdate) RemoveDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTypeUpdate {
	itu.mutation.RemoveDebriefQuestionIDs(ids...)
	return itu
}

// RemoveDebriefQuestions removes "debrief_questions" edges to IncidentDebriefQuestion entities.
func (itu *IncidentTypeUpdate) RemoveDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTypeUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return itu.RemoveDebriefQuestionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (itu *IncidentTypeUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, itu.sqlSave, itu.mutation, itu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (itu *IncidentTypeUpdate) SaveX(ctx context.Context) int {
	affected, err := itu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (itu *IncidentTypeUpdate) Exec(ctx context.Context) error {
	_, err := itu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (itu *IncidentTypeUpdate) ExecX(ctx context.Context) {
	if err := itu.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (itu *IncidentTypeUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentTypeUpdate {
	itu.modifiers = append(itu.modifiers, modifiers...)
	return itu
}

func (itu *IncidentTypeUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(incidenttype.Table, incidenttype.Columns, sqlgraph.NewFieldSpec(incidenttype.FieldID, field.TypeUUID))
	if ps := itu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := itu.mutation.ArchiveTime(); ok {
		_spec.SetField(incidenttype.FieldArchiveTime, field.TypeTime, value)
	}
	if itu.mutation.ArchiveTimeCleared() {
		_spec.ClearField(incidenttype.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := itu.mutation.Name(); ok {
		_spec.SetField(incidenttype.FieldName, field.TypeString, value)
	}
	if itu.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidenttype.IncidentsTable,
			Columns: []string{incidenttype.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := itu.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !itu.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidenttype.IncidentsTable,
			Columns: []string{incidenttype.IncidentsColumn},
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
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidenttype.IncidentsTable,
			Columns: []string{incidenttype.IncidentsColumn},
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
			Table:   incidenttype.DebriefQuestionsTable,
			Columns: incidenttype.DebriefQuestionsPrimaryKey,
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
			Table:   incidenttype.DebriefQuestionsTable,
			Columns: incidenttype.DebriefQuestionsPrimaryKey,
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
			Table:   incidenttype.DebriefQuestionsTable,
			Columns: incidenttype.DebriefQuestionsPrimaryKey,
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
			err = &NotFoundError{incidenttype.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	itu.mutation.done = true
	return n, nil
}

// IncidentTypeUpdateOne is the builder for updating a single IncidentType entity.
type IncidentTypeUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentTypeMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetArchiveTime sets the "archive_time" field.
func (ituo *IncidentTypeUpdateOne) SetArchiveTime(t time.Time) *IncidentTypeUpdateOne {
	ituo.mutation.SetArchiveTime(t)
	return ituo
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (ituo *IncidentTypeUpdateOne) SetNillableArchiveTime(t *time.Time) *IncidentTypeUpdateOne {
	if t != nil {
		ituo.SetArchiveTime(*t)
	}
	return ituo
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (ituo *IncidentTypeUpdateOne) ClearArchiveTime() *IncidentTypeUpdateOne {
	ituo.mutation.ClearArchiveTime()
	return ituo
}

// SetName sets the "name" field.
func (ituo *IncidentTypeUpdateOne) SetName(s string) *IncidentTypeUpdateOne {
	ituo.mutation.SetName(s)
	return ituo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (ituo *IncidentTypeUpdateOne) SetNillableName(s *string) *IncidentTypeUpdateOne {
	if s != nil {
		ituo.SetName(*s)
	}
	return ituo
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (ituo *IncidentTypeUpdateOne) AddIncidentIDs(ids ...uuid.UUID) *IncidentTypeUpdateOne {
	ituo.mutation.AddIncidentIDs(ids...)
	return ituo
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (ituo *IncidentTypeUpdateOne) AddIncidents(i ...*Incident) *IncidentTypeUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.AddIncidentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (ituo *IncidentTypeUpdateOne) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTypeUpdateOne {
	ituo.mutation.AddDebriefQuestionIDs(ids...)
	return ituo
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (ituo *IncidentTypeUpdateOne) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTypeUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentTypeMutation object of the builder.
func (ituo *IncidentTypeUpdateOne) Mutation() *IncidentTypeMutation {
	return ituo.mutation
}

// ClearIncidents clears all "incidents" edges to the Incident entity.
func (ituo *IncidentTypeUpdateOne) ClearIncidents() *IncidentTypeUpdateOne {
	ituo.mutation.ClearIncidents()
	return ituo
}

// RemoveIncidentIDs removes the "incidents" edge to Incident entities by IDs.
func (ituo *IncidentTypeUpdateOne) RemoveIncidentIDs(ids ...uuid.UUID) *IncidentTypeUpdateOne {
	ituo.mutation.RemoveIncidentIDs(ids...)
	return ituo
}

// RemoveIncidents removes "incidents" edges to Incident entities.
func (ituo *IncidentTypeUpdateOne) RemoveIncidents(i ...*Incident) *IncidentTypeUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.RemoveIncidentIDs(ids...)
}

// ClearDebriefQuestions clears all "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (ituo *IncidentTypeUpdateOne) ClearDebriefQuestions() *IncidentTypeUpdateOne {
	ituo.mutation.ClearDebriefQuestions()
	return ituo
}

// RemoveDebriefQuestionIDs removes the "debrief_questions" edge to IncidentDebriefQuestion entities by IDs.
func (ituo *IncidentTypeUpdateOne) RemoveDebriefQuestionIDs(ids ...uuid.UUID) *IncidentTypeUpdateOne {
	ituo.mutation.RemoveDebriefQuestionIDs(ids...)
	return ituo
}

// RemoveDebriefQuestions removes "debrief_questions" edges to IncidentDebriefQuestion entities.
func (ituo *IncidentTypeUpdateOne) RemoveDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentTypeUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ituo.RemoveDebriefQuestionIDs(ids...)
}

// Where appends a list predicates to the IncidentTypeUpdate builder.
func (ituo *IncidentTypeUpdateOne) Where(ps ...predicate.IncidentType) *IncidentTypeUpdateOne {
	ituo.mutation.Where(ps...)
	return ituo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (ituo *IncidentTypeUpdateOne) Select(field string, fields ...string) *IncidentTypeUpdateOne {
	ituo.fields = append([]string{field}, fields...)
	return ituo
}

// Save executes the query and returns the updated IncidentType entity.
func (ituo *IncidentTypeUpdateOne) Save(ctx context.Context) (*IncidentType, error) {
	return withHooks(ctx, ituo.sqlSave, ituo.mutation, ituo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ituo *IncidentTypeUpdateOne) SaveX(ctx context.Context) *IncidentType {
	node, err := ituo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (ituo *IncidentTypeUpdateOne) Exec(ctx context.Context) error {
	_, err := ituo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ituo *IncidentTypeUpdateOne) ExecX(ctx context.Context) {
	if err := ituo.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (ituo *IncidentTypeUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentTypeUpdateOne {
	ituo.modifiers = append(ituo.modifiers, modifiers...)
	return ituo
}

func (ituo *IncidentTypeUpdateOne) sqlSave(ctx context.Context) (_node *IncidentType, err error) {
	_spec := sqlgraph.NewUpdateSpec(incidenttype.Table, incidenttype.Columns, sqlgraph.NewFieldSpec(incidenttype.FieldID, field.TypeUUID))
	id, ok := ituo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentType.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := ituo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidenttype.FieldID)
		for _, f := range fields {
			if !incidenttype.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidenttype.FieldID {
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
		_spec.SetField(incidenttype.FieldArchiveTime, field.TypeTime, value)
	}
	if ituo.mutation.ArchiveTimeCleared() {
		_spec.ClearField(incidenttype.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := ituo.mutation.Name(); ok {
		_spec.SetField(incidenttype.FieldName, field.TypeString, value)
	}
	if ituo.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidenttype.IncidentsTable,
			Columns: []string{incidenttype.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ituo.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !ituo.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidenttype.IncidentsTable,
			Columns: []string{incidenttype.IncidentsColumn},
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
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidenttype.IncidentsTable,
			Columns: []string{incidenttype.IncidentsColumn},
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
			Table:   incidenttype.DebriefQuestionsTable,
			Columns: incidenttype.DebriefQuestionsPrimaryKey,
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
			Table:   incidenttype.DebriefQuestionsTable,
			Columns: incidenttype.DebriefQuestionsPrimaryKey,
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
			Table:   incidenttype.DebriefQuestionsTable,
			Columns: incidenttype.DebriefQuestionsPrimaryKey,
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
	_node = &IncidentType{config: ituo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, ituo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidenttype.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	ituo.mutation.done = true
	return _node, nil
}

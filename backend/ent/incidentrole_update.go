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
	"github.com/rezible/rezible/ent/incidentdebriefquestion"
	"github.com/rezible/rezible/ent/incidentrole"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentRoleUpdate is the builder for updating IncidentRole entities.
type IncidentRoleUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentRoleMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentRoleUpdate builder.
func (iru *IncidentRoleUpdate) Where(ps ...predicate.IncidentRole) *IncidentRoleUpdate {
	iru.mutation.Where(ps...)
	return iru
}

// SetArchiveTime sets the "archive_time" field.
func (iru *IncidentRoleUpdate) SetArchiveTime(t time.Time) *IncidentRoleUpdate {
	iru.mutation.SetArchiveTime(t)
	return iru
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (iru *IncidentRoleUpdate) SetNillableArchiveTime(t *time.Time) *IncidentRoleUpdate {
	if t != nil {
		iru.SetArchiveTime(*t)
	}
	return iru
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (iru *IncidentRoleUpdate) ClearArchiveTime() *IncidentRoleUpdate {
	iru.mutation.ClearArchiveTime()
	return iru
}

// SetName sets the "name" field.
func (iru *IncidentRoleUpdate) SetName(s string) *IncidentRoleUpdate {
	iru.mutation.SetName(s)
	return iru
}

// SetNillableName sets the "name" field if the given value is not nil.
func (iru *IncidentRoleUpdate) SetNillableName(s *string) *IncidentRoleUpdate {
	if s != nil {
		iru.SetName(*s)
	}
	return iru
}

// SetProviderID sets the "provider_id" field.
func (iru *IncidentRoleUpdate) SetProviderID(s string) *IncidentRoleUpdate {
	iru.mutation.SetProviderID(s)
	return iru
}

// SetNillableProviderID sets the "provider_id" field if the given value is not nil.
func (iru *IncidentRoleUpdate) SetNillableProviderID(s *string) *IncidentRoleUpdate {
	if s != nil {
		iru.SetProviderID(*s)
	}
	return iru
}

// SetRequired sets the "required" field.
func (iru *IncidentRoleUpdate) SetRequired(b bool) *IncidentRoleUpdate {
	iru.mutation.SetRequired(b)
	return iru
}

// SetNillableRequired sets the "required" field if the given value is not nil.
func (iru *IncidentRoleUpdate) SetNillableRequired(b *bool) *IncidentRoleUpdate {
	if b != nil {
		iru.SetRequired(*b)
	}
	return iru
}

// AddAssignmentIDs adds the "assignments" edge to the IncidentRoleAssignment entity by IDs.
func (iru *IncidentRoleUpdate) AddAssignmentIDs(ids ...uuid.UUID) *IncidentRoleUpdate {
	iru.mutation.AddAssignmentIDs(ids...)
	return iru
}

// AddAssignments adds the "assignments" edges to the IncidentRoleAssignment entity.
func (iru *IncidentRoleUpdate) AddAssignments(i ...*IncidentRoleAssignment) *IncidentRoleUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iru.AddAssignmentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (iru *IncidentRoleUpdate) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentRoleUpdate {
	iru.mutation.AddDebriefQuestionIDs(ids...)
	return iru
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (iru *IncidentRoleUpdate) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentRoleUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iru.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentRoleMutation object of the builder.
func (iru *IncidentRoleUpdate) Mutation() *IncidentRoleMutation {
	return iru.mutation
}

// ClearAssignments clears all "assignments" edges to the IncidentRoleAssignment entity.
func (iru *IncidentRoleUpdate) ClearAssignments() *IncidentRoleUpdate {
	iru.mutation.ClearAssignments()
	return iru
}

// RemoveAssignmentIDs removes the "assignments" edge to IncidentRoleAssignment entities by IDs.
func (iru *IncidentRoleUpdate) RemoveAssignmentIDs(ids ...uuid.UUID) *IncidentRoleUpdate {
	iru.mutation.RemoveAssignmentIDs(ids...)
	return iru
}

// RemoveAssignments removes "assignments" edges to IncidentRoleAssignment entities.
func (iru *IncidentRoleUpdate) RemoveAssignments(i ...*IncidentRoleAssignment) *IncidentRoleUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iru.RemoveAssignmentIDs(ids...)
}

// ClearDebriefQuestions clears all "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (iru *IncidentRoleUpdate) ClearDebriefQuestions() *IncidentRoleUpdate {
	iru.mutation.ClearDebriefQuestions()
	return iru
}

// RemoveDebriefQuestionIDs removes the "debrief_questions" edge to IncidentDebriefQuestion entities by IDs.
func (iru *IncidentRoleUpdate) RemoveDebriefQuestionIDs(ids ...uuid.UUID) *IncidentRoleUpdate {
	iru.mutation.RemoveDebriefQuestionIDs(ids...)
	return iru
}

// RemoveDebriefQuestions removes "debrief_questions" edges to IncidentDebriefQuestion entities.
func (iru *IncidentRoleUpdate) RemoveDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentRoleUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iru.RemoveDebriefQuestionIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (iru *IncidentRoleUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, iru.sqlSave, iru.mutation, iru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iru *IncidentRoleUpdate) SaveX(ctx context.Context) int {
	affected, err := iru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (iru *IncidentRoleUpdate) Exec(ctx context.Context) error {
	_, err := iru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iru *IncidentRoleUpdate) ExecX(ctx context.Context) {
	if err := iru.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iru *IncidentRoleUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentRoleUpdate {
	iru.modifiers = append(iru.modifiers, modifiers...)
	return iru
}

func (iru *IncidentRoleUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(incidentrole.Table, incidentrole.Columns, sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID))
	if ps := iru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iru.mutation.ArchiveTime(); ok {
		_spec.SetField(incidentrole.FieldArchiveTime, field.TypeTime, value)
	}
	if iru.mutation.ArchiveTimeCleared() {
		_spec.ClearField(incidentrole.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := iru.mutation.Name(); ok {
		_spec.SetField(incidentrole.FieldName, field.TypeString, value)
	}
	if value, ok := iru.mutation.ProviderID(); ok {
		_spec.SetField(incidentrole.FieldProviderID, field.TypeString, value)
	}
	if value, ok := iru.mutation.Required(); ok {
		_spec.SetField(incidentrole.FieldRequired, field.TypeBool, value)
	}
	if iru.mutation.AssignmentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iru.mutation.RemovedAssignmentsIDs(); len(nodes) > 0 && !iru.mutation.AssignmentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iru.mutation.AssignmentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if iru.mutation.DebriefQuestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefquestion.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iru.mutation.RemovedDebriefQuestionsIDs(); len(nodes) > 0 && !iru.mutation.DebriefQuestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
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
	if nodes := iru.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
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
	_spec.AddModifiers(iru.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, iru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentrole.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	iru.mutation.done = true
	return n, nil
}

// IncidentRoleUpdateOne is the builder for updating a single IncidentRole entity.
type IncidentRoleUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentRoleMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetArchiveTime sets the "archive_time" field.
func (iruo *IncidentRoleUpdateOne) SetArchiveTime(t time.Time) *IncidentRoleUpdateOne {
	iruo.mutation.SetArchiveTime(t)
	return iruo
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (iruo *IncidentRoleUpdateOne) SetNillableArchiveTime(t *time.Time) *IncidentRoleUpdateOne {
	if t != nil {
		iruo.SetArchiveTime(*t)
	}
	return iruo
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (iruo *IncidentRoleUpdateOne) ClearArchiveTime() *IncidentRoleUpdateOne {
	iruo.mutation.ClearArchiveTime()
	return iruo
}

// SetName sets the "name" field.
func (iruo *IncidentRoleUpdateOne) SetName(s string) *IncidentRoleUpdateOne {
	iruo.mutation.SetName(s)
	return iruo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (iruo *IncidentRoleUpdateOne) SetNillableName(s *string) *IncidentRoleUpdateOne {
	if s != nil {
		iruo.SetName(*s)
	}
	return iruo
}

// SetProviderID sets the "provider_id" field.
func (iruo *IncidentRoleUpdateOne) SetProviderID(s string) *IncidentRoleUpdateOne {
	iruo.mutation.SetProviderID(s)
	return iruo
}

// SetNillableProviderID sets the "provider_id" field if the given value is not nil.
func (iruo *IncidentRoleUpdateOne) SetNillableProviderID(s *string) *IncidentRoleUpdateOne {
	if s != nil {
		iruo.SetProviderID(*s)
	}
	return iruo
}

// SetRequired sets the "required" field.
func (iruo *IncidentRoleUpdateOne) SetRequired(b bool) *IncidentRoleUpdateOne {
	iruo.mutation.SetRequired(b)
	return iruo
}

// SetNillableRequired sets the "required" field if the given value is not nil.
func (iruo *IncidentRoleUpdateOne) SetNillableRequired(b *bool) *IncidentRoleUpdateOne {
	if b != nil {
		iruo.SetRequired(*b)
	}
	return iruo
}

// AddAssignmentIDs adds the "assignments" edge to the IncidentRoleAssignment entity by IDs.
func (iruo *IncidentRoleUpdateOne) AddAssignmentIDs(ids ...uuid.UUID) *IncidentRoleUpdateOne {
	iruo.mutation.AddAssignmentIDs(ids...)
	return iruo
}

// AddAssignments adds the "assignments" edges to the IncidentRoleAssignment entity.
func (iruo *IncidentRoleUpdateOne) AddAssignments(i ...*IncidentRoleAssignment) *IncidentRoleUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iruo.AddAssignmentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (iruo *IncidentRoleUpdateOne) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentRoleUpdateOne {
	iruo.mutation.AddDebriefQuestionIDs(ids...)
	return iruo
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (iruo *IncidentRoleUpdateOne) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentRoleUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iruo.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentRoleMutation object of the builder.
func (iruo *IncidentRoleUpdateOne) Mutation() *IncidentRoleMutation {
	return iruo.mutation
}

// ClearAssignments clears all "assignments" edges to the IncidentRoleAssignment entity.
func (iruo *IncidentRoleUpdateOne) ClearAssignments() *IncidentRoleUpdateOne {
	iruo.mutation.ClearAssignments()
	return iruo
}

// RemoveAssignmentIDs removes the "assignments" edge to IncidentRoleAssignment entities by IDs.
func (iruo *IncidentRoleUpdateOne) RemoveAssignmentIDs(ids ...uuid.UUID) *IncidentRoleUpdateOne {
	iruo.mutation.RemoveAssignmentIDs(ids...)
	return iruo
}

// RemoveAssignments removes "assignments" edges to IncidentRoleAssignment entities.
func (iruo *IncidentRoleUpdateOne) RemoveAssignments(i ...*IncidentRoleAssignment) *IncidentRoleUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iruo.RemoveAssignmentIDs(ids...)
}

// ClearDebriefQuestions clears all "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (iruo *IncidentRoleUpdateOne) ClearDebriefQuestions() *IncidentRoleUpdateOne {
	iruo.mutation.ClearDebriefQuestions()
	return iruo
}

// RemoveDebriefQuestionIDs removes the "debrief_questions" edge to IncidentDebriefQuestion entities by IDs.
func (iruo *IncidentRoleUpdateOne) RemoveDebriefQuestionIDs(ids ...uuid.UUID) *IncidentRoleUpdateOne {
	iruo.mutation.RemoveDebriefQuestionIDs(ids...)
	return iruo
}

// RemoveDebriefQuestions removes "debrief_questions" edges to IncidentDebriefQuestion entities.
func (iruo *IncidentRoleUpdateOne) RemoveDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentRoleUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iruo.RemoveDebriefQuestionIDs(ids...)
}

// Where appends a list predicates to the IncidentRoleUpdate builder.
func (iruo *IncidentRoleUpdateOne) Where(ps ...predicate.IncidentRole) *IncidentRoleUpdateOne {
	iruo.mutation.Where(ps...)
	return iruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (iruo *IncidentRoleUpdateOne) Select(field string, fields ...string) *IncidentRoleUpdateOne {
	iruo.fields = append([]string{field}, fields...)
	return iruo
}

// Save executes the query and returns the updated IncidentRole entity.
func (iruo *IncidentRoleUpdateOne) Save(ctx context.Context) (*IncidentRole, error) {
	return withHooks(ctx, iruo.sqlSave, iruo.mutation, iruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iruo *IncidentRoleUpdateOne) SaveX(ctx context.Context) *IncidentRole {
	node, err := iruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (iruo *IncidentRoleUpdateOne) Exec(ctx context.Context) error {
	_, err := iruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iruo *IncidentRoleUpdateOne) ExecX(ctx context.Context) {
	if err := iruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iruo *IncidentRoleUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentRoleUpdateOne {
	iruo.modifiers = append(iruo.modifiers, modifiers...)
	return iruo
}

func (iruo *IncidentRoleUpdateOne) sqlSave(ctx context.Context) (_node *IncidentRole, err error) {
	_spec := sqlgraph.NewUpdateSpec(incidentrole.Table, incidentrole.Columns, sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID))
	id, ok := iruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentRole.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := iruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentrole.FieldID)
		for _, f := range fields {
			if !incidentrole.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidentrole.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := iruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iruo.mutation.ArchiveTime(); ok {
		_spec.SetField(incidentrole.FieldArchiveTime, field.TypeTime, value)
	}
	if iruo.mutation.ArchiveTimeCleared() {
		_spec.ClearField(incidentrole.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := iruo.mutation.Name(); ok {
		_spec.SetField(incidentrole.FieldName, field.TypeString, value)
	}
	if value, ok := iruo.mutation.ProviderID(); ok {
		_spec.SetField(incidentrole.FieldProviderID, field.TypeString, value)
	}
	if value, ok := iruo.mutation.Required(); ok {
		_spec.SetField(incidentrole.FieldRequired, field.TypeBool, value)
	}
	if iruo.mutation.AssignmentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iruo.mutation.RemovedAssignmentsIDs(); len(nodes) > 0 && !iruo.mutation.AssignmentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iruo.mutation.AssignmentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if iruo.mutation.DebriefQuestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefquestion.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iruo.mutation.RemovedDebriefQuestionsIDs(); len(nodes) > 0 && !iruo.mutation.DebriefQuestionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
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
	if nodes := iruo.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
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
	_spec.AddModifiers(iruo.modifiers...)
	_node = &IncidentRole{config: iruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, iruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentrole.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	iruo.mutation.done = true
	return _node, nil
}

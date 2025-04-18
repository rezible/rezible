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
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/ent/systemanalysiscomponent"
	"github.com/rezible/rezible/ent/systemanalysisrelationship"
	"github.com/rezible/rezible/ent/systemcomponent"
)

// SystemAnalysisUpdate is the builder for updating SystemAnalysis entities.
type SystemAnalysisUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemAnalysisMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemAnalysisUpdate builder.
func (sau *SystemAnalysisUpdate) Where(ps ...predicate.SystemAnalysis) *SystemAnalysisUpdate {
	sau.mutation.Where(ps...)
	return sau
}

// SetCreatedAt sets the "created_at" field.
func (sau *SystemAnalysisUpdate) SetCreatedAt(t time.Time) *SystemAnalysisUpdate {
	sau.mutation.SetCreatedAt(t)
	return sau
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sau *SystemAnalysisUpdate) SetNillableCreatedAt(t *time.Time) *SystemAnalysisUpdate {
	if t != nil {
		sau.SetCreatedAt(*t)
	}
	return sau
}

// SetUpdatedAt sets the "updated_at" field.
func (sau *SystemAnalysisUpdate) SetUpdatedAt(t time.Time) *SystemAnalysisUpdate {
	sau.mutation.SetUpdatedAt(t)
	return sau
}

// SetRetrospectiveID sets the "retrospective" edge to the Retrospective entity by ID.
func (sau *SystemAnalysisUpdate) SetRetrospectiveID(id uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.SetRetrospectiveID(id)
	return sau
}

// SetRetrospective sets the "retrospective" edge to the Retrospective entity.
func (sau *SystemAnalysisUpdate) SetRetrospective(r *Retrospective) *SystemAnalysisUpdate {
	return sau.SetRetrospectiveID(r.ID)
}

// AddComponentIDs adds the "components" edge to the SystemComponent entity by IDs.
func (sau *SystemAnalysisUpdate) AddComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.AddComponentIDs(ids...)
	return sau
}

// AddComponents adds the "components" edges to the SystemComponent entity.
func (sau *SystemAnalysisUpdate) AddComponents(s ...*SystemComponent) *SystemAnalysisUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sau.AddComponentIDs(ids...)
}

// AddRelationshipIDs adds the "relationships" edge to the SystemAnalysisRelationship entity by IDs.
func (sau *SystemAnalysisUpdate) AddRelationshipIDs(ids ...uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.AddRelationshipIDs(ids...)
	return sau
}

// AddRelationships adds the "relationships" edges to the SystemAnalysisRelationship entity.
func (sau *SystemAnalysisUpdate) AddRelationships(s ...*SystemAnalysisRelationship) *SystemAnalysisUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sau.AddRelationshipIDs(ids...)
}

// AddAnalysisComponentIDs adds the "analysis_components" edge to the SystemAnalysisComponent entity by IDs.
func (sau *SystemAnalysisUpdate) AddAnalysisComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.AddAnalysisComponentIDs(ids...)
	return sau
}

// AddAnalysisComponents adds the "analysis_components" edges to the SystemAnalysisComponent entity.
func (sau *SystemAnalysisUpdate) AddAnalysisComponents(s ...*SystemAnalysisComponent) *SystemAnalysisUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sau.AddAnalysisComponentIDs(ids...)
}

// Mutation returns the SystemAnalysisMutation object of the builder.
func (sau *SystemAnalysisUpdate) Mutation() *SystemAnalysisMutation {
	return sau.mutation
}

// ClearRetrospective clears the "retrospective" edge to the Retrospective entity.
func (sau *SystemAnalysisUpdate) ClearRetrospective() *SystemAnalysisUpdate {
	sau.mutation.ClearRetrospective()
	return sau
}

// ClearComponents clears all "components" edges to the SystemComponent entity.
func (sau *SystemAnalysisUpdate) ClearComponents() *SystemAnalysisUpdate {
	sau.mutation.ClearComponents()
	return sau
}

// RemoveComponentIDs removes the "components" edge to SystemComponent entities by IDs.
func (sau *SystemAnalysisUpdate) RemoveComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.RemoveComponentIDs(ids...)
	return sau
}

// RemoveComponents removes "components" edges to SystemComponent entities.
func (sau *SystemAnalysisUpdate) RemoveComponents(s ...*SystemComponent) *SystemAnalysisUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sau.RemoveComponentIDs(ids...)
}

// ClearRelationships clears all "relationships" edges to the SystemAnalysisRelationship entity.
func (sau *SystemAnalysisUpdate) ClearRelationships() *SystemAnalysisUpdate {
	sau.mutation.ClearRelationships()
	return sau
}

// RemoveRelationshipIDs removes the "relationships" edge to SystemAnalysisRelationship entities by IDs.
func (sau *SystemAnalysisUpdate) RemoveRelationshipIDs(ids ...uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.RemoveRelationshipIDs(ids...)
	return sau
}

// RemoveRelationships removes "relationships" edges to SystemAnalysisRelationship entities.
func (sau *SystemAnalysisUpdate) RemoveRelationships(s ...*SystemAnalysisRelationship) *SystemAnalysisUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sau.RemoveRelationshipIDs(ids...)
}

// ClearAnalysisComponents clears all "analysis_components" edges to the SystemAnalysisComponent entity.
func (sau *SystemAnalysisUpdate) ClearAnalysisComponents() *SystemAnalysisUpdate {
	sau.mutation.ClearAnalysisComponents()
	return sau
}

// RemoveAnalysisComponentIDs removes the "analysis_components" edge to SystemAnalysisComponent entities by IDs.
func (sau *SystemAnalysisUpdate) RemoveAnalysisComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdate {
	sau.mutation.RemoveAnalysisComponentIDs(ids...)
	return sau
}

// RemoveAnalysisComponents removes "analysis_components" edges to SystemAnalysisComponent entities.
func (sau *SystemAnalysisUpdate) RemoveAnalysisComponents(s ...*SystemAnalysisComponent) *SystemAnalysisUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sau.RemoveAnalysisComponentIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (sau *SystemAnalysisUpdate) Save(ctx context.Context) (int, error) {
	sau.defaults()
	return withHooks(ctx, sau.sqlSave, sau.mutation, sau.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sau *SystemAnalysisUpdate) SaveX(ctx context.Context) int {
	affected, err := sau.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (sau *SystemAnalysisUpdate) Exec(ctx context.Context) error {
	_, err := sau.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sau *SystemAnalysisUpdate) ExecX(ctx context.Context) {
	if err := sau.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sau *SystemAnalysisUpdate) defaults() {
	if _, ok := sau.mutation.UpdatedAt(); !ok {
		v := systemanalysis.UpdateDefaultUpdatedAt()
		sau.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sau *SystemAnalysisUpdate) check() error {
	if sau.mutation.RetrospectiveCleared() && len(sau.mutation.RetrospectiveIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemAnalysis.retrospective"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sau *SystemAnalysisUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemAnalysisUpdate {
	sau.modifiers = append(sau.modifiers, modifiers...)
	return sau
}

func (sau *SystemAnalysisUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := sau.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemanalysis.Table, systemanalysis.Columns, sqlgraph.NewFieldSpec(systemanalysis.FieldID, field.TypeUUID))
	if ps := sau.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sau.mutation.CreatedAt(); ok {
		_spec.SetField(systemanalysis.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := sau.mutation.UpdatedAt(); ok {
		_spec.SetField(systemanalysis.FieldUpdatedAt, field.TypeTime, value)
	}
	if sau.mutation.RetrospectiveCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   systemanalysis.RetrospectiveTable,
			Columns: []string{systemanalysis.RetrospectiveColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.RetrospectiveIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   systemanalysis.RetrospectiveTable,
			Columns: []string{systemanalysis.RetrospectiveColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sau.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemanalysis.ComponentsTable,
			Columns: systemanalysis.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		createE := &SystemAnalysisComponentCreate{config: sau.config, mutation: newSystemAnalysisComponentMutation(sau.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.RemovedComponentsIDs(); len(nodes) > 0 && !sau.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemanalysis.ComponentsTable,
			Columns: systemanalysis.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemAnalysisComponentCreate{config: sau.config, mutation: newSystemAnalysisComponentMutation(sau.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.ComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemanalysis.ComponentsTable,
			Columns: systemanalysis.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemAnalysisComponentCreate{config: sau.config, mutation: newSystemAnalysisComponentMutation(sau.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sau.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.RelationshipsTable,
			Columns: []string{systemanalysis.RelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.RemovedRelationshipsIDs(); len(nodes) > 0 && !sau.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.RelationshipsTable,
			Columns: []string{systemanalysis.RelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.RelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.RelationshipsTable,
			Columns: []string{systemanalysis.RelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sau.mutation.AnalysisComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.AnalysisComponentsTable,
			Columns: []string{systemanalysis.AnalysisComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.RemovedAnalysisComponentsIDs(); len(nodes) > 0 && !sau.mutation.AnalysisComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.AnalysisComponentsTable,
			Columns: []string{systemanalysis.AnalysisComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sau.mutation.AnalysisComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.AnalysisComponentsTable,
			Columns: []string{systemanalysis.AnalysisComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(sau.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, sau.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemanalysis.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	sau.mutation.done = true
	return n, nil
}

// SystemAnalysisUpdateOne is the builder for updating a single SystemAnalysis entity.
type SystemAnalysisUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemAnalysisMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetCreatedAt sets the "created_at" field.
func (sauo *SystemAnalysisUpdateOne) SetCreatedAt(t time.Time) *SystemAnalysisUpdateOne {
	sauo.mutation.SetCreatedAt(t)
	return sauo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sauo *SystemAnalysisUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemAnalysisUpdateOne {
	if t != nil {
		sauo.SetCreatedAt(*t)
	}
	return sauo
}

// SetUpdatedAt sets the "updated_at" field.
func (sauo *SystemAnalysisUpdateOne) SetUpdatedAt(t time.Time) *SystemAnalysisUpdateOne {
	sauo.mutation.SetUpdatedAt(t)
	return sauo
}

// SetRetrospectiveID sets the "retrospective" edge to the Retrospective entity by ID.
func (sauo *SystemAnalysisUpdateOne) SetRetrospectiveID(id uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.SetRetrospectiveID(id)
	return sauo
}

// SetRetrospective sets the "retrospective" edge to the Retrospective entity.
func (sauo *SystemAnalysisUpdateOne) SetRetrospective(r *Retrospective) *SystemAnalysisUpdateOne {
	return sauo.SetRetrospectiveID(r.ID)
}

// AddComponentIDs adds the "components" edge to the SystemComponent entity by IDs.
func (sauo *SystemAnalysisUpdateOne) AddComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.AddComponentIDs(ids...)
	return sauo
}

// AddComponents adds the "components" edges to the SystemComponent entity.
func (sauo *SystemAnalysisUpdateOne) AddComponents(s ...*SystemComponent) *SystemAnalysisUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sauo.AddComponentIDs(ids...)
}

// AddRelationshipIDs adds the "relationships" edge to the SystemAnalysisRelationship entity by IDs.
func (sauo *SystemAnalysisUpdateOne) AddRelationshipIDs(ids ...uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.AddRelationshipIDs(ids...)
	return sauo
}

// AddRelationships adds the "relationships" edges to the SystemAnalysisRelationship entity.
func (sauo *SystemAnalysisUpdateOne) AddRelationships(s ...*SystemAnalysisRelationship) *SystemAnalysisUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sauo.AddRelationshipIDs(ids...)
}

// AddAnalysisComponentIDs adds the "analysis_components" edge to the SystemAnalysisComponent entity by IDs.
func (sauo *SystemAnalysisUpdateOne) AddAnalysisComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.AddAnalysisComponentIDs(ids...)
	return sauo
}

// AddAnalysisComponents adds the "analysis_components" edges to the SystemAnalysisComponent entity.
func (sauo *SystemAnalysisUpdateOne) AddAnalysisComponents(s ...*SystemAnalysisComponent) *SystemAnalysisUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sauo.AddAnalysisComponentIDs(ids...)
}

// Mutation returns the SystemAnalysisMutation object of the builder.
func (sauo *SystemAnalysisUpdateOne) Mutation() *SystemAnalysisMutation {
	return sauo.mutation
}

// ClearRetrospective clears the "retrospective" edge to the Retrospective entity.
func (sauo *SystemAnalysisUpdateOne) ClearRetrospective() *SystemAnalysisUpdateOne {
	sauo.mutation.ClearRetrospective()
	return sauo
}

// ClearComponents clears all "components" edges to the SystemComponent entity.
func (sauo *SystemAnalysisUpdateOne) ClearComponents() *SystemAnalysisUpdateOne {
	sauo.mutation.ClearComponents()
	return sauo
}

// RemoveComponentIDs removes the "components" edge to SystemComponent entities by IDs.
func (sauo *SystemAnalysisUpdateOne) RemoveComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.RemoveComponentIDs(ids...)
	return sauo
}

// RemoveComponents removes "components" edges to SystemComponent entities.
func (sauo *SystemAnalysisUpdateOne) RemoveComponents(s ...*SystemComponent) *SystemAnalysisUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sauo.RemoveComponentIDs(ids...)
}

// ClearRelationships clears all "relationships" edges to the SystemAnalysisRelationship entity.
func (sauo *SystemAnalysisUpdateOne) ClearRelationships() *SystemAnalysisUpdateOne {
	sauo.mutation.ClearRelationships()
	return sauo
}

// RemoveRelationshipIDs removes the "relationships" edge to SystemAnalysisRelationship entities by IDs.
func (sauo *SystemAnalysisUpdateOne) RemoveRelationshipIDs(ids ...uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.RemoveRelationshipIDs(ids...)
	return sauo
}

// RemoveRelationships removes "relationships" edges to SystemAnalysisRelationship entities.
func (sauo *SystemAnalysisUpdateOne) RemoveRelationships(s ...*SystemAnalysisRelationship) *SystemAnalysisUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sauo.RemoveRelationshipIDs(ids...)
}

// ClearAnalysisComponents clears all "analysis_components" edges to the SystemAnalysisComponent entity.
func (sauo *SystemAnalysisUpdateOne) ClearAnalysisComponents() *SystemAnalysisUpdateOne {
	sauo.mutation.ClearAnalysisComponents()
	return sauo
}

// RemoveAnalysisComponentIDs removes the "analysis_components" edge to SystemAnalysisComponent entities by IDs.
func (sauo *SystemAnalysisUpdateOne) RemoveAnalysisComponentIDs(ids ...uuid.UUID) *SystemAnalysisUpdateOne {
	sauo.mutation.RemoveAnalysisComponentIDs(ids...)
	return sauo
}

// RemoveAnalysisComponents removes "analysis_components" edges to SystemAnalysisComponent entities.
func (sauo *SystemAnalysisUpdateOne) RemoveAnalysisComponents(s ...*SystemAnalysisComponent) *SystemAnalysisUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return sauo.RemoveAnalysisComponentIDs(ids...)
}

// Where appends a list predicates to the SystemAnalysisUpdate builder.
func (sauo *SystemAnalysisUpdateOne) Where(ps ...predicate.SystemAnalysis) *SystemAnalysisUpdateOne {
	sauo.mutation.Where(ps...)
	return sauo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (sauo *SystemAnalysisUpdateOne) Select(field string, fields ...string) *SystemAnalysisUpdateOne {
	sauo.fields = append([]string{field}, fields...)
	return sauo
}

// Save executes the query and returns the updated SystemAnalysis entity.
func (sauo *SystemAnalysisUpdateOne) Save(ctx context.Context) (*SystemAnalysis, error) {
	sauo.defaults()
	return withHooks(ctx, sauo.sqlSave, sauo.mutation, sauo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sauo *SystemAnalysisUpdateOne) SaveX(ctx context.Context) *SystemAnalysis {
	node, err := sauo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (sauo *SystemAnalysisUpdateOne) Exec(ctx context.Context) error {
	_, err := sauo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sauo *SystemAnalysisUpdateOne) ExecX(ctx context.Context) {
	if err := sauo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sauo *SystemAnalysisUpdateOne) defaults() {
	if _, ok := sauo.mutation.UpdatedAt(); !ok {
		v := systemanalysis.UpdateDefaultUpdatedAt()
		sauo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sauo *SystemAnalysisUpdateOne) check() error {
	if sauo.mutation.RetrospectiveCleared() && len(sauo.mutation.RetrospectiveIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemAnalysis.retrospective"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sauo *SystemAnalysisUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemAnalysisUpdateOne {
	sauo.modifiers = append(sauo.modifiers, modifiers...)
	return sauo
}

func (sauo *SystemAnalysisUpdateOne) sqlSave(ctx context.Context) (_node *SystemAnalysis, err error) {
	if err := sauo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemanalysis.Table, systemanalysis.Columns, sqlgraph.NewFieldSpec(systemanalysis.FieldID, field.TypeUUID))
	id, ok := sauo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemAnalysis.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := sauo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemanalysis.FieldID)
		for _, f := range fields {
			if !systemanalysis.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemanalysis.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := sauo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sauo.mutation.CreatedAt(); ok {
		_spec.SetField(systemanalysis.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := sauo.mutation.UpdatedAt(); ok {
		_spec.SetField(systemanalysis.FieldUpdatedAt, field.TypeTime, value)
	}
	if sauo.mutation.RetrospectiveCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   systemanalysis.RetrospectiveTable,
			Columns: []string{systemanalysis.RetrospectiveColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.RetrospectiveIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   systemanalysis.RetrospectiveTable,
			Columns: []string{systemanalysis.RetrospectiveColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sauo.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemanalysis.ComponentsTable,
			Columns: systemanalysis.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		createE := &SystemAnalysisComponentCreate{config: sauo.config, mutation: newSystemAnalysisComponentMutation(sauo.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.RemovedComponentsIDs(); len(nodes) > 0 && !sauo.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemanalysis.ComponentsTable,
			Columns: systemanalysis.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemAnalysisComponentCreate{config: sauo.config, mutation: newSystemAnalysisComponentMutation(sauo.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.ComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemanalysis.ComponentsTable,
			Columns: systemanalysis.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemAnalysisComponentCreate{config: sauo.config, mutation: newSystemAnalysisComponentMutation(sauo.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sauo.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.RelationshipsTable,
			Columns: []string{systemanalysis.RelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.RemovedRelationshipsIDs(); len(nodes) > 0 && !sauo.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.RelationshipsTable,
			Columns: []string{systemanalysis.RelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.RelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.RelationshipsTable,
			Columns: []string{systemanalysis.RelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sauo.mutation.AnalysisComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.AnalysisComponentsTable,
			Columns: []string{systemanalysis.AnalysisComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.RemovedAnalysisComponentsIDs(); len(nodes) > 0 && !sauo.mutation.AnalysisComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.AnalysisComponentsTable,
			Columns: []string{systemanalysis.AnalysisComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sauo.mutation.AnalysisComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemanalysis.AnalysisComponentsTable,
			Columns: []string{systemanalysis.AnalysisComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(sauo.modifiers...)
	_node = &SystemAnalysis{config: sauo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, sauo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemanalysis.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	sauo.mutation.done = true
	return _node, nil
}

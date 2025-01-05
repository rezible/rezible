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
	"github.com/rezible/rezible/ent/incidentevent"
	"github.com/rezible/rezible/ent/incidenteventsystemcomponent"
	"github.com/rezible/rezible/ent/incidentsystemcomponent"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentcontrolrelationship"
	"github.com/rezible/rezible/ent/systemcomponentfeedbackrelationship"
)

// SystemComponentCreate is the builder for creating a SystemComponent entity.
type SystemComponentCreate struct {
	config
	mutation *SystemComponentMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetName sets the "name" field.
func (scc *SystemComponentCreate) SetName(s string) *SystemComponentCreate {
	scc.mutation.SetName(s)
	return scc
}

// SetType sets the "type" field.
func (scc *SystemComponentCreate) SetType(s systemcomponent.Type) *SystemComponentCreate {
	scc.mutation.SetType(s)
	return scc
}

// SetDescription sets the "description" field.
func (scc *SystemComponentCreate) SetDescription(s string) *SystemComponentCreate {
	scc.mutation.SetDescription(s)
	return scc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (scc *SystemComponentCreate) SetNillableDescription(s *string) *SystemComponentCreate {
	if s != nil {
		scc.SetDescription(*s)
	}
	return scc
}

// SetProperties sets the "properties" field.
func (scc *SystemComponentCreate) SetProperties(m map[string]interface{}) *SystemComponentCreate {
	scc.mutation.SetProperties(m)
	return scc
}

// SetCreatedAt sets the "created_at" field.
func (scc *SystemComponentCreate) SetCreatedAt(t time.Time) *SystemComponentCreate {
	scc.mutation.SetCreatedAt(t)
	return scc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (scc *SystemComponentCreate) SetNillableCreatedAt(t *time.Time) *SystemComponentCreate {
	if t != nil {
		scc.SetCreatedAt(*t)
	}
	return scc
}

// SetUpdatedAt sets the "updated_at" field.
func (scc *SystemComponentCreate) SetUpdatedAt(t time.Time) *SystemComponentCreate {
	scc.mutation.SetUpdatedAt(t)
	return scc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (scc *SystemComponentCreate) SetNillableUpdatedAt(t *time.Time) *SystemComponentCreate {
	if t != nil {
		scc.SetUpdatedAt(*t)
	}
	return scc
}

// SetID sets the "id" field.
func (scc *SystemComponentCreate) SetID(u uuid.UUID) *SystemComponentCreate {
	scc.mutation.SetID(u)
	return scc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (scc *SystemComponentCreate) SetNillableID(u *uuid.UUID) *SystemComponentCreate {
	if u != nil {
		scc.SetID(*u)
	}
	return scc
}

// SetParentID sets the "parent" edge to the SystemComponent entity by ID.
func (scc *SystemComponentCreate) SetParentID(id uuid.UUID) *SystemComponentCreate {
	scc.mutation.SetParentID(id)
	return scc
}

// SetNillableParentID sets the "parent" edge to the SystemComponent entity by ID if the given value is not nil.
func (scc *SystemComponentCreate) SetNillableParentID(id *uuid.UUID) *SystemComponentCreate {
	if id != nil {
		scc = scc.SetParentID(*id)
	}
	return scc
}

// SetParent sets the "parent" edge to the SystemComponent entity.
func (scc *SystemComponentCreate) SetParent(s *SystemComponent) *SystemComponentCreate {
	return scc.SetParentID(s.ID)
}

// AddChildIDs adds the "children" edge to the SystemComponent entity by IDs.
func (scc *SystemComponentCreate) AddChildIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddChildIDs(ids...)
	return scc
}

// AddChildren adds the "children" edges to the SystemComponent entity.
func (scc *SystemComponentCreate) AddChildren(s ...*SystemComponent) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scc.AddChildIDs(ids...)
}

// AddControlIDs adds the "controls" edge to the SystemComponent entity by IDs.
func (scc *SystemComponentCreate) AddControlIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddControlIDs(ids...)
	return scc
}

// AddControls adds the "controls" edges to the SystemComponent entity.
func (scc *SystemComponentCreate) AddControls(s ...*SystemComponent) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scc.AddControlIDs(ids...)
}

// AddFeedbackToIDs adds the "feedback_to" edge to the SystemComponent entity by IDs.
func (scc *SystemComponentCreate) AddFeedbackToIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddFeedbackToIDs(ids...)
	return scc
}

// AddFeedbackTo adds the "feedback_to" edges to the SystemComponent entity.
func (scc *SystemComponentCreate) AddFeedbackTo(s ...*SystemComponent) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scc.AddFeedbackToIDs(ids...)
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (scc *SystemComponentCreate) AddIncidentIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddIncidentIDs(ids...)
	return scc
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (scc *SystemComponentCreate) AddIncidents(i ...*Incident) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return scc.AddIncidentIDs(ids...)
}

// AddEventIDs adds the "events" edge to the IncidentEvent entity by IDs.
func (scc *SystemComponentCreate) AddEventIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddEventIDs(ids...)
	return scc
}

// AddEvents adds the "events" edges to the IncidentEvent entity.
func (scc *SystemComponentCreate) AddEvents(i ...*IncidentEvent) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return scc.AddEventIDs(ids...)
}

// AddControlRelationshipIDs adds the "control_relationships" edge to the SystemComponentControlRelationship entity by IDs.
func (scc *SystemComponentCreate) AddControlRelationshipIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddControlRelationshipIDs(ids...)
	return scc
}

// AddControlRelationships adds the "control_relationships" edges to the SystemComponentControlRelationship entity.
func (scc *SystemComponentCreate) AddControlRelationships(s ...*SystemComponentControlRelationship) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scc.AddControlRelationshipIDs(ids...)
}

// AddFeedbackRelationshipIDs adds the "feedback_relationships" edge to the SystemComponentFeedbackRelationship entity by IDs.
func (scc *SystemComponentCreate) AddFeedbackRelationshipIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddFeedbackRelationshipIDs(ids...)
	return scc
}

// AddFeedbackRelationships adds the "feedback_relationships" edges to the SystemComponentFeedbackRelationship entity.
func (scc *SystemComponentCreate) AddFeedbackRelationships(s ...*SystemComponentFeedbackRelationship) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scc.AddFeedbackRelationshipIDs(ids...)
}

// AddIncidentSystemComponentIDs adds the "incident_system_components" edge to the IncidentSystemComponent entity by IDs.
func (scc *SystemComponentCreate) AddIncidentSystemComponentIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddIncidentSystemComponentIDs(ids...)
	return scc
}

// AddIncidentSystemComponents adds the "incident_system_components" edges to the IncidentSystemComponent entity.
func (scc *SystemComponentCreate) AddIncidentSystemComponents(i ...*IncidentSystemComponent) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return scc.AddIncidentSystemComponentIDs(ids...)
}

// AddEventComponentIDs adds the "event_components" edge to the IncidentEventSystemComponent entity by IDs.
func (scc *SystemComponentCreate) AddEventComponentIDs(ids ...uuid.UUID) *SystemComponentCreate {
	scc.mutation.AddEventComponentIDs(ids...)
	return scc
}

// AddEventComponents adds the "event_components" edges to the IncidentEventSystemComponent entity.
func (scc *SystemComponentCreate) AddEventComponents(i ...*IncidentEventSystemComponent) *SystemComponentCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return scc.AddEventComponentIDs(ids...)
}

// Mutation returns the SystemComponentMutation object of the builder.
func (scc *SystemComponentCreate) Mutation() *SystemComponentMutation {
	return scc.mutation
}

// Save creates the SystemComponent in the database.
func (scc *SystemComponentCreate) Save(ctx context.Context) (*SystemComponent, error) {
	scc.defaults()
	return withHooks(ctx, scc.sqlSave, scc.mutation, scc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (scc *SystemComponentCreate) SaveX(ctx context.Context) *SystemComponent {
	v, err := scc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scc *SystemComponentCreate) Exec(ctx context.Context) error {
	_, err := scc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scc *SystemComponentCreate) ExecX(ctx context.Context) {
	if err := scc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (scc *SystemComponentCreate) defaults() {
	if _, ok := scc.mutation.CreatedAt(); !ok {
		v := systemcomponent.DefaultCreatedAt()
		scc.mutation.SetCreatedAt(v)
	}
	if _, ok := scc.mutation.UpdatedAt(); !ok {
		v := systemcomponent.DefaultUpdatedAt()
		scc.mutation.SetUpdatedAt(v)
	}
	if _, ok := scc.mutation.ID(); !ok {
		v := systemcomponent.DefaultID()
		scc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (scc *SystemComponentCreate) check() error {
	if _, ok := scc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "SystemComponent.name"`)}
	}
	if v, ok := scc.mutation.Name(); ok {
		if err := systemcomponent.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "SystemComponent.name": %w`, err)}
		}
	}
	if _, ok := scc.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "SystemComponent.type"`)}
	}
	if v, ok := scc.mutation.GetType(); ok {
		if err := systemcomponent.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "SystemComponent.type": %w`, err)}
		}
	}
	if _, ok := scc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SystemComponent.created_at"`)}
	}
	if _, ok := scc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "SystemComponent.updated_at"`)}
	}
	return nil
}

func (scc *SystemComponentCreate) sqlSave(ctx context.Context) (*SystemComponent, error) {
	if err := scc.check(); err != nil {
		return nil, err
	}
	_node, _spec := scc.createSpec()
	if err := sqlgraph.CreateNode(ctx, scc.driver, _spec); err != nil {
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
	scc.mutation.id = &_node.ID
	scc.mutation.done = true
	return _node, nil
}

func (scc *SystemComponentCreate) createSpec() (*SystemComponent, *sqlgraph.CreateSpec) {
	var (
		_node = &SystemComponent{config: scc.config}
		_spec = sqlgraph.NewCreateSpec(systemcomponent.Table, sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = scc.conflict
	if id, ok := scc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := scc.mutation.Name(); ok {
		_spec.SetField(systemcomponent.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := scc.mutation.GetType(); ok {
		_spec.SetField(systemcomponent.FieldType, field.TypeEnum, value)
		_node.Type = value
	}
	if value, ok := scc.mutation.Description(); ok {
		_spec.SetField(systemcomponent.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := scc.mutation.Properties(); ok {
		_spec.SetField(systemcomponent.FieldProperties, field.TypeJSON, value)
		_node.Properties = value
	}
	if value, ok := scc.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponent.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := scc.mutation.UpdatedAt(); ok {
		_spec.SetField(systemcomponent.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := scc.mutation.ParentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   systemcomponent.ParentTable,
			Columns: []string{systemcomponent.ParentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.system_component_children = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.ChildrenIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   systemcomponent.ChildrenTable,
			Columns: []string{systemcomponent.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.ControlsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemcomponent.ControlsTable,
			Columns: systemcomponent.ControlsPrimaryKey,
			Bidi:    true,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemComponentControlRelationshipCreate{config: scc.config, mutation: newSystemComponentControlRelationshipMutation(scc.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.FeedbackToIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemcomponent.FeedbackToTable,
			Columns: systemcomponent.FeedbackToPrimaryKey,
			Bidi:    true,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &SystemComponentFeedbackRelationshipCreate{config: scc.config, mutation: newSystemComponentFeedbackRelationshipMutation(scc.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemcomponent.IncidentsTable,
			Columns: systemcomponent.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &IncidentSystemComponentCreate{config: scc.config, mutation: newIncidentSystemComponentMutation(scc.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.EventsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   systemcomponent.EventsTable,
			Columns: systemcomponent.EventsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		createE := &IncidentEventSystemComponentCreate{config: scc.config, mutation: newIncidentEventSystemComponentMutation(scc.config, OpCreate)}
		createE.defaults()
		_, specE := createE.createSpec()
		edge.Target.Fields = specE.Fields
		if specE.ID.Value != nil {
			edge.Target.Fields = append(edge.Target.Fields, specE.ID)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.ControlRelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponent.ControlRelationshipsTable,
			Columns: []string{systemcomponent.ControlRelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentcontrolrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.FeedbackRelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponent.FeedbackRelationshipsTable,
			Columns: []string{systemcomponent.FeedbackRelationshipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentfeedbackrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.IncidentSystemComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponent.IncidentSystemComponentsTable,
			Columns: []string{systemcomponent.IncidentSystemComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentsystemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := scc.mutation.EventComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponent.EventComponentsTable,
			Columns: []string{systemcomponent.EventComponentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidenteventsystemcomponent.FieldID, field.TypeUUID),
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
//	client.SystemComponent.Create().
//		SetName(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (scc *SystemComponentCreate) OnConflict(opts ...sql.ConflictOption) *SystemComponentUpsertOne {
	scc.conflict = opts
	return &SystemComponentUpsertOne{
		create: scc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponent.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (scc *SystemComponentCreate) OnConflictColumns(columns ...string) *SystemComponentUpsertOne {
	scc.conflict = append(scc.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentUpsertOne{
		create: scc,
	}
}

type (
	// SystemComponentUpsertOne is the builder for "upsert"-ing
	//  one SystemComponent node.
	SystemComponentUpsertOne struct {
		create *SystemComponentCreate
	}

	// SystemComponentUpsert is the "OnConflict" setter.
	SystemComponentUpsert struct {
		*sql.UpdateSet
	}
)

// SetName sets the "name" field.
func (u *SystemComponentUpsert) SetName(v string) *SystemComponentUpsert {
	u.Set(systemcomponent.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *SystemComponentUpsert) UpdateName() *SystemComponentUpsert {
	u.SetExcluded(systemcomponent.FieldName)
	return u
}

// SetType sets the "type" field.
func (u *SystemComponentUpsert) SetType(v systemcomponent.Type) *SystemComponentUpsert {
	u.Set(systemcomponent.FieldType, v)
	return u
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *SystemComponentUpsert) UpdateType() *SystemComponentUpsert {
	u.SetExcluded(systemcomponent.FieldType)
	return u
}

// SetDescription sets the "description" field.
func (u *SystemComponentUpsert) SetDescription(v string) *SystemComponentUpsert {
	u.Set(systemcomponent.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentUpsert) UpdateDescription() *SystemComponentUpsert {
	u.SetExcluded(systemcomponent.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentUpsert) ClearDescription() *SystemComponentUpsert {
	u.SetNull(systemcomponent.FieldDescription)
	return u
}

// SetProperties sets the "properties" field.
func (u *SystemComponentUpsert) SetProperties(v map[string]interface{}) *SystemComponentUpsert {
	u.Set(systemcomponent.FieldProperties, v)
	return u
}

// UpdateProperties sets the "properties" field to the value that was provided on create.
func (u *SystemComponentUpsert) UpdateProperties() *SystemComponentUpsert {
	u.SetExcluded(systemcomponent.FieldProperties)
	return u
}

// ClearProperties clears the value of the "properties" field.
func (u *SystemComponentUpsert) ClearProperties() *SystemComponentUpsert {
	u.SetNull(systemcomponent.FieldProperties)
	return u
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentUpsert) SetCreatedAt(v time.Time) *SystemComponentUpsert {
	u.Set(systemcomponent.FieldCreatedAt, v)
	return u
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentUpsert) UpdateCreatedAt() *SystemComponentUpsert {
	u.SetExcluded(systemcomponent.FieldCreatedAt)
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *SystemComponentUpsert) SetUpdatedAt(v time.Time) *SystemComponentUpsert {
	u.Set(systemcomponent.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *SystemComponentUpsert) UpdateUpdatedAt() *SystemComponentUpsert {
	u.SetExcluded(systemcomponent.FieldUpdatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.SystemComponent.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponent.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentUpsertOne) UpdateNewValues() *SystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(systemcomponent.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponent.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *SystemComponentUpsertOne) Ignore() *SystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentUpsertOne) DoNothing() *SystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentCreate.OnConflict
// documentation for more info.
func (u *SystemComponentUpsertOne) Update(set func(*SystemComponentUpsert)) *SystemComponentUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *SystemComponentUpsertOne) SetName(v string) *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *SystemComponentUpsertOne) UpdateName() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateName()
	})
}

// SetType sets the "type" field.
func (u *SystemComponentUpsertOne) SetType(v systemcomponent.Type) *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *SystemComponentUpsertOne) UpdateType() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateType()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentUpsertOne) SetDescription(v string) *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentUpsertOne) UpdateDescription() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentUpsertOne) ClearDescription() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.ClearDescription()
	})
}

// SetProperties sets the "properties" field.
func (u *SystemComponentUpsertOne) SetProperties(v map[string]interface{}) *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetProperties(v)
	})
}

// UpdateProperties sets the "properties" field to the value that was provided on create.
func (u *SystemComponentUpsertOne) UpdateProperties() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateProperties()
	})
}

// ClearProperties clears the value of the "properties" field.
func (u *SystemComponentUpsertOne) ClearProperties() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.ClearProperties()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentUpsertOne) SetCreatedAt(v time.Time) *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentUpsertOne) UpdateCreatedAt() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateCreatedAt()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *SystemComponentUpsertOne) SetUpdatedAt(v time.Time) *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *SystemComponentUpsertOne) UpdateUpdatedAt() *SystemComponentUpsertOne {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *SystemComponentUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: SystemComponentUpsertOne.ID is not supported by MySQL driver. Use SystemComponentUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *SystemComponentUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// SystemComponentCreateBulk is the builder for creating many SystemComponent entities in bulk.
type SystemComponentCreateBulk struct {
	config
	err      error
	builders []*SystemComponentCreate
	conflict []sql.ConflictOption
}

// Save creates the SystemComponent entities in the database.
func (sccb *SystemComponentCreateBulk) Save(ctx context.Context) ([]*SystemComponent, error) {
	if sccb.err != nil {
		return nil, sccb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(sccb.builders))
	nodes := make([]*SystemComponent, len(sccb.builders))
	mutators := make([]Mutator, len(sccb.builders))
	for i := range sccb.builders {
		func(i int, root context.Context) {
			builder := sccb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SystemComponentMutation)
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
					_, err = mutators[i+1].Mutate(root, sccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = sccb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, sccb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, sccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (sccb *SystemComponentCreateBulk) SaveX(ctx context.Context) []*SystemComponent {
	v, err := sccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sccb *SystemComponentCreateBulk) Exec(ctx context.Context) error {
	_, err := sccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sccb *SystemComponentCreateBulk) ExecX(ctx context.Context) {
	if err := sccb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.SystemComponent.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.SystemComponentUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (sccb *SystemComponentCreateBulk) OnConflict(opts ...sql.ConflictOption) *SystemComponentUpsertBulk {
	sccb.conflict = opts
	return &SystemComponentUpsertBulk{
		create: sccb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.SystemComponent.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (sccb *SystemComponentCreateBulk) OnConflictColumns(columns ...string) *SystemComponentUpsertBulk {
	sccb.conflict = append(sccb.conflict, sql.ConflictColumns(columns...))
	return &SystemComponentUpsertBulk{
		create: sccb,
	}
}

// SystemComponentUpsertBulk is the builder for "upsert"-ing
// a bulk of SystemComponent nodes.
type SystemComponentUpsertBulk struct {
	create *SystemComponentCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.SystemComponent.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(systemcomponent.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *SystemComponentUpsertBulk) UpdateNewValues() *SystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(systemcomponent.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.SystemComponent.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *SystemComponentUpsertBulk) Ignore() *SystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *SystemComponentUpsertBulk) DoNothing() *SystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the SystemComponentCreateBulk.OnConflict
// documentation for more info.
func (u *SystemComponentUpsertBulk) Update(set func(*SystemComponentUpsert)) *SystemComponentUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&SystemComponentUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *SystemComponentUpsertBulk) SetName(v string) *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *SystemComponentUpsertBulk) UpdateName() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateName()
	})
}

// SetType sets the "type" field.
func (u *SystemComponentUpsertBulk) SetType(v systemcomponent.Type) *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *SystemComponentUpsertBulk) UpdateType() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateType()
	})
}

// SetDescription sets the "description" field.
func (u *SystemComponentUpsertBulk) SetDescription(v string) *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *SystemComponentUpsertBulk) UpdateDescription() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *SystemComponentUpsertBulk) ClearDescription() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.ClearDescription()
	})
}

// SetProperties sets the "properties" field.
func (u *SystemComponentUpsertBulk) SetProperties(v map[string]interface{}) *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetProperties(v)
	})
}

// UpdateProperties sets the "properties" field to the value that was provided on create.
func (u *SystemComponentUpsertBulk) UpdateProperties() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateProperties()
	})
}

// ClearProperties clears the value of the "properties" field.
func (u *SystemComponentUpsertBulk) ClearProperties() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.ClearProperties()
	})
}

// SetCreatedAt sets the "created_at" field.
func (u *SystemComponentUpsertBulk) SetCreatedAt(v time.Time) *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetCreatedAt(v)
	})
}

// UpdateCreatedAt sets the "created_at" field to the value that was provided on create.
func (u *SystemComponentUpsertBulk) UpdateCreatedAt() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateCreatedAt()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *SystemComponentUpsertBulk) SetUpdatedAt(v time.Time) *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *SystemComponentUpsertBulk) UpdateUpdatedAt() *SystemComponentUpsertBulk {
	return u.Update(func(s *SystemComponentUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *SystemComponentUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the SystemComponentCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for SystemComponentCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *SystemComponentUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

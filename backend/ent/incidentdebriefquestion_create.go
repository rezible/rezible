// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incidentdebriefmessage"
	"github.com/twohundreds/rezible/ent/incidentdebriefquestion"
	"github.com/twohundreds/rezible/ent/incidentfield"
	"github.com/twohundreds/rezible/ent/incidentrole"
	"github.com/twohundreds/rezible/ent/incidentseverity"
	"github.com/twohundreds/rezible/ent/incidenttag"
	"github.com/twohundreds/rezible/ent/incidenttype"
)

// IncidentDebriefQuestionCreate is the builder for creating a IncidentDebriefQuestion entity.
type IncidentDebriefQuestionCreate struct {
	config
	mutation *IncidentDebriefQuestionMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetContent sets the "content" field.
func (idqc *IncidentDebriefQuestionCreate) SetContent(s string) *IncidentDebriefQuestionCreate {
	idqc.mutation.SetContent(s)
	return idqc
}

// SetID sets the "id" field.
func (idqc *IncidentDebriefQuestionCreate) SetID(u uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.SetID(u)
	return idqc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (idqc *IncidentDebriefQuestionCreate) SetNillableID(u *uuid.UUID) *IncidentDebriefQuestionCreate {
	if u != nil {
		idqc.SetID(*u)
	}
	return idqc
}

// AddMessageIDs adds the "messages" edge to the IncidentDebriefMessage entity by IDs.
func (idqc *IncidentDebriefQuestionCreate) AddMessageIDs(ids ...uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.AddMessageIDs(ids...)
	return idqc
}

// AddMessages adds the "messages" edges to the IncidentDebriefMessage entity.
func (idqc *IncidentDebriefQuestionCreate) AddMessages(i ...*IncidentDebriefMessage) *IncidentDebriefQuestionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idqc.AddMessageIDs(ids...)
}

// AddIncidentFieldIDs adds the "incident_fields" edge to the IncidentField entity by IDs.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentFieldIDs(ids ...uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.AddIncidentFieldIDs(ids...)
	return idqc
}

// AddIncidentFields adds the "incident_fields" edges to the IncidentField entity.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentFields(i ...*IncidentField) *IncidentDebriefQuestionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idqc.AddIncidentFieldIDs(ids...)
}

// AddIncidentRoleIDs adds the "incident_roles" edge to the IncidentRole entity by IDs.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentRoleIDs(ids ...uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.AddIncidentRoleIDs(ids...)
	return idqc
}

// AddIncidentRoles adds the "incident_roles" edges to the IncidentRole entity.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentRoles(i ...*IncidentRole) *IncidentDebriefQuestionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idqc.AddIncidentRoleIDs(ids...)
}

// AddIncidentSeverityIDs adds the "incident_severities" edge to the IncidentSeverity entity by IDs.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentSeverityIDs(ids ...uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.AddIncidentSeverityIDs(ids...)
	return idqc
}

// AddIncidentSeverities adds the "incident_severities" edges to the IncidentSeverity entity.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentSeverities(i ...*IncidentSeverity) *IncidentDebriefQuestionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idqc.AddIncidentSeverityIDs(ids...)
}

// AddIncidentTagIDs adds the "incident_tags" edge to the IncidentTag entity by IDs.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentTagIDs(ids ...uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.AddIncidentTagIDs(ids...)
	return idqc
}

// AddIncidentTags adds the "incident_tags" edges to the IncidentTag entity.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentTags(i ...*IncidentTag) *IncidentDebriefQuestionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idqc.AddIncidentTagIDs(ids...)
}

// AddIncidentTypeIDs adds the "incident_types" edge to the IncidentType entity by IDs.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentTypeIDs(ids ...uuid.UUID) *IncidentDebriefQuestionCreate {
	idqc.mutation.AddIncidentTypeIDs(ids...)
	return idqc
}

// AddIncidentTypes adds the "incident_types" edges to the IncidentType entity.
func (idqc *IncidentDebriefQuestionCreate) AddIncidentTypes(i ...*IncidentType) *IncidentDebriefQuestionCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return idqc.AddIncidentTypeIDs(ids...)
}

// Mutation returns the IncidentDebriefQuestionMutation object of the builder.
func (idqc *IncidentDebriefQuestionCreate) Mutation() *IncidentDebriefQuestionMutation {
	return idqc.mutation
}

// Save creates the IncidentDebriefQuestion in the database.
func (idqc *IncidentDebriefQuestionCreate) Save(ctx context.Context) (*IncidentDebriefQuestion, error) {
	idqc.defaults()
	return withHooks(ctx, idqc.sqlSave, idqc.mutation, idqc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (idqc *IncidentDebriefQuestionCreate) SaveX(ctx context.Context) *IncidentDebriefQuestion {
	v, err := idqc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (idqc *IncidentDebriefQuestionCreate) Exec(ctx context.Context) error {
	_, err := idqc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (idqc *IncidentDebriefQuestionCreate) ExecX(ctx context.Context) {
	if err := idqc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (idqc *IncidentDebriefQuestionCreate) defaults() {
	if _, ok := idqc.mutation.ID(); !ok {
		v := incidentdebriefquestion.DefaultID()
		idqc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (idqc *IncidentDebriefQuestionCreate) check() error {
	if _, ok := idqc.mutation.Content(); !ok {
		return &ValidationError{Name: "content", err: errors.New(`ent: missing required field "IncidentDebriefQuestion.content"`)}
	}
	return nil
}

func (idqc *IncidentDebriefQuestionCreate) sqlSave(ctx context.Context) (*IncidentDebriefQuestion, error) {
	if err := idqc.check(); err != nil {
		return nil, err
	}
	_node, _spec := idqc.createSpec()
	if err := sqlgraph.CreateNode(ctx, idqc.driver, _spec); err != nil {
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
	idqc.mutation.id = &_node.ID
	idqc.mutation.done = true
	return _node, nil
}

func (idqc *IncidentDebriefQuestionCreate) createSpec() (*IncidentDebriefQuestion, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentDebriefQuestion{config: idqc.config}
		_spec = sqlgraph.NewCreateSpec(incidentdebriefquestion.Table, sqlgraph.NewFieldSpec(incidentdebriefquestion.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = idqc.conflict
	if id, ok := idqc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := idqc.mutation.Content(); ok {
		_spec.SetField(incidentdebriefquestion.FieldContent, field.TypeString, value)
		_node.Content = value
	}
	if nodes := idqc.mutation.MessagesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentdebriefquestion.MessagesTable,
			Columns: []string{incidentdebriefquestion.MessagesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefmessage.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := idqc.mutation.IncidentFieldsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   incidentdebriefquestion.IncidentFieldsTable,
			Columns: incidentdebriefquestion.IncidentFieldsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentfield.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := idqc.mutation.IncidentRolesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   incidentdebriefquestion.IncidentRolesTable,
			Columns: incidentdebriefquestion.IncidentRolesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := idqc.mutation.IncidentSeveritiesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   incidentdebriefquestion.IncidentSeveritiesTable,
			Columns: incidentdebriefquestion.IncidentSeveritiesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentseverity.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := idqc.mutation.IncidentTagsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   incidentdebriefquestion.IncidentTagsTable,
			Columns: incidentdebriefquestion.IncidentTagsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidenttag.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := idqc.mutation.IncidentTypesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   incidentdebriefquestion.IncidentTypesTable,
			Columns: incidentdebriefquestion.IncidentTypesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidenttype.FieldID, field.TypeUUID),
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
//	client.IncidentDebriefQuestion.Create().
//		SetContent(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentDebriefQuestionUpsert) {
//			SetContent(v+v).
//		}).
//		Exec(ctx)
func (idqc *IncidentDebriefQuestionCreate) OnConflict(opts ...sql.ConflictOption) *IncidentDebriefQuestionUpsertOne {
	idqc.conflict = opts
	return &IncidentDebriefQuestionUpsertOne{
		create: idqc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentDebriefQuestion.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (idqc *IncidentDebriefQuestionCreate) OnConflictColumns(columns ...string) *IncidentDebriefQuestionUpsertOne {
	idqc.conflict = append(idqc.conflict, sql.ConflictColumns(columns...))
	return &IncidentDebriefQuestionUpsertOne{
		create: idqc,
	}
}

type (
	// IncidentDebriefQuestionUpsertOne is the builder for "upsert"-ing
	//  one IncidentDebriefQuestion node.
	IncidentDebriefQuestionUpsertOne struct {
		create *IncidentDebriefQuestionCreate
	}

	// IncidentDebriefQuestionUpsert is the "OnConflict" setter.
	IncidentDebriefQuestionUpsert struct {
		*sql.UpdateSet
	}
)

// SetContent sets the "content" field.
func (u *IncidentDebriefQuestionUpsert) SetContent(v string) *IncidentDebriefQuestionUpsert {
	u.Set(incidentdebriefquestion.FieldContent, v)
	return u
}

// UpdateContent sets the "content" field to the value that was provided on create.
func (u *IncidentDebriefQuestionUpsert) UpdateContent() *IncidentDebriefQuestionUpsert {
	u.SetExcluded(incidentdebriefquestion.FieldContent)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentDebriefQuestion.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentdebriefquestion.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentDebriefQuestionUpsertOne) UpdateNewValues() *IncidentDebriefQuestionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidentdebriefquestion.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentDebriefQuestion.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentDebriefQuestionUpsertOne) Ignore() *IncidentDebriefQuestionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentDebriefQuestionUpsertOne) DoNothing() *IncidentDebriefQuestionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentDebriefQuestionCreate.OnConflict
// documentation for more info.
func (u *IncidentDebriefQuestionUpsertOne) Update(set func(*IncidentDebriefQuestionUpsert)) *IncidentDebriefQuestionUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentDebriefQuestionUpsert{UpdateSet: update})
	}))
	return u
}

// SetContent sets the "content" field.
func (u *IncidentDebriefQuestionUpsertOne) SetContent(v string) *IncidentDebriefQuestionUpsertOne {
	return u.Update(func(s *IncidentDebriefQuestionUpsert) {
		s.SetContent(v)
	})
}

// UpdateContent sets the "content" field to the value that was provided on create.
func (u *IncidentDebriefQuestionUpsertOne) UpdateContent() *IncidentDebriefQuestionUpsertOne {
	return u.Update(func(s *IncidentDebriefQuestionUpsert) {
		s.UpdateContent()
	})
}

// Exec executes the query.
func (u *IncidentDebriefQuestionUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentDebriefQuestionCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentDebriefQuestionUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentDebriefQuestionUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentDebriefQuestionUpsertOne.ID is not supported by MySQL driver. Use IncidentDebriefQuestionUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentDebriefQuestionUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentDebriefQuestionCreateBulk is the builder for creating many IncidentDebriefQuestion entities in bulk.
type IncidentDebriefQuestionCreateBulk struct {
	config
	err      error
	builders []*IncidentDebriefQuestionCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentDebriefQuestion entities in the database.
func (idqcb *IncidentDebriefQuestionCreateBulk) Save(ctx context.Context) ([]*IncidentDebriefQuestion, error) {
	if idqcb.err != nil {
		return nil, idqcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(idqcb.builders))
	nodes := make([]*IncidentDebriefQuestion, len(idqcb.builders))
	mutators := make([]Mutator, len(idqcb.builders))
	for i := range idqcb.builders {
		func(i int, root context.Context) {
			builder := idqcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentDebriefQuestionMutation)
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
					_, err = mutators[i+1].Mutate(root, idqcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = idqcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, idqcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, idqcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (idqcb *IncidentDebriefQuestionCreateBulk) SaveX(ctx context.Context) []*IncidentDebriefQuestion {
	v, err := idqcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (idqcb *IncidentDebriefQuestionCreateBulk) Exec(ctx context.Context) error {
	_, err := idqcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (idqcb *IncidentDebriefQuestionCreateBulk) ExecX(ctx context.Context) {
	if err := idqcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentDebriefQuestion.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentDebriefQuestionUpsert) {
//			SetContent(v+v).
//		}).
//		Exec(ctx)
func (idqcb *IncidentDebriefQuestionCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentDebriefQuestionUpsertBulk {
	idqcb.conflict = opts
	return &IncidentDebriefQuestionUpsertBulk{
		create: idqcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentDebriefQuestion.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (idqcb *IncidentDebriefQuestionCreateBulk) OnConflictColumns(columns ...string) *IncidentDebriefQuestionUpsertBulk {
	idqcb.conflict = append(idqcb.conflict, sql.ConflictColumns(columns...))
	return &IncidentDebriefQuestionUpsertBulk{
		create: idqcb,
	}
}

// IncidentDebriefQuestionUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentDebriefQuestion nodes.
type IncidentDebriefQuestionUpsertBulk struct {
	create *IncidentDebriefQuestionCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentDebriefQuestion.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentdebriefquestion.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentDebriefQuestionUpsertBulk) UpdateNewValues() *IncidentDebriefQuestionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidentdebriefquestion.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentDebriefQuestion.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentDebriefQuestionUpsertBulk) Ignore() *IncidentDebriefQuestionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentDebriefQuestionUpsertBulk) DoNothing() *IncidentDebriefQuestionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentDebriefQuestionCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentDebriefQuestionUpsertBulk) Update(set func(*IncidentDebriefQuestionUpsert)) *IncidentDebriefQuestionUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentDebriefQuestionUpsert{UpdateSet: update})
	}))
	return u
}

// SetContent sets the "content" field.
func (u *IncidentDebriefQuestionUpsertBulk) SetContent(v string) *IncidentDebriefQuestionUpsertBulk {
	return u.Update(func(s *IncidentDebriefQuestionUpsert) {
		s.SetContent(v)
	})
}

// UpdateContent sets the "content" field to the value that was provided on create.
func (u *IncidentDebriefQuestionUpsertBulk) UpdateContent() *IncidentDebriefQuestionUpsertBulk {
	return u.Update(func(s *IncidentDebriefQuestionUpsert) {
		s.UpdateContent()
	})
}

// Exec executes the query.
func (u *IncidentDebriefQuestionUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentDebriefQuestionCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentDebriefQuestionCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentDebriefQuestionUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
// Code generated by ent, DO NOT EDIT.

package systemcomponent

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLTE(FieldID, id))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldName, v))
}

// KindID applies equality check predicate on the "kind_id" field. It's identical to KindIDEQ.
func KindID(v uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldKindID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldUpdatedAt, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldContainsFold(FieldName, v))
}

// KindIDEQ applies the EQ predicate on the "kind_id" field.
func KindIDEQ(v uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldKindID, v))
}

// KindIDNEQ applies the NEQ predicate on the "kind_id" field.
func KindIDNEQ(v uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNEQ(FieldKindID, v))
}

// KindIDIn applies the In predicate on the "kind_id" field.
func KindIDIn(vs ...uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIn(FieldKindID, vs...))
}

// KindIDNotIn applies the NotIn predicate on the "kind_id" field.
func KindIDNotIn(vs ...uuid.UUID) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotIn(FieldKindID, vs...))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.SystemComponent {
	return predicate.SystemComponent(sql.FieldLTE(FieldUpdatedAt, v))
}

// HasKind applies the HasEdge predicate on the "kind" edge.
func HasKind() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, KindTable, KindColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasKindWith applies the HasEdge predicate on the "kind" edge with a given conditions (other predicates).
func HasKindWith(preds ...predicate.SystemComponentKind) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newKindStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAnalyses applies the HasEdge predicate on the "analyses" edge.
func HasAnalyses() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, AnalysesTable, AnalysesPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAnalysesWith applies the HasEdge predicate on the "analyses" edge with a given conditions (other predicates).
func HasAnalysesWith(preds ...predicate.SystemAnalysis) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newAnalysesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRelated applies the HasEdge predicate on the "related" edge.
func HasRelated() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, RelatedTable, RelatedPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRelatedWith applies the HasEdge predicate on the "related" edge with a given conditions (other predicates).
func HasRelatedWith(preds ...predicate.SystemComponent) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newRelatedStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasEvents applies the HasEdge predicate on the "events" edge.
func HasEvents() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, EventsTable, EventsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventsWith applies the HasEdge predicate on the "events" edge with a given conditions (other predicates).
func HasEventsWith(preds ...predicate.IncidentEvent) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newEventsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasConstraints applies the HasEdge predicate on the "constraints" edge.
func HasConstraints() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, ConstraintsTable, ConstraintsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasConstraintsWith applies the HasEdge predicate on the "constraints" edge with a given conditions (other predicates).
func HasConstraintsWith(preds ...predicate.SystemComponentConstraint) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newConstraintsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasControls applies the HasEdge predicate on the "controls" edge.
func HasControls() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, ControlsTable, ControlsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasControlsWith applies the HasEdge predicate on the "controls" edge with a given conditions (other predicates).
func HasControlsWith(preds ...predicate.SystemComponentControl) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newControlsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSignals applies the HasEdge predicate on the "signals" edge.
func HasSignals() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, SignalsTable, SignalsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSignalsWith applies the HasEdge predicate on the "signals" edge with a given conditions (other predicates).
func HasSignalsWith(preds ...predicate.SystemComponentSignal) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newSignalsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAnalysisComponents applies the HasEdge predicate on the "analysis_components" edge.
func HasAnalysisComponents() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, AnalysisComponentsTable, AnalysisComponentsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAnalysisComponentsWith applies the HasEdge predicate on the "analysis_components" edge with a given conditions (other predicates).
func HasAnalysisComponentsWith(preds ...predicate.SystemAnalysisComponent) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newAnalysisComponentsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRelationships applies the HasEdge predicate on the "relationships" edge.
func HasRelationships() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, RelationshipsTable, RelationshipsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRelationshipsWith applies the HasEdge predicate on the "relationships" edge with a given conditions (other predicates).
func HasRelationshipsWith(preds ...predicate.SystemRelationship) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newRelationshipsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasEventComponents applies the HasEdge predicate on the "event_components" edge.
func HasEventComponents() predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, EventComponentsTable, EventComponentsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventComponentsWith applies the HasEdge predicate on the "event_components" edge with a given conditions (other predicates).
func HasEventComponentsWith(preds ...predicate.IncidentEventSystemComponent) predicate.SystemComponent {
	return predicate.SystemComponent(func(s *sql.Selector) {
		step := newEventComponentsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemComponent) predicate.SystemComponent {
	return predicate.SystemComponent(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemComponent) predicate.SystemComponent {
	return predicate.SystemComponent(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemComponent) predicate.SystemComponent {
	return predicate.SystemComponent(sql.NotPredicates(p))
}

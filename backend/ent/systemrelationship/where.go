// Code generated by ent, DO NOT EDIT.

package systemrelationship

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldLTE(FieldID, id))
}

// SourceComponentID applies equality check predicate on the "source_component_id" field. It's identical to SourceComponentIDEQ.
func SourceComponentID(v uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldSourceComponentID, v))
}

// TargetComponentID applies equality check predicate on the "target_component_id" field. It's identical to TargetComponentIDEQ.
func TargetComponentID(v uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldTargetComponentID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldCreatedAt, v))
}

// SourceComponentIDEQ applies the EQ predicate on the "source_component_id" field.
func SourceComponentIDEQ(v uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldSourceComponentID, v))
}

// SourceComponentIDNEQ applies the NEQ predicate on the "source_component_id" field.
func SourceComponentIDNEQ(v uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNEQ(FieldSourceComponentID, v))
}

// SourceComponentIDIn applies the In predicate on the "source_component_id" field.
func SourceComponentIDIn(vs ...uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldIn(FieldSourceComponentID, vs...))
}

// SourceComponentIDNotIn applies the NotIn predicate on the "source_component_id" field.
func SourceComponentIDNotIn(vs ...uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNotIn(FieldSourceComponentID, vs...))
}

// TargetComponentIDEQ applies the EQ predicate on the "target_component_id" field.
func TargetComponentIDEQ(v uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldTargetComponentID, v))
}

// TargetComponentIDNEQ applies the NEQ predicate on the "target_component_id" field.
func TargetComponentIDNEQ(v uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNEQ(FieldTargetComponentID, v))
}

// TargetComponentIDIn applies the In predicate on the "target_component_id" field.
func TargetComponentIDIn(vs ...uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldIn(FieldTargetComponentID, vs...))
}

// TargetComponentIDNotIn applies the NotIn predicate on the "target_component_id" field.
func TargetComponentIDNotIn(vs ...uuid.UUID) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNotIn(FieldTargetComponentID, vs...))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.FieldLTE(FieldCreatedAt, v))
}

// HasSourceComponent applies the HasEdge predicate on the "source_component" edge.
func HasSourceComponent() predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, SourceComponentTable, SourceComponentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSourceComponentWith applies the HasEdge predicate on the "source_component" edge with a given conditions (other predicates).
func HasSourceComponentWith(preds ...predicate.SystemComponent) predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := newSourceComponentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTargetComponent applies the HasEdge predicate on the "target_component" edge.
func HasTargetComponent() predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, TargetComponentTable, TargetComponentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTargetComponentWith applies the HasEdge predicate on the "target_component" edge with a given conditions (other predicates).
func HasTargetComponentWith(preds ...predicate.SystemComponent) predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := newTargetComponentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasControls applies the HasEdge predicate on the "controls" edge.
func HasControls() predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, ControlsTable, ControlsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasControlsWith applies the HasEdge predicate on the "controls" edge with a given conditions (other predicates).
func HasControlsWith(preds ...predicate.SystemComponentControl) predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := newControlsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSignals applies the HasEdge predicate on the "signals" edge.
func HasSignals() predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, SignalsTable, SignalsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSignalsWith applies the HasEdge predicate on the "signals" edge with a given conditions (other predicates).
func HasSignalsWith(preds ...predicate.SystemComponentSignal) predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := newSignalsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasControlActions applies the HasEdge predicate on the "control_actions" edge.
func HasControlActions() predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, ControlActionsTable, ControlActionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasControlActionsWith applies the HasEdge predicate on the "control_actions" edge with a given conditions (other predicates).
func HasControlActionsWith(preds ...predicate.SystemRelationshipControlAction) predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := newControlActionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasFeedback applies the HasEdge predicate on the "feedback" edge.
func HasFeedback() predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, FeedbackTable, FeedbackColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasFeedbackWith applies the HasEdge predicate on the "feedback" edge with a given conditions (other predicates).
func HasFeedbackWith(preds ...predicate.SystemRelationshipFeedback) predicate.SystemRelationship {
	return predicate.SystemRelationship(func(s *sql.Selector) {
		step := newFeedbackStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemRelationship) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemRelationship) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemRelationship) predicate.SystemRelationship {
	return predicate.SystemRelationship(sql.NotPredicates(p))
}

// Code generated by ent, DO NOT EDIT.

package systemanalysisrelationship

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldLTE(FieldID, id))
}

// AnalysisID applies equality check predicate on the "analysis_id" field. It's identical to AnalysisIDEQ.
func AnalysisID(v uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldAnalysisID, v))
}

// ComponentRelationshipID applies equality check predicate on the "component_relationship_id" field. It's identical to ComponentRelationshipIDEQ.
func ComponentRelationshipID(v uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldComponentRelationshipID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldCreatedAt, v))
}

// AnalysisIDEQ applies the EQ predicate on the "analysis_id" field.
func AnalysisIDEQ(v uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldAnalysisID, v))
}

// AnalysisIDNEQ applies the NEQ predicate on the "analysis_id" field.
func AnalysisIDNEQ(v uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNEQ(FieldAnalysisID, v))
}

// AnalysisIDIn applies the In predicate on the "analysis_id" field.
func AnalysisIDIn(vs ...uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldIn(FieldAnalysisID, vs...))
}

// AnalysisIDNotIn applies the NotIn predicate on the "analysis_id" field.
func AnalysisIDNotIn(vs ...uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNotIn(FieldAnalysisID, vs...))
}

// ComponentRelationshipIDEQ applies the EQ predicate on the "component_relationship_id" field.
func ComponentRelationshipIDEQ(v uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldComponentRelationshipID, v))
}

// ComponentRelationshipIDNEQ applies the NEQ predicate on the "component_relationship_id" field.
func ComponentRelationshipIDNEQ(v uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNEQ(FieldComponentRelationshipID, v))
}

// ComponentRelationshipIDIn applies the In predicate on the "component_relationship_id" field.
func ComponentRelationshipIDIn(vs ...uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldIn(FieldComponentRelationshipID, vs...))
}

// ComponentRelationshipIDNotIn applies the NotIn predicate on the "component_relationship_id" field.
func ComponentRelationshipIDNotIn(vs ...uuid.UUID) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNotIn(FieldComponentRelationshipID, vs...))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.FieldLTE(FieldCreatedAt, v))
}

// HasSystemAnalysis applies the HasEdge predicate on the "system_analysis" edge.
func HasSystemAnalysis() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, SystemAnalysisTable, SystemAnalysisColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSystemAnalysisWith applies the HasEdge predicate on the "system_analysis" edge with a given conditions (other predicates).
func HasSystemAnalysisWith(preds ...predicate.SystemAnalysis) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := newSystemAnalysisStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasComponentRelationship applies the HasEdge predicate on the "component_relationship" edge.
func HasComponentRelationship() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ComponentRelationshipTable, ComponentRelationshipColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasComponentRelationshipWith applies the HasEdge predicate on the "component_relationship" edge with a given conditions (other predicates).
func HasComponentRelationshipWith(preds ...predicate.SystemComponentRelationship) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := newComponentRelationshipStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasControls applies the HasEdge predicate on the "controls" edge.
func HasControls() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, ControlsTable, ControlsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasControlsWith applies the HasEdge predicate on the "controls" edge with a given conditions (other predicates).
func HasControlsWith(preds ...predicate.SystemComponentControl) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := newControlsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSignals applies the HasEdge predicate on the "signals" edge.
func HasSignals() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, SignalsTable, SignalsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSignalsWith applies the HasEdge predicate on the "signals" edge with a given conditions (other predicates).
func HasSignalsWith(preds ...predicate.SystemComponentSignal) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := newSignalsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasControlActions applies the HasEdge predicate on the "control_actions" edge.
func HasControlActions() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, ControlActionsTable, ControlActionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasControlActionsWith applies the HasEdge predicate on the "control_actions" edge with a given conditions (other predicates).
func HasControlActionsWith(preds ...predicate.SystemRelationshipControlAction) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := newControlActionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasFeedbackSignals applies the HasEdge predicate on the "feedback_signals" edge.
func HasFeedbackSignals() predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, FeedbackSignalsTable, FeedbackSignalsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasFeedbackSignalsWith applies the HasEdge predicate on the "feedback_signals" edge with a given conditions (other predicates).
func HasFeedbackSignalsWith(preds ...predicate.SystemRelationshipFeedbackSignal) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(func(s *sql.Selector) {
		step := newFeedbackSignalsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemAnalysisRelationship) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemAnalysisRelationship) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemAnalysisRelationship) predicate.SystemAnalysisRelationship {
	return predicate.SystemAnalysisRelationship(sql.NotPredicates(p))
}

// Code generated by ent, DO NOT EDIT.

package systemcomponentrelationship

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldLTE(FieldID, id))
}

// SourceID applies equality check predicate on the "source_id" field. It's identical to SourceIDEQ.
func SourceID(v uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldSourceID, v))
}

// TargetID applies equality check predicate on the "target_id" field. It's identical to TargetIDEQ.
func TargetID(v uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldTargetID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldCreatedAt, v))
}

// SourceIDEQ applies the EQ predicate on the "source_id" field.
func SourceIDEQ(v uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldSourceID, v))
}

// SourceIDNEQ applies the NEQ predicate on the "source_id" field.
func SourceIDNEQ(v uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNEQ(FieldSourceID, v))
}

// SourceIDIn applies the In predicate on the "source_id" field.
func SourceIDIn(vs ...uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldIn(FieldSourceID, vs...))
}

// SourceIDNotIn applies the NotIn predicate on the "source_id" field.
func SourceIDNotIn(vs ...uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNotIn(FieldSourceID, vs...))
}

// TargetIDEQ applies the EQ predicate on the "target_id" field.
func TargetIDEQ(v uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldTargetID, v))
}

// TargetIDNEQ applies the NEQ predicate on the "target_id" field.
func TargetIDNEQ(v uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNEQ(FieldTargetID, v))
}

// TargetIDIn applies the In predicate on the "target_id" field.
func TargetIDIn(vs ...uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldIn(FieldTargetID, vs...))
}

// TargetIDNotIn applies the NotIn predicate on the "target_id" field.
func TargetIDNotIn(vs ...uuid.UUID) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNotIn(FieldTargetID, vs...))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.FieldLTE(FieldCreatedAt, v))
}

// HasSource applies the HasEdge predicate on the "source" edge.
func HasSource() predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, SourceTable, SourceColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSourceWith applies the HasEdge predicate on the "source" edge with a given conditions (other predicates).
func HasSourceWith(preds ...predicate.SystemComponent) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := newSourceStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTarget applies the HasEdge predicate on the "target" edge.
func HasTarget() predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, TargetTable, TargetColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTargetWith applies the HasEdge predicate on the "target" edge with a given conditions (other predicates).
func HasTargetWith(preds ...predicate.SystemComponent) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := newTargetStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasControlActions applies the HasEdge predicate on the "control_actions" edge.
func HasControlActions() predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, ControlActionsTable, ControlActionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasControlActionsWith applies the HasEdge predicate on the "control_actions" edge with a given conditions (other predicates).
func HasControlActionsWith(preds ...predicate.SystemComponentRelationshipControlAction) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := newControlActionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasFeedback applies the HasEdge predicate on the "feedback" edge.
func HasFeedback() predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, FeedbackTable, FeedbackColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasFeedbackWith applies the HasEdge predicate on the "feedback" edge with a given conditions (other predicates).
func HasFeedbackWith(preds ...predicate.SystemComponentRelationshipFeedback) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(func(s *sql.Selector) {
		step := newFeedbackStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemComponentRelationship) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemComponentRelationship) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemComponentRelationship) predicate.SystemComponentRelationship {
	return predicate.SystemComponentRelationship(sql.NotPredicates(p))
}

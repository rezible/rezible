// Code generated by ent, DO NOT EDIT.

package systemrelationshipfeedbacksignal

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLTE(FieldID, id))
}

// RelationshipID applies equality check predicate on the "relationship_id" field. It's identical to RelationshipIDEQ.
func RelationshipID(v uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldRelationshipID, v))
}

// SignalID applies equality check predicate on the "signal_id" field. It's identical to SignalIDEQ.
func SignalID(v uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldSignalID, v))
}

// Type applies equality check predicate on the "type" field. It's identical to TypeEQ.
func Type(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldType, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldCreatedAt, v))
}

// RelationshipIDEQ applies the EQ predicate on the "relationship_id" field.
func RelationshipIDEQ(v uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldRelationshipID, v))
}

// RelationshipIDNEQ applies the NEQ predicate on the "relationship_id" field.
func RelationshipIDNEQ(v uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNEQ(FieldRelationshipID, v))
}

// RelationshipIDIn applies the In predicate on the "relationship_id" field.
func RelationshipIDIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIn(FieldRelationshipID, vs...))
}

// RelationshipIDNotIn applies the NotIn predicate on the "relationship_id" field.
func RelationshipIDNotIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotIn(FieldRelationshipID, vs...))
}

// SignalIDEQ applies the EQ predicate on the "signal_id" field.
func SignalIDEQ(v uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldSignalID, v))
}

// SignalIDNEQ applies the NEQ predicate on the "signal_id" field.
func SignalIDNEQ(v uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNEQ(FieldSignalID, v))
}

// SignalIDIn applies the In predicate on the "signal_id" field.
func SignalIDIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIn(FieldSignalID, vs...))
}

// SignalIDNotIn applies the NotIn predicate on the "signal_id" field.
func SignalIDNotIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotIn(FieldSignalID, vs...))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotIn(FieldType, vs...))
}

// TypeGT applies the GT predicate on the "type" field.
func TypeGT(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGT(FieldType, v))
}

// TypeGTE applies the GTE predicate on the "type" field.
func TypeGTE(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGTE(FieldType, v))
}

// TypeLT applies the LT predicate on the "type" field.
func TypeLT(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLT(FieldType, v))
}

// TypeLTE applies the LTE predicate on the "type" field.
func TypeLTE(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLTE(FieldType, v))
}

// TypeContains applies the Contains predicate on the "type" field.
func TypeContains(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldContains(FieldType, v))
}

// TypeHasPrefix applies the HasPrefix predicate on the "type" field.
func TypeHasPrefix(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldHasPrefix(FieldType, v))
}

// TypeHasSuffix applies the HasSuffix predicate on the "type" field.
func TypeHasSuffix(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldHasSuffix(FieldType, v))
}

// TypeEqualFold applies the EqualFold predicate on the "type" field.
func TypeEqualFold(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEqualFold(FieldType, v))
}

// TypeContainsFold applies the ContainsFold predicate on the "type" field.
func TypeContainsFold(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldContainsFold(FieldType, v))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.FieldLTE(FieldCreatedAt, v))
}

// HasRelationship applies the HasEdge predicate on the "relationship" edge.
func HasRelationship() predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, RelationshipTable, RelationshipColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRelationshipWith applies the HasEdge predicate on the "relationship" edge with a given conditions (other predicates).
func HasRelationshipWith(preds ...predicate.SystemAnalysisRelationship) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(func(s *sql.Selector) {
		step := newRelationshipStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSignal applies the HasEdge predicate on the "signal" edge.
func HasSignal() predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, SignalTable, SignalColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSignalWith applies the HasEdge predicate on the "signal" edge with a given conditions (other predicates).
func HasSignalWith(preds ...predicate.SystemComponentSignal) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(func(s *sql.Selector) {
		step := newSignalStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemRelationshipFeedbackSignal) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemRelationshipFeedbackSignal) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemRelationshipFeedbackSignal) predicate.SystemRelationshipFeedbackSignal {
	return predicate.SystemRelationshipFeedbackSignal(sql.NotPredicates(p))
}

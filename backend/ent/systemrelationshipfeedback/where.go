// Code generated by ent, DO NOT EDIT.

package systemrelationshipfeedback

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLTE(FieldID, id))
}

// RelationshipID applies equality check predicate on the "relationship_id" field. It's identical to RelationshipIDEQ.
func RelationshipID(v uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldRelationshipID, v))
}

// SignalID applies equality check predicate on the "signal_id" field. It's identical to SignalIDEQ.
func SignalID(v uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldSignalID, v))
}

// Type applies equality check predicate on the "type" field. It's identical to TypeEQ.
func Type(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldType, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldCreatedAt, v))
}

// RelationshipIDEQ applies the EQ predicate on the "relationship_id" field.
func RelationshipIDEQ(v uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldRelationshipID, v))
}

// RelationshipIDNEQ applies the NEQ predicate on the "relationship_id" field.
func RelationshipIDNEQ(v uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNEQ(FieldRelationshipID, v))
}

// RelationshipIDIn applies the In predicate on the "relationship_id" field.
func RelationshipIDIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIn(FieldRelationshipID, vs...))
}

// RelationshipIDNotIn applies the NotIn predicate on the "relationship_id" field.
func RelationshipIDNotIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotIn(FieldRelationshipID, vs...))
}

// SignalIDEQ applies the EQ predicate on the "signal_id" field.
func SignalIDEQ(v uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldSignalID, v))
}

// SignalIDNEQ applies the NEQ predicate on the "signal_id" field.
func SignalIDNEQ(v uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNEQ(FieldSignalID, v))
}

// SignalIDIn applies the In predicate on the "signal_id" field.
func SignalIDIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIn(FieldSignalID, vs...))
}

// SignalIDNotIn applies the NotIn predicate on the "signal_id" field.
func SignalIDNotIn(vs ...uuid.UUID) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotIn(FieldSignalID, vs...))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotIn(FieldType, vs...))
}

// TypeGT applies the GT predicate on the "type" field.
func TypeGT(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGT(FieldType, v))
}

// TypeGTE applies the GTE predicate on the "type" field.
func TypeGTE(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGTE(FieldType, v))
}

// TypeLT applies the LT predicate on the "type" field.
func TypeLT(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLT(FieldType, v))
}

// TypeLTE applies the LTE predicate on the "type" field.
func TypeLTE(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLTE(FieldType, v))
}

// TypeContains applies the Contains predicate on the "type" field.
func TypeContains(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldContains(FieldType, v))
}

// TypeHasPrefix applies the HasPrefix predicate on the "type" field.
func TypeHasPrefix(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldHasPrefix(FieldType, v))
}

// TypeHasSuffix applies the HasSuffix predicate on the "type" field.
func TypeHasSuffix(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldHasSuffix(FieldType, v))
}

// TypeEqualFold applies the EqualFold predicate on the "type" field.
func TypeEqualFold(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEqualFold(FieldType, v))
}

// TypeContainsFold applies the ContainsFold predicate on the "type" field.
func TypeContainsFold(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldContainsFold(FieldType, v))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.FieldLTE(FieldCreatedAt, v))
}

// HasSignal applies the HasEdge predicate on the "signal" edge.
func HasSignal() predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, SignalTable, SignalColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSignalWith applies the HasEdge predicate on the "signal" edge with a given conditions (other predicates).
func HasSignalWith(preds ...predicate.SystemComponentSignal) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(func(s *sql.Selector) {
		step := newSignalStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRelationship applies the HasEdge predicate on the "relationship" edge.
func HasRelationship() predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, RelationshipTable, RelationshipColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRelationshipWith applies the HasEdge predicate on the "relationship" edge with a given conditions (other predicates).
func HasRelationshipWith(preds ...predicate.SystemRelationship) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(func(s *sql.Selector) {
		step := newRelationshipStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemRelationshipFeedback) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemRelationshipFeedback) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemRelationshipFeedback) predicate.SystemRelationshipFeedback {
	return predicate.SystemRelationshipFeedback(sql.NotPredicates(p))
}

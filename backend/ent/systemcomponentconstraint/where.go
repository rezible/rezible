// Code generated by ent, DO NOT EDIT.

package systemcomponentconstraint

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLTE(FieldID, id))
}

// ComponentID applies equality check predicate on the "component_id" field. It's identical to ComponentIDEQ.
func ComponentID(v uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldComponentID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldDescription, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldCreatedAt, v))
}

// ComponentIDEQ applies the EQ predicate on the "component_id" field.
func ComponentIDEQ(v uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldComponentID, v))
}

// ComponentIDNEQ applies the NEQ predicate on the "component_id" field.
func ComponentIDNEQ(v uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNEQ(FieldComponentID, v))
}

// ComponentIDIn applies the In predicate on the "component_id" field.
func ComponentIDIn(vs ...uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldIn(FieldComponentID, vs...))
}

// ComponentIDNotIn applies the NotIn predicate on the "component_id" field.
func ComponentIDNotIn(vs ...uuid.UUID) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNotIn(FieldComponentID, vs...))
}

// LabelEQ applies the EQ predicate on the "label" field.
func LabelEQ(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldLabel, v))
}

// LabelNEQ applies the NEQ predicate on the "label" field.
func LabelNEQ(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNEQ(FieldLabel, v))
}

// LabelIn applies the In predicate on the "label" field.
func LabelIn(vs ...string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldIn(FieldLabel, vs...))
}

// LabelNotIn applies the NotIn predicate on the "label" field.
func LabelNotIn(vs ...string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNotIn(FieldLabel, vs...))
}

// LabelGT applies the GT predicate on the "label" field.
func LabelGT(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGT(FieldLabel, v))
}

// LabelGTE applies the GTE predicate on the "label" field.
func LabelGTE(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGTE(FieldLabel, v))
}

// LabelLT applies the LT predicate on the "label" field.
func LabelLT(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLT(FieldLabel, v))
}

// LabelLTE applies the LTE predicate on the "label" field.
func LabelLTE(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLTE(FieldLabel, v))
}

// LabelContains applies the Contains predicate on the "label" field.
func LabelContains(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldContains(FieldLabel, v))
}

// LabelHasPrefix applies the HasPrefix predicate on the "label" field.
func LabelHasPrefix(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldHasPrefix(FieldLabel, v))
}

// LabelHasSuffix applies the HasSuffix predicate on the "label" field.
func LabelHasSuffix(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldHasSuffix(FieldLabel, v))
}

// LabelEqualFold applies the EqualFold predicate on the "label" field.
func LabelEqualFold(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEqualFold(FieldLabel, v))
}

// LabelContainsFold applies the ContainsFold predicate on the "label" field.
func LabelContainsFold(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldContainsFold(FieldLabel, v))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldContainsFold(FieldDescription, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.FieldLTE(FieldCreatedAt, v))
}

// HasComponent applies the HasEdge predicate on the "component" edge.
func HasComponent() predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ComponentTable, ComponentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasComponentWith applies the HasEdge predicate on the "component" edge with a given conditions (other predicates).
func HasComponentWith(preds ...predicate.SystemComponent) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(func(s *sql.Selector) {
		step := newComponentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemComponentConstraint) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemComponentConstraint) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemComponentConstraint) predicate.SystemComponentConstraint {
	return predicate.SystemComponentConstraint(sql.NotPredicates(p))
}

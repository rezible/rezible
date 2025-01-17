// Code generated by ent, DO NOT EDIT.

package systemanalysiscomponent

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLTE(FieldID, id))
}

// AnalysisID applies equality check predicate on the "analysis_id" field. It's identical to AnalysisIDEQ.
func AnalysisID(v uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldAnalysisID, v))
}

// ComponentID applies equality check predicate on the "component_id" field. It's identical to ComponentIDEQ.
func ComponentID(v uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldComponentID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldDescription, v))
}

// PosX applies equality check predicate on the "pos_x" field. It's identical to PosXEQ.
func PosX(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldPosX, v))
}

// PosY applies equality check predicate on the "pos_y" field. It's identical to PosYEQ.
func PosY(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldPosY, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldCreatedAt, v))
}

// AnalysisIDEQ applies the EQ predicate on the "analysis_id" field.
func AnalysisIDEQ(v uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldAnalysisID, v))
}

// AnalysisIDNEQ applies the NEQ predicate on the "analysis_id" field.
func AnalysisIDNEQ(v uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldAnalysisID, v))
}

// AnalysisIDIn applies the In predicate on the "analysis_id" field.
func AnalysisIDIn(vs ...uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldAnalysisID, vs...))
}

// AnalysisIDNotIn applies the NotIn predicate on the "analysis_id" field.
func AnalysisIDNotIn(vs ...uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldAnalysisID, vs...))
}

// ComponentIDEQ applies the EQ predicate on the "component_id" field.
func ComponentIDEQ(v uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldComponentID, v))
}

// ComponentIDNEQ applies the NEQ predicate on the "component_id" field.
func ComponentIDNEQ(v uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldComponentID, v))
}

// ComponentIDIn applies the In predicate on the "component_id" field.
func ComponentIDIn(vs ...uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldComponentID, vs...))
}

// ComponentIDNotIn applies the NotIn predicate on the "component_id" field.
func ComponentIDNotIn(vs ...uuid.UUID) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldComponentID, vs...))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldContainsFold(FieldDescription, v))
}

// PosXEQ applies the EQ predicate on the "pos_x" field.
func PosXEQ(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldPosX, v))
}

// PosXNEQ applies the NEQ predicate on the "pos_x" field.
func PosXNEQ(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldPosX, v))
}

// PosXIn applies the In predicate on the "pos_x" field.
func PosXIn(vs ...int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldPosX, vs...))
}

// PosXNotIn applies the NotIn predicate on the "pos_x" field.
func PosXNotIn(vs ...int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldPosX, vs...))
}

// PosXGT applies the GT predicate on the "pos_x" field.
func PosXGT(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGT(FieldPosX, v))
}

// PosXGTE applies the GTE predicate on the "pos_x" field.
func PosXGTE(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGTE(FieldPosX, v))
}

// PosXLT applies the LT predicate on the "pos_x" field.
func PosXLT(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLT(FieldPosX, v))
}

// PosXLTE applies the LTE predicate on the "pos_x" field.
func PosXLTE(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLTE(FieldPosX, v))
}

// PosYEQ applies the EQ predicate on the "pos_y" field.
func PosYEQ(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldPosY, v))
}

// PosYNEQ applies the NEQ predicate on the "pos_y" field.
func PosYNEQ(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldPosY, v))
}

// PosYIn applies the In predicate on the "pos_y" field.
func PosYIn(vs ...int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldPosY, vs...))
}

// PosYNotIn applies the NotIn predicate on the "pos_y" field.
func PosYNotIn(vs ...int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldPosY, vs...))
}

// PosYGT applies the GT predicate on the "pos_y" field.
func PosYGT(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGT(FieldPosY, v))
}

// PosYGTE applies the GTE predicate on the "pos_y" field.
func PosYGTE(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGTE(FieldPosY, v))
}

// PosYLT applies the LT predicate on the "pos_y" field.
func PosYLT(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLT(FieldPosY, v))
}

// PosYLTE applies the LTE predicate on the "pos_y" field.
func PosYLTE(v int) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLTE(FieldPosY, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.FieldLTE(FieldCreatedAt, v))
}

// HasAnalysis applies the HasEdge predicate on the "analysis" edge.
func HasAnalysis() predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, AnalysisTable, AnalysisColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAnalysisWith applies the HasEdge predicate on the "analysis" edge with a given conditions (other predicates).
func HasAnalysisWith(preds ...predicate.SystemAnalysis) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(func(s *sql.Selector) {
		step := newAnalysisStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasComponent applies the HasEdge predicate on the "component" edge.
func HasComponent() predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ComponentTable, ComponentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasComponentWith applies the HasEdge predicate on the "component" edge with a given conditions (other predicates).
func HasComponentWith(preds ...predicate.SystemComponent) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(func(s *sql.Selector) {
		step := newComponentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.SystemAnalysisComponent) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.SystemAnalysisComponent) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.SystemAnalysisComponent) predicate.SystemAnalysisComponent {
	return predicate.SystemAnalysisComponent(sql.NotPredicates(p))
}

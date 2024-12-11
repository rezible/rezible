// Code generated by ent, DO NOT EDIT.

package oncallusershifthandover

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLTE(FieldID, id))
}

// ShiftID applies equality check predicate on the "shift_id" field. It's identical to ShiftIDEQ.
func ShiftID(v uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldShiftID, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldUpdatedAt, v))
}

// SentAt applies equality check predicate on the "sent_at" field. It's identical to SentAtEQ.
func SentAt(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldSentAt, v))
}

// Contents applies equality check predicate on the "contents" field. It's identical to ContentsEQ.
func Contents(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldContents, v))
}

// ShiftIDEQ applies the EQ predicate on the "shift_id" field.
func ShiftIDEQ(v uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldShiftID, v))
}

// ShiftIDNEQ applies the NEQ predicate on the "shift_id" field.
func ShiftIDNEQ(v uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNEQ(FieldShiftID, v))
}

// ShiftIDIn applies the In predicate on the "shift_id" field.
func ShiftIDIn(vs ...uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIn(FieldShiftID, vs...))
}

// ShiftIDNotIn applies the NotIn predicate on the "shift_id" field.
func ShiftIDNotIn(vs ...uuid.UUID) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotIn(FieldShiftID, vs...))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLTE(FieldUpdatedAt, v))
}

// SentAtEQ applies the EQ predicate on the "sent_at" field.
func SentAtEQ(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldSentAt, v))
}

// SentAtNEQ applies the NEQ predicate on the "sent_at" field.
func SentAtNEQ(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNEQ(FieldSentAt, v))
}

// SentAtIn applies the In predicate on the "sent_at" field.
func SentAtIn(vs ...time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIn(FieldSentAt, vs...))
}

// SentAtNotIn applies the NotIn predicate on the "sent_at" field.
func SentAtNotIn(vs ...time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotIn(FieldSentAt, vs...))
}

// SentAtGT applies the GT predicate on the "sent_at" field.
func SentAtGT(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGT(FieldSentAt, v))
}

// SentAtGTE applies the GTE predicate on the "sent_at" field.
func SentAtGTE(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGTE(FieldSentAt, v))
}

// SentAtLT applies the LT predicate on the "sent_at" field.
func SentAtLT(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLT(FieldSentAt, v))
}

// SentAtLTE applies the LTE predicate on the "sent_at" field.
func SentAtLTE(v time.Time) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLTE(FieldSentAt, v))
}

// SentAtIsNil applies the IsNil predicate on the "sent_at" field.
func SentAtIsNil() predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIsNull(FieldSentAt))
}

// SentAtNotNil applies the NotNil predicate on the "sent_at" field.
func SentAtNotNil() predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotNull(FieldSentAt))
}

// ContentsEQ applies the EQ predicate on the "contents" field.
func ContentsEQ(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldEQ(FieldContents, v))
}

// ContentsNEQ applies the NEQ predicate on the "contents" field.
func ContentsNEQ(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNEQ(FieldContents, v))
}

// ContentsIn applies the In predicate on the "contents" field.
func ContentsIn(vs ...[]byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldIn(FieldContents, vs...))
}

// ContentsNotIn applies the NotIn predicate on the "contents" field.
func ContentsNotIn(vs ...[]byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldNotIn(FieldContents, vs...))
}

// ContentsGT applies the GT predicate on the "contents" field.
func ContentsGT(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGT(FieldContents, v))
}

// ContentsGTE applies the GTE predicate on the "contents" field.
func ContentsGTE(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldGTE(FieldContents, v))
}

// ContentsLT applies the LT predicate on the "contents" field.
func ContentsLT(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLT(FieldContents, v))
}

// ContentsLTE applies the LTE predicate on the "contents" field.
func ContentsLTE(v []byte) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.FieldLTE(FieldContents, v))
}

// HasShift applies the HasEdge predicate on the "shift" edge.
func HasShift() predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, ShiftTable, ShiftColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasShiftWith applies the HasEdge predicate on the "shift" edge with a given conditions (other predicates).
func HasShiftWith(preds ...predicate.OncallUserShift) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(func(s *sql.Selector) {
		step := newShiftStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallUserShiftHandover) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallUserShiftHandover) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallUserShiftHandover) predicate.OncallUserShiftHandover {
	return predicate.OncallUserShiftHandover(sql.NotPredicates(p))
}

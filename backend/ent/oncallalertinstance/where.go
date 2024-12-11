// Code generated by ent, DO NOT EDIT.

package oncallalertinstance

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldLTE(FieldID, id))
}

// AlertID applies equality check predicate on the "alert_id" field. It's identical to AlertIDEQ.
func AlertID(v uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldAlertID, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldCreatedAt, v))
}

// AckedAt applies equality check predicate on the "acked_at" field. It's identical to AckedAtEQ.
func AckedAt(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldAckedAt, v))
}

// ReceiverUserID applies equality check predicate on the "receiver_user_id" field. It's identical to ReceiverUserIDEQ.
func ReceiverUserID(v uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldReceiverUserID, v))
}

// AlertIDEQ applies the EQ predicate on the "alert_id" field.
func AlertIDEQ(v uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldAlertID, v))
}

// AlertIDNEQ applies the NEQ predicate on the "alert_id" field.
func AlertIDNEQ(v uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNEQ(FieldAlertID, v))
}

// AlertIDIn applies the In predicate on the "alert_id" field.
func AlertIDIn(vs ...uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldIn(FieldAlertID, vs...))
}

// AlertIDNotIn applies the NotIn predicate on the "alert_id" field.
func AlertIDNotIn(vs ...uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNotIn(FieldAlertID, vs...))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldLTE(FieldCreatedAt, v))
}

// AckedAtEQ applies the EQ predicate on the "acked_at" field.
func AckedAtEQ(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldAckedAt, v))
}

// AckedAtNEQ applies the NEQ predicate on the "acked_at" field.
func AckedAtNEQ(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNEQ(FieldAckedAt, v))
}

// AckedAtIn applies the In predicate on the "acked_at" field.
func AckedAtIn(vs ...time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldIn(FieldAckedAt, vs...))
}

// AckedAtNotIn applies the NotIn predicate on the "acked_at" field.
func AckedAtNotIn(vs ...time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNotIn(FieldAckedAt, vs...))
}

// AckedAtGT applies the GT predicate on the "acked_at" field.
func AckedAtGT(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldGT(FieldAckedAt, v))
}

// AckedAtGTE applies the GTE predicate on the "acked_at" field.
func AckedAtGTE(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldGTE(FieldAckedAt, v))
}

// AckedAtLT applies the LT predicate on the "acked_at" field.
func AckedAtLT(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldLT(FieldAckedAt, v))
}

// AckedAtLTE applies the LTE predicate on the "acked_at" field.
func AckedAtLTE(v time.Time) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldLTE(FieldAckedAt, v))
}

// ReceiverUserIDEQ applies the EQ predicate on the "receiver_user_id" field.
func ReceiverUserIDEQ(v uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldEQ(FieldReceiverUserID, v))
}

// ReceiverUserIDNEQ applies the NEQ predicate on the "receiver_user_id" field.
func ReceiverUserIDNEQ(v uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNEQ(FieldReceiverUserID, v))
}

// ReceiverUserIDIn applies the In predicate on the "receiver_user_id" field.
func ReceiverUserIDIn(vs ...uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldIn(FieldReceiverUserID, vs...))
}

// ReceiverUserIDNotIn applies the NotIn predicate on the "receiver_user_id" field.
func ReceiverUserIDNotIn(vs ...uuid.UUID) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.FieldNotIn(FieldReceiverUserID, vs...))
}

// HasAlert applies the HasEdge predicate on the "alert" edge.
func HasAlert() predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, AlertTable, AlertColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAlertWith applies the HasEdge predicate on the "alert" edge with a given conditions (other predicates).
func HasAlertWith(preds ...predicate.OncallAlert) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(func(s *sql.Selector) {
		step := newAlertStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasReceiver applies the HasEdge predicate on the "receiver" edge.
func HasReceiver() predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ReceiverTable, ReceiverColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasReceiverWith applies the HasEdge predicate on the "receiver" edge with a given conditions (other predicates).
func HasReceiverWith(preds ...predicate.User) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(func(s *sql.Selector) {
		step := newReceiverStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallAlertInstance) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallAlertInstance) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallAlertInstance) predicate.OncallAlertInstance {
	return predicate.OncallAlertInstance(sql.NotPredicates(p))
}

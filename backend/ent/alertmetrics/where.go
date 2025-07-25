// Code generated by ent, DO NOT EDIT.

package alertmetrics

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldLTE(FieldID, id))
}

// AlertID applies equality check predicate on the "alert_id" field. It's identical to AlertIDEQ.
func AlertID(v uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldEQ(FieldAlertID, v))
}

// AlertIDEQ applies the EQ predicate on the "alert_id" field.
func AlertIDEQ(v uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldEQ(FieldAlertID, v))
}

// AlertIDNEQ applies the NEQ predicate on the "alert_id" field.
func AlertIDNEQ(v uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldNEQ(FieldAlertID, v))
}

// AlertIDIn applies the In predicate on the "alert_id" field.
func AlertIDIn(vs ...uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldIn(FieldAlertID, vs...))
}

// AlertIDNotIn applies the NotIn predicate on the "alert_id" field.
func AlertIDNotIn(vs ...uuid.UUID) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.FieldNotIn(FieldAlertID, vs...))
}

// HasAlert applies the HasEdge predicate on the "alert" edge.
func HasAlert() predicate.AlertMetrics {
	return predicate.AlertMetrics(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, AlertTable, AlertColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAlertWith applies the HasEdge predicate on the "alert" edge with a given conditions (other predicates).
func HasAlertWith(preds ...predicate.Alert) predicate.AlertMetrics {
	return predicate.AlertMetrics(func(s *sql.Selector) {
		step := newAlertStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.AlertMetrics) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.AlertMetrics) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.AlertMetrics) predicate.AlertMetrics {
	return predicate.AlertMetrics(sql.NotPredicates(p))
}

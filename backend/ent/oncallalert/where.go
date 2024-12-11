// Code generated by ent, DO NOT EDIT.

package oncallalert

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldLTE(FieldID, id))
}

// RosterID applies equality check predicate on the "roster_id" field. It's identical to RosterIDEQ.
func RosterID(v uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldRosterID, v))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldName, v))
}

// Timestamp applies equality check predicate on the "timestamp" field. It's identical to TimestampEQ.
func Timestamp(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldTimestamp, v))
}

// RosterIDEQ applies the EQ predicate on the "roster_id" field.
func RosterIDEQ(v uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldRosterID, v))
}

// RosterIDNEQ applies the NEQ predicate on the "roster_id" field.
func RosterIDNEQ(v uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNEQ(FieldRosterID, v))
}

// RosterIDIn applies the In predicate on the "roster_id" field.
func RosterIDIn(vs ...uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldIn(FieldRosterID, vs...))
}

// RosterIDNotIn applies the NotIn predicate on the "roster_id" field.
func RosterIDNotIn(vs ...uuid.UUID) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNotIn(FieldRosterID, vs...))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldContainsFold(FieldName, v))
}

// TimestampEQ applies the EQ predicate on the "timestamp" field.
func TimestampEQ(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldEQ(FieldTimestamp, v))
}

// TimestampNEQ applies the NEQ predicate on the "timestamp" field.
func TimestampNEQ(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNEQ(FieldTimestamp, v))
}

// TimestampIn applies the In predicate on the "timestamp" field.
func TimestampIn(vs ...time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldIn(FieldTimestamp, vs...))
}

// TimestampNotIn applies the NotIn predicate on the "timestamp" field.
func TimestampNotIn(vs ...time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldNotIn(FieldTimestamp, vs...))
}

// TimestampGT applies the GT predicate on the "timestamp" field.
func TimestampGT(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldGT(FieldTimestamp, v))
}

// TimestampGTE applies the GTE predicate on the "timestamp" field.
func TimestampGTE(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldGTE(FieldTimestamp, v))
}

// TimestampLT applies the LT predicate on the "timestamp" field.
func TimestampLT(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldLT(FieldTimestamp, v))
}

// TimestampLTE applies the LTE predicate on the "timestamp" field.
func TimestampLTE(v time.Time) predicate.OncallAlert {
	return predicate.OncallAlert(sql.FieldLTE(FieldTimestamp, v))
}

// HasInstances applies the HasEdge predicate on the "instances" edge.
func HasInstances() predicate.OncallAlert {
	return predicate.OncallAlert(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, InstancesTable, InstancesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasInstancesWith applies the HasEdge predicate on the "instances" edge with a given conditions (other predicates).
func HasInstancesWith(preds ...predicate.OncallAlertInstance) predicate.OncallAlert {
	return predicate.OncallAlert(func(s *sql.Selector) {
		step := newInstancesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRoster applies the HasEdge predicate on the "roster" edge.
func HasRoster() predicate.OncallAlert {
	return predicate.OncallAlert(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, RosterTable, RosterColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRosterWith applies the HasEdge predicate on the "roster" edge with a given conditions (other predicates).
func HasRosterWith(preds ...predicate.OncallRoster) predicate.OncallAlert {
	return predicate.OncallAlert(func(s *sql.Selector) {
		step := newRosterStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallAlert) predicate.OncallAlert {
	return predicate.OncallAlert(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallAlert) predicate.OncallAlert {
	return predicate.OncallAlert(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallAlert) predicate.OncallAlert {
	return predicate.OncallAlert(sql.NotPredicates(p))
}

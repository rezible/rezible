// Code generated by ent, DO NOT EDIT.

package subscription

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Subscription {
	return predicate.Subscription(sql.FieldLTE(FieldID, id))
}

// Event applies equality check predicate on the "event" field. It's identical to EventEQ.
func Event(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldEQ(FieldEvent, v))
}

// Active applies equality check predicate on the "active" field. It's identical to ActiveEQ.
func Active(v bool) predicate.Subscription {
	return predicate.Subscription(sql.FieldEQ(FieldActive, v))
}

// EventEQ applies the EQ predicate on the "event" field.
func EventEQ(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldEQ(FieldEvent, v))
}

// EventNEQ applies the NEQ predicate on the "event" field.
func EventNEQ(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldNEQ(FieldEvent, v))
}

// EventIn applies the In predicate on the "event" field.
func EventIn(vs ...string) predicate.Subscription {
	return predicate.Subscription(sql.FieldIn(FieldEvent, vs...))
}

// EventNotIn applies the NotIn predicate on the "event" field.
func EventNotIn(vs ...string) predicate.Subscription {
	return predicate.Subscription(sql.FieldNotIn(FieldEvent, vs...))
}

// EventGT applies the GT predicate on the "event" field.
func EventGT(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldGT(FieldEvent, v))
}

// EventGTE applies the GTE predicate on the "event" field.
func EventGTE(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldGTE(FieldEvent, v))
}

// EventLT applies the LT predicate on the "event" field.
func EventLT(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldLT(FieldEvent, v))
}

// EventLTE applies the LTE predicate on the "event" field.
func EventLTE(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldLTE(FieldEvent, v))
}

// EventContains applies the Contains predicate on the "event" field.
func EventContains(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldContains(FieldEvent, v))
}

// EventHasPrefix applies the HasPrefix predicate on the "event" field.
func EventHasPrefix(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldHasPrefix(FieldEvent, v))
}

// EventHasSuffix applies the HasSuffix predicate on the "event" field.
func EventHasSuffix(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldHasSuffix(FieldEvent, v))
}

// EventEqualFold applies the EqualFold predicate on the "event" field.
func EventEqualFold(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldEqualFold(FieldEvent, v))
}

// EventContainsFold applies the ContainsFold predicate on the "event" field.
func EventContainsFold(v string) predicate.Subscription {
	return predicate.Subscription(sql.FieldContainsFold(FieldEvent, v))
}

// ActiveEQ applies the EQ predicate on the "active" field.
func ActiveEQ(v bool) predicate.Subscription {
	return predicate.Subscription(sql.FieldEQ(FieldActive, v))
}

// ActiveNEQ applies the NEQ predicate on the "active" field.
func ActiveNEQ(v bool) predicate.Subscription {
	return predicate.Subscription(sql.FieldNEQ(FieldActive, v))
}

// HasUser applies the HasEdge predicate on the "user" edge.
func HasUser() predicate.Subscription {
	return predicate.Subscription(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, UserTable, UserColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUserWith applies the HasEdge predicate on the "user" edge with a given conditions (other predicates).
func HasUserWith(preds ...predicate.User) predicate.Subscription {
	return predicate.Subscription(func(s *sql.Selector) {
		step := newUserStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTeam applies the HasEdge predicate on the "team" edge.
func HasTeam() predicate.Subscription {
	return predicate.Subscription(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, TeamTable, TeamColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTeamWith applies the HasEdge predicate on the "team" edge with a given conditions (other predicates).
func HasTeamWith(preds ...predicate.Team) predicate.Subscription {
	return predicate.Subscription(func(s *sql.Selector) {
		step := newTeamStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasIncident applies the HasEdge predicate on the "incident" edge.
func HasIncident() predicate.Subscription {
	return predicate.Subscription(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, IncidentTable, IncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentWith applies the HasEdge predicate on the "incident" edge with a given conditions (other predicates).
func HasIncidentWith(preds ...predicate.Incident) predicate.Subscription {
	return predicate.Subscription(func(s *sql.Selector) {
		step := newIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Subscription) predicate.Subscription {
	return predicate.Subscription(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Subscription) predicate.Subscription {
	return predicate.Subscription(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Subscription) predicate.Subscription {
	return predicate.Subscription(sql.NotPredicates(p))
}

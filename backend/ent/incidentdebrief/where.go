// Code generated by ent, DO NOT EDIT.

package incidentdebrief

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldLTE(FieldID, id))
}

// IncidentID applies equality check predicate on the "incident_id" field. It's identical to IncidentIDEQ.
func IncidentID(v uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldIncidentID, v))
}

// UserID applies equality check predicate on the "user_id" field. It's identical to UserIDEQ.
func UserID(v uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldUserID, v))
}

// Required applies equality check predicate on the "required" field. It's identical to RequiredEQ.
func Required(v bool) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldRequired, v))
}

// Started applies equality check predicate on the "started" field. It's identical to StartedEQ.
func Started(v bool) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldStarted, v))
}

// IncidentIDEQ applies the EQ predicate on the "incident_id" field.
func IncidentIDEQ(v uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldIncidentID, v))
}

// IncidentIDNEQ applies the NEQ predicate on the "incident_id" field.
func IncidentIDNEQ(v uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNEQ(FieldIncidentID, v))
}

// IncidentIDIn applies the In predicate on the "incident_id" field.
func IncidentIDIn(vs ...uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldIn(FieldIncidentID, vs...))
}

// IncidentIDNotIn applies the NotIn predicate on the "incident_id" field.
func IncidentIDNotIn(vs ...uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNotIn(FieldIncidentID, vs...))
}

// UserIDEQ applies the EQ predicate on the "user_id" field.
func UserIDEQ(v uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldUserID, v))
}

// UserIDNEQ applies the NEQ predicate on the "user_id" field.
func UserIDNEQ(v uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNEQ(FieldUserID, v))
}

// UserIDIn applies the In predicate on the "user_id" field.
func UserIDIn(vs ...uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldIn(FieldUserID, vs...))
}

// UserIDNotIn applies the NotIn predicate on the "user_id" field.
func UserIDNotIn(vs ...uuid.UUID) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNotIn(FieldUserID, vs...))
}

// RequiredEQ applies the EQ predicate on the "required" field.
func RequiredEQ(v bool) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldRequired, v))
}

// RequiredNEQ applies the NEQ predicate on the "required" field.
func RequiredNEQ(v bool) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNEQ(FieldRequired, v))
}

// StartedEQ applies the EQ predicate on the "started" field.
func StartedEQ(v bool) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldEQ(FieldStarted, v))
}

// StartedNEQ applies the NEQ predicate on the "started" field.
func StartedNEQ(v bool) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.FieldNEQ(FieldStarted, v))
}

// HasIncident applies the HasEdge predicate on the "incident" edge.
func HasIncident() predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, IncidentTable, IncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentWith applies the HasEdge predicate on the "incident" edge with a given conditions (other predicates).
func HasIncidentWith(preds ...predicate.Incident) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := newIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasUser applies the HasEdge predicate on the "user" edge.
func HasUser() predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, UserTable, UserColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUserWith applies the HasEdge predicate on the "user" edge with a given conditions (other predicates).
func HasUserWith(preds ...predicate.User) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := newUserStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasMessages applies the HasEdge predicate on the "messages" edge.
func HasMessages() predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, MessagesTable, MessagesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasMessagesWith applies the HasEdge predicate on the "messages" edge with a given conditions (other predicates).
func HasMessagesWith(preds ...predicate.IncidentDebriefMessage) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := newMessagesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasSuggestions applies the HasEdge predicate on the "suggestions" edge.
func HasSuggestions() predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, SuggestionsTable, SuggestionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSuggestionsWith applies the HasEdge predicate on the "suggestions" edge with a given conditions (other predicates).
func HasSuggestionsWith(preds ...predicate.IncidentDebriefSuggestion) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(func(s *sql.Selector) {
		step := newSuggestionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.IncidentDebrief) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.IncidentDebrief) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.IncidentDebrief) predicate.IncidentDebrief {
	return predicate.IncidentDebrief(sql.NotPredicates(p))
}

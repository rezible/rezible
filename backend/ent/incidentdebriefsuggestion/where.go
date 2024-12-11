// Code generated by ent, DO NOT EDIT.

package incidentdebriefsuggestion

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldLTE(FieldID, id))
}

// Content applies equality check predicate on the "content" field. It's identical to ContentEQ.
func Content(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldEQ(FieldContent, v))
}

// ContentEQ applies the EQ predicate on the "content" field.
func ContentEQ(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldEQ(FieldContent, v))
}

// ContentNEQ applies the NEQ predicate on the "content" field.
func ContentNEQ(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldNEQ(FieldContent, v))
}

// ContentIn applies the In predicate on the "content" field.
func ContentIn(vs ...string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldIn(FieldContent, vs...))
}

// ContentNotIn applies the NotIn predicate on the "content" field.
func ContentNotIn(vs ...string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldNotIn(FieldContent, vs...))
}

// ContentGT applies the GT predicate on the "content" field.
func ContentGT(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldGT(FieldContent, v))
}

// ContentGTE applies the GTE predicate on the "content" field.
func ContentGTE(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldGTE(FieldContent, v))
}

// ContentLT applies the LT predicate on the "content" field.
func ContentLT(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldLT(FieldContent, v))
}

// ContentLTE applies the LTE predicate on the "content" field.
func ContentLTE(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldLTE(FieldContent, v))
}

// ContentContains applies the Contains predicate on the "content" field.
func ContentContains(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldContains(FieldContent, v))
}

// ContentHasPrefix applies the HasPrefix predicate on the "content" field.
func ContentHasPrefix(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldHasPrefix(FieldContent, v))
}

// ContentHasSuffix applies the HasSuffix predicate on the "content" field.
func ContentHasSuffix(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldHasSuffix(FieldContent, v))
}

// ContentEqualFold applies the EqualFold predicate on the "content" field.
func ContentEqualFold(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldEqualFold(FieldContent, v))
}

// ContentContainsFold applies the ContainsFold predicate on the "content" field.
func ContentContainsFold(v string) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.FieldContainsFold(FieldContent, v))
}

// HasDebrief applies the HasEdge predicate on the "debrief" edge.
func HasDebrief() predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, DebriefTable, DebriefColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasDebriefWith applies the HasEdge predicate on the "debrief" edge with a given conditions (other predicates).
func HasDebriefWith(preds ...predicate.IncidentDebrief) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(func(s *sql.Selector) {
		step := newDebriefStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.IncidentDebriefSuggestion) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.IncidentDebriefSuggestion) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.IncidentDebriefSuggestion) predicate.IncidentDebriefSuggestion {
	return predicate.IncidentDebriefSuggestion(sql.NotPredicates(p))
}
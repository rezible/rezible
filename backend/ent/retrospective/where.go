// Code generated by ent, DO NOT EDIT.

package retrospective

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldLTE(FieldID, id))
}

// DocumentName applies equality check predicate on the "document_name" field. It's identical to DocumentNameEQ.
func DocumentName(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEQ(FieldDocumentName, v))
}

// DocumentNameEQ applies the EQ predicate on the "document_name" field.
func DocumentNameEQ(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEQ(FieldDocumentName, v))
}

// DocumentNameNEQ applies the NEQ predicate on the "document_name" field.
func DocumentNameNEQ(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNEQ(FieldDocumentName, v))
}

// DocumentNameIn applies the In predicate on the "document_name" field.
func DocumentNameIn(vs ...string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldIn(FieldDocumentName, vs...))
}

// DocumentNameNotIn applies the NotIn predicate on the "document_name" field.
func DocumentNameNotIn(vs ...string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNotIn(FieldDocumentName, vs...))
}

// DocumentNameGT applies the GT predicate on the "document_name" field.
func DocumentNameGT(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldGT(FieldDocumentName, v))
}

// DocumentNameGTE applies the GTE predicate on the "document_name" field.
func DocumentNameGTE(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldGTE(FieldDocumentName, v))
}

// DocumentNameLT applies the LT predicate on the "document_name" field.
func DocumentNameLT(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldLT(FieldDocumentName, v))
}

// DocumentNameLTE applies the LTE predicate on the "document_name" field.
func DocumentNameLTE(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldLTE(FieldDocumentName, v))
}

// DocumentNameContains applies the Contains predicate on the "document_name" field.
func DocumentNameContains(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldContains(FieldDocumentName, v))
}

// DocumentNameHasPrefix applies the HasPrefix predicate on the "document_name" field.
func DocumentNameHasPrefix(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldHasPrefix(FieldDocumentName, v))
}

// DocumentNameHasSuffix applies the HasSuffix predicate on the "document_name" field.
func DocumentNameHasSuffix(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldHasSuffix(FieldDocumentName, v))
}

// DocumentNameEqualFold applies the EqualFold predicate on the "document_name" field.
func DocumentNameEqualFold(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEqualFold(FieldDocumentName, v))
}

// DocumentNameContainsFold applies the ContainsFold predicate on the "document_name" field.
func DocumentNameContainsFold(v string) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldContainsFold(FieldDocumentName, v))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v Type) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v Type) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...Type) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...Type) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNotIn(FieldType, vs...))
}

// StateEQ applies the EQ predicate on the "state" field.
func StateEQ(v State) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldEQ(FieldState, v))
}

// StateNEQ applies the NEQ predicate on the "state" field.
func StateNEQ(v State) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNEQ(FieldState, v))
}

// StateIn applies the In predicate on the "state" field.
func StateIn(vs ...State) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldIn(FieldState, vs...))
}

// StateNotIn applies the NotIn predicate on the "state" field.
func StateNotIn(vs ...State) predicate.Retrospective {
	return predicate.Retrospective(sql.FieldNotIn(FieldState, vs...))
}

// HasIncident applies the HasEdge predicate on the "incident" edge.
func HasIncident() predicate.Retrospective {
	return predicate.Retrospective(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, IncidentTable, IncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentWith applies the HasEdge predicate on the "incident" edge with a given conditions (other predicates).
func HasIncidentWith(preds ...predicate.Incident) predicate.Retrospective {
	return predicate.Retrospective(func(s *sql.Selector) {
		step := newIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasDiscussions applies the HasEdge predicate on the "discussions" edge.
func HasDiscussions() predicate.Retrospective {
	return predicate.Retrospective(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, DiscussionsTable, DiscussionsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasDiscussionsWith applies the HasEdge predicate on the "discussions" edge with a given conditions (other predicates).
func HasDiscussionsWith(preds ...predicate.RetrospectiveDiscussion) predicate.Retrospective {
	return predicate.Retrospective(func(s *sql.Selector) {
		step := newDiscussionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Retrospective) predicate.Retrospective {
	return predicate.Retrospective(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Retrospective) predicate.Retrospective {
	return predicate.Retrospective(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Retrospective) predicate.Retrospective {
	return predicate.Retrospective(sql.NotPredicates(p))
}

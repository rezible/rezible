// Code generated by ent, DO NOT EDIT.

package retrospectivediscussion

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldLTE(FieldID, id))
}

// RetrospectiveID applies equality check predicate on the "retrospective_id" field. It's identical to RetrospectiveIDEQ.
func RetrospectiveID(v uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldEQ(FieldRetrospectiveID, v))
}

// Content applies equality check predicate on the "content" field. It's identical to ContentEQ.
func Content(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldEQ(FieldContent, v))
}

// RetrospectiveIDEQ applies the EQ predicate on the "retrospective_id" field.
func RetrospectiveIDEQ(v uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldEQ(FieldRetrospectiveID, v))
}

// RetrospectiveIDNEQ applies the NEQ predicate on the "retrospective_id" field.
func RetrospectiveIDNEQ(v uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldNEQ(FieldRetrospectiveID, v))
}

// RetrospectiveIDIn applies the In predicate on the "retrospective_id" field.
func RetrospectiveIDIn(vs ...uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldIn(FieldRetrospectiveID, vs...))
}

// RetrospectiveIDNotIn applies the NotIn predicate on the "retrospective_id" field.
func RetrospectiveIDNotIn(vs ...uuid.UUID) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldNotIn(FieldRetrospectiveID, vs...))
}

// ContentEQ applies the EQ predicate on the "content" field.
func ContentEQ(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldEQ(FieldContent, v))
}

// ContentNEQ applies the NEQ predicate on the "content" field.
func ContentNEQ(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldNEQ(FieldContent, v))
}

// ContentIn applies the In predicate on the "content" field.
func ContentIn(vs ...[]byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldIn(FieldContent, vs...))
}

// ContentNotIn applies the NotIn predicate on the "content" field.
func ContentNotIn(vs ...[]byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldNotIn(FieldContent, vs...))
}

// ContentGT applies the GT predicate on the "content" field.
func ContentGT(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldGT(FieldContent, v))
}

// ContentGTE applies the GTE predicate on the "content" field.
func ContentGTE(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldGTE(FieldContent, v))
}

// ContentLT applies the LT predicate on the "content" field.
func ContentLT(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldLT(FieldContent, v))
}

// ContentLTE applies the LTE predicate on the "content" field.
func ContentLTE(v []byte) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.FieldLTE(FieldContent, v))
}

// HasRetrospective applies the HasEdge predicate on the "retrospective" edge.
func HasRetrospective() predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, RetrospectiveTable, RetrospectiveColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRetrospectiveWith applies the HasEdge predicate on the "retrospective" edge with a given conditions (other predicates).
func HasRetrospectiveWith(preds ...predicate.Retrospective) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		step := newRetrospectiveStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasReplies applies the HasEdge predicate on the "replies" edge.
func HasReplies() predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, RepliesTable, RepliesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRepliesWith applies the HasEdge predicate on the "replies" edge with a given conditions (other predicates).
func HasRepliesWith(preds ...predicate.RetrospectiveDiscussionReply) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		step := newRepliesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasReview applies the HasEdge predicate on the "review" edge.
func HasReview() predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, ReviewTable, ReviewColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasReviewWith applies the HasEdge predicate on the "review" edge with a given conditions (other predicates).
func HasReviewWith(preds ...predicate.RetrospectiveReview) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		step := newReviewStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.RetrospectiveDiscussion) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.RetrospectiveDiscussion) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.RetrospectiveDiscussion) predicate.RetrospectiveDiscussion {
	return predicate.RetrospectiveDiscussion(sql.NotPredicates(p))
}

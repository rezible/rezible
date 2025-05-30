// Code generated by ent, DO NOT EDIT.

package oncallannotation

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLTE(FieldID, id))
}

// EventID applies equality check predicate on the "event_id" field. It's identical to EventIDEQ.
func EventID(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldEventID, v))
}

// RosterID applies equality check predicate on the "roster_id" field. It's identical to RosterIDEQ.
func RosterID(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldRosterID, v))
}

// CreatorID applies equality check predicate on the "creator_id" field. It's identical to CreatorIDEQ.
func CreatorID(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldCreatorID, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldCreatedAt, v))
}

// MinutesOccupied applies equality check predicate on the "minutes_occupied" field. It's identical to MinutesOccupiedEQ.
func MinutesOccupied(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldMinutesOccupied, v))
}

// Notes applies equality check predicate on the "notes" field. It's identical to NotesEQ.
func Notes(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldNotes, v))
}

// EventIDEQ applies the EQ predicate on the "event_id" field.
func EventIDEQ(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldEventID, v))
}

// EventIDNEQ applies the NEQ predicate on the "event_id" field.
func EventIDNEQ(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldEventID, v))
}

// EventIDIn applies the In predicate on the "event_id" field.
func EventIDIn(vs ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldEventID, vs...))
}

// EventIDNotIn applies the NotIn predicate on the "event_id" field.
func EventIDNotIn(vs ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldEventID, vs...))
}

// RosterIDEQ applies the EQ predicate on the "roster_id" field.
func RosterIDEQ(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldRosterID, v))
}

// RosterIDNEQ applies the NEQ predicate on the "roster_id" field.
func RosterIDNEQ(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldRosterID, v))
}

// RosterIDIn applies the In predicate on the "roster_id" field.
func RosterIDIn(vs ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldRosterID, vs...))
}

// RosterIDNotIn applies the NotIn predicate on the "roster_id" field.
func RosterIDNotIn(vs ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldRosterID, vs...))
}

// CreatorIDEQ applies the EQ predicate on the "creator_id" field.
func CreatorIDEQ(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldCreatorID, v))
}

// CreatorIDNEQ applies the NEQ predicate on the "creator_id" field.
func CreatorIDNEQ(v uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldCreatorID, v))
}

// CreatorIDIn applies the In predicate on the "creator_id" field.
func CreatorIDIn(vs ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldCreatorID, vs...))
}

// CreatorIDNotIn applies the NotIn predicate on the "creator_id" field.
func CreatorIDNotIn(vs ...uuid.UUID) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldCreatorID, vs...))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLTE(FieldCreatedAt, v))
}

// MinutesOccupiedEQ applies the EQ predicate on the "minutes_occupied" field.
func MinutesOccupiedEQ(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldMinutesOccupied, v))
}

// MinutesOccupiedNEQ applies the NEQ predicate on the "minutes_occupied" field.
func MinutesOccupiedNEQ(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldMinutesOccupied, v))
}

// MinutesOccupiedIn applies the In predicate on the "minutes_occupied" field.
func MinutesOccupiedIn(vs ...int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldMinutesOccupied, vs...))
}

// MinutesOccupiedNotIn applies the NotIn predicate on the "minutes_occupied" field.
func MinutesOccupiedNotIn(vs ...int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldMinutesOccupied, vs...))
}

// MinutesOccupiedGT applies the GT predicate on the "minutes_occupied" field.
func MinutesOccupiedGT(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGT(FieldMinutesOccupied, v))
}

// MinutesOccupiedGTE applies the GTE predicate on the "minutes_occupied" field.
func MinutesOccupiedGTE(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGTE(FieldMinutesOccupied, v))
}

// MinutesOccupiedLT applies the LT predicate on the "minutes_occupied" field.
func MinutesOccupiedLT(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLT(FieldMinutesOccupied, v))
}

// MinutesOccupiedLTE applies the LTE predicate on the "minutes_occupied" field.
func MinutesOccupiedLTE(v int) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLTE(FieldMinutesOccupied, v))
}

// NotesEQ applies the EQ predicate on the "notes" field.
func NotesEQ(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEQ(FieldNotes, v))
}

// NotesNEQ applies the NEQ predicate on the "notes" field.
func NotesNEQ(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNEQ(FieldNotes, v))
}

// NotesIn applies the In predicate on the "notes" field.
func NotesIn(vs ...string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldIn(FieldNotes, vs...))
}

// NotesNotIn applies the NotIn predicate on the "notes" field.
func NotesNotIn(vs ...string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldNotIn(FieldNotes, vs...))
}

// NotesGT applies the GT predicate on the "notes" field.
func NotesGT(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGT(FieldNotes, v))
}

// NotesGTE applies the GTE predicate on the "notes" field.
func NotesGTE(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldGTE(FieldNotes, v))
}

// NotesLT applies the LT predicate on the "notes" field.
func NotesLT(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLT(FieldNotes, v))
}

// NotesLTE applies the LTE predicate on the "notes" field.
func NotesLTE(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldLTE(FieldNotes, v))
}

// NotesContains applies the Contains predicate on the "notes" field.
func NotesContains(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldContains(FieldNotes, v))
}

// NotesHasPrefix applies the HasPrefix predicate on the "notes" field.
func NotesHasPrefix(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldHasPrefix(FieldNotes, v))
}

// NotesHasSuffix applies the HasSuffix predicate on the "notes" field.
func NotesHasSuffix(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldHasSuffix(FieldNotes, v))
}

// NotesEqualFold applies the EqualFold predicate on the "notes" field.
func NotesEqualFold(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldEqualFold(FieldNotes, v))
}

// NotesContainsFold applies the ContainsFold predicate on the "notes" field.
func NotesContainsFold(v string) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.FieldContainsFold(FieldNotes, v))
}

// HasEvent applies the HasEdge predicate on the "event" edge.
func HasEvent() predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, EventTable, EventColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventWith applies the HasEdge predicate on the "event" edge with a given conditions (other predicates).
func HasEventWith(preds ...predicate.OncallEvent) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := newEventStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRoster applies the HasEdge predicate on the "roster" edge.
func HasRoster() predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, RosterTable, RosterColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRosterWith applies the HasEdge predicate on the "roster" edge with a given conditions (other predicates).
func HasRosterWith(preds ...predicate.OncallRoster) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := newRosterStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasCreator applies the HasEdge predicate on the "creator" edge.
func HasCreator() predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, CreatorTable, CreatorColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCreatorWith applies the HasEdge predicate on the "creator" edge with a given conditions (other predicates).
func HasCreatorWith(preds ...predicate.User) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := newCreatorStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAlertFeedback applies the HasEdge predicate on the "alert_feedback" edge.
func HasAlertFeedback() predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, AlertFeedbackTable, AlertFeedbackColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAlertFeedbackWith applies the HasEdge predicate on the "alert_feedback" edge with a given conditions (other predicates).
func HasAlertFeedbackWith(preds ...predicate.OncallAnnotationAlertFeedback) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := newAlertFeedbackStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasHandovers applies the HasEdge predicate on the "handovers" edge.
func HasHandovers() predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, HandoversTable, HandoversPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasHandoversWith applies the HasEdge predicate on the "handovers" edge with a given conditions (other predicates).
func HasHandoversWith(preds ...predicate.OncallUserShiftHandover) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(func(s *sql.Selector) {
		step := newHandoversStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallAnnotation) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallAnnotation) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallAnnotation) predicate.OncallAnnotation {
	return predicate.OncallAnnotation(sql.NotPredicates(p))
}

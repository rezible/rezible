// Code generated by ent, DO NOT EDIT.

package oncallusershiftannotation

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLTE(FieldID, id))
}

// ShiftID applies equality check predicate on the "shift_id" field. It's identical to ShiftIDEQ.
func ShiftID(v uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldShiftID, v))
}

// EventID applies equality check predicate on the "event_id" field. It's identical to EventIDEQ.
func EventID(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldEventID, v))
}

// MinutesOccupied applies equality check predicate on the "minutes_occupied" field. It's identical to MinutesOccupiedEQ.
func MinutesOccupied(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldMinutesOccupied, v))
}

// Notes applies equality check predicate on the "notes" field. It's identical to NotesEQ.
func Notes(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldNotes, v))
}

// Pinned applies equality check predicate on the "pinned" field. It's identical to PinnedEQ.
func Pinned(v bool) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldPinned, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldCreatedAt, v))
}

// ShiftIDEQ applies the EQ predicate on the "shift_id" field.
func ShiftIDEQ(v uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldShiftID, v))
}

// ShiftIDNEQ applies the NEQ predicate on the "shift_id" field.
func ShiftIDNEQ(v uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldShiftID, v))
}

// ShiftIDIn applies the In predicate on the "shift_id" field.
func ShiftIDIn(vs ...uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldIn(FieldShiftID, vs...))
}

// ShiftIDNotIn applies the NotIn predicate on the "shift_id" field.
func ShiftIDNotIn(vs ...uuid.UUID) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNotIn(FieldShiftID, vs...))
}

// EventIDEQ applies the EQ predicate on the "event_id" field.
func EventIDEQ(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldEventID, v))
}

// EventIDNEQ applies the NEQ predicate on the "event_id" field.
func EventIDNEQ(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldEventID, v))
}

// EventIDIn applies the In predicate on the "event_id" field.
func EventIDIn(vs ...string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldIn(FieldEventID, vs...))
}

// EventIDNotIn applies the NotIn predicate on the "event_id" field.
func EventIDNotIn(vs ...string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNotIn(FieldEventID, vs...))
}

// EventIDGT applies the GT predicate on the "event_id" field.
func EventIDGT(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGT(FieldEventID, v))
}

// EventIDGTE applies the GTE predicate on the "event_id" field.
func EventIDGTE(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGTE(FieldEventID, v))
}

// EventIDLT applies the LT predicate on the "event_id" field.
func EventIDLT(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLT(FieldEventID, v))
}

// EventIDLTE applies the LTE predicate on the "event_id" field.
func EventIDLTE(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLTE(FieldEventID, v))
}

// EventIDContains applies the Contains predicate on the "event_id" field.
func EventIDContains(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldContains(FieldEventID, v))
}

// EventIDHasPrefix applies the HasPrefix predicate on the "event_id" field.
func EventIDHasPrefix(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldHasPrefix(FieldEventID, v))
}

// EventIDHasSuffix applies the HasSuffix predicate on the "event_id" field.
func EventIDHasSuffix(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldHasSuffix(FieldEventID, v))
}

// EventIDEqualFold applies the EqualFold predicate on the "event_id" field.
func EventIDEqualFold(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEqualFold(FieldEventID, v))
}

// EventIDContainsFold applies the ContainsFold predicate on the "event_id" field.
func EventIDContainsFold(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldContainsFold(FieldEventID, v))
}

// MinutesOccupiedEQ applies the EQ predicate on the "minutes_occupied" field.
func MinutesOccupiedEQ(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldMinutesOccupied, v))
}

// MinutesOccupiedNEQ applies the NEQ predicate on the "minutes_occupied" field.
func MinutesOccupiedNEQ(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldMinutesOccupied, v))
}

// MinutesOccupiedIn applies the In predicate on the "minutes_occupied" field.
func MinutesOccupiedIn(vs ...int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldIn(FieldMinutesOccupied, vs...))
}

// MinutesOccupiedNotIn applies the NotIn predicate on the "minutes_occupied" field.
func MinutesOccupiedNotIn(vs ...int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNotIn(FieldMinutesOccupied, vs...))
}

// MinutesOccupiedGT applies the GT predicate on the "minutes_occupied" field.
func MinutesOccupiedGT(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGT(FieldMinutesOccupied, v))
}

// MinutesOccupiedGTE applies the GTE predicate on the "minutes_occupied" field.
func MinutesOccupiedGTE(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGTE(FieldMinutesOccupied, v))
}

// MinutesOccupiedLT applies the LT predicate on the "minutes_occupied" field.
func MinutesOccupiedLT(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLT(FieldMinutesOccupied, v))
}

// MinutesOccupiedLTE applies the LTE predicate on the "minutes_occupied" field.
func MinutesOccupiedLTE(v int) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLTE(FieldMinutesOccupied, v))
}

// NotesEQ applies the EQ predicate on the "notes" field.
func NotesEQ(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldNotes, v))
}

// NotesNEQ applies the NEQ predicate on the "notes" field.
func NotesNEQ(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldNotes, v))
}

// NotesIn applies the In predicate on the "notes" field.
func NotesIn(vs ...string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldIn(FieldNotes, vs...))
}

// NotesNotIn applies the NotIn predicate on the "notes" field.
func NotesNotIn(vs ...string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNotIn(FieldNotes, vs...))
}

// NotesGT applies the GT predicate on the "notes" field.
func NotesGT(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGT(FieldNotes, v))
}

// NotesGTE applies the GTE predicate on the "notes" field.
func NotesGTE(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGTE(FieldNotes, v))
}

// NotesLT applies the LT predicate on the "notes" field.
func NotesLT(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLT(FieldNotes, v))
}

// NotesLTE applies the LTE predicate on the "notes" field.
func NotesLTE(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLTE(FieldNotes, v))
}

// NotesContains applies the Contains predicate on the "notes" field.
func NotesContains(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldContains(FieldNotes, v))
}

// NotesHasPrefix applies the HasPrefix predicate on the "notes" field.
func NotesHasPrefix(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldHasPrefix(FieldNotes, v))
}

// NotesHasSuffix applies the HasSuffix predicate on the "notes" field.
func NotesHasSuffix(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldHasSuffix(FieldNotes, v))
}

// NotesEqualFold applies the EqualFold predicate on the "notes" field.
func NotesEqualFold(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEqualFold(FieldNotes, v))
}

// NotesContainsFold applies the ContainsFold predicate on the "notes" field.
func NotesContainsFold(v string) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldContainsFold(FieldNotes, v))
}

// PinnedEQ applies the EQ predicate on the "pinned" field.
func PinnedEQ(v bool) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldPinned, v))
}

// PinnedNEQ applies the NEQ predicate on the "pinned" field.
func PinnedNEQ(v bool) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldPinned, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.FieldLTE(FieldCreatedAt, v))
}

// HasShift applies the HasEdge predicate on the "shift" edge.
func HasShift() predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, ShiftTable, ShiftColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasShiftWith applies the HasEdge predicate on the "shift" edge with a given conditions (other predicates).
func HasShiftWith(preds ...predicate.OncallUserShift) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(func(s *sql.Selector) {
		step := newShiftStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallUserShiftAnnotation) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallUserShiftAnnotation) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallUserShiftAnnotation) predicate.OncallUserShiftAnnotation {
	return predicate.OncallUserShiftAnnotation(sql.NotPredicates(p))
}

// Code generated by ent, DO NOT EDIT.

package oncallusershift

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLTE(FieldID, id))
}

// UserID applies equality check predicate on the "user_id" field. It's identical to UserIDEQ.
func UserID(v uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldUserID, v))
}

// RosterID applies equality check predicate on the "roster_id" field. It's identical to RosterIDEQ.
func RosterID(v uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldRosterID, v))
}

// StartAt applies equality check predicate on the "start_at" field. It's identical to StartAtEQ.
func StartAt(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldStartAt, v))
}

// EndAt applies equality check predicate on the "end_at" field. It's identical to EndAtEQ.
func EndAt(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldEndAt, v))
}

// ProviderID applies equality check predicate on the "provider_id" field. It's identical to ProviderIDEQ.
func ProviderID(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldProviderID, v))
}

// UserIDEQ applies the EQ predicate on the "user_id" field.
func UserIDEQ(v uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldUserID, v))
}

// UserIDNEQ applies the NEQ predicate on the "user_id" field.
func UserIDNEQ(v uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNEQ(FieldUserID, v))
}

// UserIDIn applies the In predicate on the "user_id" field.
func UserIDIn(vs ...uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIn(FieldUserID, vs...))
}

// UserIDNotIn applies the NotIn predicate on the "user_id" field.
func UserIDNotIn(vs ...uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotIn(FieldUserID, vs...))
}

// RosterIDEQ applies the EQ predicate on the "roster_id" field.
func RosterIDEQ(v uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldRosterID, v))
}

// RosterIDNEQ applies the NEQ predicate on the "roster_id" field.
func RosterIDNEQ(v uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNEQ(FieldRosterID, v))
}

// RosterIDIn applies the In predicate on the "roster_id" field.
func RosterIDIn(vs ...uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIn(FieldRosterID, vs...))
}

// RosterIDNotIn applies the NotIn predicate on the "roster_id" field.
func RosterIDNotIn(vs ...uuid.UUID) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotIn(FieldRosterID, vs...))
}

// StartAtEQ applies the EQ predicate on the "start_at" field.
func StartAtEQ(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldStartAt, v))
}

// StartAtNEQ applies the NEQ predicate on the "start_at" field.
func StartAtNEQ(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNEQ(FieldStartAt, v))
}

// StartAtIn applies the In predicate on the "start_at" field.
func StartAtIn(vs ...time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIn(FieldStartAt, vs...))
}

// StartAtNotIn applies the NotIn predicate on the "start_at" field.
func StartAtNotIn(vs ...time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotIn(FieldStartAt, vs...))
}

// StartAtGT applies the GT predicate on the "start_at" field.
func StartAtGT(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGT(FieldStartAt, v))
}

// StartAtGTE applies the GTE predicate on the "start_at" field.
func StartAtGTE(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGTE(FieldStartAt, v))
}

// StartAtLT applies the LT predicate on the "start_at" field.
func StartAtLT(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLT(FieldStartAt, v))
}

// StartAtLTE applies the LTE predicate on the "start_at" field.
func StartAtLTE(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLTE(FieldStartAt, v))
}

// EndAtEQ applies the EQ predicate on the "end_at" field.
func EndAtEQ(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldEndAt, v))
}

// EndAtNEQ applies the NEQ predicate on the "end_at" field.
func EndAtNEQ(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNEQ(FieldEndAt, v))
}

// EndAtIn applies the In predicate on the "end_at" field.
func EndAtIn(vs ...time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIn(FieldEndAt, vs...))
}

// EndAtNotIn applies the NotIn predicate on the "end_at" field.
func EndAtNotIn(vs ...time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotIn(FieldEndAt, vs...))
}

// EndAtGT applies the GT predicate on the "end_at" field.
func EndAtGT(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGT(FieldEndAt, v))
}

// EndAtGTE applies the GTE predicate on the "end_at" field.
func EndAtGTE(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGTE(FieldEndAt, v))
}

// EndAtLT applies the LT predicate on the "end_at" field.
func EndAtLT(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLT(FieldEndAt, v))
}

// EndAtLTE applies the LTE predicate on the "end_at" field.
func EndAtLTE(v time.Time) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLTE(FieldEndAt, v))
}

// ProviderIDEQ applies the EQ predicate on the "provider_id" field.
func ProviderIDEQ(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEQ(FieldProviderID, v))
}

// ProviderIDNEQ applies the NEQ predicate on the "provider_id" field.
func ProviderIDNEQ(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNEQ(FieldProviderID, v))
}

// ProviderIDIn applies the In predicate on the "provider_id" field.
func ProviderIDIn(vs ...string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIn(FieldProviderID, vs...))
}

// ProviderIDNotIn applies the NotIn predicate on the "provider_id" field.
func ProviderIDNotIn(vs ...string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotIn(FieldProviderID, vs...))
}

// ProviderIDGT applies the GT predicate on the "provider_id" field.
func ProviderIDGT(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGT(FieldProviderID, v))
}

// ProviderIDGTE applies the GTE predicate on the "provider_id" field.
func ProviderIDGTE(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldGTE(FieldProviderID, v))
}

// ProviderIDLT applies the LT predicate on the "provider_id" field.
func ProviderIDLT(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLT(FieldProviderID, v))
}

// ProviderIDLTE applies the LTE predicate on the "provider_id" field.
func ProviderIDLTE(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldLTE(FieldProviderID, v))
}

// ProviderIDContains applies the Contains predicate on the "provider_id" field.
func ProviderIDContains(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldContains(FieldProviderID, v))
}

// ProviderIDHasPrefix applies the HasPrefix predicate on the "provider_id" field.
func ProviderIDHasPrefix(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldHasPrefix(FieldProviderID, v))
}

// ProviderIDHasSuffix applies the HasSuffix predicate on the "provider_id" field.
func ProviderIDHasSuffix(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldHasSuffix(FieldProviderID, v))
}

// ProviderIDIsNil applies the IsNil predicate on the "provider_id" field.
func ProviderIDIsNil() predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldIsNull(FieldProviderID))
}

// ProviderIDNotNil applies the NotNil predicate on the "provider_id" field.
func ProviderIDNotNil() predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldNotNull(FieldProviderID))
}

// ProviderIDEqualFold applies the EqualFold predicate on the "provider_id" field.
func ProviderIDEqualFold(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldEqualFold(FieldProviderID, v))
}

// ProviderIDContainsFold applies the ContainsFold predicate on the "provider_id" field.
func ProviderIDContainsFold(v string) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.FieldContainsFold(FieldProviderID, v))
}

// HasUser applies the HasEdge predicate on the "user" edge.
func HasUser() predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, UserTable, UserColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUserWith applies the HasEdge predicate on the "user" edge with a given conditions (other predicates).
func HasUserWith(preds ...predicate.User) predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := newUserStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRoster applies the HasEdge predicate on the "roster" edge.
func HasRoster() predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, RosterTable, RosterColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRosterWith applies the HasEdge predicate on the "roster" edge with a given conditions (other predicates).
func HasRosterWith(preds ...predicate.OncallRoster) predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := newRosterStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasCovers applies the HasEdge predicate on the "covers" edge.
func HasCovers() predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, CoversTable, CoversColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCoversWith applies the HasEdge predicate on the "covers" edge with a given conditions (other predicates).
func HasCoversWith(preds ...predicate.OncallUserShiftCover) predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := newCoversStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasHandover applies the HasEdge predicate on the "handover" edge.
func HasHandover() predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, HandoverTable, HandoverColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasHandoverWith applies the HasEdge predicate on the "handover" edge with a given conditions (other predicates).
func HasHandoverWith(preds ...predicate.OncallUserShiftHandover) predicate.OncallUserShift {
	return predicate.OncallUserShift(func(s *sql.Selector) {
		step := newHandoverStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallUserShift) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallUserShift) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallUserShift) predicate.OncallUserShift {
	return predicate.OncallUserShift(sql.NotPredicates(p))
}

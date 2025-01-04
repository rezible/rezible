// Code generated by ent, DO NOT EDIT.

package team

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Team {
	return predicate.Team(sql.FieldLTE(FieldID, id))
}

// Slug applies equality check predicate on the "slug" field. It's identical to SlugEQ.
func Slug(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldSlug, v))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldName, v))
}

// ChatChannelID applies equality check predicate on the "chat_channel_id" field. It's identical to ChatChannelIDEQ.
func ChatChannelID(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldChatChannelID, v))
}

// Timezone applies equality check predicate on the "timezone" field. It's identical to TimezoneEQ.
func Timezone(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldTimezone, v))
}

// SlugEQ applies the EQ predicate on the "slug" field.
func SlugEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldSlug, v))
}

// SlugNEQ applies the NEQ predicate on the "slug" field.
func SlugNEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldNEQ(FieldSlug, v))
}

// SlugIn applies the In predicate on the "slug" field.
func SlugIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldIn(FieldSlug, vs...))
}

// SlugNotIn applies the NotIn predicate on the "slug" field.
func SlugNotIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldNotIn(FieldSlug, vs...))
}

// SlugGT applies the GT predicate on the "slug" field.
func SlugGT(v string) predicate.Team {
	return predicate.Team(sql.FieldGT(FieldSlug, v))
}

// SlugGTE applies the GTE predicate on the "slug" field.
func SlugGTE(v string) predicate.Team {
	return predicate.Team(sql.FieldGTE(FieldSlug, v))
}

// SlugLT applies the LT predicate on the "slug" field.
func SlugLT(v string) predicate.Team {
	return predicate.Team(sql.FieldLT(FieldSlug, v))
}

// SlugLTE applies the LTE predicate on the "slug" field.
func SlugLTE(v string) predicate.Team {
	return predicate.Team(sql.FieldLTE(FieldSlug, v))
}

// SlugContains applies the Contains predicate on the "slug" field.
func SlugContains(v string) predicate.Team {
	return predicate.Team(sql.FieldContains(FieldSlug, v))
}

// SlugHasPrefix applies the HasPrefix predicate on the "slug" field.
func SlugHasPrefix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasPrefix(FieldSlug, v))
}

// SlugHasSuffix applies the HasSuffix predicate on the "slug" field.
func SlugHasSuffix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasSuffix(FieldSlug, v))
}

// SlugEqualFold applies the EqualFold predicate on the "slug" field.
func SlugEqualFold(v string) predicate.Team {
	return predicate.Team(sql.FieldEqualFold(FieldSlug, v))
}

// SlugContainsFold applies the ContainsFold predicate on the "slug" field.
func SlugContainsFold(v string) predicate.Team {
	return predicate.Team(sql.FieldContainsFold(FieldSlug, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Team {
	return predicate.Team(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Team {
	return predicate.Team(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Team {
	return predicate.Team(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Team {
	return predicate.Team(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Team {
	return predicate.Team(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Team {
	return predicate.Team(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Team {
	return predicate.Team(sql.FieldContainsFold(FieldName, v))
}

// ChatChannelIDEQ applies the EQ predicate on the "chat_channel_id" field.
func ChatChannelIDEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldChatChannelID, v))
}

// ChatChannelIDNEQ applies the NEQ predicate on the "chat_channel_id" field.
func ChatChannelIDNEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldNEQ(FieldChatChannelID, v))
}

// ChatChannelIDIn applies the In predicate on the "chat_channel_id" field.
func ChatChannelIDIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldIn(FieldChatChannelID, vs...))
}

// ChatChannelIDNotIn applies the NotIn predicate on the "chat_channel_id" field.
func ChatChannelIDNotIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldNotIn(FieldChatChannelID, vs...))
}

// ChatChannelIDGT applies the GT predicate on the "chat_channel_id" field.
func ChatChannelIDGT(v string) predicate.Team {
	return predicate.Team(sql.FieldGT(FieldChatChannelID, v))
}

// ChatChannelIDGTE applies the GTE predicate on the "chat_channel_id" field.
func ChatChannelIDGTE(v string) predicate.Team {
	return predicate.Team(sql.FieldGTE(FieldChatChannelID, v))
}

// ChatChannelIDLT applies the LT predicate on the "chat_channel_id" field.
func ChatChannelIDLT(v string) predicate.Team {
	return predicate.Team(sql.FieldLT(FieldChatChannelID, v))
}

// ChatChannelIDLTE applies the LTE predicate on the "chat_channel_id" field.
func ChatChannelIDLTE(v string) predicate.Team {
	return predicate.Team(sql.FieldLTE(FieldChatChannelID, v))
}

// ChatChannelIDContains applies the Contains predicate on the "chat_channel_id" field.
func ChatChannelIDContains(v string) predicate.Team {
	return predicate.Team(sql.FieldContains(FieldChatChannelID, v))
}

// ChatChannelIDHasPrefix applies the HasPrefix predicate on the "chat_channel_id" field.
func ChatChannelIDHasPrefix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasPrefix(FieldChatChannelID, v))
}

// ChatChannelIDHasSuffix applies the HasSuffix predicate on the "chat_channel_id" field.
func ChatChannelIDHasSuffix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasSuffix(FieldChatChannelID, v))
}

// ChatChannelIDIsNil applies the IsNil predicate on the "chat_channel_id" field.
func ChatChannelIDIsNil() predicate.Team {
	return predicate.Team(sql.FieldIsNull(FieldChatChannelID))
}

// ChatChannelIDNotNil applies the NotNil predicate on the "chat_channel_id" field.
func ChatChannelIDNotNil() predicate.Team {
	return predicate.Team(sql.FieldNotNull(FieldChatChannelID))
}

// ChatChannelIDEqualFold applies the EqualFold predicate on the "chat_channel_id" field.
func ChatChannelIDEqualFold(v string) predicate.Team {
	return predicate.Team(sql.FieldEqualFold(FieldChatChannelID, v))
}

// ChatChannelIDContainsFold applies the ContainsFold predicate on the "chat_channel_id" field.
func ChatChannelIDContainsFold(v string) predicate.Team {
	return predicate.Team(sql.FieldContainsFold(FieldChatChannelID, v))
}

// TimezoneEQ applies the EQ predicate on the "timezone" field.
func TimezoneEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldEQ(FieldTimezone, v))
}

// TimezoneNEQ applies the NEQ predicate on the "timezone" field.
func TimezoneNEQ(v string) predicate.Team {
	return predicate.Team(sql.FieldNEQ(FieldTimezone, v))
}

// TimezoneIn applies the In predicate on the "timezone" field.
func TimezoneIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldIn(FieldTimezone, vs...))
}

// TimezoneNotIn applies the NotIn predicate on the "timezone" field.
func TimezoneNotIn(vs ...string) predicate.Team {
	return predicate.Team(sql.FieldNotIn(FieldTimezone, vs...))
}

// TimezoneGT applies the GT predicate on the "timezone" field.
func TimezoneGT(v string) predicate.Team {
	return predicate.Team(sql.FieldGT(FieldTimezone, v))
}

// TimezoneGTE applies the GTE predicate on the "timezone" field.
func TimezoneGTE(v string) predicate.Team {
	return predicate.Team(sql.FieldGTE(FieldTimezone, v))
}

// TimezoneLT applies the LT predicate on the "timezone" field.
func TimezoneLT(v string) predicate.Team {
	return predicate.Team(sql.FieldLT(FieldTimezone, v))
}

// TimezoneLTE applies the LTE predicate on the "timezone" field.
func TimezoneLTE(v string) predicate.Team {
	return predicate.Team(sql.FieldLTE(FieldTimezone, v))
}

// TimezoneContains applies the Contains predicate on the "timezone" field.
func TimezoneContains(v string) predicate.Team {
	return predicate.Team(sql.FieldContains(FieldTimezone, v))
}

// TimezoneHasPrefix applies the HasPrefix predicate on the "timezone" field.
func TimezoneHasPrefix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasPrefix(FieldTimezone, v))
}

// TimezoneHasSuffix applies the HasSuffix predicate on the "timezone" field.
func TimezoneHasSuffix(v string) predicate.Team {
	return predicate.Team(sql.FieldHasSuffix(FieldTimezone, v))
}

// TimezoneIsNil applies the IsNil predicate on the "timezone" field.
func TimezoneIsNil() predicate.Team {
	return predicate.Team(sql.FieldIsNull(FieldTimezone))
}

// TimezoneNotNil applies the NotNil predicate on the "timezone" field.
func TimezoneNotNil() predicate.Team {
	return predicate.Team(sql.FieldNotNull(FieldTimezone))
}

// TimezoneEqualFold applies the EqualFold predicate on the "timezone" field.
func TimezoneEqualFold(v string) predicate.Team {
	return predicate.Team(sql.FieldEqualFold(FieldTimezone, v))
}

// TimezoneContainsFold applies the ContainsFold predicate on the "timezone" field.
func TimezoneContainsFold(v string) predicate.Team {
	return predicate.Team(sql.FieldContainsFold(FieldTimezone, v))
}

// HasUsers applies the HasEdge predicate on the "users" edge.
func HasUsers() predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, UsersTable, UsersPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUsersWith applies the HasEdge predicate on the "users" edge with a given conditions (other predicates).
func HasUsersWith(preds ...predicate.User) predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := newUsersStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasOncallRosters applies the HasEdge predicate on the "oncall_rosters" edge.
func HasOncallRosters() predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, OncallRostersTable, OncallRostersPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOncallRostersWith applies the HasEdge predicate on the "oncall_rosters" edge with a given conditions (other predicates).
func HasOncallRostersWith(preds ...predicate.OncallRoster) predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := newOncallRostersStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasIncidentAssignments applies the HasEdge predicate on the "incident_assignments" edge.
func HasIncidentAssignments() predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, IncidentAssignmentsTable, IncidentAssignmentsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentAssignmentsWith applies the HasEdge predicate on the "incident_assignments" edge with a given conditions (other predicates).
func HasIncidentAssignmentsWith(preds ...predicate.IncidentTeamAssignment) predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := newIncidentAssignmentsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasScheduledMeetings applies the HasEdge predicate on the "scheduled_meetings" edge.
func HasScheduledMeetings() predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, ScheduledMeetingsTable, ScheduledMeetingsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasScheduledMeetingsWith applies the HasEdge predicate on the "scheduled_meetings" edge with a given conditions (other predicates).
func HasScheduledMeetingsWith(preds ...predicate.MeetingSchedule) predicate.Team {
	return predicate.Team(func(s *sql.Selector) {
		step := newScheduledMeetingsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Team) predicate.Team {
	return predicate.Team(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Team) predicate.Team {
	return predicate.Team(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Team) predicate.Team {
	return predicate.Team(sql.NotPredicates(p))
}

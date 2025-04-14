// Code generated by ent, DO NOT EDIT.

package user

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.User {
	return predicate.User(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldLTE(FieldID, id))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldName, v))
}

// Email applies equality check predicate on the "email" field. It's identical to EmailEQ.
func Email(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldEmail, v))
}

// ChatID applies equality check predicate on the "chat_id" field. It's identical to ChatIDEQ.
func ChatID(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldChatID, v))
}

// Timezone applies equality check predicate on the "timezone" field. It's identical to TimezoneEQ.
func Timezone(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldTimezone, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldName, v))
}

// EmailEQ applies the EQ predicate on the "email" field.
func EmailEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldEmail, v))
}

// EmailNEQ applies the NEQ predicate on the "email" field.
func EmailNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldEmail, v))
}

// EmailIn applies the In predicate on the "email" field.
func EmailIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldEmail, vs...))
}

// EmailNotIn applies the NotIn predicate on the "email" field.
func EmailNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldEmail, vs...))
}

// EmailGT applies the GT predicate on the "email" field.
func EmailGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldEmail, v))
}

// EmailGTE applies the GTE predicate on the "email" field.
func EmailGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldEmail, v))
}

// EmailLT applies the LT predicate on the "email" field.
func EmailLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldEmail, v))
}

// EmailLTE applies the LTE predicate on the "email" field.
func EmailLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldEmail, v))
}

// EmailContains applies the Contains predicate on the "email" field.
func EmailContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldEmail, v))
}

// EmailHasPrefix applies the HasPrefix predicate on the "email" field.
func EmailHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldEmail, v))
}

// EmailHasSuffix applies the HasSuffix predicate on the "email" field.
func EmailHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldEmail, v))
}

// EmailEqualFold applies the EqualFold predicate on the "email" field.
func EmailEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldEmail, v))
}

// EmailContainsFold applies the ContainsFold predicate on the "email" field.
func EmailContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldEmail, v))
}

// ChatIDEQ applies the EQ predicate on the "chat_id" field.
func ChatIDEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldChatID, v))
}

// ChatIDNEQ applies the NEQ predicate on the "chat_id" field.
func ChatIDNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldChatID, v))
}

// ChatIDIn applies the In predicate on the "chat_id" field.
func ChatIDIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldChatID, vs...))
}

// ChatIDNotIn applies the NotIn predicate on the "chat_id" field.
func ChatIDNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldChatID, vs...))
}

// ChatIDGT applies the GT predicate on the "chat_id" field.
func ChatIDGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldChatID, v))
}

// ChatIDGTE applies the GTE predicate on the "chat_id" field.
func ChatIDGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldChatID, v))
}

// ChatIDLT applies the LT predicate on the "chat_id" field.
func ChatIDLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldChatID, v))
}

// ChatIDLTE applies the LTE predicate on the "chat_id" field.
func ChatIDLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldChatID, v))
}

// ChatIDContains applies the Contains predicate on the "chat_id" field.
func ChatIDContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldChatID, v))
}

// ChatIDHasPrefix applies the HasPrefix predicate on the "chat_id" field.
func ChatIDHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldChatID, v))
}

// ChatIDHasSuffix applies the HasSuffix predicate on the "chat_id" field.
func ChatIDHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldChatID, v))
}

// ChatIDIsNil applies the IsNil predicate on the "chat_id" field.
func ChatIDIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldChatID))
}

// ChatIDNotNil applies the NotNil predicate on the "chat_id" field.
func ChatIDNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldChatID))
}

// ChatIDEqualFold applies the EqualFold predicate on the "chat_id" field.
func ChatIDEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldChatID, v))
}

// ChatIDContainsFold applies the ContainsFold predicate on the "chat_id" field.
func ChatIDContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldChatID, v))
}

// TimezoneEQ applies the EQ predicate on the "timezone" field.
func TimezoneEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldTimezone, v))
}

// TimezoneNEQ applies the NEQ predicate on the "timezone" field.
func TimezoneNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldTimezone, v))
}

// TimezoneIn applies the In predicate on the "timezone" field.
func TimezoneIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldTimezone, vs...))
}

// TimezoneNotIn applies the NotIn predicate on the "timezone" field.
func TimezoneNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldTimezone, vs...))
}

// TimezoneGT applies the GT predicate on the "timezone" field.
func TimezoneGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldTimezone, v))
}

// TimezoneGTE applies the GTE predicate on the "timezone" field.
func TimezoneGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldTimezone, v))
}

// TimezoneLT applies the LT predicate on the "timezone" field.
func TimezoneLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldTimezone, v))
}

// TimezoneLTE applies the LTE predicate on the "timezone" field.
func TimezoneLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldTimezone, v))
}

// TimezoneContains applies the Contains predicate on the "timezone" field.
func TimezoneContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldTimezone, v))
}

// TimezoneHasPrefix applies the HasPrefix predicate on the "timezone" field.
func TimezoneHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldTimezone, v))
}

// TimezoneHasSuffix applies the HasSuffix predicate on the "timezone" field.
func TimezoneHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldTimezone, v))
}

// TimezoneIsNil applies the IsNil predicate on the "timezone" field.
func TimezoneIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldTimezone))
}

// TimezoneNotNil applies the NotNil predicate on the "timezone" field.
func TimezoneNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldTimezone))
}

// TimezoneEqualFold applies the EqualFold predicate on the "timezone" field.
func TimezoneEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldTimezone, v))
}

// TimezoneContainsFold applies the ContainsFold predicate on the "timezone" field.
func TimezoneContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldTimezone, v))
}

// HasTeams applies the HasEdge predicate on the "teams" edge.
func HasTeams() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, TeamsTable, TeamsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTeamsWith applies the HasEdge predicate on the "teams" edge with a given conditions (other predicates).
func HasTeamsWith(preds ...predicate.Team) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newTeamsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasWatchedOncallRosters applies the HasEdge predicate on the "watched_oncall_rosters" edge.
func HasWatchedOncallRosters() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, WatchedOncallRostersTable, WatchedOncallRostersPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasWatchedOncallRostersWith applies the HasEdge predicate on the "watched_oncall_rosters" edge with a given conditions (other predicates).
func HasWatchedOncallRostersWith(preds ...predicate.OncallRoster) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newWatchedOncallRostersStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasOncallSchedules applies the HasEdge predicate on the "oncall_schedules" edge.
func HasOncallSchedules() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, OncallSchedulesTable, OncallSchedulesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOncallSchedulesWith applies the HasEdge predicate on the "oncall_schedules" edge with a given conditions (other predicates).
func HasOncallSchedulesWith(preds ...predicate.OncallScheduleParticipant) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newOncallSchedulesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasOncallShifts applies the HasEdge predicate on the "oncall_shifts" edge.
func HasOncallShifts() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, OncallShiftsTable, OncallShiftsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOncallShiftsWith applies the HasEdge predicate on the "oncall_shifts" edge with a given conditions (other predicates).
func HasOncallShiftsWith(preds ...predicate.OncallUserShift) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newOncallShiftsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasOncallShiftCovers applies the HasEdge predicate on the "oncall_shift_covers" edge.
func HasOncallShiftCovers() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, OncallShiftCoversTable, OncallShiftCoversColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOncallShiftCoversWith applies the HasEdge predicate on the "oncall_shift_covers" edge with a given conditions (other predicates).
func HasOncallShiftCoversWith(preds ...predicate.OncallUserShiftCover) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newOncallShiftCoversStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasOncallEventAnnotations applies the HasEdge predicate on the "oncall_event_annotations" edge.
func HasOncallEventAnnotations() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, OncallEventAnnotationsTable, OncallEventAnnotationsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOncallEventAnnotationsWith applies the HasEdge predicate on the "oncall_event_annotations" edge with a given conditions (other predicates).
func HasOncallEventAnnotationsWith(preds ...predicate.OncallEventAnnotation) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newOncallEventAnnotationsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasIncidentRoleAssignments applies the HasEdge predicate on the "incident_role_assignments" edge.
func HasIncidentRoleAssignments() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, IncidentRoleAssignmentsTable, IncidentRoleAssignmentsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentRoleAssignmentsWith applies the HasEdge predicate on the "incident_role_assignments" edge with a given conditions (other predicates).
func HasIncidentRoleAssignmentsWith(preds ...predicate.IncidentRoleAssignment) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newIncidentRoleAssignmentsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasIncidentDebriefs applies the HasEdge predicate on the "incident_debriefs" edge.
func HasIncidentDebriefs() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, IncidentDebriefsTable, IncidentDebriefsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentDebriefsWith applies the HasEdge predicate on the "incident_debriefs" edge with a given conditions (other predicates).
func HasIncidentDebriefsWith(preds ...predicate.IncidentDebrief) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newIncidentDebriefsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAssignedTasks applies the HasEdge predicate on the "assigned_tasks" edge.
func HasAssignedTasks() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, AssignedTasksTable, AssignedTasksColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAssignedTasksWith applies the HasEdge predicate on the "assigned_tasks" edge with a given conditions (other predicates).
func HasAssignedTasksWith(preds ...predicate.Task) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newAssignedTasksStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasCreatedTasks applies the HasEdge predicate on the "created_tasks" edge.
func HasCreatedTasks() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, CreatedTasksTable, CreatedTasksColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCreatedTasksWith applies the HasEdge predicate on the "created_tasks" edge with a given conditions (other predicates).
func HasCreatedTasksWith(preds ...predicate.Task) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newCreatedTasksStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRetrospectiveReviewRequests applies the HasEdge predicate on the "retrospective_review_requests" edge.
func HasRetrospectiveReviewRequests() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, RetrospectiveReviewRequestsTable, RetrospectiveReviewRequestsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRetrospectiveReviewRequestsWith applies the HasEdge predicate on the "retrospective_review_requests" edge with a given conditions (other predicates).
func HasRetrospectiveReviewRequestsWith(preds ...predicate.RetrospectiveReview) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newRetrospectiveReviewRequestsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRetrospectiveReviewResponses applies the HasEdge predicate on the "retrospective_review_responses" edge.
func HasRetrospectiveReviewResponses() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, RetrospectiveReviewResponsesTable, RetrospectiveReviewResponsesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRetrospectiveReviewResponsesWith applies the HasEdge predicate on the "retrospective_review_responses" edge with a given conditions (other predicates).
func HasRetrospectiveReviewResponsesWith(preds ...predicate.RetrospectiveReview) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := newRetrospectiveReviewResponsesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.User) predicate.User {
	return predicate.User(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.User) predicate.User {
	return predicate.User(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.User) predicate.User {
	return predicate.User(sql.NotPredicates(p))
}

// Code generated by ent, DO NOT EDIT.

package user

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the user type in the database.
	Label = "user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldEmail holds the string denoting the email field in the database.
	FieldEmail = "email"
	// FieldChatID holds the string denoting the chat_id field in the database.
	FieldChatID = "chat_id"
	// FieldTimezone holds the string denoting the timezone field in the database.
	FieldTimezone = "timezone"
	// EdgeTeams holds the string denoting the teams edge name in mutations.
	EdgeTeams = "teams"
	// EdgeWatchedOncallRosters holds the string denoting the watched_oncall_rosters edge name in mutations.
	EdgeWatchedOncallRosters = "watched_oncall_rosters"
	// EdgeOncallSchedules holds the string denoting the oncall_schedules edge name in mutations.
	EdgeOncallSchedules = "oncall_schedules"
	// EdgeOncallShifts holds the string denoting the oncall_shifts edge name in mutations.
	EdgeOncallShifts = "oncall_shifts"
	// EdgeOncallShiftCovers holds the string denoting the oncall_shift_covers edge name in mutations.
	EdgeOncallShiftCovers = "oncall_shift_covers"
	// EdgeOncallEventAnnotations holds the string denoting the oncall_event_annotations edge name in mutations.
	EdgeOncallEventAnnotations = "oncall_event_annotations"
	// EdgeIncidentRoleAssignments holds the string denoting the incident_role_assignments edge name in mutations.
	EdgeIncidentRoleAssignments = "incident_role_assignments"
	// EdgeIncidentDebriefs holds the string denoting the incident_debriefs edge name in mutations.
	EdgeIncidentDebriefs = "incident_debriefs"
	// EdgeAssignedTasks holds the string denoting the assigned_tasks edge name in mutations.
	EdgeAssignedTasks = "assigned_tasks"
	// EdgeCreatedTasks holds the string denoting the created_tasks edge name in mutations.
	EdgeCreatedTasks = "created_tasks"
	// EdgeRetrospectiveReviewRequests holds the string denoting the retrospective_review_requests edge name in mutations.
	EdgeRetrospectiveReviewRequests = "retrospective_review_requests"
	// EdgeRetrospectiveReviewResponses holds the string denoting the retrospective_review_responses edge name in mutations.
	EdgeRetrospectiveReviewResponses = "retrospective_review_responses"
	// Table holds the table name of the user in the database.
	Table = "users"
	// TeamsTable is the table that holds the teams relation/edge. The primary key declared below.
	TeamsTable = "team_users"
	// TeamsInverseTable is the table name for the Team entity.
	// It exists in this package in order to avoid circular dependency with the "team" package.
	TeamsInverseTable = "teams"
	// WatchedOncallRostersTable is the table that holds the watched_oncall_rosters relation/edge. The primary key declared below.
	WatchedOncallRostersTable = "user_watched_oncall_rosters"
	// WatchedOncallRostersInverseTable is the table name for the OncallRoster entity.
	// It exists in this package in order to avoid circular dependency with the "oncallroster" package.
	WatchedOncallRostersInverseTable = "oncall_rosters"
	// OncallSchedulesTable is the table that holds the oncall_schedules relation/edge.
	OncallSchedulesTable = "oncall_schedule_participants"
	// OncallSchedulesInverseTable is the table name for the OncallScheduleParticipant entity.
	// It exists in this package in order to avoid circular dependency with the "oncallscheduleparticipant" package.
	OncallSchedulesInverseTable = "oncall_schedule_participants"
	// OncallSchedulesColumn is the table column denoting the oncall_schedules relation/edge.
	OncallSchedulesColumn = "user_id"
	// OncallShiftsTable is the table that holds the oncall_shifts relation/edge.
	OncallShiftsTable = "oncall_user_shifts"
	// OncallShiftsInverseTable is the table name for the OncallUserShift entity.
	// It exists in this package in order to avoid circular dependency with the "oncallusershift" package.
	OncallShiftsInverseTable = "oncall_user_shifts"
	// OncallShiftsColumn is the table column denoting the oncall_shifts relation/edge.
	OncallShiftsColumn = "user_id"
	// OncallShiftCoversTable is the table that holds the oncall_shift_covers relation/edge.
	OncallShiftCoversTable = "oncall_user_shift_covers"
	// OncallShiftCoversInverseTable is the table name for the OncallUserShiftCover entity.
	// It exists in this package in order to avoid circular dependency with the "oncallusershiftcover" package.
	OncallShiftCoversInverseTable = "oncall_user_shift_covers"
	// OncallShiftCoversColumn is the table column denoting the oncall_shift_covers relation/edge.
	OncallShiftCoversColumn = "user_id"
	// OncallEventAnnotationsTable is the table that holds the oncall_event_annotations relation/edge.
	OncallEventAnnotationsTable = "oncall_event_annotations"
	// OncallEventAnnotationsInverseTable is the table name for the OncallEventAnnotation entity.
	// It exists in this package in order to avoid circular dependency with the "oncalleventannotation" package.
	OncallEventAnnotationsInverseTable = "oncall_event_annotations"
	// OncallEventAnnotationsColumn is the table column denoting the oncall_event_annotations relation/edge.
	OncallEventAnnotationsColumn = "creator_id"
	// IncidentRoleAssignmentsTable is the table that holds the incident_role_assignments relation/edge.
	IncidentRoleAssignmentsTable = "incident_role_assignments"
	// IncidentRoleAssignmentsInverseTable is the table name for the IncidentRoleAssignment entity.
	// It exists in this package in order to avoid circular dependency with the "incidentroleassignment" package.
	IncidentRoleAssignmentsInverseTable = "incident_role_assignments"
	// IncidentRoleAssignmentsColumn is the table column denoting the incident_role_assignments relation/edge.
	IncidentRoleAssignmentsColumn = "user_id"
	// IncidentDebriefsTable is the table that holds the incident_debriefs relation/edge.
	IncidentDebriefsTable = "incident_debriefs"
	// IncidentDebriefsInverseTable is the table name for the IncidentDebrief entity.
	// It exists in this package in order to avoid circular dependency with the "incidentdebrief" package.
	IncidentDebriefsInverseTable = "incident_debriefs"
	// IncidentDebriefsColumn is the table column denoting the incident_debriefs relation/edge.
	IncidentDebriefsColumn = "user_id"
	// AssignedTasksTable is the table that holds the assigned_tasks relation/edge.
	AssignedTasksTable = "tasks"
	// AssignedTasksInverseTable is the table name for the Task entity.
	// It exists in this package in order to avoid circular dependency with the "task" package.
	AssignedTasksInverseTable = "tasks"
	// AssignedTasksColumn is the table column denoting the assigned_tasks relation/edge.
	AssignedTasksColumn = "assignee_id"
	// CreatedTasksTable is the table that holds the created_tasks relation/edge.
	CreatedTasksTable = "tasks"
	// CreatedTasksInverseTable is the table name for the Task entity.
	// It exists in this package in order to avoid circular dependency with the "task" package.
	CreatedTasksInverseTable = "tasks"
	// CreatedTasksColumn is the table column denoting the created_tasks relation/edge.
	CreatedTasksColumn = "creator_id"
	// RetrospectiveReviewRequestsTable is the table that holds the retrospective_review_requests relation/edge.
	RetrospectiveReviewRequestsTable = "retrospective_reviews"
	// RetrospectiveReviewRequestsInverseTable is the table name for the RetrospectiveReview entity.
	// It exists in this package in order to avoid circular dependency with the "retrospectivereview" package.
	RetrospectiveReviewRequestsInverseTable = "retrospective_reviews"
	// RetrospectiveReviewRequestsColumn is the table column denoting the retrospective_review_requests relation/edge.
	RetrospectiveReviewRequestsColumn = "requester_id"
	// RetrospectiveReviewResponsesTable is the table that holds the retrospective_review_responses relation/edge.
	RetrospectiveReviewResponsesTable = "retrospective_reviews"
	// RetrospectiveReviewResponsesInverseTable is the table name for the RetrospectiveReview entity.
	// It exists in this package in order to avoid circular dependency with the "retrospectivereview" package.
	RetrospectiveReviewResponsesInverseTable = "retrospective_reviews"
	// RetrospectiveReviewResponsesColumn is the table column denoting the retrospective_review_responses relation/edge.
	RetrospectiveReviewResponsesColumn = "reviewer_id"
)

// Columns holds all SQL columns for user fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldEmail,
	FieldChatID,
	FieldTimezone,
}

var (
	// TeamsPrimaryKey and TeamsColumn2 are the table columns denoting the
	// primary key for the teams relation (M2M).
	TeamsPrimaryKey = []string{"team_id", "user_id"}
	// WatchedOncallRostersPrimaryKey and WatchedOncallRostersColumn2 are the table columns denoting the
	// primary key for the watched_oncall_rosters relation (M2M).
	WatchedOncallRostersPrimaryKey = []string{"user_id", "oncall_roster_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the User queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByEmail orders the results by the email field.
func ByEmail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEmail, opts...).ToFunc()
}

// ByChatID orders the results by the chat_id field.
func ByChatID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldChatID, opts...).ToFunc()
}

// ByTimezone orders the results by the timezone field.
func ByTimezone(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTimezone, opts...).ToFunc()
}

// ByTeamsCount orders the results by teams count.
func ByTeamsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newTeamsStep(), opts...)
	}
}

// ByTeams orders the results by teams terms.
func ByTeams(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newTeamsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByWatchedOncallRostersCount orders the results by watched_oncall_rosters count.
func ByWatchedOncallRostersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newWatchedOncallRostersStep(), opts...)
	}
}

// ByWatchedOncallRosters orders the results by watched_oncall_rosters terms.
func ByWatchedOncallRosters(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWatchedOncallRostersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOncallSchedulesCount orders the results by oncall_schedules count.
func ByOncallSchedulesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOncallSchedulesStep(), opts...)
	}
}

// ByOncallSchedules orders the results by oncall_schedules terms.
func ByOncallSchedules(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOncallSchedulesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOncallShiftsCount orders the results by oncall_shifts count.
func ByOncallShiftsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOncallShiftsStep(), opts...)
	}
}

// ByOncallShifts orders the results by oncall_shifts terms.
func ByOncallShifts(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOncallShiftsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOncallShiftCoversCount orders the results by oncall_shift_covers count.
func ByOncallShiftCoversCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOncallShiftCoversStep(), opts...)
	}
}

// ByOncallShiftCovers orders the results by oncall_shift_covers terms.
func ByOncallShiftCovers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOncallShiftCoversStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOncallEventAnnotationsCount orders the results by oncall_event_annotations count.
func ByOncallEventAnnotationsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOncallEventAnnotationsStep(), opts...)
	}
}

// ByOncallEventAnnotations orders the results by oncall_event_annotations terms.
func ByOncallEventAnnotations(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOncallEventAnnotationsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByIncidentRoleAssignmentsCount orders the results by incident_role_assignments count.
func ByIncidentRoleAssignmentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newIncidentRoleAssignmentsStep(), opts...)
	}
}

// ByIncidentRoleAssignments orders the results by incident_role_assignments terms.
func ByIncidentRoleAssignments(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newIncidentRoleAssignmentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByIncidentDebriefsCount orders the results by incident_debriefs count.
func ByIncidentDebriefsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newIncidentDebriefsStep(), opts...)
	}
}

// ByIncidentDebriefs orders the results by incident_debriefs terms.
func ByIncidentDebriefs(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newIncidentDebriefsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByAssignedTasksCount orders the results by assigned_tasks count.
func ByAssignedTasksCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAssignedTasksStep(), opts...)
	}
}

// ByAssignedTasks orders the results by assigned_tasks terms.
func ByAssignedTasks(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAssignedTasksStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByCreatedTasksCount orders the results by created_tasks count.
func ByCreatedTasksCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newCreatedTasksStep(), opts...)
	}
}

// ByCreatedTasks orders the results by created_tasks terms.
func ByCreatedTasks(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newCreatedTasksStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByRetrospectiveReviewRequestsCount orders the results by retrospective_review_requests count.
func ByRetrospectiveReviewRequestsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRetrospectiveReviewRequestsStep(), opts...)
	}
}

// ByRetrospectiveReviewRequests orders the results by retrospective_review_requests terms.
func ByRetrospectiveReviewRequests(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRetrospectiveReviewRequestsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByRetrospectiveReviewResponsesCount orders the results by retrospective_review_responses count.
func ByRetrospectiveReviewResponsesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRetrospectiveReviewResponsesStep(), opts...)
	}
}

// ByRetrospectiveReviewResponses orders the results by retrospective_review_responses terms.
func ByRetrospectiveReviewResponses(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRetrospectiveReviewResponsesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newTeamsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(TeamsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, TeamsTable, TeamsPrimaryKey...),
	)
}
func newWatchedOncallRostersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WatchedOncallRostersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, WatchedOncallRostersTable, WatchedOncallRostersPrimaryKey...),
	)
}
func newOncallSchedulesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OncallSchedulesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, OncallSchedulesTable, OncallSchedulesColumn),
	)
}
func newOncallShiftsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OncallShiftsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, OncallShiftsTable, OncallShiftsColumn),
	)
}
func newOncallShiftCoversStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OncallShiftCoversInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, OncallShiftCoversTable, OncallShiftCoversColumn),
	)
}
func newOncallEventAnnotationsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OncallEventAnnotationsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, OncallEventAnnotationsTable, OncallEventAnnotationsColumn),
	)
}
func newIncidentRoleAssignmentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentRoleAssignmentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, IncidentRoleAssignmentsTable, IncidentRoleAssignmentsColumn),
	)
}
func newIncidentDebriefsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentDebriefsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, IncidentDebriefsTable, IncidentDebriefsColumn),
	)
}
func newAssignedTasksStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AssignedTasksInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, AssignedTasksTable, AssignedTasksColumn),
	)
}
func newCreatedTasksStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(CreatedTasksInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, CreatedTasksTable, CreatedTasksColumn),
	)
}
func newRetrospectiveReviewRequestsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RetrospectiveReviewRequestsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, RetrospectiveReviewRequestsTable, RetrospectiveReviewRequestsColumn),
	)
}
func newRetrospectiveReviewResponsesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RetrospectiveReviewResponsesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, RetrospectiveReviewResponsesTable, RetrospectiveReviewResponsesColumn),
	)
}

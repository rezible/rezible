// Code generated by ent, DO NOT EDIT.

package team

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the team type in the database.
	Label = "team"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldSlug holds the string denoting the slug field in the database.
	FieldSlug = "slug"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldChatChannelID holds the string denoting the chat_channel_id field in the database.
	FieldChatChannelID = "chat_channel_id"
	// FieldTimezone holds the string denoting the timezone field in the database.
	FieldTimezone = "timezone"
	// EdgeUsers holds the string denoting the users edge name in mutations.
	EdgeUsers = "users"
	// EdgeServices holds the string denoting the services edge name in mutations.
	EdgeServices = "services"
	// EdgeOncallRosters holds the string denoting the oncall_rosters edge name in mutations.
	EdgeOncallRosters = "oncall_rosters"
	// EdgeSubscriptions holds the string denoting the subscriptions edge name in mutations.
	EdgeSubscriptions = "subscriptions"
	// EdgeIncidentAssignments holds the string denoting the incident_assignments edge name in mutations.
	EdgeIncidentAssignments = "incident_assignments"
	// EdgeScheduledMeetings holds the string denoting the scheduled_meetings edge name in mutations.
	EdgeScheduledMeetings = "scheduled_meetings"
	// Table holds the table name of the team in the database.
	Table = "teams"
	// UsersTable is the table that holds the users relation/edge. The primary key declared below.
	UsersTable = "team_users"
	// UsersInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UsersInverseTable = "users"
	// ServicesTable is the table that holds the services relation/edge.
	ServicesTable = "services"
	// ServicesInverseTable is the table name for the Service entity.
	// It exists in this package in order to avoid circular dependency with the "service" package.
	ServicesInverseTable = "services"
	// ServicesColumn is the table column denoting the services relation/edge.
	ServicesColumn = "service_owner_team"
	// OncallRostersTable is the table that holds the oncall_rosters relation/edge. The primary key declared below.
	OncallRostersTable = "team_oncall_rosters"
	// OncallRostersInverseTable is the table name for the OncallRoster entity.
	// It exists in this package in order to avoid circular dependency with the "oncallroster" package.
	OncallRostersInverseTable = "oncall_rosters"
	// SubscriptionsTable is the table that holds the subscriptions relation/edge.
	SubscriptionsTable = "subscriptions"
	// SubscriptionsInverseTable is the table name for the Subscription entity.
	// It exists in this package in order to avoid circular dependency with the "subscription" package.
	SubscriptionsInverseTable = "subscriptions"
	// SubscriptionsColumn is the table column denoting the subscriptions relation/edge.
	SubscriptionsColumn = "subscription_team"
	// IncidentAssignmentsTable is the table that holds the incident_assignments relation/edge.
	IncidentAssignmentsTable = "incident_team_assignments"
	// IncidentAssignmentsInverseTable is the table name for the IncidentTeamAssignment entity.
	// It exists in this package in order to avoid circular dependency with the "incidentteamassignment" package.
	IncidentAssignmentsInverseTable = "incident_team_assignments"
	// IncidentAssignmentsColumn is the table column denoting the incident_assignments relation/edge.
	IncidentAssignmentsColumn = "team_id"
	// ScheduledMeetingsTable is the table that holds the scheduled_meetings relation/edge. The primary key declared below.
	ScheduledMeetingsTable = "meeting_schedule_owning_team"
	// ScheduledMeetingsInverseTable is the table name for the MeetingSchedule entity.
	// It exists in this package in order to avoid circular dependency with the "meetingschedule" package.
	ScheduledMeetingsInverseTable = "meeting_schedules"
)

// Columns holds all SQL columns for team fields.
var Columns = []string{
	FieldID,
	FieldSlug,
	FieldName,
	FieldChatChannelID,
	FieldTimezone,
}

var (
	// UsersPrimaryKey and UsersColumn2 are the table columns denoting the
	// primary key for the users relation (M2M).
	UsersPrimaryKey = []string{"team_id", "user_id"}
	// OncallRostersPrimaryKey and OncallRostersColumn2 are the table columns denoting the
	// primary key for the oncall_rosters relation (M2M).
	OncallRostersPrimaryKey = []string{"team_id", "oncall_roster_id"}
	// ScheduledMeetingsPrimaryKey and ScheduledMeetingsColumn2 are the table columns denoting the
	// primary key for the scheduled_meetings relation (M2M).
	ScheduledMeetingsPrimaryKey = []string{"meeting_schedule_id", "team_id"}
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

// OrderOption defines the ordering options for the Team queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// BySlug orders the results by the slug field.
func BySlug(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSlug, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByChatChannelID orders the results by the chat_channel_id field.
func ByChatChannelID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldChatChannelID, opts...).ToFunc()
}

// ByTimezone orders the results by the timezone field.
func ByTimezone(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTimezone, opts...).ToFunc()
}

// ByUsersCount orders the results by users count.
func ByUsersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newUsersStep(), opts...)
	}
}

// ByUsers orders the results by users terms.
func ByUsers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUsersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByServicesCount orders the results by services count.
func ByServicesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newServicesStep(), opts...)
	}
}

// ByServices orders the results by services terms.
func ByServices(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newServicesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOncallRostersCount orders the results by oncall_rosters count.
func ByOncallRostersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOncallRostersStep(), opts...)
	}
}

// ByOncallRosters orders the results by oncall_rosters terms.
func ByOncallRosters(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOncallRostersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySubscriptionsCount orders the results by subscriptions count.
func BySubscriptionsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSubscriptionsStep(), opts...)
	}
}

// BySubscriptions orders the results by subscriptions terms.
func BySubscriptions(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSubscriptionsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByIncidentAssignmentsCount orders the results by incident_assignments count.
func ByIncidentAssignmentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newIncidentAssignmentsStep(), opts...)
	}
}

// ByIncidentAssignments orders the results by incident_assignments terms.
func ByIncidentAssignments(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newIncidentAssignmentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByScheduledMeetingsCount orders the results by scheduled_meetings count.
func ByScheduledMeetingsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newScheduledMeetingsStep(), opts...)
	}
}

// ByScheduledMeetings orders the results by scheduled_meetings terms.
func ByScheduledMeetings(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newScheduledMeetingsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newUsersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UsersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, UsersTable, UsersPrimaryKey...),
	)
}
func newServicesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ServicesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, ServicesTable, ServicesColumn),
	)
}
func newOncallRostersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OncallRostersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, OncallRostersTable, OncallRostersPrimaryKey...),
	)
}
func newSubscriptionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SubscriptionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, SubscriptionsTable, SubscriptionsColumn),
	)
}
func newIncidentAssignmentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentAssignmentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, IncidentAssignmentsTable, IncidentAssignmentsColumn),
	)
}
func newScheduledMeetingsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ScheduledMeetingsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, ScheduledMeetingsTable, ScheduledMeetingsPrimaryKey...),
	)
}
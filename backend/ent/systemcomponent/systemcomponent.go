// Code generated by ent, DO NOT EDIT.

package systemcomponent

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the systemcomponent type in the database.
	Label = "system_component"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldProviderID holds the string denoting the provider_id field in the database.
	FieldProviderID = "provider_id"
	// FieldKindID holds the string denoting the kind_id field in the database.
	FieldKindID = "kind_id"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldProperties holds the string denoting the properties field in the database.
	FieldProperties = "properties"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeKind holds the string denoting the kind edge name in mutations.
	EdgeKind = "kind"
	// EdgeRelated holds the string denoting the related edge name in mutations.
	EdgeRelated = "related"
	// EdgeSystemAnalyses holds the string denoting the system_analyses edge name in mutations.
	EdgeSystemAnalyses = "system_analyses"
	// EdgeEvents holds the string denoting the events edge name in mutations.
	EdgeEvents = "events"
	// EdgeConstraints holds the string denoting the constraints edge name in mutations.
	EdgeConstraints = "constraints"
	// EdgeControls holds the string denoting the controls edge name in mutations.
	EdgeControls = "controls"
	// EdgeSignals holds the string denoting the signals edge name in mutations.
	EdgeSignals = "signals"
	// EdgeComponentRelationships holds the string denoting the component_relationships edge name in mutations.
	EdgeComponentRelationships = "component_relationships"
	// EdgeSystemAnalysisComponents holds the string denoting the system_analysis_components edge name in mutations.
	EdgeSystemAnalysisComponents = "system_analysis_components"
	// EdgeEventComponents holds the string denoting the event_components edge name in mutations.
	EdgeEventComponents = "event_components"
	// Table holds the table name of the systemcomponent in the database.
	Table = "system_components"
	// KindTable is the table that holds the kind relation/edge.
	KindTable = "system_components"
	// KindInverseTable is the table name for the SystemComponentKind entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentkind" package.
	KindInverseTable = "system_component_kinds"
	// KindColumn is the table column denoting the kind relation/edge.
	KindColumn = "kind_id"
	// RelatedTable is the table that holds the related relation/edge. The primary key declared below.
	RelatedTable = "system_component_relationships"
	// SystemAnalysesTable is the table that holds the system_analyses relation/edge. The primary key declared below.
	SystemAnalysesTable = "system_analysis_components"
	// SystemAnalysesInverseTable is the table name for the SystemAnalysis entity.
	// It exists in this package in order to avoid circular dependency with the "systemanalysis" package.
	SystemAnalysesInverseTable = "system_analyses"
	// EventsTable is the table that holds the events relation/edge. The primary key declared below.
	EventsTable = "incident_event_system_components"
	// EventsInverseTable is the table name for the IncidentEvent entity.
	// It exists in this package in order to avoid circular dependency with the "incidentevent" package.
	EventsInverseTable = "incident_events"
	// ConstraintsTable is the table that holds the constraints relation/edge.
	ConstraintsTable = "system_component_constraints"
	// ConstraintsInverseTable is the table name for the SystemComponentConstraint entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentconstraint" package.
	ConstraintsInverseTable = "system_component_constraints"
	// ConstraintsColumn is the table column denoting the constraints relation/edge.
	ConstraintsColumn = "component_id"
	// ControlsTable is the table that holds the controls relation/edge.
	ControlsTable = "system_component_controls"
	// ControlsInverseTable is the table name for the SystemComponentControl entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentcontrol" package.
	ControlsInverseTable = "system_component_controls"
	// ControlsColumn is the table column denoting the controls relation/edge.
	ControlsColumn = "component_id"
	// SignalsTable is the table that holds the signals relation/edge.
	SignalsTable = "system_component_signals"
	// SignalsInverseTable is the table name for the SystemComponentSignal entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentsignal" package.
	SignalsInverseTable = "system_component_signals"
	// SignalsColumn is the table column denoting the signals relation/edge.
	SignalsColumn = "component_id"
	// ComponentRelationshipsTable is the table that holds the component_relationships relation/edge.
	ComponentRelationshipsTable = "system_component_relationships"
	// ComponentRelationshipsInverseTable is the table name for the SystemComponentRelationship entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentrelationship" package.
	ComponentRelationshipsInverseTable = "system_component_relationships"
	// ComponentRelationshipsColumn is the table column denoting the component_relationships relation/edge.
	ComponentRelationshipsColumn = "source_id"
	// SystemAnalysisComponentsTable is the table that holds the system_analysis_components relation/edge.
	SystemAnalysisComponentsTable = "system_analysis_components"
	// SystemAnalysisComponentsInverseTable is the table name for the SystemAnalysisComponent entity.
	// It exists in this package in order to avoid circular dependency with the "systemanalysiscomponent" package.
	SystemAnalysisComponentsInverseTable = "system_analysis_components"
	// SystemAnalysisComponentsColumn is the table column denoting the system_analysis_components relation/edge.
	SystemAnalysisComponentsColumn = "component_id"
	// EventComponentsTable is the table that holds the event_components relation/edge.
	EventComponentsTable = "incident_event_system_components"
	// EventComponentsInverseTable is the table name for the IncidentEventSystemComponent entity.
	// It exists in this package in order to avoid circular dependency with the "incidenteventsystemcomponent" package.
	EventComponentsInverseTable = "incident_event_system_components"
	// EventComponentsColumn is the table column denoting the event_components relation/edge.
	EventComponentsColumn = "system_component_id"
)

// Columns holds all SQL columns for systemcomponent fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldProviderID,
	FieldKindID,
	FieldDescription,
	FieldProperties,
	FieldCreatedAt,
	FieldUpdatedAt,
}

var (
	// RelatedPrimaryKey and RelatedColumn2 are the table columns denoting the
	// primary key for the related relation (M2M).
	RelatedPrimaryKey = []string{"source_id", "target_id"}
	// SystemAnalysesPrimaryKey and SystemAnalysesColumn2 are the table columns denoting the
	// primary key for the system_analyses relation (M2M).
	SystemAnalysesPrimaryKey = []string{"component_id", "analysis_id"}
	// EventsPrimaryKey and EventsColumn2 are the table columns denoting the
	// primary key for the events relation (M2M).
	EventsPrimaryKey = []string{"incident_event_id", "system_component_id"}
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
	// NameValidator is a validator for the "name" field. It is called by the builders before save.
	NameValidator func(string) error
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the SystemComponent queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByProviderID orders the results by the provider_id field.
func ByProviderID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProviderID, opts...).ToFunc()
}

// ByKindID orders the results by the kind_id field.
func ByKindID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldKindID, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByKindField orders the results by kind field.
func ByKindField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newKindStep(), sql.OrderByField(field, opts...))
	}
}

// ByRelatedCount orders the results by related count.
func ByRelatedCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRelatedStep(), opts...)
	}
}

// ByRelated orders the results by related terms.
func ByRelated(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRelatedStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySystemAnalysesCount orders the results by system_analyses count.
func BySystemAnalysesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSystemAnalysesStep(), opts...)
	}
}

// BySystemAnalyses orders the results by system_analyses terms.
func BySystemAnalyses(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSystemAnalysesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByEventsCount orders the results by events count.
func ByEventsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEventsStep(), opts...)
	}
}

// ByEvents orders the results by events terms.
func ByEvents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEventsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByConstraintsCount orders the results by constraints count.
func ByConstraintsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newConstraintsStep(), opts...)
	}
}

// ByConstraints orders the results by constraints terms.
func ByConstraints(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newConstraintsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByControlsCount orders the results by controls count.
func ByControlsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newControlsStep(), opts...)
	}
}

// ByControls orders the results by controls terms.
func ByControls(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySignalsCount orders the results by signals count.
func BySignalsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSignalsStep(), opts...)
	}
}

// BySignals orders the results by signals terms.
func BySignals(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSignalsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByComponentRelationshipsCount orders the results by component_relationships count.
func ByComponentRelationshipsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newComponentRelationshipsStep(), opts...)
	}
}

// ByComponentRelationships orders the results by component_relationships terms.
func ByComponentRelationships(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newComponentRelationshipsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySystemAnalysisComponentsCount orders the results by system_analysis_components count.
func BySystemAnalysisComponentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSystemAnalysisComponentsStep(), opts...)
	}
}

// BySystemAnalysisComponents orders the results by system_analysis_components terms.
func BySystemAnalysisComponents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSystemAnalysisComponentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByEventComponentsCount orders the results by event_components count.
func ByEventComponentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEventComponentsStep(), opts...)
	}
}

// ByEventComponents orders the results by event_components terms.
func ByEventComponents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEventComponentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newKindStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(KindInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, KindTable, KindColumn),
	)
}
func newRelatedStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(Table, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, RelatedTable, RelatedPrimaryKey...),
	)
}
func newSystemAnalysesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SystemAnalysesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, SystemAnalysesTable, SystemAnalysesPrimaryKey...),
	)
}
func newEventsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EventsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, EventsTable, EventsPrimaryKey...),
	)
}
func newConstraintsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ConstraintsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, ConstraintsTable, ConstraintsColumn),
	)
}
func newControlsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, ControlsTable, ControlsColumn),
	)
}
func newSignalsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SignalsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, SignalsTable, SignalsColumn),
	)
}
func newComponentRelationshipsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ComponentRelationshipsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, ComponentRelationshipsTable, ComponentRelationshipsColumn),
	)
}
func newSystemAnalysisComponentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SystemAnalysisComponentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, SystemAnalysisComponentsTable, SystemAnalysisComponentsColumn),
	)
}
func newEventComponentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EventComponentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, EventComponentsTable, EventComponentsColumn),
	)
}

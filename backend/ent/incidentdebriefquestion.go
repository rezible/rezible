// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incidentdebriefquestion"
)

// IncidentDebriefQuestion is the model entity for the IncidentDebriefQuestion schema.
type IncidentDebriefQuestion struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Content holds the value of the "content" field.
	Content string `json:"content,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentDebriefQuestionQuery when eager-loading is set.
	Edges        IncidentDebriefQuestionEdges `json:"edges"`
	selectValues sql.SelectValues
}

// IncidentDebriefQuestionEdges holds the relations/edges for other nodes in the graph.
type IncidentDebriefQuestionEdges struct {
	// Messages holds the value of the messages edge.
	Messages []*IncidentDebriefMessage `json:"messages,omitempty"`
	// IncidentFields holds the value of the incident_fields edge.
	IncidentFields []*IncidentField `json:"incident_fields,omitempty"`
	// IncidentRoles holds the value of the incident_roles edge.
	IncidentRoles []*IncidentRole `json:"incident_roles,omitempty"`
	// IncidentSeverities holds the value of the incident_severities edge.
	IncidentSeverities []*IncidentSeverity `json:"incident_severities,omitempty"`
	// IncidentTags holds the value of the incident_tags edge.
	IncidentTags []*IncidentTag `json:"incident_tags,omitempty"`
	// IncidentTypes holds the value of the incident_types edge.
	IncidentTypes []*IncidentType `json:"incident_types,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [6]bool
}

// MessagesOrErr returns the Messages value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentDebriefQuestionEdges) MessagesOrErr() ([]*IncidentDebriefMessage, error) {
	if e.loadedTypes[0] {
		return e.Messages, nil
	}
	return nil, &NotLoadedError{edge: "messages"}
}

// IncidentFieldsOrErr returns the IncidentFields value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentDebriefQuestionEdges) IncidentFieldsOrErr() ([]*IncidentField, error) {
	if e.loadedTypes[1] {
		return e.IncidentFields, nil
	}
	return nil, &NotLoadedError{edge: "incident_fields"}
}

// IncidentRolesOrErr returns the IncidentRoles value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentDebriefQuestionEdges) IncidentRolesOrErr() ([]*IncidentRole, error) {
	if e.loadedTypes[2] {
		return e.IncidentRoles, nil
	}
	return nil, &NotLoadedError{edge: "incident_roles"}
}

// IncidentSeveritiesOrErr returns the IncidentSeverities value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentDebriefQuestionEdges) IncidentSeveritiesOrErr() ([]*IncidentSeverity, error) {
	if e.loadedTypes[3] {
		return e.IncidentSeverities, nil
	}
	return nil, &NotLoadedError{edge: "incident_severities"}
}

// IncidentTagsOrErr returns the IncidentTags value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentDebriefQuestionEdges) IncidentTagsOrErr() ([]*IncidentTag, error) {
	if e.loadedTypes[4] {
		return e.IncidentTags, nil
	}
	return nil, &NotLoadedError{edge: "incident_tags"}
}

// IncidentTypesOrErr returns the IncidentTypes value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentDebriefQuestionEdges) IncidentTypesOrErr() ([]*IncidentType, error) {
	if e.loadedTypes[5] {
		return e.IncidentTypes, nil
	}
	return nil, &NotLoadedError{edge: "incident_types"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentDebriefQuestion) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidentdebriefquestion.FieldContent:
			values[i] = new(sql.NullString)
		case incidentdebriefquestion.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentDebriefQuestion fields.
func (idq *IncidentDebriefQuestion) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidentdebriefquestion.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				idq.ID = *value
			}
		case incidentdebriefquestion.FieldContent:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field content", values[i])
			} else if value.Valid {
				idq.Content = value.String
			}
		default:
			idq.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the IncidentDebriefQuestion.
// This includes values selected through modifiers, order, etc.
func (idq *IncidentDebriefQuestion) Value(name string) (ent.Value, error) {
	return idq.selectValues.Get(name)
}

// QueryMessages queries the "messages" edge of the IncidentDebriefQuestion entity.
func (idq *IncidentDebriefQuestion) QueryMessages() *IncidentDebriefMessageQuery {
	return NewIncidentDebriefQuestionClient(idq.config).QueryMessages(idq)
}

// QueryIncidentFields queries the "incident_fields" edge of the IncidentDebriefQuestion entity.
func (idq *IncidentDebriefQuestion) QueryIncidentFields() *IncidentFieldQuery {
	return NewIncidentDebriefQuestionClient(idq.config).QueryIncidentFields(idq)
}

// QueryIncidentRoles queries the "incident_roles" edge of the IncidentDebriefQuestion entity.
func (idq *IncidentDebriefQuestion) QueryIncidentRoles() *IncidentRoleQuery {
	return NewIncidentDebriefQuestionClient(idq.config).QueryIncidentRoles(idq)
}

// QueryIncidentSeverities queries the "incident_severities" edge of the IncidentDebriefQuestion entity.
func (idq *IncidentDebriefQuestion) QueryIncidentSeverities() *IncidentSeverityQuery {
	return NewIncidentDebriefQuestionClient(idq.config).QueryIncidentSeverities(idq)
}

// QueryIncidentTags queries the "incident_tags" edge of the IncidentDebriefQuestion entity.
func (idq *IncidentDebriefQuestion) QueryIncidentTags() *IncidentTagQuery {
	return NewIncidentDebriefQuestionClient(idq.config).QueryIncidentTags(idq)
}

// QueryIncidentTypes queries the "incident_types" edge of the IncidentDebriefQuestion entity.
func (idq *IncidentDebriefQuestion) QueryIncidentTypes() *IncidentTypeQuery {
	return NewIncidentDebriefQuestionClient(idq.config).QueryIncidentTypes(idq)
}

// Update returns a builder for updating this IncidentDebriefQuestion.
// Note that you need to call IncidentDebriefQuestion.Unwrap() before calling this method if this IncidentDebriefQuestion
// was returned from a transaction, and the transaction was committed or rolled back.
func (idq *IncidentDebriefQuestion) Update() *IncidentDebriefQuestionUpdateOne {
	return NewIncidentDebriefQuestionClient(idq.config).UpdateOne(idq)
}

// Unwrap unwraps the IncidentDebriefQuestion entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (idq *IncidentDebriefQuestion) Unwrap() *IncidentDebriefQuestion {
	_tx, ok := idq.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentDebriefQuestion is not a transactional entity")
	}
	idq.config.driver = _tx.drv
	return idq
}

// String implements the fmt.Stringer.
func (idq *IncidentDebriefQuestion) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentDebriefQuestion(")
	builder.WriteString(fmt.Sprintf("id=%v, ", idq.ID))
	builder.WriteString("content=")
	builder.WriteString(idq.Content)
	builder.WriteByte(')')
	return builder.String()
}

// IncidentDebriefQuestions is a parsable slice of IncidentDebriefQuestion.
type IncidentDebriefQuestions []*IncidentDebriefQuestion
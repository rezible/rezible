// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rezible/rezible/ent/oncallusershifthandover"
	"github.com/rezible/rezible/ent/user"
)

// OncallUserShift is the model entity for the OncallUserShift schema.
type OncallUserShift struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID uuid.UUID `json:"user_id,omitempty"`
	// RosterID holds the value of the "roster_id" field.
	RosterID uuid.UUID `json:"roster_id,omitempty"`
	// StartAt holds the value of the "start_at" field.
	StartAt time.Time `json:"start_at,omitempty"`
	// EndAt holds the value of the "end_at" field.
	EndAt time.Time `json:"end_at,omitempty"`
	// ProviderID holds the value of the "provider_id" field.
	ProviderID string `json:"provider_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the OncallUserShiftQuery when eager-loading is set.
	Edges        OncallUserShiftEdges `json:"edges"`
	selectValues sql.SelectValues
}

// OncallUserShiftEdges holds the relations/edges for other nodes in the graph.
type OncallUserShiftEdges struct {
	// User holds the value of the user edge.
	User *User `json:"user,omitempty"`
	// Roster holds the value of the roster edge.
	Roster *OncallRoster `json:"roster,omitempty"`
	// Covers holds the value of the covers edge.
	Covers []*OncallUserShiftCover `json:"covers,omitempty"`
	// Handover holds the value of the handover edge.
	Handover *OncallUserShiftHandover `json:"handover,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [4]bool
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallUserShiftEdges) UserOrErr() (*User, error) {
	if e.User != nil {
		return e.User, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "user"}
}

// RosterOrErr returns the Roster value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallUserShiftEdges) RosterOrErr() (*OncallRoster, error) {
	if e.Roster != nil {
		return e.Roster, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: oncallroster.Label}
	}
	return nil, &NotLoadedError{edge: "roster"}
}

// CoversOrErr returns the Covers value or an error if the edge
// was not loaded in eager-loading.
func (e OncallUserShiftEdges) CoversOrErr() ([]*OncallUserShiftCover, error) {
	if e.loadedTypes[2] {
		return e.Covers, nil
	}
	return nil, &NotLoadedError{edge: "covers"}
}

// HandoverOrErr returns the Handover value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallUserShiftEdges) HandoverOrErr() (*OncallUserShiftHandover, error) {
	if e.Handover != nil {
		return e.Handover, nil
	} else if e.loadedTypes[3] {
		return nil, &NotFoundError{label: oncallusershifthandover.Label}
	}
	return nil, &NotLoadedError{edge: "handover"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*OncallUserShift) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case oncallusershift.FieldProviderID:
			values[i] = new(sql.NullString)
		case oncallusershift.FieldStartAt, oncallusershift.FieldEndAt:
			values[i] = new(sql.NullTime)
		case oncallusershift.FieldID, oncallusershift.FieldUserID, oncallusershift.FieldRosterID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the OncallUserShift fields.
func (ous *OncallUserShift) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case oncallusershift.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ous.ID = *value
			}
		case oncallusershift.FieldUserID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value != nil {
				ous.UserID = *value
			}
		case oncallusershift.FieldRosterID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field roster_id", values[i])
			} else if value != nil {
				ous.RosterID = *value
			}
		case oncallusershift.FieldStartAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field start_at", values[i])
			} else if value.Valid {
				ous.StartAt = value.Time
			}
		case oncallusershift.FieldEndAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field end_at", values[i])
			} else if value.Valid {
				ous.EndAt = value.Time
			}
		case oncallusershift.FieldProviderID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field provider_id", values[i])
			} else if value.Valid {
				ous.ProviderID = value.String
			}
		default:
			ous.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the OncallUserShift.
// This includes values selected through modifiers, order, etc.
func (ous *OncallUserShift) Value(name string) (ent.Value, error) {
	return ous.selectValues.Get(name)
}

// QueryUser queries the "user" edge of the OncallUserShift entity.
func (ous *OncallUserShift) QueryUser() *UserQuery {
	return NewOncallUserShiftClient(ous.config).QueryUser(ous)
}

// QueryRoster queries the "roster" edge of the OncallUserShift entity.
func (ous *OncallUserShift) QueryRoster() *OncallRosterQuery {
	return NewOncallUserShiftClient(ous.config).QueryRoster(ous)
}

// QueryCovers queries the "covers" edge of the OncallUserShift entity.
func (ous *OncallUserShift) QueryCovers() *OncallUserShiftCoverQuery {
	return NewOncallUserShiftClient(ous.config).QueryCovers(ous)
}

// QueryHandover queries the "handover" edge of the OncallUserShift entity.
func (ous *OncallUserShift) QueryHandover() *OncallUserShiftHandoverQuery {
	return NewOncallUserShiftClient(ous.config).QueryHandover(ous)
}

// Update returns a builder for updating this OncallUserShift.
// Note that you need to call OncallUserShift.Unwrap() before calling this method if this OncallUserShift
// was returned from a transaction, and the transaction was committed or rolled back.
func (ous *OncallUserShift) Update() *OncallUserShiftUpdateOne {
	return NewOncallUserShiftClient(ous.config).UpdateOne(ous)
}

// Unwrap unwraps the OncallUserShift entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ous *OncallUserShift) Unwrap() *OncallUserShift {
	_tx, ok := ous.config.driver.(*txDriver)
	if !ok {
		panic("ent: OncallUserShift is not a transactional entity")
	}
	ous.config.driver = _tx.drv
	return ous
}

// String implements the fmt.Stringer.
func (ous *OncallUserShift) String() string {
	var builder strings.Builder
	builder.WriteString("OncallUserShift(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ous.ID))
	builder.WriteString("user_id=")
	builder.WriteString(fmt.Sprintf("%v", ous.UserID))
	builder.WriteString(", ")
	builder.WriteString("roster_id=")
	builder.WriteString(fmt.Sprintf("%v", ous.RosterID))
	builder.WriteString(", ")
	builder.WriteString("start_at=")
	builder.WriteString(ous.StartAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("end_at=")
	builder.WriteString(ous.EndAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("provider_id=")
	builder.WriteString(ous.ProviderID)
	builder.WriteByte(')')
	return builder.String()
}

// OncallUserShifts is a parsable slice of OncallUserShift.
type OncallUserShifts []*OncallUserShift

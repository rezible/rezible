// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivediscussion"
	"github.com/rezible/rezible/ent/retrospectivereview"
	"github.com/rezible/rezible/ent/user"
)

// RetrospectiveReview is the model entity for the RetrospectiveReview schema.
type RetrospectiveReview struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// RetrospectiveID holds the value of the "retrospective_id" field.
	RetrospectiveID uuid.UUID `json:"retrospective_id,omitempty"`
	// RequesterID holds the value of the "requester_id" field.
	RequesterID uuid.UUID `json:"requester_id,omitempty"`
	// ReviewerID holds the value of the "reviewer_id" field.
	ReviewerID uuid.UUID `json:"reviewer_id,omitempty"`
	// State holds the value of the "state" field.
	State retrospectivereview.State `json:"state,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RetrospectiveReviewQuery when eager-loading is set.
	Edges                           RetrospectiveReviewEdges `json:"edges"`
	retrospective_review_discussion *uuid.UUID
	selectValues                    sql.SelectValues
}

// RetrospectiveReviewEdges holds the relations/edges for other nodes in the graph.
type RetrospectiveReviewEdges struct {
	// Retrospective holds the value of the retrospective edge.
	Retrospective *Retrospective `json:"retrospective,omitempty"`
	// Requester holds the value of the requester edge.
	Requester *User `json:"requester,omitempty"`
	// Reviewer holds the value of the reviewer edge.
	Reviewer *User `json:"reviewer,omitempty"`
	// Discussion holds the value of the discussion edge.
	Discussion *RetrospectiveDiscussion `json:"discussion,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [4]bool
}

// RetrospectiveOrErr returns the Retrospective value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RetrospectiveReviewEdges) RetrospectiveOrErr() (*Retrospective, error) {
	if e.Retrospective != nil {
		return e.Retrospective, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: retrospective.Label}
	}
	return nil, &NotLoadedError{edge: "retrospective"}
}

// RequesterOrErr returns the Requester value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RetrospectiveReviewEdges) RequesterOrErr() (*User, error) {
	if e.Requester != nil {
		return e.Requester, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "requester"}
}

// ReviewerOrErr returns the Reviewer value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RetrospectiveReviewEdges) ReviewerOrErr() (*User, error) {
	if e.Reviewer != nil {
		return e.Reviewer, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "reviewer"}
}

// DiscussionOrErr returns the Discussion value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RetrospectiveReviewEdges) DiscussionOrErr() (*RetrospectiveDiscussion, error) {
	if e.Discussion != nil {
		return e.Discussion, nil
	} else if e.loadedTypes[3] {
		return nil, &NotFoundError{label: retrospectivediscussion.Label}
	}
	return nil, &NotLoadedError{edge: "discussion"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*RetrospectiveReview) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case retrospectivereview.FieldState:
			values[i] = new(sql.NullString)
		case retrospectivereview.FieldID, retrospectivereview.FieldRetrospectiveID, retrospectivereview.FieldRequesterID, retrospectivereview.FieldReviewerID:
			values[i] = new(uuid.UUID)
		case retrospectivereview.ForeignKeys[0]: // retrospective_review_discussion
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the RetrospectiveReview fields.
func (rr *RetrospectiveReview) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case retrospectivereview.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				rr.ID = *value
			}
		case retrospectivereview.FieldRetrospectiveID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field retrospective_id", values[i])
			} else if value != nil {
				rr.RetrospectiveID = *value
			}
		case retrospectivereview.FieldRequesterID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field requester_id", values[i])
			} else if value != nil {
				rr.RequesterID = *value
			}
		case retrospectivereview.FieldReviewerID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field reviewer_id", values[i])
			} else if value != nil {
				rr.ReviewerID = *value
			}
		case retrospectivereview.FieldState:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field state", values[i])
			} else if value.Valid {
				rr.State = retrospectivereview.State(value.String)
			}
		case retrospectivereview.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field retrospective_review_discussion", values[i])
			} else if value.Valid {
				rr.retrospective_review_discussion = new(uuid.UUID)
				*rr.retrospective_review_discussion = *value.S.(*uuid.UUID)
			}
		default:
			rr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the RetrospectiveReview.
// This includes values selected through modifiers, order, etc.
func (rr *RetrospectiveReview) Value(name string) (ent.Value, error) {
	return rr.selectValues.Get(name)
}

// QueryRetrospective queries the "retrospective" edge of the RetrospectiveReview entity.
func (rr *RetrospectiveReview) QueryRetrospective() *RetrospectiveQuery {
	return NewRetrospectiveReviewClient(rr.config).QueryRetrospective(rr)
}

// QueryRequester queries the "requester" edge of the RetrospectiveReview entity.
func (rr *RetrospectiveReview) QueryRequester() *UserQuery {
	return NewRetrospectiveReviewClient(rr.config).QueryRequester(rr)
}

// QueryReviewer queries the "reviewer" edge of the RetrospectiveReview entity.
func (rr *RetrospectiveReview) QueryReviewer() *UserQuery {
	return NewRetrospectiveReviewClient(rr.config).QueryReviewer(rr)
}

// QueryDiscussion queries the "discussion" edge of the RetrospectiveReview entity.
func (rr *RetrospectiveReview) QueryDiscussion() *RetrospectiveDiscussionQuery {
	return NewRetrospectiveReviewClient(rr.config).QueryDiscussion(rr)
}

// Update returns a builder for updating this RetrospectiveReview.
// Note that you need to call RetrospectiveReview.Unwrap() before calling this method if this RetrospectiveReview
// was returned from a transaction, and the transaction was committed or rolled back.
func (rr *RetrospectiveReview) Update() *RetrospectiveReviewUpdateOne {
	return NewRetrospectiveReviewClient(rr.config).UpdateOne(rr)
}

// Unwrap unwraps the RetrospectiveReview entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (rr *RetrospectiveReview) Unwrap() *RetrospectiveReview {
	_tx, ok := rr.config.driver.(*txDriver)
	if !ok {
		panic("ent: RetrospectiveReview is not a transactional entity")
	}
	rr.config.driver = _tx.drv
	return rr
}

// String implements the fmt.Stringer.
func (rr *RetrospectiveReview) String() string {
	var builder strings.Builder
	builder.WriteString("RetrospectiveReview(")
	builder.WriteString(fmt.Sprintf("id=%v, ", rr.ID))
	builder.WriteString("retrospective_id=")
	builder.WriteString(fmt.Sprintf("%v", rr.RetrospectiveID))
	builder.WriteString(", ")
	builder.WriteString("requester_id=")
	builder.WriteString(fmt.Sprintf("%v", rr.RequesterID))
	builder.WriteString(", ")
	builder.WriteString("reviewer_id=")
	builder.WriteString(fmt.Sprintf("%v", rr.ReviewerID))
	builder.WriteString(", ")
	builder.WriteString("state=")
	builder.WriteString(fmt.Sprintf("%v", rr.State))
	builder.WriteByte(')')
	return builder.String()
}

// RetrospectiveReviews is a parsable slice of RetrospectiveReview.
type RetrospectiveReviews []*RetrospectiveReview

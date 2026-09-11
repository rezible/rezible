package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type ActivityHandler interface {
	ListActivity(context.Context, *ListActivityRequest) (*ListActivityResponse, error)
}

func (o operations) RegisterActivity(api huma.API) {
	huma.Register(api, ListActivity, o.ListActivity)
}

type (
	ActivityRecord struct {
		Id         uuid.UUID                `json:"id"`
		Attributes ActivityRecordAttributes `json:"attributes"`
	}

	ActivityRecordAttributes struct {
		OccurredAt  time.Time  `json:"occurredAt"`
		Explanation string     `json:"explanation"`
		RecordKind  string     `json:"recordKind" enum:"incident-update,situation-investigation,inbox-item"`
		RecordId    uuid.UUID  `json:"recordId"`
		Scope       string     `json:"scope"`
		IncidentId  *uuid.UUID `json:"incidentId,omitempty"`
	}
)

var ListActivity = huma.Operation{
	OperationID: "list-activity",
	Method:      http.MethodGet,
	Path:        "/activity",
	Summary:     "List Activity",
	Errors:      ErrorCodes(),
}

type ListActivityRequest struct {
	PaginationRequest
	From  time.Time `query:"from,omitempty"`
	To    time.Time `query:"to,omitempty"`
	Scope string    `query:"scope,omitempty" enum:"team"`
}
type ListActivityResponse PaginatedResponse[ActivityRecord]

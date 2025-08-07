package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type AlertsHandler interface {
	ListAlerts(context.Context, *ListAlertsRequest) (*ListAlertsResponse, error)
	GetAlert(context.Context, *GetAlertRequest) (*GetAlertResponse, error)
}

func (o operations) RegisterAlerts(api huma.API) {
	huma.Register(api, ListAlerts, o.ListAlerts)
	huma.Register(api, GetAlert, o.GetAlert)
}

type (
	Alert struct {
		Id         uuid.UUID       `json:"id"`
		Attributes AlertAttributes `json:"attributes"`
	}

	AlertAttributes struct {
		Title           string     `json:"title"`
		Description     string     `json:"description"`
		LinkedPlaybooks []Playbook `json:"linkedPlaybooks"`
	}
)

func AlertFromEnt(a *ent.Alert) Alert {
	attrs := AlertAttributes{
		Title:       a.Title,
		Description: "",
	}

	attrs.LinkedPlaybooks = make([]Playbook, len(a.Edges.Playbooks))
	for i, playbook := range a.Edges.Playbooks {
		attrs.LinkedPlaybooks[i] = PlaybookFromEnt(playbook)
	}

	return Alert{
		Id:         a.ID,
		Attributes: attrs,
	}
}

var alertsTags = []string{"Alerts"}

// ops

var ListAlerts = huma.Operation{
	OperationID: "list-alerts",
	Method:      http.MethodGet,
	Path:        "/alerts",
	Summary:     "List Alerts",
	Tags:        alertsTags,
	Errors:      errorCodes(),
}

type ListAlertsRequest struct {
	ListRequest
	RosterId uuid.UUID `query:"rosterId" required:"false"`
}
type ListAlertsResponse PaginatedResponse[Alert]

var GetAlert = huma.Operation{
	OperationID: "get-alert",
	Method:      http.MethodGet,
	Path:        "/alerts/{id}",
	Summary:     "Get Alert",
	Tags:        alertsTags,
	Errors:      errorCodes(),
}

type GetAlertRequest struct {
	GetIdRequest
	IncludeAnnotations bool `query:"includeAnnotations" default:"true"`
}
type GetAlertResponse ItemResponse[Alert]

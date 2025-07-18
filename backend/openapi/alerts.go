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
		Name        string `json:"name"`
		Description string `json:"description"`
	}
)

func AlertFromEnt(alerts *ent.Alert) Alert {
	return Alert{
		Id:         alerts.ID,
		Attributes: AlertAttributes{},
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
	TeamId uuid.UUID `query:"teamId" required:"false"`
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

type GetAlertRequest GetIdRequest
type GetAlertResponse ItemResponse[Alert]

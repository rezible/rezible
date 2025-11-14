package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type AlertsHandler interface {
	ListAlerts(context.Context, *ListAlertsRequest) (*ListAlertsResponse, error)
	GetAlert(context.Context, *GetAlertRequest) (*GetAlertResponse, error)
	GetAlertMetrics(context.Context, *GetAlertMetricsRequest) (*GetAlertMetricsResponse, error)
	ListAlertIncidentLinks(context.Context, *ListAlertIncidentLinksRequest) (*ListAlertIncidentLinksResponse, error)
}

func (o operations) RegisterAlerts(api huma.API) {
	huma.Register(api, ListAlerts, o.ListAlerts)
	huma.Register(api, GetAlert, o.GetAlert)
	huma.Register(api, GetAlertMetrics, o.GetAlertMetrics)
	huma.Register(api, ListAlertIncidentLinks, o.ListAlertIncidentLinks)
}

type (
	Alert struct {
		Id         uuid.UUID       `json:"id"`
		Attributes AlertAttributes `json:"attributes"`
	}

	AlertAttributes struct {
		Title       string                              `json:"title"`
		Description string                              `json:"description"`
		Definition  string                              `json:"definition"`
		Roster      *Expandable[OncallRosterAttributes] `json:"roster,omitempty"`
	}

	AlertInstance struct {
		Id         uuid.UUID               `json:"id"`
		Attributes AlertInstanceAttributes `json:"attributes"`
	}

	AlertInstanceAttributes struct {
		Timestamp time.Time              `json:"timestamp"`
		Feedback  *AlertInstanceFeedback `json:"feedback,omitempty"`
	}

	AlertInstanceFeedback struct {
		UserId                   uuid.UUID `json:"userId"`
		Actionable               bool      `json:"actionable"`
		Accurate                 string    `json:"accurate" enum:"yes,no,unknown"`
		DocumentationAvailable   bool      `json:"documentationAvailable"`
		DocumentationNeedsUpdate bool      `json:"documentationNeedsUpdate"`
	}

	AlertMetrics struct {
		Triggers                         int `json:"triggers"`
		Interrupts                       int `json:"interrupts"`
		NightInterrupts                  int `json:"nightInterrupts"`
		IncidentLinks                    int `json:"incidentLinks"`
		Feedbacks                        int `json:"feedbacks"`
		FeedbackActionable               int `json:"actionable"`
		FeedbackAccurate                 int `json:"accurate"`
		FeedbackAccurateUnknown          int `json:"accurateUnknown"`
		FeedbackDocumentationAvailable   int `json:"docsAvailable"`
		FeedbackDocumentationNeedsUpdate int `json:"docsNeedsUpdate"`
	}

	AlertIncidentLink struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes AlertIncidentLinkAttributes `json:"attributes"`
	}

	AlertIncidentLinkAttributes struct {
		AlertId     uuid.UUID `json:"alertId"`
		IncidentId  uuid.UUID `json:"incidentId"`
		Description string    `json:"description"`
	}
)

func AlertFromEnt(a *ent.Alert) Alert {
	attrs := AlertAttributes{
		Title:       a.Title,
		Description: a.Description,
		Definition:  a.Definition,
	}

	if a.Edges.Roster != nil {
		r := OncallRosterFromEnt(a.Edges.Roster)
		attrs.Roster = &Expandable[OncallRosterAttributes]{
			Id:         a.RosterID,
			Attributes: &r.Attributes,
		}
	}

	return Alert{
		Id:         a.ID,
		Attributes: attrs,
	}
}

func AlertMetricsFromEnt(m *ent.AlertMetrics) AlertMetrics {
	return AlertMetrics{
		Triggers:                         m.EventCount,
		Interrupts:                       m.InterruptCount,
		NightInterrupts:                  m.NightInterruptCount,
		IncidentLinks:                    m.Incidents,
		Feedbacks:                        m.FeedbackCount,
		FeedbackActionable:               m.FeedbackActionable,
		FeedbackAccurate:                 m.FeedbackAccurate,
		FeedbackAccurateUnknown:          m.FeedbackAccurateUnknown,
		FeedbackDocumentationAvailable:   m.FeedbackDocsAvailable,
		FeedbackDocumentationNeedsUpdate: m.FeedbackDocsNeedUpdate,
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
}
type GetAlertResponse ItemResponse[Alert]

var GetAlertMetrics = huma.Operation{
	OperationID: "get-alert-metrics",
	Method:      http.MethodGet,
	Path:        "/alerts/{id}/metrics",
	Summary:     "Get Alert Metrics",
	Tags:        alertsTags,
	Errors:      errorCodes(),
}

type GetAlertMetricsRequest struct {
	GetIdRequest
	RosterId uuid.UUID    `query:"rosterId"`
	From     CalendarDate `query:"from" format:"date" required:"true"`
	To       CalendarDate `query:"to" format:"date" required:"true"`
}
type GetAlertMetricsResponse ItemResponse[AlertMetrics]

var ListAlertIncidentLinks = huma.Operation{
	OperationID: "list-alert-incident-links",
	Method:      http.MethodGet,
	Path:        "/alerts/{id}/incident_links",
	Summary:     "List Incident Links for an Alert",
	Tags:        alertsTags,
	Errors:      errorCodes(),
}

type ListAlertIncidentLinksRequest struct {
	ListIdRequest
}
type ListAlertIncidentLinksResponse PaginatedResponse[AlertIncidentLink]

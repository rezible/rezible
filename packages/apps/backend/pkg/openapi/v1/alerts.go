package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/openapi"
)

type AlertsHandler interface {
	ListAlertDefinitions(context.Context, *ListAlertDefinitionsRequest) (*ListAlertDefinitionsResponse, error)
	GetAlertDefinition(context.Context, *GetAlertDefinitionRequest) (*GetAlertDefinitionResponse, error)

	GetAlertMetrics(context.Context, *GetAlertMetricsRequest) (*GetAlertMetricsResponse, error)
	ListAlertIncidentLinks(context.Context, *ListAlertIncidentLinksRequest) (*ListAlertIncidentLinksResponse, error)
}

func (o operations) RegisterAlerts(api huma.API) {
	huma.Register(api, ListAlerts, o.ListAlertDefinitions)
	huma.Register(api, GetAlert, o.GetAlertDefinition)
	huma.Register(api, GetAlertMetrics, o.GetAlertMetrics)
	huma.Register(api, ListAlertIncidentLinks, o.ListAlertIncidentLinks)
}

type (
	AlertDefinition struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes AlertDefinitionAttributes `json:"attributes"`
	}

	AlertDefinitionAttributes struct {
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

	AlertEpisode struct {
		Id         uuid.UUID              `json:"id"`
		Attributes AlertEpisodeAttributes `json:"attributes"`
	}

	AlertEpisodeAttributes struct {
		Status         string           `json:"status" enum:"open,closed"`
		Definition     *AlertDefinition `json:"definition,omitempty"`
		StartedAt      time.Time        `json:"startedAt"`
		LastObservedAt time.Time        `json:"lastObservedAt"`
		ClosedAt       *time.Time       `json:"closedAt,omitempty"`
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

func AlertDefinitionFromEnt(a *ent.AlertDefinition) AlertDefinition {
	attrs := AlertDefinitionAttributes{
		Title:       a.Title,
		Description: a.Description,
		Definition:  a.Definition,
	}

	return AlertDefinition{Id: a.ID, Attributes: attrs}
}

func AlertEpisodeFromEnt(ep *ent.AlertEpisode) AlertEpisode {
	attrs := AlertEpisodeAttributes{
		Status:         string(ep.Status),
		StartedAt:      ep.StartedAt,
		LastObservedAt: ep.LastObservedAt,
		ClosedAt:       ep.ClosedAt,
	}
	if def := ep.Edges.AlertDefinition; def != nil {
		attrs.Definition = new(AlertDefinitionFromEnt(def))
	}
	return AlertEpisode{Id: ep.ID, Attributes: attrs}
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

var ListAlerts = openapi.Operation{
	OperationID: "list-alerts",
	Method:      http.MethodGet,
	Path:        "/alerts",
	Summary:     "List Alerts",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type ListAlertDefinitionsRequest struct {
	PaginationRequest
	Search string `query:"search" required:"false" nullable:"false"`
}
type ListAlertDefinitionsResponse PaginatedResponse[AlertDefinition]

var GetAlert = openapi.Operation{
	OperationID: "get-alert",
	Method:      http.MethodGet,
	Path:        "/alerts/{id}",
	Summary:     "Get Alert",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type GetAlertDefinitionRequest struct {
	IdRequest
}
type GetAlertDefinitionResponse ItemResponse[AlertDefinition]

var GetAlertMetrics = openapi.Operation{
	OperationID: "get-alert-metrics",
	Method:      http.MethodGet,
	Path:        "/alerts/{id}/metrics",
	Summary:     "Get Alert Metrics",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type GetAlertMetricsRequest struct {
	IdRequest
	RosterId uuid.UUID    `query:"rosterId"`
	From     CalendarDate `query:"from" format:"date" required:"true"`
	To       CalendarDate `query:"to" format:"date" required:"true"`
}
type GetAlertMetricsResponse ItemResponse[AlertMetrics]

var ListAlertIncidentLinks = openapi.Operation{
	OperationID: "list-alert-incident-links",
	Method:      http.MethodGet,
	Path:        "/alerts/{id}/incident_links",
	Summary:     "List Incident Links for an Alert",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type ListAlertIncidentLinksRequest struct {
	IdRequest
}
type ListAlertIncidentLinksResponse CollectionResponse[AlertIncidentLink]

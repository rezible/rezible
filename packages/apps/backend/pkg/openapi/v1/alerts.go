package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/situationsignalattention"
	"github.com/rezible/rezible/pkg/openapi"
)

type AlertsHandler interface {
	SetAlertIdentityGroupLabels(context.Context, *SetAlertIdentityGroupLabelsRequest) (*SetAlertIdentityGroupLabelsResponse, error)
	SetAlertSituationSignalAttention(context.Context, *SetAlertSituationSignalAttentionRequest) (*SetAlertSituationSignalAttentionResponse, error)
	ListAlertDefinitions(context.Context, *ListAlertDefinitionsRequest) (*ListAlertDefinitionsResponse, error)
	GetAlertDefinition(context.Context, *GetAlertDefinitionRequest) (*GetAlertDefinitionResponse, error)

	GetAlertMetrics(context.Context, *GetAlertMetricsRequest) (*GetAlertMetricsResponse, error)
	ListAlertIncidentLinks(context.Context, *ListAlertIncidentLinksRequest) (*ListAlertIncidentLinksResponse, error)
	ListSituationAlertEpisodes(context.Context, *ListSituationAlertEpisodesRequest) (*ListSituationAlertEpisodesResponse, error)
}

func (o operations) RegisterAlerts(api huma.API) {
	huma.Register(api, SetAlertIdentityGroupLabels, o.SetAlertIdentityGroupLabels)
	huma.Register(api, SetAlertSituationSignalAttention, o.SetAlertSituationSignalAttention)
	huma.Register(api, ListAlerts, o.ListAlertDefinitions)
	huma.Register(api, GetAlert, o.GetAlertDefinition)
	huma.Register(api, GetAlertMetrics, o.GetAlertMetrics)
	huma.Register(api, ListAlertIncidentLinks, o.ListAlertIncidentLinks)
	huma.Register(api, ListSituationAlertEpisodes, o.ListSituationAlertEpisodes)
}

type (
	AlertDefinition struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes AlertDefinitionAttributes `json:"attributes"`
	}

	AlertDefinitionAttributes struct {
		Title                    string                              `json:"title"`
		Description              string                              `json:"description"`
		Definition               string                              `json:"definition"`
		KnowledgeEntityId        *uuid.UUID                          `json:"knowledgeEntityId,omitempty"`
		ResolutionTimeoutSeconds int                                 `json:"resolutionTimeoutSeconds" doc:"How long after its last firing notification a window is assumed ended; 0 never"`
		IdentityGroupLabels      []string                            `json:"identityGroupLabels" doc:"Label names that group the alert's instances"`
		SituationSignalAttention AlertSituationSignalAttention       `json:"situationSignalAttention"`
		Roster                   *Expandable[OncallRosterAttributes] `json:"roster,omitempty"`
	}

	AlertSituationSignalAttention struct {
		Level string     `json:"level" enum:"default,watch_only,join_only"`
		SetAt *time.Time `json:"setAt,omitempty"`
	}

	AlertInstance struct {
		Id         uuid.UUID               `json:"id"`
		Attributes AlertInstanceAttributes `json:"attributes"`
	}

	AlertInstanceAttributes struct {
		InstanceKey    string                 `json:"instanceKey"`
		GroupingKey    string                 `json:"groupingKey"`
		Labels         map[string]string      `json:"labels"`
		Summary        string                 `json:"summary"`
		Severity       string                 `json:"severity" enum:"unknown,info,warning,critical"`
		FiredAt        time.Time              `json:"firedAt"`
		LastObservedAt time.Time              `json:"lastObservedAt"`
		EndedAt        *time.Time             `json:"endedAt,omitempty"`
		EndReason      *string                `json:"endReason,omitempty" enum:"resolved,superseded,timeout"`
		Feedback       *AlertInstanceFeedback `json:"feedback,omitempty"`
	}

	AlertEpisode struct {
		Id         uuid.UUID              `json:"id"`
		Attributes AlertEpisodeAttributes `json:"attributes"`
	}

	AlertEpisodeAttributes struct {
		KnowledgeEntityId *uuid.UUID       `json:"knowledgeEntityId,omitempty"`
		Instances         []AlertInstance  `json:"instances"`
		Status            string           `json:"status" enum:"open,closed"`
		Definition        *AlertDefinition `json:"definition,omitempty"`
		StartedAt         time.Time        `json:"startedAt"`
		ClosedAt          *time.Time       `json:"closedAt,omitempty"`
		HighestSeverity   string           `json:"highestSeverity" enum:"unknown,info,warning,critical"`
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
		Title:                    a.Title,
		Description:              a.Description,
		Definition:               a.Definition,
		KnowledgeEntityId:        a.KnowledgeEntityID,
		ResolutionTimeoutSeconds: a.ResolutionTimeoutSeconds,
		IdentityGroupLabels:      a.IdentityGroupLabels,
		SituationSignalAttention: AlertSituationSignalAttention{Level: situationsignalattention.LevelDefault.String()},
	}
	if attrs.IdentityGroupLabels == nil {
		attrs.IdentityGroupLabels = []string{}
	}
	if attention := a.Edges.SituationSignalAttention; attention != nil {
		attrs.SituationSignalAttention = AlertSituationSignalAttention{Level: attention.Level.String(), SetAt: &attention.SetAt}
	}

	return AlertDefinition{Id: a.ID, Attributes: attrs}
}

func AlertInstanceFromEnt(instance *ent.AlertInstance) AlertInstance {
	attrs := AlertInstanceAttributes{
		InstanceKey:    instance.InstanceKey,
		GroupingKey:    instance.GroupingKey,
		Labels:         instance.Labels,
		Summary:        instance.Summary,
		Severity:       string(instance.Severity),
		FiredAt:        instance.FiredAt,
		LastObservedAt: instance.LastObservedAt,
		EndedAt:        instance.EndedAt,
	}
	if instance.EndReason != nil {
		attrs.EndReason = new(instance.EndReason.String())
	}
	return AlertInstance{Id: instance.ID, Attributes: attrs}
}

func AlertEpisodeFromEnt(ep *ent.AlertEpisode) AlertEpisode {
	attrs := AlertEpisodeAttributes{
		KnowledgeEntityId: ep.KnowledgeEntityID,
		Instances:         ConvertSlice(ep.Edges.Instances, AlertInstanceFromEnt),
		Status:            "open",
		StartedAt:         ep.StartedAt,
		ClosedAt:          ep.ClosedAt,
		HighestSeverity:   string(ep.HighestSeverity),
	}
	if ep.ClosedAt != nil {
		attrs.Status = "closed"
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

var ListSituationAlertEpisodes = openapi.Operation{
	OperationID: "list-situation-alert-episodes",
	Method:      http.MethodGet,
	Path:        "/situations/{id}/alert_episodes",
	Summary:     "List the Alert Episodes of a Situation",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type ListSituationAlertEpisodesRequest struct {
	IdRequest
	PaginationRequest
}
type ListSituationAlertEpisodesResponse PaginatedResponse[AlertEpisode]

var SetAlertIdentityGroupLabels = openapi.Operation{
	OperationID: "set-alert-identity-group-labels",
	Method:      http.MethodPut,
	Path:        "/alerts/{id}/identity_group_labels",
	Summary:     "Set Alert Identity Group Labels",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type SetAlertIdentityGroupLabelsAttributes struct {
	Labels []string `json:"labels"`
}
type SetAlertIdentityGroupLabelsRequest IdRequestWithBody[SetAlertIdentityGroupLabelsAttributes]
type SetAlertIdentityGroupLabelsResponse ItemResponse[AlertDefinition]

var SetAlertSituationSignalAttention = openapi.Operation{
	OperationID: "set-alert-situation-signal-attention",
	Method:      http.MethodPut,
	Path:        "/alerts/{id}/situation_signal_attention",
	Summary:     "Set Alert Situation Signal Attention",
	Tags:        alertsTags,
	Errors:      ErrorCodes(),
}

type SetAlertSituationSignalAttentionAttributes struct {
	Level situationsignalattention.Level `json:"level" enum:"default,watch_only,join_only"`
}
type SetAlertSituationSignalAttentionRequest IdRequestWithBody[SetAlertSituationSignalAttentionAttributes]
type SetAlertSituationSignalAttentionResponse ItemResponse[AlertDefinition]

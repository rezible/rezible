package v1

import (
	"context"

	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type IncidentTimelineHandler interface {
	ListIncidentTimelineEvents(context.Context, *ListIncidentTimelineEventsRequest) (*ListIncidentTimelineEventsResponse, error)
	CreateIncidentTimelineEvent(context.Context, *CreateIncidentTimelineEventRequest) (*CreateIncidentTimelineEventResponse, error)
	UpdateIncidentTimelineEvent(context.Context, *UpdateIncidentTimelineEventRequest) (*UpdateIncidentTimelineEventResponse, error)
	DeleteIncidentTimelineEvent(context.Context, *DeleteIncidentTimelineEventRequest) (*DeleteIncidentTimelineEventResponse, error)

	GetIncidentTimelineEventMetadata(context.Context, *GetIncidentTimelineEventMetadataRequest) (*GetIncidentTimelineEventMetadataResponse, error)
}

func (o operations) RegisterIncidentTimeline(api huma.API) {
	huma.Register(api, ListIncidentTimelineEvents, o.ListIncidentTimelineEvents)
	huma.Register(api, CreateIncidentTimelineEvent, o.CreateIncidentTimelineEvent)
	huma.Register(api, UpdateIncidentTimelineEvent, o.UpdateIncidentTimelineEvent)
	huma.Register(api, DeleteIncidentTimelineEvent, o.DeleteIncidentTimelineEvent)

	huma.Register(api, GetIncidentTimelineEventMetadata, o.GetIncidentTimelineEventMetadata)
}

type (
	IncidentTimelineEvent struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes IncidentTimelineEventAttributes `json:"attributes"`
	}
	IncidentTimelineEventAttributes struct {
		IncidentId          uuid.UUID                                 `json:"incidentId"`
		Kind                string                                    `json:"kind" enum:"observation,action,decision,context"`
		Timestamp           time.Time                                 `json:"timestamp"`
		IsKey               bool                                      `json:"isKey"`
		Title               string                                    `json:"title"`
		Description         string                                    `json:"description,omitempty"`
		Sequence            int                                       `json:"sequence"`
		DecisionContext     *IncidentTimelineEventDecisionContext     `json:"decisionContext,omitempty"`
		ContributingFactors []IncidentTimelineEventContributingFactor `json:"contributingFactors"`
		Evidence            []IncidentTimelineEventEvidence           `json:"evidence"`
		SystemContext       []IncidentTimelineEventSystemContext      `json:"systemContext"`
	}

	IncidentTimelineEventDecisionContext struct {
		OptionsConsidered []string `json:"optionsConsidered"`
		Constraints       []string `json:"constraints"`
		DecisionRationale string   `json:"decisionRationale"`
	}

	IncidentTimelineEventContributingFactor struct {
		Id         uuid.UUID                                         `json:"id"`
		Attributes IncidentTimelineEventContributingFactorAttributes `json:"attributes"`
	}

	IncidentTimelineEventContributingFactorAttributes struct {
		FactorTypeId uuid.UUID `json:"factorTypeId"`
		Description  string    `json:"description"`
		Links        []string  `json:"links"`
	}

	IncidentTimelineEventEvidence struct {
		Id         uuid.UUID                               `json:"id"`
		Attributes IncidentTimelineEventEvidenceAttributes `json:"attributes"`
	}

	IncidentTimelineEventEvidenceAttributes struct {
		Source     string             `json:"source"`
		Value      string             `json:"value"`
		Properties *map[string]string `json:"properties,omitempty"`
	}

	IncidentTimelineEventSystemContext struct {
		Id         uuid.UUID                                    `json:"id"`
		Attributes IncidentTimelineEventSystemContextAttributes `json:"attributes"`
	}

	IncidentTimelineEventSystemContextAttributes struct {
		SystemAnalysisNodeId uuid.UUID `json:"systemAnalysisNodeId"`
		Relationship         string    `json:"relationship" enum:"primary,affected,contributing"`
	}

	IncidentTimelineEventMetadata struct {
		ContributingFactorCategories []IncidentTimelineEventContributingFactorCategory `json:"contributingFactorCategories"`
	}

	IncidentTimelineEventContributingFactorCategory struct {
		Id         uuid.UUID                                                 `json:"id"`
		Attributes IncidentTimelineEventContributingFactorCategoryAttributes `json:"attributes"`
	}

	IncidentTimelineEventContributingFactorCategoryAttributes struct {
		Label       string                                        `json:"name"`
		Description string                                        `json:"description"`
		FactorTypes []IncidentTimelineEventContributingFactorType `json:"factorTypes"`
	}

	IncidentTimelineEventContributingFactorType struct {
		Id         uuid.UUID                                             `json:"id"`
		Attributes IncidentTimelineEventContributingFactorTypeAttributes `json:"attributes"`
	}

	IncidentTimelineEventContributingFactorTypeAttributes struct {
		Label       string   `json:"name"`
		Description string   `json:"description"`
		Examples    []string `json:"examples"`
	}
)

func IncidentTimelineEventFromEnt(e *ent.IncidentTimelineEvent) IncidentTimelineEvent {
	attr := IncidentTimelineEventAttributes{
		IncidentId:  e.IncidentID,
		Kind:        e.Kind.String(),
		Timestamp:   e.Timestamp,
		IsKey:       e.IsKey,
		Title:       e.Title,
		Description: e.Description,
		Sequence:    e.Sequence,
	}

	if e.Edges.Context != nil {
		attr.DecisionContext = new(IncidentTimelineEventDecisionContextFromEnt(e.Edges.Context))
	}

	attr.ContributingFactors = make([]IncidentTimelineEventContributingFactor, len(e.Edges.Factors))
	for i, f := range e.Edges.Factors {
		attr.ContributingFactors[i] = IncidentTimelineEventContributingFactorFromEnt(f)
	}

	attr.Evidence = make([]IncidentTimelineEventEvidence, len(e.Edges.Evidence))
	for i, evi := range e.Edges.Evidence {
		attr.Evidence[i] = IncidentTimelineEventEvidenceFromEnt(evi)
	}
	attr.SystemContext = make([]IncidentTimelineEventSystemContext, len(e.Edges.SystemContext))
	for i, systemContext := range e.Edges.SystemContext {
		attr.SystemContext[i] = IncidentTimelineEventSystemContextFromEnt(systemContext)
	}

	return IncidentTimelineEvent{Id: e.ID, Attributes: attr}
}

func IncidentTimelineEventSystemContextFromEnt(context *ent.IncidentTimelineEventSystemContext) IncidentTimelineEventSystemContext {
	return IncidentTimelineEventSystemContext{
		Id: context.ID,
		Attributes: IncidentTimelineEventSystemContextAttributes{
			SystemAnalysisNodeId: context.SystemAnalysisNodeID,
			Relationship:         context.Relationship.String(),
		},
	}
}

func IncidentTimelineEventDecisionContextFromEnt(c *ent.IncidentTimelineEventContext) IncidentTimelineEventDecisionContext {
	return IncidentTimelineEventDecisionContext{
		OptionsConsidered: c.DecisionOptions,
		//Constraints:       nil,
		DecisionRationale: c.DecisionRationale,
	}
}

func IncidentTimelineEventContributingFactorFromEnt(f *ent.IncidentTimelineEventContributingFactor) IncidentTimelineEventContributingFactor {
	return IncidentTimelineEventContributingFactor{
		Id: f.ID,
		Attributes: IncidentTimelineEventContributingFactorAttributes{
			// FactorTypeId: f.FactorType,
			Description: f.Description,
			// Links:        nil,
		},
	}
}

func IncidentTimelineEventEvidenceFromEnt(evi *ent.IncidentTimelineEventEvidence) IncidentTimelineEventEvidence {
	return IncidentTimelineEventEvidence{
		Id:         evi.ID,
		Attributes: IncidentTimelineEventEvidenceAttributes{
			//Source:     "",
			//Value:      "",
			//Properties: nil,
		},
	}
}

var incidentTimelineTags = []string{"Incident Timeline"}

// ops

var ListIncidentTimelineEvents = huma.Operation{
	OperationID: "list-incident-timeline-events",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/timeline_events",
	Summary:     "List Events for Incident",
	Tags:        append(incidentsTags, incidentTimelineTags...),
	Errors:      ErrorCodes(),
}

type ListIncidentTimelineEventsRequest ListIdRequest
type ListIncidentTimelineEventsResponse ListResponse[IncidentTimelineEvent]

var CreateIncidentTimelineEvent = huma.Operation{
	OperationID: "create-incident-timeline-event",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/timeline_events",
	Summary:     "Create an Incident Event",
	Tags:        incidentTimelineTags,
	Errors:      ErrorCodes(),
}

type CreateIncidentTimelineEventAttributes struct {
	Title         string                                            `json:"title"`
	Kind          string                                            `json:"kind" enum:"observation,action,decision,context"`
	IsKey         bool                                              `json:"isKey" required:"false"`
	Timestamp     time.Time                                         `json:"timestamp"`
	SystemContext []SetIncidentTimelineEventSystemContextAttributes `json:"systemContext,omitempty"`
}
type CreateIncidentTimelineEventRequest IdRequestWithBody[CreateIncidentTimelineEventAttributes]
type CreateIncidentTimelineEventResponse ItemResponse[IncidentTimelineEvent]

var UpdateIncidentTimelineEvent = huma.Operation{
	OperationID: "update-incident-timeline-event",
	Method:      http.MethodPatch,
	Path:        "/incident_timeline/events/{id}",
	Summary:     "Update an Incident Event",
	Tags:        incidentTimelineTags,
	Errors:      ErrorCodes(),
}

type UpdateIncidentTimelineEventAttributes struct {
	Title         *string                                            `json:"title,omitempty"`
	Kind          *string                                            `json:"kind,omitempty" enum:"observation,action,decision,context"`
	Timestamp     *time.Time                                         `json:"timestamp,omitempty"`
	SystemContext *[]SetIncidentTimelineEventSystemContextAttributes `json:"systemContext,omitempty"`
}

type SetIncidentTimelineEventSystemContextAttributes struct {
	SystemAnalysisNodeId uuid.UUID `json:"systemAnalysisNodeId"`
	Relationship         string    `json:"relationship" enum:"primary,affected,contributing"`
}
type UpdateIncidentTimelineEventRequest IdRequestWithBody[UpdateIncidentTimelineEventAttributes]
type UpdateIncidentTimelineEventResponse ItemResponse[IncidentTimelineEvent]

var DeleteIncidentTimelineEvent = huma.Operation{
	OperationID: "delete-incident-timeline-event",
	Method:      http.MethodDelete,
	Path:        "/incident_timeline/events/{id}",
	Summary:     "Delete an Incident Event",
	Tags:        incidentTimelineTags,
	Errors:      ErrorCodes(),
}

type DeleteIncidentTimelineEventRequest IdRequest
type DeleteIncidentTimelineEventResponse EmptyResponse

var GetIncidentTimelineEventMetadata = huma.Operation{
	OperationID: "list-incident-timeline-event-metadata",
	Method:      http.MethodGet,
	Path:        "/incident_timeline/event_metadata",
	Summary:     "Get metadata available for incident timeline events",
	Tags:        incidentTimelineTags,
	Errors:      ErrorCodes(),
}

type GetIncidentTimelineEventMetadataRequest EmptyRequest
type GetIncidentTimelineEventMetadataResponse ItemResponse[IncidentTimelineEventMetadata]

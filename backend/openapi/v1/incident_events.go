package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"
	"time"
)

type IncidentEventsHandler interface {
	ListIncidentEvents(context.Context, *ListIncidentEventsRequest) (*ListIncidentEventsResponse, error)
	CreateIncidentEvent(context.Context, *CreateIncidentEventRequest) (*CreateIncidentEventResponse, error)
	UpdateIncidentEvent(context.Context, *UpdateIncidentEventRequest) (*UpdateIncidentEventResponse, error)
	DeleteIncidentEvent(context.Context, *DeleteIncidentEventRequest) (*DeleteIncidentEventResponse, error)

	ListIncidentEventContributingFactors(context.Context, *ListIncidentEventContributingFactorsRequest) (*ListIncidentEventContributingFactorsResponse, error)
}

func (o operations) RegisterIncidentEvents(api huma.API) {
	huma.Register(api, ListIncidentEvents, o.ListIncidentEvents)
	huma.Register(api, CreateIncidentEvent, o.CreateIncidentEvent)
	huma.Register(api, UpdateIncidentEvent, o.UpdateIncidentEvent)
	huma.Register(api, DeleteIncidentEvent, o.DeleteIncidentEvent)

	huma.Register(api, ListIncidentEventContributingFactors, o.ListIncidentEventContributingFactors)
}

type (
	IncidentEvent struct {
		Id         uuid.UUID               `json:"id"`
		Attributes IncidentEventAttributes `json:"attributes"`
	}
	IncidentEventAttributes struct {
		IncidentId          uuid.UUID                         `json:"incidentId"`
		Kind                string                            `json:"kind" enum:"observation,action,decision,context"`
		Timestamp           time.Time                         `json:"timestamp"`
		IsKey               bool                              `json:"isKey"`
		Title               string                            `json:"title"`
		Description         string                            `json:"description,omitempty"`
		Sequence            int                               `json:"sequence"`
		DecisionContext     *IncidentEventDecisionContext     `json:"decisionContext,omitempty"`
		ContributingFactors []IncidentEventContributingFactor `json:"contributingFactors"`
		Evidence            []IncidentEventEvidence           `json:"evidence"`
		SystemContext       []IncidentEventSystemComponent    `json:"systemContext"`
	}

	IncidentEventDecisionContext struct {
		OptionsConsidered []string `json:"optionsConsidered"`
		Constraints       []string `json:"constraints"`
		DecisionRationale string   `json:"decisionRationale"`
	}

	IncidentEventContributingFactor struct {
		Id         uuid.UUID                                 `json:"id"`
		Attributes IncidentEventContributingFactorAttributes `json:"attributes"`
	}

	IncidentEventContributingFactorAttributes struct {
		FactorTypeId uuid.UUID `json:"factorTypeId"`
		Description  string    `json:"description"`
		Links        []string  `json:"links"`
	}

	IncidentEventEvidence struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes IncidentEventEvidenceAttributes `json:"attributes"`
	}

	IncidentEventEvidenceAttributes struct {
		Source     string             `json:"source"`
		Value      string             `json:"value"`
		Properties *map[string]string `json:"properties,omitempty"`
	}

	IncidentEventSystemComponent struct {
		Id         uuid.UUID                              `json:"id"`
		Attributes IncidentEventSystemComponentAttributes `json:"attributes"`
	}

	IncidentEventSystemComponentAttributes struct {
		AnalysisComponentId uuid.UUID `json:"analysisComponentId"`
		Status              string    `json:"status"`
		Description         string    `json:"description"`
		// TODO: what else do we want as context?
	}

	IncidentEventContributingFactorCategory struct {
		Id         uuid.UUID                                         `json:"id"`
		Attributes IncidentEventContributingFactorCategoryAttributes `json:"attributes"`
	}

	IncidentEventContributingFactorCategoryAttributes struct {
		Label       string                                `json:"name"`
		Description string                                `json:"description"`
		FactorTypes []IncidentEventContributingFactorType `json:"factorTypes"`
	}

	IncidentEventContributingFactorType struct {
		Id         uuid.UUID                                     `json:"id"`
		Attributes IncidentEventContributingFactorTypeAttributes `json:"attributes"`
	}

	IncidentEventContributingFactorTypeAttributes struct {
		Label       string   `json:"name"`
		Description string   `json:"description"`
		Examples    []string `json:"examples"`
	}
)

func IncidentEventFromEnt(e *ent.IncidentEvent) IncidentEvent {
	attr := IncidentEventAttributes{
		IncidentId:  e.IncidentID,
		Kind:        e.Kind.String(),
		Timestamp:   e.Timestamp,
		IsKey:       e.IsKey,
		Title:       e.Title,
		Description: e.Description,
		Sequence:    e.Sequence,
	}

	if e.Edges.Context != nil {
		c := IncidentEventDecisionContextFromEnt(e.Edges.Context)
		attr.DecisionContext = &c
	}

	attr.ContributingFactors = make([]IncidentEventContributingFactor, len(e.Edges.Factors))
	for i, f := range e.Edges.Factors {
		attr.ContributingFactors[i] = IncidentEventContributingFactorFromEnt(f)
	}

	attr.Evidence = make([]IncidentEventEvidence, len(e.Edges.Evidence))
	for i, evi := range e.Edges.Evidence {
		attr.Evidence[i] = IncidentEventEvidenceFromEnt(evi)
	}

	attr.SystemContext = make([]IncidentEventSystemComponent, len(e.Edges.EventComponents))
	for i, c := range e.Edges.EventComponents {
		attr.SystemContext[i] = IncidentEventSystemComponentFromEnt(c)
	}

	return IncidentEvent{Id: e.ID, Attributes: attr}
}

func IncidentEventDecisionContextFromEnt(c *ent.IncidentEventContext) IncidentEventDecisionContext {
	return IncidentEventDecisionContext{
		OptionsConsidered: c.DecisionOptions,
		//Constraints:       nil,
		DecisionRationale: c.DecisionRationale,
	}
}

func IncidentEventContributingFactorFromEnt(f *ent.IncidentEventContributingFactor) IncidentEventContributingFactor {
	return IncidentEventContributingFactor{
		Id: f.ID,
		Attributes: IncidentEventContributingFactorAttributes{
			// FactorTypeId: f.FactorType,
			Description: f.Description,
			// Links:        nil,
		},
	}
}

func IncidentEventEvidenceFromEnt(evi *ent.IncidentEventEvidence) IncidentEventEvidence {
	return IncidentEventEvidence{
		Id:         evi.ID,
		Attributes: IncidentEventEvidenceAttributes{
			//Source:     "",
			//Value:      "",
			//Properties: nil,
		},
	}
}

func IncidentEventSystemComponentFromEnt(c *ent.IncidentEventSystemComponent) IncidentEventSystemComponent {
	return IncidentEventSystemComponent{
		Id:         c.ID,
		Attributes: IncidentEventSystemComponentAttributes{
			//AnalysisComponentId: uuid.UUID{},
			//Status:              "",
			//Description:         "",
		},
	}
}

var incidentEventsTags = []string{"Incident Events"}

// ops

var ListIncidentEvents = huma.Operation{
	OperationID: "list-incident-events",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/events",
	Summary:     "List Events for Incident",
	Tags:        append(incidentsTags, incidentEventsTags...),
	Errors:      errorCodes(),
}

type ListIncidentEventsRequest ListIdRequest
type ListIncidentEventsResponse PaginatedResponse[IncidentEvent]

var CreateIncidentEvent = huma.Operation{
	OperationID: "create-incident-event",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/events",
	Summary:     "Create an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type CreateIncidentEventAttributes struct {
	Title     string    `json:"title"`
	Kind      string    `json:"kind" enum:"observation,action,decision,context"`
	IsKey     bool      `json:"isKey" required:"false"`
	Timestamp time.Time `json:"timestamp"`
}
type CreateIncidentEventRequest CreateIdRequest[CreateIncidentEventAttributes]
type CreateIncidentEventResponse ItemResponse[IncidentEvent]

var UpdateIncidentEvent = huma.Operation{
	OperationID: "update-incident-event",
	Method:      http.MethodPatch,
	Path:        "/incident_events/{id}",
	Summary:     "Update an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentEventAttributes struct {
	Title     *string    `json:"title,omitempty"`
	Kind      *string    `json:"kind,omitempty" enum:"observation,action,decision,context"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}
type UpdateIncidentEventRequest UpdateIdRequest[UpdateIncidentEventAttributes]
type UpdateIncidentEventResponse ItemResponse[IncidentEvent]

var DeleteIncidentEvent = huma.Operation{
	OperationID: "delete-incident-event",
	Method:      http.MethodDelete,
	Path:        "/incident_events/{id}",
	Summary:     "Delete an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type DeleteIncidentEventRequest DeleteIdRequest
type DeleteIncidentEventResponse EmptyResponse

var ListIncidentEventContributingFactors = huma.Operation{
	OperationID: "list-incident-event-contributing-factor-categories",
	Method:      http.MethodGet,
	Path:        "/incident_event_contributing_factor_categories",
	Summary:     "List Categories of Contributing Factors used in Incident Events",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type ListIncidentEventContributingFactorsRequest ListRequest
type ListIncidentEventContributingFactorsResponse PaginatedResponse[IncidentEventContributingFactorCategory]

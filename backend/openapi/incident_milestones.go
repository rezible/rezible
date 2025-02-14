package openapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type IncidentMilestonesHandler interface {
	ListIncidentMilestones(context.Context, *ListIncidentMilestonesRequest) (*ListIncidentMilestonesResponse, error)
	CreateIncidentMilestone(context.Context, *CreateIncidentMilestoneRequest) (*CreateIncidentMilestoneResponse, error)
	UpdateIncidentMilestone(context.Context, *UpdateIncidentMilestoneRequest) (*UpdateIncidentMilestoneResponse, error)
	ArchiveIncidentMilestone(context.Context, *ArchiveIncidentMilestoneRequest) (*ArchiveIncidentMilestoneResponse, error)
}

func (o operations) RegisterIncidentMilestones(api huma.API) {
	huma.Register(api, ListIncidentMilestones, o.ListIncidentMilestones)
	huma.Register(api, CreateIncidentMilestone, o.CreateIncidentMilestone)
	huma.Register(api, UpdateIncidentMilestone, o.UpdateIncidentMilestone)
	huma.Register(api, ArchiveIncidentMilestone, o.ArchiveIncidentMilestone)
}

type (
	IncidentMilestone struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes IncidentMilestoneAttributes `json:"attributes"`
	}
	IncidentMilestoneAttributes struct {
		Type      string    `json:"type" enum:"default,incident"`
		Title     string    `json:"title"`
		Timestamp time.Time `json:"timestamp"`
	}
)

func IncidentMilestoneFromEnt(m *ent.IncidentMilestone) IncidentMilestone {
	return IncidentMilestone{
		Id: m.ID,
		Attributes: IncidentMilestoneAttributes{
			Type:      m.Type.String(),
			Timestamp: m.Time,
		},
	}
}

var incidentMilestonesTags = []string{"Incident Milestones"}

// ops

var ListIncidentMilestones = huma.Operation{
	OperationID: "list-incident-milestones",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/milestones",
	Summary:     "List Milestones for Incident",
	Tags:        append(incidentsTags, incidentMilestonesTags...),
	Errors:      errorCodes(),
}

type ListIncidentMilestonesRequest GetIdRequest
type ListIncidentMilestonesResponse PaginatedResponse[IncidentMilestone]

var CreateIncidentMilestone = huma.Operation{
	OperationID: "create-incident-milestone",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/milestones",
	Summary:     "Create an Incident Milestone",
	Tags:        incidentMilestonesTags,
	Errors:      errorCodes(),
}

type CreateIncidentMilestoneAttributes struct {
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
type CreateIncidentMilestoneRequest CreateIdRequest[CreateIncidentMilestoneAttributes]
type CreateIncidentMilestoneResponse ItemResponse[IncidentMilestone]

var UpdateIncidentMilestone = huma.Operation{
	OperationID: "update-incident-milestone",
	Method:      http.MethodPatch,
	Path:        "/incident_milestones/{id}",
	Summary:     "Update an Incident Milestone",
	Tags:        incidentMilestonesTags,
	Errors:      errorCodes(),
}

type UpdateIncidentMilestoneAttributes struct {
	Title     *string    `json:"title"`
	Type      *string    `json:"type"`
	Timestamp *time.Time `json:"timestamp"`
}
type UpdateIncidentMilestoneRequest UpdateIdRequest[UpdateIncidentMilestoneAttributes]
type UpdateIncidentMilestoneResponse ItemResponse[IncidentMilestone]

var ArchiveIncidentMilestone = huma.Operation{
	OperationID: "archive-incident-milestone",
	Method:      http.MethodDelete,
	Path:        "/incident_milestones/{id}",
	Summary:     "Archive an Incident Milestone",
	Tags:        incidentMilestonesTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentMilestoneRequest ArchiveIdRequest
type ArchiveIncidentMilestoneResponse EmptyResponse

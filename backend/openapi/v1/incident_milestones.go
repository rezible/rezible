package v1

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
	DeleteIncidentMilestone(context.Context, *DeleteIncidentMilestoneRequest) (*DeleteIncidentMilestoneResponse, error)
}

func (o operations) RegisterIncidentMilestones(api huma.API) {
	huma.Register(api, ListIncidentMilestones, o.ListIncidentMilestones)
	huma.Register(api, CreateIncidentMilestone, o.CreateIncidentMilestone)
	huma.Register(api, UpdateIncidentMilestone, o.UpdateIncidentMilestone)
	huma.Register(api, DeleteIncidentMilestone, o.DeleteIncidentMilestone)
}

type (
	IncidentMilestone struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes IncidentMilestoneAttributes `json:"attributes"`
	}
	IncidentMilestoneAttributes struct {
		Kind        string    `json:"kind" enum:"impact,detection,investigation,mitigation,resolution"`
		Description string    `json:"description"`
		Timestamp   time.Time `json:"timestamp"`
	}
)

func IncidentMilestoneFromEnt(m *ent.IncidentMilestone) IncidentMilestone {
	return IncidentMilestone{
		Id: m.ID,
		Attributes: IncidentMilestoneAttributes{
			Kind:      m.Kind.String(),
			Timestamp: m.Timestamp,
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
	Kind        string    `json:"kind" enum:"impact,detection,investigation,mitigation,resolution"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
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
	Kind        *string    `json:"kind,omitempty" enum:"impact,detection,investigation,mitigation,resolution"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	Description *string    `json:"description,omitempty"`
}
type UpdateIncidentMilestoneRequest UpdateIdRequest[UpdateIncidentMilestoneAttributes]
type UpdateIncidentMilestoneResponse ItemResponse[IncidentMilestone]

var DeleteIncidentMilestone = huma.Operation{
	OperationID: "delete-incident-milestone",
	Method:      http.MethodDelete,
	Path:        "/incident_milestones/{id}",
	Summary:     "Delete an Incident Milestone",
	Tags:        incidentMilestonesTags,
	Errors:      errorCodes(),
}

type DeleteIncidentMilestoneRequest DeleteIdRequest
type DeleteIncidentMilestoneResponse EmptyResponse

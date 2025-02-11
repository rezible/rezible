package openapi

import (
	"context"
	"github.com/rezible/rezible/ent"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type IncidentsHandler interface {
	ListIncidents(context.Context, *ListIncidentsRequest) (*ListIncidentsResponse, error)
	CreateIncident(context.Context, *CreateIncidentRequest) (*CreateIncidentResponse, error)
	GetIncident(context.Context, *GetIncidentRequest) (*GetIncidentResponse, error)
	UpdateIncident(context.Context, *UpdateIncidentRequest) (*UpdateIncidentResponse, error)
	ArchiveIncident(context.Context, *ArchiveIncidentRequest) (*ArchiveIncidentResponse, error)
}

func (o operations) RegisterIncidents(api huma.API) {
	huma.Register(api, ListIncidents, o.ListIncidents)
	huma.Register(api, CreateIncident, o.CreateIncident)
	huma.Register(api, GetIncident, o.GetIncident)
	huma.Register(api, UpdateIncident, o.UpdateIncident)
	huma.Register(api, ArchiveIncident, o.ArchiveIncident)
}

type (
	Incident struct {
		Id         uuid.UUID          `json:"id"`
		Attributes IncidentAttributes `json:"attributes"`
	}

	IncidentAttributes struct {
		Slug             string                   `json:"slug"`
		Title            string                   `json:"title"`
		Summary          string                   `json:"summary"`
		Private          bool                     `json:"private"`
		CurrentStatus    string                   `json:"current_status" enum:"started,mitigated,resolved,closed"`
		OpenedAt         time.Time                `json:"opened_at"`
		ClosedAt         time.Time                `json:"closed_at"`
		Severity         IncidentSeverity         `json:"severity"`
		Type             IncidentType             `json:"type"`
		SystemAnalysisId uuid.UUID                `json:"system_analysis_id"`
		Environments     []Environment            `json:"environments"`
		Tags             []IncidentTag            `json:"tags"`
		Ticket           ExternalTicket           `json:"ticket"`
		Tasks            []Task                   `json:"tasks"`
		RoleAssignments  []IncidentRoleAssignment `json:"roles"`
		TeamAssignments  []IncidentTeamAssignment `json:"teams"`
		LinkedIncidents  []IncidentLink           `json:"linked_incidents"`
		ChatChannel      IncidentChatChannel      `json:"chat_channel"`
	}

	IncidentLink struct {
		IncidentId      uuid.UUID        `json:"incident_id"`
		IncidentTitle   string           `json:"incident_title"`
		IncidentSummary string           `json:"incident_summary"`
		LinkType        IncidentLinkType `json:"link_type" enum:"duplicate_of,parent,sibling,child"`
	}
	IncidentLinkType string

	IncidentRoleAssignment struct {
		User      User         `json:"user"`
		Role      IncidentRole `json:"role"`
		Active    bool         `json:"active"`
		StartedAt time.Time    `json:"started_at"`
		EndedAt   time.Time    `json:"ended_at"`
	}

	IncidentTeamAssignment struct {
		Team      Team      `json:"team"`
		Active    bool      `json:"active"`
		StartedAt time.Time `json:"started_at"`
		EndedAt   time.Time `json:"ended_at"`
	}

	IncidentChatChannel struct {
		Provider IncidentChatChannelProvider `json:"provider" enum:"slack,ms_teams"`
		Id       string                      `json:"id"`
		Url      string                      `json:"url"`
		Private  bool                        `json:"private"`
	}
	IncidentChatChannelProvider string

	IncidentResponderImpact struct {
		Timezone        string `json:"timezone"`
		BusinessMinutes int    `json:"business_minutes"`
		PersonalMinutes int    `json:"personal_minutes"`
		SleepMinutes    int    `json:"sleep_minutes"`
	}
)

func IncidentFromEnt(inc *ent.Incident) Incident {
	attributes := IncidentAttributes{
		Slug:     inc.Slug,
		Title:    inc.Title,
		Summary:  inc.Summary,
		Private:  inc.Private,
		OpenedAt: inc.OpenedAt,
		ClosedAt: inc.ClosedAt,
	}

	if sev, sevErr := inc.Edges.SeverityOrErr(); sevErr == nil {
		attributes.Severity = IncidentSeverityFromEnt(sev)
	}

	if assns, rolesErr := inc.Edges.RoleAssignmentsOrErr(); rolesErr == nil {
		attributes.RoleAssignments = make([]IncidentRoleAssignment, len(assns))
		for i, assignment := range assns {
			attributes.RoleAssignments[i] = IncidentRoleAssignmentFromEnt(assignment)
		}
	}

	if teams, teamsErr := inc.Edges.TeamAssignmentsOrErr(); teamsErr == nil {
		attributes.TeamAssignments = make([]IncidentTeamAssignment, len(teams))
		for i, t := range teams {
			attributes.TeamAssignments[i] = IncidentTeamAssignmentFromEnt(t)
		}
	}

	attributes.Environments = make([]Environment, len(inc.Edges.Environments))
	for i, env := range inc.Edges.Environments {
		attributes.Environments[i] = EnvironmentFromEnt(env)
	}

	attributes.LinkedIncidents = make([]IncidentLink, 0)

	return Incident{
		Id:         inc.ID,
		Attributes: attributes,
	}
}

func IncidentRoleAssignmentFromEnt(assn *ent.IncidentRoleAssignment) IncidentRoleAssignment {
	return IncidentRoleAssignment{
		User:      UserFromEnt(assn.Edges.User),
		Role:      IncidentRoleFromEnt(assn.Edges.Role),
		Active:    false,
		StartedAt: time.Time{},
		EndedAt:   time.Time{},
	}
}

func IncidentTeamAssignmentFromEnt(assn *ent.IncidentTeamAssignment) IncidentTeamAssignment {
	return IncidentTeamAssignment{
		Team:      TeamFromEnt(assn.Edges.Team),
		Active:    false,
		StartedAt: time.Time{},
		EndedAt:   time.Time{},
	}
}

// Operations

var incidentsTags = []string{"Incidents"}

var ListIncidents = huma.Operation{
	OperationID: "list-incidents",
	Method:      http.MethodGet,
	Path:        "/incidents",
	Summary:     "List Incidents",
	Tags:        incidentsTags,
	Errors:      errorCodes(),
}

type ListIncidentsRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"team_id" required:"false"`
}
type ListIncidentsResponse PaginatedResponse[Incident]

var CreateIncident = huma.Operation{
	OperationID: "create-incident",
	Method:      http.MethodPost,
	Path:        "/incidents",
	Summary:     "Create an Incident",
	Tags:        incidentsTags,
	Errors:      errorCodes(),
}

type CreateIncidentAttributes struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
type CreateIncidentRequest RequestWithBodyAttributes[CreateIncidentAttributes]
type CreateIncidentResponse ItemResponse[Incident]

var GetIncident = huma.Operation{
	OperationID: "get-incident",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}",
	Summary:     "Get Incident",
	Tags:        incidentsTags,
	Errors:      errorCodes(),
}

type GetIncidentRequest = GetFlexibleIdRequest
type GetIncidentResponse ItemResponse[Incident]

var UpdateIncident = huma.Operation{
	OperationID: "update-incident",
	Method:      http.MethodPatch,
	Path:        "/incidents/{id}",
	Summary:     "Update an Incident",
	Tags:        incidentsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentAttributes struct {
	Title        *string   `json:"title,omitempty"`
	Summary      *string   `json:"summary,omitempty"`
	SeverityId   *string   `json:"severity_id,omitempty"`
	Private      *bool     `json:"private,omitempty"`
	Environments *[]string `json:"environments,omitempty"`
}
type UpdateIncidentRequest UpdateIdRequest[UpdateIncidentAttributes]
type UpdateIncidentResponse ItemResponse[Incident]

var ArchiveIncident = huma.Operation{
	OperationID: "archive-incident",
	Method:      http.MethodDelete,
	Path:        "/incidents/{id}",
	Summary:     "Archive an Incident",
	Tags:        incidentsTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentRequest ArchiveIdRequest
type ArchiveIncidentResponse EmptyResponse

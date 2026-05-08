package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
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
		Slug                   string                   `json:"slug"`
		Title                  string                   `json:"title"`
		Summary                string                   `json:"summary"`
		Private                bool                     `json:"private"`
		CurrentStatus          string                   `json:"currentStatus" enum:"started,mitigated,resolved,closed"`
		OpenedAt               time.Time                `json:"openedAt"`
		ClosedAt               time.Time                `json:"closedAt"`
		RetrospectiveId        *uuid.UUID               `json:"retrospectiveId,omitempty"`
		Severity               IncidentSeverity         `json:"severity"`
		Type                   IncidentType             `json:"type"`
		Tags                   []IncidentTag            `json:"tags"`
		Ticket                 *ExternalTicket          `json:"ticket,omitempty"`
		Tasks                  []Task                   `json:"tasks"`
		RoleAssignments        []IncidentRoleAssignment `json:"roles"`
		TeamAssignments        []IncidentTeamAssignment `json:"teams"`
		FieldSelections        []IncidentFieldSelection `json:"fieldSelections"`
		LinkedIncidents        []IncidentLink           `json:"linkedIncidents"`
		ChatChannel            IncidentChatChannel      `json:"chatChannel"`
		PrimaryVideoConference *VideoConference         `json:"primaryVideoConference,omitempty"`
	}

	IncidentLink struct {
		IncidentId      uuid.UUID        `json:"incidentId"`
		IncidentTitle   string           `json:"incidentTitle"`
		IncidentSummary string           `json:"incidentSummary"`
		LinkType        IncidentLinkType `json:"linkType" enum:"duplicate_of,parent,sibling,child"`
	}
	IncidentLinkType string

	IncidentRoleAssignment struct {
		User      User         `json:"user"`
		Role      IncidentRole `json:"role"`
		Active    bool         `json:"active"`
		StartedAt time.Time    `json:"startedAt"`
		EndedAt   time.Time    `json:"endedAt"`
	}

	IncidentTeamAssignment struct {
		Team      Team      `json:"team"`
		Active    bool      `json:"active"`
		StartedAt time.Time `json:"startedAt"`
		EndedAt   time.Time `json:"endedAt"`
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
		BusinessMinutes int    `json:"businessMinutes"`
		PersonalMinutes int    `json:"personalMinutes"`
		SleepMinutes    int    `json:"sleepMinutes"`
	}

	IncidentFieldSelection struct {
		FieldId   uuid.UUID           `json:"fieldId"`
		FieldName string              `json:"fieldName"`
		Option    IncidentFieldOption `json:"option"`
	}
)

func IncidentFromEnt(inc *ent.Incident) Incident {
	attr := IncidentAttributes{
		Slug:            inc.Slug,
		Title:           inc.Title,
		Summary:         inc.Summary,
		OpenedAt:        inc.OpenedAt,
		Tags:            make([]IncidentTag, 0),
		FieldSelections: make([]IncidentFieldSelection, 0),
		LinkedIncidents: make([]IncidentLink, 0),
	}

	if inc.Edges.Retrospective != nil {
		attr.RetrospectiveId = &inc.Edges.Retrospective.ID
	}

	if sev, sevErr := inc.Edges.SeverityOrErr(); sevErr == nil {
		attr.Severity = IncidentSeverityFromEnt(sev)
	}
	if t, typeErr := inc.Edges.TypeOrErr(); typeErr == nil {
		attr.Type = IncidentTypeFromEnt(t)
	}
	if tags, tagsErr := inc.Edges.TagAssignmentsOrErr(); tagsErr == nil {
		attr.Tags = make([]IncidentTag, len(tags))
		for i, tag := range tags {
			attr.Tags[i] = IncidentTagFromEnt(tag)
		}
	}
	if selections, selectionsErr := inc.Edges.FieldSelectionsOrErr(); selectionsErr == nil {
		attr.FieldSelections = make([]IncidentFieldSelection, len(selections))
		for i, selection := range selections {
			attr.FieldSelections[i] = IncidentFieldSelectionFromEnt(selection)
		}
	}

	if assns, rolesErr := inc.Edges.RoleAssignmentsOrErr(); rolesErr == nil {
		attr.RoleAssignments = make([]IncidentRoleAssignment, len(assns))
		for i, assignment := range assns {
			attr.RoleAssignments[i] = IncidentRoleAssignmentFromEnt(assignment)
		}
	}
	if primaryVc := inc.Edges.GetPrimaryVideoConference(); primaryVc != nil {
		primaryConf := VideoConferenceFromEnt(primaryVc)
		attr.PrimaryVideoConference = &primaryConf
	}

	attr.CurrentStatus = currentStatusFromEnt(inc)

	return Incident{Id: inc.ID, Attributes: attr}
}

func IncidentFieldSelectionFromEnt(opt *ent.IncidentFieldOption) IncidentFieldSelection {
	field := opt.Edges.IncidentField
	if field == nil {
		field = &ent.IncidentField{}
	}

	return IncidentFieldSelection{
		FieldId:   opt.IncidentFieldID,
		FieldName: field.Name,
		Option:    IncidentFieldOptionFromEnt(opt),
	}
}

func currentStatusFromEnt(inc *ent.Incident) string {
	status := "started"
	for _, milestone := range inc.Edges.Milestones {
		switch milestone.Kind {
		case im.KindMitigation:
			return "mitigated"
		case im.KindResolution:
			return "resolved"
		}
	}
	return status
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

// Operations

var incidentsTags = []string{"Incidents"}

var ListIncidents = huma.Operation{
	OperationID: "list-incidents",
	Method:      http.MethodGet,
	Path:        "/incidents",
	Summary:     "List Incidents",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type ListIncidentsRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"teamId" required:"false"`
}
type ListIncidentsResponse ListResponse[Incident]

var CreateIncident = huma.Operation{
	OperationID: "create-incident",
	Method:      http.MethodPost,
	Path:        "/incidents",
	Summary:     "Create an Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type CreateIncidentAttributes struct {
	Title             string      `json:"title"`
	Summary           *string     `json:"summary,omitempty" required:"false"`
	SeverityId        uuid.UUID   `json:"severityId"`
	TypeId            uuid.UUID   `json:"typeId"`
	TagIds            []uuid.UUID `json:"tagIds,omitempty"`
	FieldSelectionIds []uuid.UUID `json:"fieldSelectionIds,omitempty"`
}
type CreateIncidentRequest RequestWithBodyAttributes[CreateIncidentAttributes]
type CreateIncidentResponse ItemResponse[Incident]

var GetIncident = huma.Operation{
	OperationID: "get-incident",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}",
	Summary:     "Get Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type GetIncidentRequest = GetFlexibleIdRequest
type GetIncidentResponse ItemResponse[Incident]

var UpdateIncident = huma.Operation{
	OperationID: "update-incident",
	Method:      http.MethodPatch,
	Path:        "/incidents/{id}",
	Summary:     "Update an Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type UpdateIncidentAttributes struct {
	Title      *string   `json:"title,omitempty"`
	Summary    *string   `json:"summary,omitempty"`
	SeverityId uuid.UUID `json:"severityId,omitempty" required:"false"`
	TypeId     uuid.UUID `json:"typeId,omitempty" required:"false"`
}
type UpdateIncidentRequest UpdateIdRequest[UpdateIncidentAttributes]
type UpdateIncidentResponse ItemResponse[Incident]

var ArchiveIncident = huma.Operation{
	OperationID: "archive-incident",
	Method:      http.MethodDelete,
	Path:        "/incidents/{id}",
	Summary:     "Archive an Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentRequest ArchiveIdRequest
type ArchiveIncidentResponse EmptyResponse

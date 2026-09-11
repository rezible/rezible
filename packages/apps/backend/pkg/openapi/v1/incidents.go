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

	LinkIncidentSituation(context.Context, *LinkIncidentSituationRequest) (*LinkIncidentSituationResponse, error)
	UnlinkIncidentSituation(context.Context, *UnlinkIncidentSituationRequest) (*UnlinkIncidentSituationResponse, error)

	ListIncidentUpdates(context.Context, *ListIncidentUpdatesRequest) (*ListIncidentUpdatesResponse, error)
	CreateIncidentUpdate(context.Context, *CreateIncidentUpdateRequest) (*CreateIncidentUpdateResponse, error)
}

func (o operations) RegisterIncidents(api huma.API) {
	huma.Register(api, ListIncidents, o.ListIncidents)
	huma.Register(api, CreateIncident, o.CreateIncident)
	huma.Register(api, GetIncident, o.GetIncident)
	huma.Register(api, UpdateIncident, o.UpdateIncident)
	huma.Register(api, ArchiveIncident, o.ArchiveIncident)

	huma.Register(api, LinkIncidentSituation, o.LinkIncidentSituation)
	huma.Register(api, UnlinkIncidentSituation, o.UnlinkIncidentSituation)

	huma.Register(api, ListIncidentUpdates, o.ListIncidentUpdates)
	huma.Register(api, CreateIncidentUpdate, o.CreateIncidentUpdate)
}

type (
	Incident struct {
		Id         uuid.UUID          `json:"id"`
		Attributes IncidentAttributes `json:"attributes"`
	}

	IncidentAttributes struct {
		Title                  string                   `json:"title"`
		Summary                string                   `json:"summary"`
		Slug                   string                   `json:"slug"`
		CurrentStatus          string                   `json:"currentStatus" enum:"started,mitigated,resolved"`
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
		LinkedSituationIds     []uuid.UUID              `json:"linkedSituationIds"`
		RelatedTaskIds         []uuid.UUID              `json:"relatedTaskIds"`
		ChatChannel            IncidentChatChannel      `json:"chatChannel"`
		PrimaryVideoConference *VideoConference         `json:"primaryVideoConference,omitempty"`
	}

	IncidentLink struct {
		IncidentId      uuid.UUID        `json:"incidentId"`
		IncidentTitle   string           `json:"incidentTitle"`
		IncidentSummary string           `json:"incidentSummary"`
		LinkType        IncidentLinkType `json:"linkType" enum:"parent,child,similar"`
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
		Slug:               inc.Slug,
		Title:              inc.Title,
		Summary:            inc.Summary,
		OpenedAt:           inc.OpenedAt,
		Tags:               make([]IncidentTag, 0),
		FieldSelections:    make([]IncidentFieldSelection, 0),
		LinkedIncidents:    make([]IncidentLink, 0),
		LinkedSituationIds: make([]uuid.UUID, 0),
		RelatedTaskIds:     make([]uuid.UUID, 0),
	}
	for _, situation := range inc.Edges.Situations {
		attr.LinkedSituationIds = append(attr.LinkedSituationIds, situation.ID)
	}
	for _, task := range inc.Edges.Tasks {
		attr.RelatedTaskIds = append(attr.RelatedTaskIds, task.ID)
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
		attr.PrimaryVideoConference = new(VideoConferenceFromEnt(primaryVc))
	}

	status := im.KindOpened
	if latestMilestone := inc.Edges.GetLatestMilestone(); latestMilestone != nil {
		status = latestMilestone.Kind
	}
	attr.CurrentStatus = status.String()

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
	PaginationRequest
	Search     string    `query:"search" required:"false" nullable:"false"`
	Statuses   []string  `query:"statuses" required:"false" enum:"started,mitigated,resolved"`
	SeverityId uuid.UUID `query:"severityId" required:"false"`
}

type ListIncidentsResponse PaginatedResponse[Incident]

var CreateIncident = huma.Operation{
	OperationID: "create-incident",
	Method:      http.MethodPost,
	Path:        "/incidents",
	Summary:     "Create an Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type CreateIncidentAttributes struct {
	Title             string      `json:"title"`
	Summary           *string     `json:"summary,omitempty" required:"false"`
	SeverityId        uuid.UUID   `json:"severityId"`
	TypeId            uuid.UUID   `json:"typeId"`
	TagIds            []uuid.UUID `json:"tagIds,omitempty"`
	FieldSelectionIds []uuid.UUID `json:"fieldSelectionIds,omitempty"`
	SituationIds      []uuid.UUID `json:"situationIds,omitempty"`
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

type GetIncidentRequest = FlexibleIdRequest
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

type UpdateIncidentRequest IdRequestWithBody[UpdateIncidentAttributes]
type UpdateIncidentResponse ItemResponse[Incident]

var ArchiveIncident = huma.Operation{
	OperationID: "archive-incident",
	Method:      http.MethodDelete,
	Path:        "/incidents/{id}",
	Summary:     "Archive an Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentRequest IdRequest
type ArchiveIncidentResponse EmptyResponse

type IncidentSituationLink struct {
	Id         uuid.UUID                       `json:"id"`
	Attributes IncidentSituationLinkAttributes `json:"attributes"`
}

type IncidentSituationLinkAttributes struct {
	IncidentId  uuid.UUID `json:"incidentId"`
	SituationId uuid.UUID `json:"situationId"`
	CreatedAt   time.Time `json:"createdAt"`
}

type IncidentUpdate struct {
	Id         uuid.UUID                `json:"id"`
	Attributes IncidentUpdateAttributes `json:"attributes"`
}

type IncidentUpdateAttributes struct {
	IncidentId uuid.UUID  `json:"incidentId"`
	AuthorId   *uuid.UUID `json:"authorId,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	Body       string     `json:"body"`
}

var LinkIncidentSituation = huma.Operation{
	OperationID: "link-incident-situation",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/situations",
	Summary:     "Link Incident Situation",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type LinkIncidentSituationRequest struct {
	Id   uuid.UUID `path:"id"`
	Body struct {
		Attributes struct {
			SituationId uuid.UUID `json:"situationId"`
		} `json:"attributes"`
	}
}

type LinkIncidentSituationResponse ItemResponse[IncidentSituationLink]

var UnlinkIncidentSituation = huma.Operation{
	OperationID: "unlink-incident-situation",
	Method:      http.MethodDelete,
	Path:        "/incidents/{id}/situations",
	Summary:     "Unlink Incident Situation",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type UnlinkIncidentSituationRequest struct {
	Id          uuid.UUID `path:"id"`
	SituationId uuid.UUID `query:"situationId"`
}

type UnlinkIncidentSituationResponse ItemResponse[IncidentSituationLink]

var ListIncidentUpdates = huma.Operation{
	OperationID: "list-incident-updates",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/updates",
	Summary:     "List Incident Updates",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type ListIncidentUpdatesRequest struct {
	PaginationRequest
	Id uuid.UUID `path:"id"`
}

type ListIncidentUpdatesResponse PaginatedResponse[IncidentUpdate]

var CreateIncidentUpdate = huma.Operation{
	OperationID: "create-incident-update",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/updates",
	Summary:     "Create Incident Update",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type CreateIncidentUpdateRequest struct {
	Id   uuid.UUID `path:"id"`
	Body struct {
		Attributes struct {
			Body string `json:"body"`
		} `json:"attributes"`
	}
}

type CreateIncidentUpdateResponse ItemResponse[IncidentUpdate]

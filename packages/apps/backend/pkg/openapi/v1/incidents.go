package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
)

type IncidentsHandler interface {
	ListIncidents(context.Context, *ListIncidentsRequest) (*ListIncidentsResponse, error)
	CreateIncident(context.Context, *CreateIncidentRequest) (*CreateIncidentResponse, error)
	GetIncident(context.Context, *GetIncidentRequest) (*GetIncidentResponse, error)
	UpdateIncident(context.Context, *UpdateIncidentRequest) (*UpdateIncidentResponse, error)
	ArchiveIncident(context.Context, *ArchiveIncidentRequest) (*ArchiveIncidentResponse, error)

	ListIncidentMilestones(context.Context, *ListIncidentMilestonesRequest) (*ListIncidentMilestonesResponse, error)

	CreateIncidentRoleAssignment(context.Context, *CreateIncidentRoleAssignmentRequest) (*CreateIncidentRoleAssignmentResponse, error)
	DeleteIncidentRoleAssignment(context.Context, *DeleteIncidentRoleAssignmentRequest) (*DeleteIncidentRoleAssignmentResponse, error)

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

	huma.Register(api, ListIncidentMilestones, o.ListIncidentMilestones)

	huma.Register(api, CreateIncidentRoleAssignment, o.CreateIncidentRoleAssignment)
	huma.Register(api, DeleteIncidentRoleAssignment, o.DeleteIncidentRoleAssignment)

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
		Title         string                                  `json:"title"`
		Summary       string                                  `json:"summary"`
		Slug          string                                  `json:"slug"`
		ResponseState incident.ResponseState                  `json:"responseState"`
		OpenedAt      time.Time                               `json:"openedAt"`
		UpdatedAt     time.Time                               `json:"updatedAt"`
		ClosedAt      *time.Time                              `json:"closedAt"`
		ResolvedAt    *time.Time                              `json:"resolvedAt"`
		Milestones    []IncidentMilestone                     `json:"milestones"`
		Severity      *Expandable[IncidentSeverityAttributes] `json:"severity,omitempty"`
		Type          *Expandable[IncidentTypeAttributes]     `json:"type,omitempty"`

		Tags            []IncidentTag                `json:"tags"`
		Situations      []IncidentSituation          `json:"situations"`
		RoleAssignments []IncidentUserRoleAssignment `json:"roles"`
		FieldSelections []IncidentFieldSelection     `json:"fieldSelections"`

		LinkedTeams     []IncidentTeamLink `json:"teams"`
		LinkedIncidents []IncidentLink     `json:"linkedIncidents"`
		Impacts         []IncidentImpact   `json:"impacts"`
		Tasks           []Task             `json:"tasks"`

		Retrospective *Expandable[RetrospectiveAttributes] `json:"retrospective,omitempty"`

		ExternalTicket         *ExternalTicket      `json:"externalTicket,omitempty"`
		ChatChannel            *IncidentChatChannel `json:"chatChannel,omitempty"`
		PrimaryVideoConference *VideoConference     `json:"primaryVideoConference,omitempty"`
	}

	IncidentMilestone struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes IncidentMilestoneAttributes `json:"attributes"`
	}
	IncidentMilestoneAttributes struct {
		Kind        string                      `json:"kind" enum:"impact,detection,investigation,mitigation,resolution"`
		Description string                      `json:"description"`
		Timestamp   time.Time                   `json:"timestamp"`
		Source      string                      `json:"source"`
		User        *Expandable[UserAttributes] `json:"user"`
	}

	IncidentLink struct {
		IncidentId      uuid.UUID `json:"incidentId"`
		IncidentTitle   string    `json:"incidentTitle"`
		IncidentSummary string    `json:"incidentSummary"`
		LinkType        string    `json:"linkType" enum:"parent,child,similar"`
	}

	IncidentUserRoleAssignment struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes IncidentUserRoleAssignmentAttributes `json:"attributes"`
	}

	IncidentUserRoleAssignmentAttributes struct {
		User Expandable[UserAttributes]         `json:"user"`
		Role Expandable[IncidentRoleAttributes] `json:"role"`
	}

	IncidentTeamLink struct {
		Team Team   `json:"team"`
		Link string `json:"link"`
	}

	IncidentChatChannel struct {
		Provider string `json:"provider" enum:"slack,ms_teams"`
		Id       string `json:"id"`
		Url      string `json:"url"`
		Private  bool   `json:"private"`
	}

	IncidentFieldSelection struct {
		FieldId   uuid.UUID           `json:"fieldId"`
		FieldName string              `json:"fieldName"`
		Option    IncidentFieldOption `json:"option"`
	}

	IncidentSituation struct {
		Id      uuid.UUID `json:"id"`
		Title   string    `json:"title"`
		Summary string    `json:"summary"`
	}

	IncidentImpact struct {
		Id              uuid.UUID                                  `json:"id"`
		KnowledgeEntity Expandable[KnowledgeGraphEntityAttributes] `json:"knowledgeEntity"`
		Source          string                                     `json:"source"`
		Note            string                                     `json:"note"`
	}

	IncidentSituationLink struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes IncidentSituationLinkAttributes `json:"attributes"`
	}

	IncidentSituationLinkAttributes struct {
		IncidentId  uuid.UUID `json:"incidentId"`
		SituationId uuid.UUID `json:"situationId"`
		CreatedAt   time.Time `json:"createdAt"`
	}
)

func (o operations) RegisterIncidentEnums(api huma.API) {
	registerEnumAlias[incident.ResponseState, incidentResponseStateSchema](api)
}

type incidentResponseStateSchema incident.ResponseState

func (incidentResponseStateSchema) Schema(huma.Registry) *huma.Schema {
	return makeEnumStringSchema(incident.ResponseStateValues)
}

func IncidentFromEnt(inc *ent.Incident) Incident {
	attr := IncidentAttributes{
		Title:           inc.Title,
		Summary:         inc.Summary,
		Slug:            inc.Slug,
		ResponseState:   inc.ResponseState,
		OpenedAt:        inc.OpenedAt,
		UpdatedAt:       inc.UpdatedAt,
		ClosedAt:        inc.ResolvedAt,
		ResolvedAt:      inc.ResolvedAt,
		Milestones:      make([]IncidentMilestone, 0),
		Tags:            make([]IncidentTag, 0),
		Situations:      make([]IncidentSituation, 0),
		FieldSelections: make([]IncidentFieldSelection, 0),
		LinkedTeams:     make([]IncidentTeamLink, 0),
		LinkedIncidents: make([]IncidentLink, 0),
		Impacts:         make([]IncidentImpact, 0),
		Tasks:           make([]Task, 0),
	}

	if milestones, milestonesErr := inc.Edges.MilestonesOrErr(); milestonesErr == nil {
		attr.Milestones = ConvertSlice(milestones, IncidentMilestoneFromEnt)
	}

	if retro := inc.Edges.Retrospective; retro != nil {
		attr.Retrospective = &Expandable[RetrospectiveAttributes]{Id: retro.ID}
	}

	if inc.SeverityID != nil {
		attr.Severity = &Expandable[IncidentSeverityAttributes]{Id: *inc.SeverityID}
		if sev := inc.Edges.Severity; sev != nil {
			attr.Severity.Attributes = new(IncidentSeverityFromEnt(sev).Attributes)
		}
	}

	if inc.TypeID != nil {
		attr.Type = &Expandable[IncidentTypeAttributes]{Id: *inc.TypeID}
		if t := inc.Edges.Type; t != nil {
			attr.Type.Attributes = new(IncidentTypeFromEnt(t).Attributes)
		}
	}

	if sits, sitsErr := inc.Edges.SituationsOrErr(); sitsErr == nil {
		attr.Situations = make([]IncidentSituation, len(sits))
		for i, s := range sits {
			attr.Situations[i] = IncidentSituation{
				Id:      s.ID,
				Title:   s.Title,
				Summary: s.Summary,
			}
		}
	}

	if tags, tagsErr := inc.Edges.TagAssignmentsOrErr(); tagsErr == nil {
		attr.Tags = make([]IncidentTag, len(tags))
		for i, ta := range tags {
			attr.Tags[i] = IncidentTagFromEnt(ta)
		}
	}

	if selections, selectionsErr := inc.Edges.FieldSelectionsOrErr(); selectionsErr == nil {
		attr.FieldSelections = make([]IncidentFieldSelection, len(selections))
		for i, selection := range selections {
			attr.FieldSelections[i] = IncidentFieldSelectionFromEnt(selection)
		}
	}

	if roles, rolesErr := inc.Edges.RoleAssignmentsOrErr(); rolesErr == nil {
		attr.RoleAssignments = make([]IncidentUserRoleAssignment, len(roles))
		for i, assignment := range roles {
			attr.RoleAssignments[i] = IncidentUserRoleAssignmentFromEnt(assignment)
		}
	}
	if impacts, impactsErr := inc.Edges.ImpactsOrErr(); impactsErr == nil {
		attr.Impacts = ConvertSlice(impacts, IncidentImpactFromEnt)
	}
	if primaryVc := inc.Edges.GetPrimaryVideoConference(); primaryVc != nil {
		attr.PrimaryVideoConference = new(VideoConferenceFromEnt(primaryVc))
	}

	return Incident{Id: inc.ID, Attributes: attr}
}

func IncidentMilestoneFromEnt(m *ent.IncidentMilestone) IncidentMilestone {
	attrs := IncidentMilestoneAttributes{
		Kind:        m.Kind.String(),
		Description: m.Description,
		Timestamp:   m.Timestamp,
		Source:      m.Source,
	}
	if m.UserID != nil {
		attrs.User = &Expandable[UserAttributes]{Id: *m.UserID}
		if m.Edges.User != nil {
			attrs.User.Attributes = new(UserFromEnt(m.Edges.User).Attributes)
		}
	}
	return IncidentMilestone{Id: m.ID, Attributes: attrs}
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

func IncidentUserRoleAssignmentFromEnt(ra *ent.IncidentRoleAssignment) IncidentUserRoleAssignment {
	attrs := IncidentUserRoleAssignmentAttributes{
		Role: Expandable[IncidentRoleAttributes]{Id: ra.RoleID},
		User: Expandable[UserAttributes]{Id: ra.UserID},
	}
	if ra.Edges.Role != nil {
		attrs.Role.Attributes = new(IncidentRoleFromEnt(ra.Edges.Role).Attributes)
	}
	if ra.Edges.User != nil {
		attrs.User.Attributes = new(UserFromEnt(ra.Edges.User).Attributes)
	}
	return IncidentUserRoleAssignment{Id: ra.ID, Attributes: attrs}
}

func IncidentImpactFromEnt(impact *ent.IncidentImpact) IncidentImpact {
	result := IncidentImpact{
		Id:              impact.ID,
		KnowledgeEntity: Expandable[KnowledgeGraphEntityAttributes]{Id: impact.KnowledgeEntityID},
		Source:          impact.Source,
		Note:            impact.Note,
	}
	if kent := impact.Edges.KnowledgeEntity; kent != nil {
		result.KnowledgeEntity.Attributes = new(KnowledgeGraphEntityFromEnt(impact.Edges.KnowledgeEntity).Attributes)
	}
	return result
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
	Search         string                   `query:"search" required:"false" nullable:"false"`
	ResponseStates []incident.ResponseState `query:"responseStates" required:"false"`
	SeverityId     uuid.UUID                `query:"severityId" required:"false"`
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
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type UpdateIncidentRequest IdRequest
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

var ListIncidentMilestones = huma.Operation{
	OperationID: "list-incident-milestones",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/milestones",
	Summary:     "List Milestones for Incident",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type ListIncidentMilestonesRequest IdRequest
type ListIncidentMilestonesResponse CollectionResponse[IncidentMilestone]

var CreateIncidentRoleAssignment = huma.Operation{
	OperationID: "create-incident-role-assignment",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/role-assignments",
	Summary:     "Assign an Incident Role",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type CreateIncidentRoleAssignmentAttributes struct {
	RoleId uuid.UUID `json:"roleId"`
	UserId uuid.UUID `json:"userId"`
}

type CreateIncidentRoleAssignmentRequest IdRequestWithBody[CreateIncidentRoleAssignmentAttributes]
type CreateIncidentRoleAssignmentResponse ItemResponse[IncidentUserRoleAssignment]

var DeleteIncidentRoleAssignment = huma.Operation{
	OperationID: "delete-incident-role-assignment",
	Method:      http.MethodDelete,
	Path:        "/incident-role-assignments/{id}",
	Summary:     "Remove an Incident Role Assignment",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type DeleteIncidentRoleAssignmentRequest IdRequest
type DeleteIncidentRoleAssignmentResponse EmptyResponse

var LinkIncidentSituation = huma.Operation{
	OperationID: "link-incident-situation",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/situations",
	Summary:     "Link Incident Situation",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(http.StatusNotImplemented),
}

type LinkIncidentSituationRequestAttributes struct {
	SituationId uuid.UUID `json:"situationId"`
}

type LinkIncidentSituationRequest IdRequestWithBody[LinkIncidentSituationRequestAttributes]
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

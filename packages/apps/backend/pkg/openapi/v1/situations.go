package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationaction"
	"github.com/rezible/rezible/ent/situationhazardassessment"
	"github.com/rezible/rezible/pkg/openapi"
)

type SituationsHandler interface {
	ListSituations(context.Context, *ListSituationsRequest) (*ListSituationsResponse, error)
	GetSituation(context.Context, *GetSituationRequest) (*GetSituationResponse, error)
	GetSituationBrief(context.Context, *GetSituationBriefRequest) (*GetSituationBriefResponse, error)
	RaiseSituation(context.Context, *RaiseSituationRequest) (*RaiseSituationResponse, error)
	ListSituationJudgments(context.Context, *ListSituationJudgmentsRequest) (*ListSituationJudgmentsResponse, error)
	CloseSituation(context.Context, *CloseSituationRequest) (*CloseSituationResponse, error)
	MergeSituation(context.Context, *MergeSituationRequest) (*MergeSituationResponse, error)

	SetSituationMute(context.Context, *SetSituationMuteRequest) (*SetSituationMuteResponse, error)
	ClearSituationMute(context.Context, *ClearSituationMuteRequest) (*ClearSituationMuteResponse, error)

	SetSituationHold(context.Context, *SetSituationHoldRequest) (*SetSituationHoldResponse, error)
	ClearSituationHold(context.Context, *ClearSituationHoldRequest) (*ClearSituationHoldResponse, error)

	ListSituationHazardAssessments(context.Context, *ListSituationHazardAssessmentsRequest) (*ListSituationHazardAssessmentsResponse, error)
	AddSituationHazardAssessment(context.Context, *AddSituationHazardAssessmentRequest) (*AddSituationHazardAssessmentResponse, error)
}

func (o operations) RegisterSituations(api huma.API) {
	huma.Register(api, ListSituations, o.ListSituations)
	huma.Register(api, GetSituation, o.GetSituation)
	huma.Register(api, GetSituationBrief, o.GetSituationBrief)
	huma.Register(api, RaiseSituation, o.RaiseSituation)
	huma.Register(api, ListSituationJudgments, o.ListSituationJudgments)
	huma.Register(api, CloseSituation, o.CloseSituation)
	huma.Register(api, MergeSituation, o.MergeSituation)

	huma.Register(api, SetSituationMute, o.SetSituationMute)
	huma.Register(api, ClearSituationMute, o.ClearSituationMute)

	huma.Register(api, SetSituationHold, o.SetSituationHold)
	huma.Register(api, ClearSituationHold, o.ClearSituationHold)

	huma.Register(api, ListSituationHazardAssessments, o.ListSituationHazardAssessments)
	huma.Register(api, AddSituationHazardAssessment, o.AddSituationHazardAssessment)
}

type (
	Situation struct {
		Id         uuid.UUID           `json:"id"`
		Attributes SituationAttributes `json:"attributes"`
	}

	SituationAttributes struct {
		Title             string                      `json:"title"`
		Summary           string                      `json:"summary"`
		Stage             string                      `json:"stage" enum:"candidate,raised,closed"`
		SignalCount       int                         `json:"signalCount"`
		SeedEntityId      uuid.UUID                   `json:"seedEntityId"`
		OpenedAt          time.Time                   `json:"openedAt"`
		CreatedAt         time.Time                   `json:"createdAt" doc:"When watching began"`
		RaisedBy          *SituationRaisedBy          `json:"raisedBy,omitempty"`
		RaisedAt          *time.Time                  `json:"raisedAt,omitempty"`
		MutedAt           *time.Time                  `json:"mutedAt,omitempty"`
		MuteReason        *string                     `json:"muteReason,omitempty" enum:"not_noteworthy,expected"`
		HoldUntil         *time.Time                  `json:"holdUntil,omitempty"`
		ClosedAt          *time.Time                  `json:"closedAt,omitempty"`
		CloseReason       *string                     `json:"closeReason,omitempty" enum:"stabilized,expired,merged,dismissed"`
		UpdatedAt         time.Time                   `json:"updatedAt"`
		LinkedIncidentIds []uuid.UUID                 `json:"linkedIncidentIds"`
		Investigation     *SituationInvestigation     `json:"investigation,omitempty"`
		ObservationGroups []SituationObservationGroup `json:"observationGroups"`
		Links             []SituationLink             `json:"links"`
		Entities          []SituationEntity           `json:"entities"`
		LatestJudgment    *SituationJudgment          `json:"latestJudgment,omitempty"`
	}

	SituationRaisedBy struct {
		UserId *uuid.UUID `json:"userId,omitempty"`
		Reason string     `json:"reason"`
	}

	SituationJudgmentHistoryItem struct {
		Id         uuid.UUID                          `json:"id"`
		Attributes SituationJudgmentHistoryAttributes `json:"attributes"`
	}

	SituationJudgmentHistoryAttributes struct {
		SituationJudgment
		Facts schematypes.SituationFacts `json:"facts"`
	}

	SituationJudgment struct {
		JudgedAt     time.Time         `json:"judgedAt"`
		Outcome      string            `json:"outcome" enum:"no_reason,needs_decision,raise"`
		Decision     string            `json:"decision" enum:"hold,raise"`
		Reasons      []SituationReason `json:"reasons"`
		CitedReasons []string          `json:"citedReasons"`
		Explanation  string            `json:"explanation"`
		Judge        string            `json:"judge"`
	}

	SituationReason struct {
		Reason string `json:"reason" enum:"linked_incident,breadth,novelty,persistence,past_incident"`
		Met    bool   `json:"met"`
		Hard   bool   `json:"hard"`
		Detail string `json:"detail"`
	}

	SituationLink struct {
		Kind        string    `json:"kind" enum:"recurrence_of,merged_into"`
		SituationId uuid.UUID `json:"situationId"`
		Title       string    `json:"title"`
		Stage       string    `json:"stage" enum:"candidate,raised,closed"`
	}

	SituationEntity struct {
		Id          uuid.UUID `json:"id" doc:"The knowledge entity"`
		Matching    bool      `json:"matching" doc:"Whether it attracts related signals"`
		Category    string    `json:"category"`
		Kind        string    `json:"kind"`
		DisplayName string    `json:"displayName"`
	}

	SituationInvestigation struct {
		Investigation Expandable[InvestigationAttributes] `json:"investigation"`
	}

	SituationHazardAssessment struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes SituationHazardAssessmentAttrs `json:"attributes"`
	}

	SituationHazardAssessmentAttrs struct {
		SystemHazardId uuid.UUID  `json:"systemHazardId"`
		Revision       int        `json:"revision"`
		Status         string     `json:"status" enum:"suspected,confirmed,disproven"`
		Summary        string     `json:"summary"`
		AssessedAt     time.Time  `json:"assessedAt"`
		UserId         *uuid.UUID `json:"userId,omitempty"`
		AgentTurnId    *uuid.UUID `json:"agentTurnId,omitempty"`
	}

	SituationObservationGroup struct {
		Id         uuid.UUID                           `json:"id"`
		Attributes SituationObservationGroupAttributes `json:"attributes"`
	}

	SituationObservationGroupAttributes struct {
		SituationId uuid.UUID         `json:"situationId"`
		Title       string            `json:"title"`
		Body        string            `json:"body,omitempty"`
		Signals     []SituationSignal `json:"signals"`
	}

	SituationSignal struct {
		KnowledgeEntityId uuid.UUID  `json:"knowledgeEntityId"`
		Kind              string     `json:"kind"`
		SourceEntityId    *uuid.UUID `json:"sourceEntityId,omitempty"`
		AttachedAt        time.Time  `json:"attachedAt"`
		MatchKind         string     `json:"matchKind" enum:"seed,shared_entity,dependency,dependent,adjacent,manual"`
	}
)

// situationStage derives the stage from stored times.
func situationStage(s *ent.Situation) string {
	if s.ClosedAt != nil {
		return "closed"
	}
	if s.RaisedAt != nil {
		return "raised"
	}
	return "candidate"
}

func SituationObservationGroupFromEnt(g *ent.SituationObservationGroup) SituationObservationGroup {
	attrs := SituationObservationGroupAttributes{
		SituationId: g.SituationID,
		Title:       g.Title,
		Body:        g.Body,
		Signals:     ConvertSlice(g.Edges.Signals, SituationSignalFromEnt),
	}
	return SituationObservationGroup{Id: g.ID, Attributes: attrs}
}

func SituationSignalFromEnt(m *ent.SituationSignal) SituationSignal {
	return SituationSignal{
		KnowledgeEntityId: m.KnowledgeEntityID,
		Kind:              m.Kind,
		SourceEntityId:    m.SourceEntityID,
		AttachedAt:        m.AttachedAt,
		MatchKind:         m.MatchKind.String(),
	}
}

func SituationEntityFromEnt(e *ent.SituationEntity) SituationEntity {
	entity := SituationEntity{Id: e.KnowledgeEntityID, Matching: e.Matching}
	if knowledgeEntity := e.Edges.KnowledgeEntity; knowledgeEntity != nil {
		entity.Category = knowledgeEntity.Category.String()
		entity.Kind = knowledgeEntity.Kind
		entity.DisplayName = knowledgeEntity.State.DisplayName
	}
	return entity
}

func SituationLinkFromEnt(l *ent.SituationLink) SituationLink {
	link := SituationLink{Kind: l.Kind.String(), SituationId: l.LinkedSituationID}
	if linked := l.Edges.LinkedSituation; linked != nil {
		link.Title = linked.Title
		link.Stage = situationStage(linked)
	}
	return link
}

func SituationHazardAssessmentFromEnt(a *ent.SituationHazardAssessment) SituationHazardAssessment {
	return SituationHazardAssessment{
		Id: a.ID,
		Attributes: SituationHazardAssessmentAttrs{
			SystemHazardId: a.SystemHazardID,
			Revision:       a.Revision,
			Status:         string(a.Status),
			Summary:        a.Summary,
			AssessedAt:     a.AssessedAt,
			UserId:         a.UserID,
			AgentTurnId:    a.AgentTurnID,
		},
	}
}

func SituationFromEnt(s *ent.Situation) Situation {
	attrs := SituationAttributes{
		Title:             s.Title,
		Summary:           s.Summary,
		Stage:             situationStage(s),
		SeedEntityId:      s.SeedEntityID,
		OpenedAt:          s.OpenedAt,
		CreatedAt:         s.CreatedAt,
		RaisedAt:          s.RaisedAt,
		MutedAt:           s.MutedAt,
		HoldUntil:         s.HoldUntil,
		ClosedAt:          s.ClosedAt,
		UpdatedAt:         s.UpdatedAt,
		ObservationGroups: ConvertSlice(s.Edges.ObservationGroups, SituationObservationGroupFromEnt),
		Links:             ConvertSlice(s.Edges.Links, SituationLinkFromEnt),
		Entities:          ConvertSlice(s.Edges.Entities, SituationEntityFromEnt),
		LinkedIncidentIds: make([]uuid.UUID, len(s.Edges.Incidents)),
	}
	for _, group := range s.Edges.ObservationGroups {
		attrs.SignalCount += len(group.Edges.Signals)
	}
	for i, incident := range s.Edges.Incidents {
		attrs.LinkedIncidentIds[i] = incident.ID
	}
	if s.MuteReason != nil {
		attrs.MuteReason = new(s.MuteReason.String())
	}
	if s.CloseReason != nil {
		attrs.CloseReason = new(s.CloseReason.String())
	}
	if s.Edges.Investigation != nil {
		attrs.Investigation = new(SituationInvestigationFromEnt(s.Edges.Investigation))
	}
	if s.RaisedAt != nil {
		for _, action := range s.Edges.Actions {
			if action.Action == situationaction.ActionRaised {
				attrs.RaisedBy = &SituationRaisedBy{UserId: action.UserID, Reason: action.Reason}
				break
			}
		}
	}
	if s.Edges.LatestJudgment != nil {
		attrs.LatestJudgment = new(SituationJudgmentFromEnt(s.Edges.LatestJudgment))
	}
	return Situation{Id: s.ID, Attributes: attrs}
}

func SituationJudgmentFromEnt(j *ent.SituationJudgment) SituationJudgment {
	judgment := SituationJudgment{
		JudgedAt:     j.JudgedAt,
		Outcome:      j.Outcome.String(),
		Decision:     j.Decision.String(),
		Reasons:      make([]SituationReason, len(j.Reasons)),
		CitedReasons: make([]string, len(j.CitedReasons)),
		Explanation:  j.Explanation,
		Judge:        j.Judge,
	}
	for i, reason := range j.Reasons {
		judgment.Reasons[i] = SituationReason{
			Reason: string(reason.Reason),
			Met:    reason.Met,
			Hard:   reason.Hard,
			Detail: reason.Detail,
		}
	}
	for i, reason := range j.CitedReasons {
		judgment.CitedReasons[i] = string(reason)
	}
	return judgment
}

func SituationJudgmentHistoryItemFromEnt(j *ent.SituationJudgment) SituationJudgmentHistoryItem {
	attrs := SituationJudgmentHistoryAttributes{
		SituationJudgment: SituationJudgmentFromEnt(j),
		Facts:             j.Facts,
	}
	return SituationJudgmentHistoryItem{Id: j.ID, Attributes: attrs}
}

func SituationInvestigationFromEnt(si *ent.SituationInvestigation) SituationInvestigation {
	return SituationInvestigation{
		Investigation: AsExpandable[InvestigationAttributes](si.InvestigationID, nil),
	}
}

var situationsTags = []string{"Situations"}

var ListSituations = openapi.Operation{
	OperationID: "list-situations",
	Method:      http.MethodGet,
	Path:        "/situations",
	Summary:     "List Situations",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type ListSituationsRequest struct {
	PaginationRequest
	Search      string              `query:"search" required:"false" nullable:"false"`
	Stage       []string            `query:"stage" required:"false" enum:"candidate,raised,closed"`
	Muted       OptionalParam[bool] `query:"muted" required:"false"`
	OpenedAfter time.Time           `query:"openedAfter" required:"false" format:"date-time"`
}

type ListSituationsResponse PaginatedResponse[Situation]

var GetSituation = openapi.Operation{
	OperationID: "get-situation",
	Method:      http.MethodGet,
	Path:        "/situations/{id}",
	Summary:     "Get Situation",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type GetSituationRequest IdRequest
type GetSituationResponse ItemResponse[Situation]

// GetSituationBrief returns what Rezible states about a situation: its services, the changes before and after
// it opened, and what Rezible cannot see.
var GetSituationBrief = openapi.Operation{
	OperationID: "get-situation-brief",
	Method:      http.MethodGet,
	Path:        "/situations/{id}/brief",
	Summary:     "Get Situation Brief",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type GetSituationBriefRequest IdRequest
type GetSituationBriefResponse ItemResponse[SituationBrief]

type (
	SituationBrief struct {
		Services []SituationBriefService `json:"services" doc:"Matching services first, then by name, at most 10"`
		Changes  SituationBriefChanges   `json:"changes"`
		Unseen   []SituationBriefUnseen  `json:"unseen"`
	}

	SituationBriefService struct {
		EntityId           uuid.UUID                   `json:"entityId"`
		DisplayName        string                      `json:"displayName"`
		UnconfirmedMatches []SituationUnconfirmedMatch `json:"unconfirmedMatches"`
	}

	// SituationUnconfirmedMatch is a pending proposal that the service may be another one under another name.
	SituationUnconfirmedMatch struct {
		ProposalId  uuid.UUID `json:"proposalId"`
		Name        string    `json:"name"`
		EntityId    uuid.UUID `json:"entityId"`
		DisplayName string    `json:"displayName"`
		Rationale   string    `json:"rationale"`
	}

	SituationBriefChanges struct {
		Available         bool                           `json:"available" doc:"Whether any integration reports deployments"`
		UnavailableReason string                         `json:"unavailableReason"`
		Start             time.Time                      `json:"start"`
		OpenedAt          time.Time                      `json:"openedAt"`
		End               time.Time                      `json:"end"`
		Services          []SituationBriefServiceChanges `json:"services"`
	}

	SituationBriefServiceChanges struct {
		ServiceId   uuid.UUID                  `json:"serviceId"`
		Deployments []SituationBriefDeployment `json:"deployments" doc:"After the situation opened, then before it, each newest first"`
		Limits      []string                   `json:"limits"`
	}

	SituationBriefDeployment struct {
		Id               uuid.UUID  `json:"id" doc:"The deployment's knowledge entity"`
		DeployedAt       time.Time  `json:"deployedAt"`
		AfterOpened      bool       `json:"afterOpened"`
		Environment      string     `json:"environment"`
		Status           string     `json:"status"`
		Sha              string     `json:"sha"`
		Version          string     `json:"version"`
		Url              string     `json:"url"`
		LatestEvidenceId *uuid.UUID `json:"latestEvidenceId,omitempty"`
		RepositoryName   string     `json:"repositoryName"`
		// MergedChangesUnlinked says why merged changes cannot be linked; empty when they can.
		MergedChangesUnlinked string                       `json:"mergedChangesUnlinked" enum:",no_repository,repository_not_from_code_forge"`
		MergedChanges         []SituationBriefMergedChange `json:"mergedChanges"`
		MergedChangesCut      bool                         `json:"mergedChangesCut"`
	}

	SituationBriefMergedChange struct {
		Id       uuid.UUID `json:"id" doc:"The change's knowledge entity"`
		Number   int       `json:"number"`
		Title    string    `json:"title"`
		Author   string    `json:"author"`
		MergedAt time.Time `json:"mergedAt"`
		Url      string    `json:"url"`
		Link     string    `json:"link" enum:"merge_commit,merged_since_previous_deployment,merged_in_lookback"`
	}

	SituationBriefUnseen struct {
		Kind string `json:"kind" enum:"source_unavailable,services_cut,not_read"`
		Text string `json:"text"`
	}
)

func SituationBriefFromRez(brief *rez.SituationBrief) SituationBrief {
	converted := SituationBrief{
		Services: make([]SituationBriefService, len(brief.Services)),
		Changes: SituationBriefChanges{
			Available:         brief.Changes.Available,
			UnavailableReason: brief.Changes.UnavailableReason,
			Start:             brief.Changes.Start,
			OpenedAt:          brief.Changes.OpenedAt,
			End:               brief.Changes.End,
			Services:          make([]SituationBriefServiceChanges, len(brief.Changes.Services)),
		},
		Unseen: make([]SituationBriefUnseen, len(brief.Unseen)),
	}
	for index, service := range brief.Services {
		converted.Services[index] = situationBriefServiceFromRez(service)
	}
	for index, changes := range brief.Changes.Services {
		converted.Changes.Services[index] = situationBriefServiceChangesFromRez(changes)
	}
	for index, unseen := range brief.Unseen {
		converted.Unseen[index] = SituationBriefUnseen{Kind: string(unseen.Kind), Text: unseen.Text}
	}
	return converted
}

func situationBriefServiceFromRez(service rez.BriefService) SituationBriefService {
	converted := SituationBriefService{
		EntityId:           service.Entity.ID,
		DisplayName:        service.Entity.State.DisplayName,
		UnconfirmedMatches: make([]SituationUnconfirmedMatch, len(service.UnconfirmedMatches)),
	}
	for index, match := range service.UnconfirmedMatches {
		converted.UnconfirmedMatches[index] = SituationUnconfirmedMatch{
			ProposalId:  match.ProposalID,
			Name:        match.Name,
			EntityId:    match.Entity.ID,
			DisplayName: match.Entity.State.DisplayName,
			Rationale:   match.Rationale,
		}
	}
	return converted
}

func situationBriefServiceChangesFromRez(changes rez.SituationBriefServiceChanges) SituationBriefServiceChanges {
	converted := SituationBriefServiceChanges{
		ServiceId:   changes.ServiceEntityID,
		Deployments: make([]SituationBriefDeployment, len(changes.Deployments)),
		Limits:      changes.Limits,
	}
	if converted.Limits == nil {
		converted.Limits = []string{}
	}
	for index, deployment := range changes.Deployments {
		converted.Deployments[index] = situationBriefDeploymentFromRez(deployment)
	}
	return converted
}

func situationBriefDeploymentFromRez(deployment rez.SituationBriefDeployment) SituationBriefDeployment {
	converted := SituationBriefDeployment{
		Id:                    deployment.Deployment.ID,
		DeployedAt:            deployment.DeployedAt,
		AfterOpened:           deployment.AfterOpened,
		Environment:           deployment.Environment,
		Status:                deployment.Status,
		Sha:                   deployment.Sha,
		Version:               deployment.Version,
		Url:                   deployment.URL,
		LatestEvidenceId:      deployment.LatestEvidenceID,
		RepositoryName:        deployment.RepositoryName,
		MergedChangesUnlinked: string(deployment.MergedChangesUnlinked),
		MergedChanges:         make([]SituationBriefMergedChange, len(deployment.MergedChanges)),
		MergedChangesCut:      deployment.MergedChangesCut,
	}
	for index, change := range deployment.MergedChanges {
		converted.MergedChanges[index] = SituationBriefMergedChange{
			Id:       change.Change.ID,
			Number:   change.Number,
			Title:    change.Change.State.DisplayName,
			Author:   change.Author,
			MergedAt: change.MergedAt,
			Url:      change.URL,
			Link:     string(change.Link),
		}
	}
	return converted
}

var RaiseSituation = openapi.Operation{
	OperationID: "raise-situation",
	Method:      http.MethodPost,
	Path:        "/situations/{id}/raise",
	Summary:     "Raise Situation",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type RaiseSituationRequest IdRequest
type RaiseSituationResponse ItemResponse[Situation]

var ListSituationHazardAssessments = openapi.Operation{
	OperationID: "list-situation-hazard-assessments",
	Method:      http.MethodGet,
	Path:        "/situations/{id}/hazard_assessments",
	Summary:     "List Situation Hazard Assessments",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type ListSituationHazardAssessmentsRequest struct {
	IdRequest
	PaginationRequest
}

type ListSituationHazardAssessmentsResponse PaginatedResponse[SituationHazardAssessment]

var AddSituationHazardAssessment = openapi.Operation{
	OperationID: "add-situation-hazard-assessment",
	Method:      http.MethodPost,
	Path:        "/situations/{id}/hazard_assessments",
	Summary:     "Add Situation Hazard Assessment",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type AddSituationHazardAssessmentRequest struct {
	IdRequest
	Body struct {
		Attributes struct {
			SystemHazardId uuid.UUID                        `json:"systemHazardId"`
			Status         situationhazardassessment.Status `json:"status" enum:"suspected,confirmed,disproven"`
			Summary        string                           `json:"summary"`
		} `json:"attributes"`
	}
}

type AddSituationHazardAssessmentResponse ItemResponse[SituationHazardAssessment]

var SetSituationMute = openapi.Operation{
	OperationID: "set-situation-mute",
	Method:      http.MethodPut,
	Path:        "/situations/{id}/mute",
	Summary:     "Set Situation Mute",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type SetSituationMuteAttributes struct {
	Reason situation.MuteReason `json:"reason" enum:"not_noteworthy,expected"`
}
type SetSituationMuteRequest IdRequestWithBody[SetSituationMuteAttributes]
type SetSituationMuteResponse ItemResponse[Situation]

var ClearSituationMute = openapi.Operation{
	OperationID: "clear-situation-mute",
	Method:      http.MethodDelete,
	Path:        "/situations/{id}/mute",
	Summary:     "Clear Situation Mute",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type ClearSituationMuteRequest IdRequest
type ClearSituationMuteResponse ItemResponse[Situation]

var SetSituationHold = openapi.Operation{
	OperationID: "set-situation-hold",
	Method:      http.MethodPut,
	Path:        "/situations/{id}/hold",
	Summary:     "Set Situation Hold",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type SetSituationHoldAttributes struct {
	Until *time.Time `json:"until,omitempty"`
}
type SetSituationHoldRequest IdRequestWithBody[SetSituationHoldAttributes]
type SetSituationHoldResponse ItemResponse[Situation]

var ClearSituationHold = openapi.Operation{
	OperationID: "clear-situation-hold",
	Method:      http.MethodDelete,
	Path:        "/situations/{id}/hold",
	Summary:     "Clear Situation Hold",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type ClearSituationHoldRequest IdRequest
type ClearSituationHoldResponse ItemResponse[Situation]

var CloseSituation = openapi.Operation{
	OperationID: "close-situation",
	Method:      http.MethodPost,
	Path:        "/situations/{id}/close",
	Summary:     "Close Situation",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type CloseSituationAttributes struct {
	Note string `json:"note,omitempty"`
}
type CloseSituationRequest IdRequestWithBody[CloseSituationAttributes]
type CloseSituationResponse ItemResponse[Situation]

var MergeSituation = openapi.Operation{
	OperationID: "merge-situation",
	Method:      http.MethodPost,
	Path:        "/situations/{id}/merge",
	Summary:     "Merge Situation",
	Tags:        situationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type MergeSituationAttributes struct {
	TargetId    uuid.UUID `json:"targetId"`
	Explanation string    `json:"explanation,omitempty"`
}
type MergeSituationRequest IdRequestWithBody[MergeSituationAttributes]
type MergeSituationResponse ItemResponse[Situation]

var ListSituationJudgments = openapi.Operation{
	OperationID: "list-situation-judgments",
	Method:      http.MethodGet,
	Path:        "/situations/{id}/judgments",
	Summary:     "List Situation Judgments",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type ListSituationJudgmentsRequest PaginatedIdRequest
type ListSituationJudgmentsResponse PaginatedResponse[SituationJudgmentHistoryItem]

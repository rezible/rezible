package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
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

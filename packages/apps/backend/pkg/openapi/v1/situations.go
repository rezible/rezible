package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationhazardassessment"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/openapi"
)

type SituationsHandler interface {
	ListSituations(context.Context, *ListSituationsRequest) (*ListSituationsResponse, error)
	GetSituation(context.Context, *GetSituationRequest) (*GetSituationResponse, error)

	ListSituationHazardAssessments(context.Context, *ListSituationHazardAssessmentsRequest) (*ListSituationHazardAssessmentsResponse, error)
	AddSituationHazardAssessment(context.Context, *AddSituationHazardAssessmentRequest) (*AddSituationHazardAssessmentResponse, error)

	ListSituationInvestigations(context.Context, *ListSituationInvestigationsRequest) (*ListSituationInvestigationsResponse, error)
	GetSituationInvestigation(context.Context, *GetSituationInvestigationRequest) (*GetSituationInvestigationResponse, error)
	StartSituationInvestigation(context.Context, *StartSituationInvestigationRequest) (*StartSituationInvestigationResponse, error)
}

func (o operations) RegisterSituations(api huma.API) {
	huma.Register(api, ListSituations, o.ListSituations)
	huma.Register(api, GetSituation, o.GetSituation)

	huma.Register(api, ListSituationHazardAssessments, o.ListSituationHazardAssessments)
	huma.Register(api, AddSituationHazardAssessment, o.AddSituationHazardAssessment)

	huma.Register(api, ListSituationInvestigations, o.ListSituationInvestigations)
	huma.Register(api, GetSituationInvestigation, o.GetSituationInvestigation)
	huma.Register(api, StartSituationInvestigation, o.StartSituationInvestigation)
}

type (
	Situation struct {
		Id         uuid.UUID           `json:"id"`
		Attributes SituationAttributes `json:"attributes"`
	}

	SituationAttributes struct {
		Title             string                      `json:"title"`
		Summary           string                      `json:"summary"`
		Status            string                      `json:"status" enum:"open,closed"`
		CloseReason       *string                     `json:"closeReason,omitempty" enum:"stabilized,dismissed"`
		EvidenceRevision  int                         `json:"evidenceRevision"`
		KnowledgeEntityId uuid.UUID                   `json:"knowledgeEntityId"`
		LinkedIncidentIds []uuid.UUID                 `json:"linkedIncidentIds"`
		Investigations    []SituationInvestigation    `json:"investigations"`
		ObservationGroups []SituationObservationGroup `json:"observationGroups"`
		OpenedAt          time.Time                   `json:"openedAt"`
		ClosedAt          *time.Time                  `json:"closedAt,omitempty"`
		UpdatedAt         time.Time                   `json:"updatedAt"`
	}
	SituationInvestigationReport struct {
		Text               string   `json:"text"`
		LikelyCause        string   `json:"likelyCause,omitempty"`
		BestNextStep       string   `json:"bestNextStep,omitempty"`
		Limitations        []string `json:"limitations,omitempty"`
		RecommendedActions []string `json:"recommendedActions,omitempty"`
		SuggestedChecks    []string `json:"suggestedChecks,omitempty"`
	}

	SituationInvestigation struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes SituationInvestigationAttrs `json:"attributes"`
	}

	SituationInvestigationAttrs struct {
		Query             *string                       `json:"query,omitempty"`
		SituationId       uuid.UUID                     `json:"situationId"`
		AnalysisId        uuid.UUID                     `json:"analysisId"`
		SessionId         uuid.UUID                     `json:"sessionId"`
		RequestedRevision int                           `json:"requestedRevision"`
		CompletedRevision int                           `json:"completedRevision"`
		EvidenceRevision  int                           `json:"evidenceRevision"`
		Report            *SituationInvestigationReport `json:"report,omitempty"`
		UpdatedAt         time.Time                     `json:"updatedAt"`
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
		SituationId   uuid.UUID      `json:"situationId"`
		Title         string         `json:"title"`
		Body          string         `json:"body,omitempty"`
		Events        []Event        `json:"events"`
		AlertEpisodes []AlertEpisode `json:"alertEpisodes"`
	}
)

func SituationObservationGroupFromEnt(g *ent.SituationObservationGroup) SituationObservationGroup {
	attrs := SituationObservationGroupAttributes{
		SituationId:   g.SituationID,
		Title:         g.Title,
		Body:          g.Body,
		Events:        ConvertSlice(g.Edges.Events, EventFromEnt),
		AlertEpisodes: ConvertSlice(g.Edges.AlertEpisodes, AlertEpisodeFromEnt),
	}
	return SituationObservationGroup{Id: g.ID, Attributes: attrs}
}

// TODO: don't require evidenceRevision
func SituationInvestigationFromEnt(inv *ent.SituationInvestigation, evidenceRevision int) (*SituationInvestigation, error) {
	attrs := SituationInvestigationAttrs{
		SituationId:       inv.SituationID,
		AnalysisId:        inv.SystemAnalysisID,
		SessionId:         inv.AgentSessionID,
		RequestedRevision: inv.RequestedRevision,
		CompletedRevision: inv.CompletedRevision,
		EvidenceRevision:  evidenceRevision,
		UpdatedAt:         inv.UpdatedAt,
	}
	if session := inv.Edges.AgentSession; session != nil {
		var input rezai.InvestigationAgentSessionInput
		if decodeErr := json.Unmarshal(session.Input, &input); decodeErr != nil {
			return nil, fmt.Errorf("decode investigation %s input: %w", inv.ID, decodeErr)
		}
		attrs.Query = input.Query
	}
	if inv.Report != nil {
		attrs.Report = SituationInvestigationReportFromSchema(inv.Report)
	}
	return &SituationInvestigation{Id: inv.ID, Attributes: attrs}, nil
}

func SituationInvestigationReportFromSchema(report *schematypes.SituationInvestigationReport) *SituationInvestigationReport {
	if report == nil {
		return nil
	}
	return &SituationInvestigationReport{
		Text:               report.Text,
		LikelyCause:        report.LikelyCause,
		BestNextStep:       report.BestNextStep,
		Limitations:        report.Limitations,
		RecommendedActions: report.RecommendedActions,
		SuggestedChecks:    report.SuggestedChecks,
	}
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

func SituationFromEnt(s *ent.Situation) (Situation, error) {
	attrs := SituationAttributes{
		Title:             s.Title,
		Summary:           s.Summary,
		Status:            string(s.Status),
		EvidenceRevision:  s.EvidenceRevision,
		KnowledgeEntityId: s.KnowledgeEntityID,
		ObservationGroups: ConvertSlice(s.Edges.ObservationGroups, SituationObservationGroupFromEnt),
		OpenedAt:          s.OpenedAt,
		ClosedAt:          s.ClosedAt,
		UpdatedAt:         s.UpdatedAt,
	}
	attrs.LinkedIncidentIds = make([]uuid.UUID, len(s.Edges.Incidents))
	for i, incident := range s.Edges.Incidents {
		attrs.LinkedIncidentIds[i] = incident.ID
	}
	if s.CloseReason != nil {
		attrs.CloseReason = new(string(*s.CloseReason))
	}
	attrs.Investigations = make([]SituationInvestigation, 0, len(s.Edges.Investigations))
	for _, inv := range s.Edges.Investigations {
		investigation, conversionErr := SituationInvestigationFromEnt(inv, s.EvidenceRevision)
		if conversionErr != nil {
			return Situation{}, conversionErr
		}
		attrs.Investigations = append(attrs.Investigations, *investigation)
	}
	return Situation{Id: s.ID, Attributes: attrs}, nil
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
	Search      string           `query:"search" required:"false" nullable:"false"`
	Status      situation.Status `query:"status" required:"false" enum:"open,closed"`
	OpenedAfter time.Time        `query:"openedAfter" required:"false" format:"date-time"`
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

var ListSituationInvestigations = openapi.Operation{
	OperationID: "list-situation-investigations",
	Method:      http.MethodGet,
	Path:        "/situations/{id}/investigations",
	Summary:     "List Situation Investigations",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type ListSituationInvestigationsRequest struct {
	PaginationRequest
	Id uuid.UUID `path:"id"`
}
type ListSituationInvestigationsResponse PaginatedResponse[SituationInvestigation]

var GetSituationInvestigation = openapi.Operation{
	OperationID: "get-situation-investigation",
	Method:      http.MethodGet,
	Path:        "/situation-investigations/{id}",
	Summary:     "Get Situation Investigation",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type GetSituationInvestigationRequest IdRequest
type GetSituationInvestigationResponse ItemResponse[SituationInvestigation]

var StartSituationInvestigation = openapi.Operation{
	OperationID: "start-situation-investigation",
	Method:      http.MethodPost,
	Path:        "/situations/{id}/investigations",
	Summary:     "Start Situation Investigation",
	Tags:        situationsTags,
	Errors:      ErrorCodes(),
}

type StartSituationInvestigationAttributes struct {
	Query *string `json:"query,omitempty"`
}
type StartSituationInvestigationRequest struct {
	Id uuid.UUID `path:"id"`
	RequestWithBodyAttributes[StartSituationInvestigationAttributes]
}
type StartSituationInvestigationResponse ItemResponse[SituationInvestigation]

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

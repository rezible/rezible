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
	"github.com/rezible/rezible/ent/situationhazardassessment"
	"github.com/rezible/rezible/pkg/openapi"
)

type SituationsHandler interface {
	ListSituations(context.Context, *ListSituationsRequest) (*ListSituationsResponse, error)
	GetSituation(context.Context, *GetSituationRequest) (*GetSituationResponse, error)

	ListSituationHazardAssessments(context.Context, *ListSituationHazardAssessmentsRequest) (*ListSituationHazardAssessmentsResponse, error)
	AddSituationHazardAssessment(context.Context, *AddSituationHazardAssessmentRequest) (*AddSituationHazardAssessmentResponse, error)
}

func (o operations) RegisterSituations(api huma.API) {
	huma.Register(api, ListSituations, o.ListSituations)
	huma.Register(api, GetSituation, o.GetSituation)

	huma.Register(api, ListSituationHazardAssessments, o.ListSituationHazardAssessments)
	huma.Register(api, AddSituationHazardAssessment, o.AddSituationHazardAssessment)
}

type (
	SituationInvestigationReport struct {
		Text               string   `json:"text"`
		LikelyCause        string   `json:"likelyCause,omitempty"`
		BestNextStep       string   `json:"bestNextStep,omitempty"`
		Limitations        []string `json:"limitations,omitempty"`
		RecommendedActions []string `json:"recommendedActions,omitempty"`
		SuggestedChecks    []string `json:"suggestedChecks,omitempty"`
	}

	SituationInvestigation struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes SituationInvestigationAttrs `json:"attributes"`
	}

	SituationInvestigationAttrs struct {
		RequestedRevision int                            `json:"requestedRevision"`
		CompletedRevision int                            `json:"completedRevision"`
		EvidenceRevision  int                            `json:"evidenceRevision"`
		Report            *SituationInvestigationReport  `json:"report,omitempty"`
		UpdatedAt         time.Time                      `json:"updatedAt"`
	}

	SituationAlertEpisode struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes SituationAlertEpisodeAttrs   `json:"attributes"`
	}

	SituationAlertEpisodeAttrs struct {
		Status          string                       `json:"status" enum:"open,closed"`
		AlertDefinition *SituationAlertEpisodeAlert  `json:"alertDefinition,omitempty"`
		StartedAt       time.Time                    `json:"startedAt"`
		LastObservedAt  time.Time                    `json:"lastObservedAt"`
		ClosedAt        *time.Time                   `json:"closedAt,omitempty"`
	}

	SituationAlertEpisodeAlert struct {
		Id    uuid.UUID `json:"id"`
		Title string    `json:"title"`
	}

	SituationHazardAssessment struct {
		Id         uuid.UUID                           `json:"id"`
		Attributes SituationHazardAssessmentAttrs     `json:"attributes"`
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

	Situation struct {
		Id         uuid.UUID            `json:"id"`
		Attributes SituationAttributes `json:"attributes"`
	}

	SituationAttributes struct {
		Title             string                      `json:"title"`
		Summary           string                      `json:"summary"`
		Status            string                      `json:"status" enum:"open,closed"`
		CloseReason       *string                     `json:"closeReason,omitempty" enum:"stabilized,dismissed"`
		EvidenceRevision  int                         `json:"evidenceRevision"`
		KnowledgeEntityId uuid.UUID                   `json:"knowledgeEntityId"`
		AlertEpisodes     []SituationAlertEpisode     `json:"alertEpisodes"`
		Investigation     *SituationInvestigation     `json:"investigation,omitempty"`
		OpenedAt          time.Time                   `json:"openedAt"`
		ClosedAt          *time.Time                  `json:"closedAt,omitempty"`
		UpdatedAt         time.Time                   `json:"updatedAt"`
	}
)

func SituationInvestigationReportFromSchema(report schematypes.SituationInvestigationReport) *SituationInvestigationReport {
	return &SituationInvestigationReport{
		Text:               report.Text,
		LikelyCause:        report.LikelyCause,
		BestNextStep:       report.BestNextStep,
		Limitations:        report.Limitations,
		RecommendedActions: report.RecommendedActions,
		SuggestedChecks:    report.SuggestedChecks,
	}
}

func SituationInvestigationFromEnt(inv *ent.SituationInvestigation, evidenceRevision int) *SituationInvestigation {
	attrs := SituationInvestigationAttrs{
		RequestedRevision: inv.RequestedRevision,
		CompletedRevision: inv.CompletedRevision,
		EvidenceRevision:  evidenceRevision,
		UpdatedAt:         inv.UpdatedAt,
	}
	if inv.Report.Text != "" || len(inv.Report.Limitations) > 0 {
		attrs.Report = SituationInvestigationReportFromSchema(inv.Report)
	}
	return &SituationInvestigation{Id: inv.ID, Attributes: attrs}
}

func SituationAlertEpisodeFromEnt(ep *ent.AlertEpisode) SituationAlertEpisode {
	attrs := SituationAlertEpisodeAttrs{
		Status:         string(ep.Status),
		StartedAt:      ep.StartedAt,
		LastObservedAt: ep.LastObservedAt,
		ClosedAt:       ep.ClosedAt,
	}
	if def := ep.Edges.AlertDefinition; def != nil {
		attrs.AlertDefinition = &SituationAlertEpisodeAlert{Id: def.ID, Title: def.Title}
	}
	return SituationAlertEpisode{Id: ep.ID, Attributes: attrs}
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
		Status:            string(s.Status),
		EvidenceRevision:  s.EvidenceRevision,
		KnowledgeEntityId: s.KnowledgeEntityID,
		AlertEpisodes:     make([]SituationAlertEpisode, len(s.Edges.AlertEpisodes)),
		OpenedAt:          s.OpenedAt,
		ClosedAt:          s.ClosedAt,
		UpdatedAt:         s.UpdatedAt,
	}
	if s.CloseReason != nil {
		reason := string(*s.CloseReason)
		attrs.CloseReason = &reason
	}
	for i, ep := range s.Edges.AlertEpisodes {
		attrs.AlertEpisodes[i] = SituationAlertEpisodeFromEnt(ep)
	}
	if inv := s.Edges.Investigation; inv != nil {
		attrs.Investigation = SituationInvestigationFromEnt(inv, s.EvidenceRevision)
	}
	return Situation{Id: s.ID, Attributes: attrs}
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
	Search      string    `query:"search" required:"false" nullable:"false"`
	Status      situation.Status `query:"status" required:"false" enum:"open,closed"`
	OpenedAfter time.Time `query:"openedAfter" required:"false" format:"date-time"`
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
			SystemHazardId uuid.UUID                              `json:"systemHazardId"`
			Status         situationhazardassessment.Status       `json:"status" enum:"suspected,confirmed,disproven"`
			Summary        string                                 `json:"summary"`
		} `json:"attributes"`
	}
}
type AddSituationHazardAssessmentResponse ItemResponse[SituationHazardAssessment]

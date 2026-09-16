package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/openapi"
)

type InvestigationsHandler interface {
	GetInvestigation(context.Context, *GetInvestigationRequest) (*GetInvestigationResponse, error)
}

func (o operations) RegisterInvestigations(api huma.API) {
	huma.Register(api, GetInvestigation, o.GetInvestigation)
}

type (
	Investigation struct {
		Id         uuid.UUID               `json:"id"`
		Attributes InvestigationAttributes `json:"attributes"`
	}

	InvestigationAttributes struct {
		Query      *string              `json:"query,omitempty"`
		AnalysisId uuid.UUID            `json:"analysisId"`
		SessionId  uuid.UUID            `json:"sessionId"`
		Report     *InvestigationReport `json:"report,omitempty"`
		CreatedAt  time.Time            `json:"createdAt"`
		UpdatedAt  time.Time            `json:"updatedAt"`
	}

	InvestigationReport struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes InvestigationReportAttributes `json:"attributes"`
	}

	InvestigationReportAttributes struct {
		Text               string    `json:"text"`
		LikelyCause        *string   `json:"likelyCause,omitempty"`
		BestNextStep       *string   `json:"bestNextStep,omitempty"`
		Limitations        []string  `json:"limitations,omitempty"`
		RecommendedActions []string  `json:"recommendedActions,omitempty"`
		SuggestedChecks    []string  `json:"suggestedChecks,omitempty"`
		CreatedAt          time.Time `json:"createdAt"`
	}
)

func InvestigationFromEnt(inv *ent.Investigation) Investigation {
	attrs := InvestigationAttributes{
		AnalysisId: inv.SystemAnalysisID,
		SessionId:  inv.AgentSessionID,
		CreatedAt:  inv.CreatedAt,
		UpdatedAt:  inv.UpdatedAt,
	}
	if session := inv.Edges.AgentSession; session != nil {
		var input rezai.InvestigationAgentSessionInput
		if decodeErr := json.Unmarshal(session.Input, &input); decodeErr != nil {
			//return nil, fmt.Errorf("decode investigation %s input: %w", inv.ID, decodeErr)
		}
		attrs.Query = input.Query
	}
	if inv.Edges.Report != nil {
		attrs.Report = new(InvestigationReportFromEnt(inv.Edges.Report))
	}
	return Investigation{Id: inv.ID, Attributes: attrs}
}

func InvestigationReportFromEnt(report *ent.InvestigationReport) InvestigationReport {
	reportAttrs := InvestigationReportAttributes{
		Text:               report.Text,
		Limitations:        report.Limitations,
		RecommendedActions: report.RecommendedActions,
		SuggestedChecks:    report.SuggestedChecks,
		CreatedAt:          report.CreatedAt,
	}
	if report.LikelyCause != "" {
		reportAttrs.LikelyCause = &report.LikelyCause
	}
	if report.BestNextStep != "" {
		reportAttrs.BestNextStep = &report.BestNextStep
	}
	return InvestigationReport{Id: report.ID, Attributes: reportAttrs}
}

var investigationsTags = []string{"Investigations"}

var GetInvestigation = openapi.Operation{
	OperationID: "get-investigation",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}",
	Summary:     "Get Investigation",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type GetInvestigationRequest IdRequest
type GetInvestigationResponse ItemResponse[Investigation]

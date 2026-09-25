package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/pkg/openapi"
)

type InvestigationsHandler interface {
	GetInvestigation(context.Context, *GetInvestigationRequest) (*GetInvestigationResponse, error)

	SubmitInvestigationUserInput(context.Context, *SubmitInvestigationUserInputRequest) (*SubmitInvestigationUserInputResponse, error)
	ListInvestigationUserInputs(context.Context, *ListInvestigationUserInputsRequest) (*ListInvestigationUserInputsResponse, error)
	ListInvestigationEvidenceRevisions(context.Context, *ListInvestigationEvidenceRevisionsRequest) (*ListInvestigationEvidenceRevisionsResponse, error)

	ListInvestigationFindings(context.Context, *ListInvestigationFindingsRequest) (*ListInvestigationFindingsResponse, error)
	GetInvestigationFinding(context.Context, *GetInvestigationFindingRequest) (*GetInvestigationFindingResponse, error)

	ListInvestigationHypotheses(context.Context, *ListInvestigationHypothesesRequest) (*ListInvestigationHypothesesResponse, error)
	GetInvestigationHypothesis(context.Context, *GetInvestigationHypothesisRequest) (*GetInvestigationHypothesisResponse, error)

	GetInvestigationReport(context.Context, *GetInvestigationReportRequest) (*GetInvestigationReportResponse, error)
}

func (o operations) RegisterInvestigations(api huma.API) {
	huma.Register(api, GetInvestigation, o.GetInvestigation)

	huma.Register(api, SubmitInvestigationUserInput, o.SubmitInvestigationUserInput)
	huma.Register(api, ListInvestigationUserInputs, o.ListInvestigationUserInputs)
	huma.Register(api, ListInvestigationEvidenceRevisions, o.ListInvestigationEvidenceRevisions)

	huma.Register(api, ListInvestigationFindings, o.ListInvestigationFindings)
	huma.Register(api, GetInvestigationFinding, o.GetInvestigationFinding)

	huma.Register(api, ListInvestigationHypotheses, o.ListInvestigationHypotheses)
	huma.Register(api, GetInvestigationHypothesis, o.GetInvestigationHypothesis)

	huma.Register(api, GetInvestigationReport, o.GetInvestigationReport)
}

type (
	Investigation struct {
		Id         uuid.UUID               `json:"id"`
		Attributes InvestigationAttributes `json:"attributes"`
	}

	InvestigationAttributes struct {
		Query            string     `json:"query"`
		AnalysisId       uuid.UUID  `json:"analysisId"`
		SessionId        uuid.UUID  `json:"sessionId"`
		LatestTurnId     *uuid.UUID `json:"latestTurnId" nullable:"true"`
		LatestTurnStatus *string    `json:"latestTurnStatus" nullable:"true"`
		CreatedAt        time.Time  `json:"createdAt"`
		UpdatedAt        time.Time  `json:"updatedAt"`
	}

	InvestigationReport struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes InvestigationReportAttributes `json:"attributes"`
	}

	InvestigationReportAttributes struct {
		Text        string                   `json:"text"`
		References  []InvestigationReference `json:"references"`
		AgentTurnId uuid.UUID                `json:"agentTurnId"`
		TurnStatus  string                   `json:"turnStatus"`
		Provisional bool                     `json:"provisional"`
		CreatedAt   time.Time                `json:"createdAt"`
	}

	InvestigationFinding struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes InvestigationFindingAttributes `json:"attributes"`
	}

	InvestigationFindingAttributes struct {
		FindingId               uuid.UUID                 `json:"findingId"`
		Key                     string                    `json:"key"`
		UserInputId             *uuid.UUID                `json:"userInputId" nullable:"true"`
		Title                   string                    `json:"title"`
		Body                    string                    `json:"body"`
		References              []InvestigationReference  `json:"references"`
		FindingReferences       []FindingVersionReference `json:"findingReferences"`
		InvalidatedByVersionIds []uuid.UUID               `json:"invalidatedByVersionIds"`
		AgentTurnId             uuid.UUID                 `json:"agentTurnId"`
		TurnStatus              string                    `json:"turnStatus"`
		Provisional             bool                      `json:"provisional"`
		CreatedAt               time.Time                 `json:"createdAt"`
	}

	InvestigationHypothesis struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes InvestigationHypothesisAttributes `json:"attributes"`
	}

	InvestigationHypothesisAttributes struct {
		HypothesisId  uuid.UUID                `json:"hypothesisId"`
		Key           string                   `json:"key"`
		Title         string                   `json:"title"`
		Justification string                   `json:"justification"`
		Status        string                   `json:"status"`
		References    []InvestigationReference `json:"references"`
		AgentTurnId   uuid.UUID                `json:"agentTurnId"`
		TurnStatus    string                   `json:"turnStatus"`
		Provisional   bool                     `json:"provisional"`
		CreatedAt     time.Time                `json:"createdAt"`
	}

	InvestigationReference struct {
		Id   uuid.UUID `json:"id"`
		Kind string    `json:"kind" enum:"knowledge_evidence"`
	}

	FindingVersionReference struct {
		VersionId uuid.UUID `json:"versionId"`
		Relation  string    `json:"relation"`
	}

	InvestigationUserInput struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes InvestigationUserInputAttributes `json:"attributes"`
	}

	InvestigationUserInputAttributes struct {
		Text            string     `json:"text"`
		UserId          uuid.UUID  `json:"userId"`
		SubmissionKey   string     `json:"submissionKey"`
		CreatedAt       time.Time  `json:"createdAt"`
		AgentTurnId     *uuid.UUID `json:"agentTurnId" nullable:"true"`
		AnswerVersionId *uuid.UUID `json:"answerVersionId" nullable:"true"`
	}

	InvestigationEvidenceRevision struct {
		Id         uuid.UUID                               `json:"id"`
		Attributes InvestigationEvidenceRevisionAttributes `json:"attributes"`
	}

	InvestigationEvidenceRevisionAttributes struct {
		Explanation string     `json:"explanation"`
		CallerKey   string     `json:"callerKey"`
		AgentTurnId *uuid.UUID `json:"agentTurnId" nullable:"true"`
		CreatedAt   time.Time  `json:"createdAt"`
	}
)

func InvestigationFromDetail(detail *rez.InvestigationDetail) Investigation {
	inv := detail.Investigation
	var latestTurnID *uuid.UUID
	var latestTurnStatus *string
	if detail.LatestTurn != nil {
		latestTurnID = new(detail.LatestTurn.ID)
		latestTurnStatusValue := string(detail.LatestTurn.Status)
		latestTurnStatus = &latestTurnStatusValue
	}
	attrs := InvestigationAttributes{
		Query:            detail.Query,
		AnalysisId:       inv.SystemAnalysisID,
		SessionId:        inv.AgentSessionID,
		LatestTurnId:     latestTurnID,
		LatestTurnStatus: latestTurnStatus,
		CreatedAt:        inv.CreatedAt,
		UpdatedAt:        inv.UpdatedAt,
	}
	return Investigation{Id: inv.ID, Attributes: attrs}
}

func InvestigationReportFromResult(report *rez.InvestigationReportResult) InvestigationReport {
	attrs := InvestigationReportAttributes{
		Text:        report.Text,
		References:  ConvertSlice(report.EvidenceIDs, InvestigationReferenceFromEvidenceID),
		AgentTurnId: report.AgentTurnID,
		TurnStatus:  string(report.TurnStatus),
		Provisional: report.TurnStatus == agentturn.StatusRunning,
		CreatedAt:   report.CreatedAt,
	}
	return InvestigationReport{Id: report.ID, Attributes: attrs}
}

func InvestigationFindingFromResult(finding *rez.InvestigationFindingVersion) InvestigationFinding {
	attrs := InvestigationFindingAttributes{
		FindingId:               finding.FindingID,
		Key:                     finding.Key,
		UserInputId:             finding.UserInputID,
		Title:                   finding.Title,
		Body:                    finding.Body,
		References:              ConvertSlice(finding.EvidenceIDs, InvestigationReferenceFromEvidenceID),
		FindingReferences:       ConvertSlice(finding.FindingReferences, FindingVersionReferenceFromRez),
		InvalidatedByVersionIds: append([]uuid.UUID{}, finding.InvalidatedByVersionIDs...),
		AgentTurnId:             finding.AgentTurnID,
		TurnStatus:              string(finding.TurnStatus),
		Provisional:             finding.TurnStatus == agentturn.StatusRunning,
		CreatedAt:               finding.CreatedAt,
	}
	return InvestigationFinding{Id: finding.ID, Attributes: attrs}
}

func InvestigationHypothesisFromResult(hypo *rez.InvestigationHypothesisVersion) InvestigationHypothesis {
	attrs := InvestigationHypothesisAttributes{
		HypothesisId:  hypo.HypothesisID,
		Key:           hypo.Key,
		Title:         hypo.Title,
		Justification: hypo.Justification,
		Status:        string(hypo.Status),
		References:    ConvertSlice(hypo.EvidenceIDs, InvestigationReferenceFromEvidenceID),
		AgentTurnId:   hypo.AgentTurnID,
		TurnStatus:    string(hypo.TurnStatus),
		Provisional:   hypo.TurnStatus == agentturn.StatusRunning,
		CreatedAt:     hypo.CreatedAt,
	}
	return InvestigationHypothesis{Id: hypo.ID, Attributes: attrs}
}

func InvestigationReferenceFromEvidenceID(evidenceID uuid.UUID) InvestigationReference {
	return InvestigationReference{Kind: "knowledge_evidence", Id: evidenceID}
}

func FindingVersionReferenceFromRez(reference rez.FindingVersionReference) FindingVersionReference {
	return FindingVersionReference{VersionId: reference.VersionID, Relation: string(reference.Relation)}
}

func InvestigationUserInputFromRez(input *rez.InvestigationUserInput) InvestigationUserInput {
	attrs := InvestigationUserInputAttributes{
		Text:            input.Text,
		UserId:          input.UserID,
		SubmissionKey:   input.SubmissionKey,
		CreatedAt:       input.CreatedAt,
		AgentTurnId:     input.AgentTurnID,
		AnswerVersionId: input.AnswerVersionID,
	}
	return InvestigationUserInput{Id: input.ID, Attributes: attrs}
}

func InvestigationEvidenceRevisionFromEnt(revision *ent.InvestigationEvidenceRevision) InvestigationEvidenceRevision {
	attrs := InvestigationEvidenceRevisionAttributes{
		Explanation: revision.Explanation,
		CallerKey:   revision.Key,
		AgentTurnId: revision.AgentTurnID,
		CreatedAt:   revision.CreatedAt,
	}
	return InvestigationEvidenceRevision{Id: revision.ID, Attributes: attrs}
}

var investigationsTags = []string{"Investigations"}

var GetInvestigation = openapi.Operation{OperationID: "get-investigation", Method: http.MethodGet, Path: "/investigations/{id}", Summary: "Get Investigation", Tags: investigationsTags, Errors: ErrorCodes()}

type GetInvestigationRequest IdRequest
type GetInvestigationResponse ItemResponse[Investigation]

var GetInvestigationReport = openapi.Operation{OperationID: "get-investigation-report", Method: http.MethodGet, Path: "/investigations/{id}/report", Summary: "Get Investigation Report", Tags: investigationsTags, Errors: ErrorCodes()}

type GetInvestigationReportRequest struct {
	Id        uuid.UUID `path:"id"`
	Selection string    `query:"selection" enum:"latest,completed" default:"latest" required:"false"`
}
type GetInvestigationReportResponse ItemResponse[InvestigationReport]

var ListInvestigationFindings = openapi.Operation{OperationID: "list-investigation-findings", Method: http.MethodGet, Path: "/investigations/{id}/findings", Summary: "List Investigation Findings", Tags: investigationsTags, Errors: ErrorCodes()}

type ListInvestigationFindingsRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationFindingsResponse PaginatedResponse[InvestigationFinding]

var GetInvestigationFinding = openapi.Operation{OperationID: "get-investigation-finding", Method: http.MethodGet, Path: "/investigations/{id}/findings/{versionId}", Summary: "Get Investigation Finding", Tags: investigationsTags, Errors: ErrorCodes()}

type GetInvestigationFindingRequest struct {
	Id        uuid.UUID `path:"id"`
	VersionId uuid.UUID `path:"versionId"`
}
type GetInvestigationFindingResponse ItemResponse[InvestigationFinding]

var ListInvestigationHypotheses = openapi.Operation{OperationID: "list-investigation-hypotheses", Method: http.MethodGet, Path: "/investigations/{id}/hypotheses", Summary: "List Investigation Hypotheses", Tags: investigationsTags, Errors: ErrorCodes()}

type ListInvestigationHypothesesRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationHypothesesResponse PaginatedResponse[InvestigationHypothesis]

var GetInvestigationHypothesis = openapi.Operation{OperationID: "get-investigation-hypothesis", Method: http.MethodGet, Path: "/investigations/{id}/hypotheses/{versionId}", Summary: "Get Investigation Hypothesis", Tags: investigationsTags, Errors: ErrorCodes()}

type GetInvestigationHypothesisRequest struct {
	Id        uuid.UUID `path:"id"`
	VersionId uuid.UUID `path:"versionId"`
}
type GetInvestigationHypothesisResponse ItemResponse[InvestigationHypothesis]

var SubmitInvestigationUserInput = openapi.Operation{OperationID: "submit-investigation-user-input", Method: http.MethodPost, Path: "/investigations/{id}/user-inputs", Summary: "Submit Investigation User Input", Tags: investigationsTags, Errors: ErrorCodes(http.StatusConflict)}

type SubmitInvestigationUserInputRequest struct {
	Id   uuid.UUID `path:"id"`
	Body struct {
		Text          string `json:"text"`
		SubmissionKey string `json:"submissionKey"`
	}
}
type SubmitInvestigationUserInputResponse ItemResponse[InvestigationUserInput]

var ListInvestigationUserInputs = openapi.Operation{OperationID: "list-investigation-user-inputs", Method: http.MethodGet, Path: "/investigations/{id}/user-inputs", Summary: "List Investigation User Inputs", Tags: investigationsTags, Errors: ErrorCodes()}

type ListInvestigationUserInputsRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationUserInputsResponse PaginatedResponse[InvestigationUserInput]

var ListInvestigationEvidenceRevisions = openapi.Operation{OperationID: "list-investigation-evidence-revisions", Method: http.MethodGet, Path: "/investigations/{id}/evidence-revisions", Summary: "List Investigation Evidence Revisions", Tags: investigationsTags, Errors: ErrorCodes()}

type ListInvestigationEvidenceRevisionsRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationEvidenceRevisionsResponse PaginatedResponse[InvestigationEvidenceRevision]

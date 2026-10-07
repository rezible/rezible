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
	UpdateInvestigation(context.Context, *UpdateInvestigationRequest) (*UpdateInvestigationResponse, error)

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
	huma.Register(api, UpdateInvestigation, o.UpdateInvestigation)

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
		Query          string                   `json:"query"`
		AnalysisId     uuid.UUID                `json:"analysisId"`
		SessionId      uuid.UUID                `json:"sessionId"`
		LatestTurn     *AgentTurnStatusOverview `json:"latestTurn"`
		ActiveTurn     *AgentTurnStatusOverview `json:"activeTurn"`
		HasPendingWork bool                     `json:"hasPendingWork"`
		CreatedAt      time.Time                `json:"createdAt"`
		UpdatedAt      time.Time                `json:"updatedAt"`

		PendingEvidenceRevisions int       `json:"pendingEvidenceRevisions"`
		AutomaticUpdatesPaused   bool      `json:"automaticUpdatesPaused"`
		EvidenceCurrentAsOf      time.Time `json:"evidenceCurrentAsOf"`
	}

	InvestigationReport struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes InvestigationReportAttributes `json:"attributes"`
	}

	InvestigationReportAttributes struct {
		Text        string                   `json:"text"`
		Summary     string                   `json:"summary"`
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
		Text            string                   `json:"text"`
		UserId          uuid.UUID                `json:"userId"`
		SubmissionKey   string                   `json:"submissionKey"`
		CreatedAt       time.Time                `json:"createdAt"`
		AgentTurn       *AgentTurnStatusOverview `json:"agentTurn"`
		AnswerVersionId *uuid.UUID               `json:"answerVersionId" nullable:"true"`
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
	attrs := InvestigationAttributes{
		Query:          detail.Query,
		AnalysisId:     inv.SystemAnalysisID,
		SessionId:      inv.AgentSessionID,
		LatestTurn:     AgentTurnStatusOverviewFromDetail(detail.LatestTurn),
		ActiveTurn:     AgentTurnStatusOverviewFromDetail(detail.ActiveTurn),
		HasPendingWork: detail.HasPendingWork,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,

		PendingEvidenceRevisions: detail.PendingEvidenceRevisions,
		AutomaticUpdatesPaused:   detail.AutomaticUpdatesPaused,
		EvidenceCurrentAsOf:      detail.EvidenceCurrentAsOf,
	}
	return Investigation{Id: inv.ID, Attributes: attrs}
}
func InvestigationReportFromEnt(report *ent.InvestigationReport) InvestigationReport {
	turn := report.Edges.AgentTurn
	attrs := InvestigationReportAttributes{
		Text:        report.Text,
		Summary:     report.Summary,
		References:  investigationReferencesFromEnt(report.Edges.OutputReferences),
		AgentTurnId: report.AgentTurnID,
		TurnStatus:  investigationTurnStatus(turn),
		Provisional: investigationTurnProvisional(turn),
		CreatedAt:   report.CreatedAt,
	}
	return InvestigationReport{Id: report.ID, Attributes: attrs}
}

func InvestigationFindingFromEnt(version *ent.InvestigationFindingVersion) InvestigationFinding {
	turn := version.Edges.AgentTurn
	attrs := InvestigationFindingAttributes{
		FindingId:               version.FindingID,
		Title:                   version.Title,
		Body:                    version.Body,
		References:              investigationReferencesFromEnt(version.Edges.OutputReferences),
		FindingReferences:       ConvertSlice(version.Edges.OutgoingLinks, FindingVersionReferenceFromEnt),
		InvalidatedByVersionIds: version.InvalidatedByVersionIDs(),
		AgentTurnId:             version.AgentTurnID,
		TurnStatus:              investigationTurnStatus(turn),
		Provisional:             investigationTurnProvisional(turn),
		CreatedAt:               version.CreatedAt,
	}
	if finding := version.Edges.Finding; finding != nil {
		attrs.Key = finding.Key
		attrs.UserInputId = finding.UserInputID
	}
	return InvestigationFinding{Id: version.ID, Attributes: attrs}
}

func InvestigationHypothesisFromEnt(version *ent.InvestigationHypothesisVersion) InvestigationHypothesis {
	turn := version.Edges.AgentTurn
	attrs := InvestigationHypothesisAttributes{
		HypothesisId:  version.HypothesisID,
		Title:         version.Title,
		Justification: version.Justification,
		Status:        string(version.Status),
		References:    investigationReferencesFromEnt(version.Edges.OutputReferences),
		AgentTurnId:   version.AgentTurnID,
		TurnStatus:    investigationTurnStatus(turn),
		Provisional:   investigationTurnProvisional(turn),
		CreatedAt:     version.CreatedAt,
	}
	if hypothesis := version.Edges.Hypothesis; hypothesis != nil {
		attrs.Key = hypothesis.Key
	}
	return InvestigationHypothesis{Id: version.ID, Attributes: attrs}
}

func investigationTurnStatus(turn *ent.AgentTurn) string {
	if turn == nil {
		return ""
	}
	return string(turn.Status)
}

func investigationTurnProvisional(turn *ent.AgentTurn) bool {
	return turn != nil && turn.Status == agentturn.StatusRunning
}

func investigationReferencesFromEnt(refs ent.InvestigationOutputReferences) []InvestigationReference {
	return ConvertSlice(refs.KnowledgeEvidenceIDs(), InvestigationReferenceFromEvidenceID)
}

func InvestigationReferenceFromEvidenceID(evidenceID uuid.UUID) InvestigationReference {
	return InvestigationReference{Kind: "knowledge_evidence", Id: evidenceID}
}

func FindingVersionReferenceFromEnt(link *ent.InvestigationFindingVersionLink) FindingVersionReference {
	return FindingVersionReference{VersionId: link.TargetVersionID, Relation: string(link.Relation)}
}

func InvestigationUserInputFromEnt(input *ent.InvestigationUserInput) InvestigationUserInput {
	attrs := InvestigationUserInputAttributes{
		Text:          input.Text,
		UserId:        input.UserID,
		SubmissionKey: input.Key,
		CreatedAt:     input.CreatedAt,
		AgentTurn:     AgentTurnStatusOverviewFromDetail(input.Edges.AgentTurn),
	}
	if answer := input.CurrentAnswerVersion(); answer != nil {
		attrs.AnswerVersionId = &answer.ID
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

// UpdateInvestigation starts one turn for pending work, even when automatic updates are paused. It does
// nothing while a turn is queued or running, or when nothing is pending.
var UpdateInvestigation = openapi.Operation{
	OperationID: "update-investigation",
	Method:      http.MethodPost,
	Path:        "/investigations/{id}/update",
	Summary:     "Update Investigation",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type UpdateInvestigationRequest IdRequest
type UpdateInvestigationResponse ItemResponse[Investigation]

var GetInvestigationReport = openapi.Operation{
	OperationID: "get-investigation-report",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/report",
	Summary:     "Get Investigation Report",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type GetInvestigationReportRequest struct {
	Id        uuid.UUID `path:"id"`
	Selection string    `query:"selection" enum:"latest,completed" default:"latest" required:"false"`
}
type GetInvestigationReportResponse ItemResponse[InvestigationReport]

var ListInvestigationFindings = openapi.Operation{
	OperationID: "list-investigation-findings",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/findings",
	Summary:     "List Investigation Findings",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type ListInvestigationFindingsRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationFindingsResponse PaginatedResponse[InvestigationFinding]

var GetInvestigationFinding = openapi.Operation{
	OperationID: "get-investigation-finding",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/findings/{versionId}",
	Summary:     "Get Investigation Finding",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type GetInvestigationFindingRequest struct {
	Id        uuid.UUID `path:"id"`
	VersionId uuid.UUID `path:"versionId"`
}
type GetInvestigationFindingResponse ItemResponse[InvestigationFinding]

var ListInvestigationHypotheses = openapi.Operation{
	OperationID: "list-investigation-hypotheses",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/hypotheses",
	Summary:     "List Investigation Hypotheses",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type ListInvestigationHypothesesRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationHypothesesResponse PaginatedResponse[InvestigationHypothesis]

var GetInvestigationHypothesis = openapi.Operation{
	OperationID: "get-investigation-hypothesis",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/hypotheses/{versionId}",
	Summary:     "Get Investigation Hypothesis",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type GetInvestigationHypothesisRequest struct {
	Id        uuid.UUID `path:"id"`
	VersionId uuid.UUID `path:"versionId"`
}
type GetInvestigationHypothesisResponse ItemResponse[InvestigationHypothesis]

var SubmitInvestigationUserInput = openapi.Operation{
	OperationID: "submit-investigation-user-input",
	Method:      http.MethodPost,
	Path:        "/investigations/{id}/user-inputs",
	Summary:     "Submit Investigation User Input",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type SubmitInvestigationUserInputRequest struct {
	Id   uuid.UUID `path:"id"`
	Body struct {
		Text          string `json:"text"`
		SubmissionKey string `json:"submissionKey"`
	}
}
type SubmitInvestigationUserInputResponse ItemResponse[InvestigationUserInput]

var ListInvestigationUserInputs = openapi.Operation{
	OperationID: "list-investigation-user-inputs",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/user-inputs",
	Summary:     "List Investigation User Inputs",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type ListInvestigationUserInputsRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationUserInputsResponse PaginatedResponse[InvestigationUserInput]

var ListInvestigationEvidenceRevisions = openapi.Operation{
	OperationID: "list-investigation-evidence-revisions",
	Method:      http.MethodGet,
	Path:        "/investigations/{id}/evidence-revisions",
	Summary:     "List Investigation Evidence Revisions",
	Tags:        investigationsTags,
	Errors:      ErrorCodes(),
}

type ListInvestigationEvidenceRevisionsRequest struct {
	Id uuid.UUID `path:"id"`
	PaginationRequest
}
type ListInvestigationEvidenceRevisionsResponse PaginatedResponse[InvestigationEvidenceRevision]

// Huma does not support nullable tags on referenced objects. Describe absent
// turns explicitly while retaining the shared status overview schema.
func (InvestigationAttributes) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	for _, name := range []string{"activeTurn", "latestTurn"} {
		schema.Properties[name] = &huma.Schema{AnyOf: []*huma.Schema{
			schema.Properties[name],
			{Type: "null"},
		}}
	}
	return schema
}

func (InvestigationUserInputAttributes) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	schema.Properties["agentTurn"] = &huma.Schema{AnyOf: []*huma.Schema{
		schema.Properties["agentTurn"],
		{Type: "null"},
	}}
	return schema
}

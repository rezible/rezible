package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type SystemAnalysisHandler interface {
	GetSystemAnalysis(context.Context, *GetSystemAnalysisRequest) (*GetSystemAnalysisResponse, error)
	UpdateSystemAnalysis(context.Context, *UpdateSystemAnalysisRequest) (*UpdateSystemAnalysisResponse, error)
	GetSystemAnalysisGraph(context.Context, *GetSystemAnalysisGraphRequest) (*GetSystemAnalysisGraphResponse, error)

	ListSystemAnalysisEntries(context.Context, *ListSystemAnalysisEntriesRequest) (*ListSystemAnalysisEntriesResponse, error)
	CreateSystemAnalysisEntry(context.Context, *CreateSystemAnalysisEntryRequest) (*CreateSystemAnalysisEntryResponse, error)
	UpdateSystemAnalysisEntry(context.Context, *UpdateSystemAnalysisEntryRequest) (*UpdateSystemAnalysisEntryResponse, error)
	DeleteSystemAnalysisEntry(context.Context, *DeleteSystemAnalysisEntryRequest) (*DeleteSystemAnalysisEntryResponse, error)

	AddSystemAnalysisEntrySubject(context.Context, *AddSystemAnalysisEntrySubjectRequest) (*AddSystemAnalysisEntrySubjectResponse, error)
	UpdateSystemAnalysisEntrySubject(context.Context, *UpdateSystemAnalysisEntrySubjectRequest) (*UpdateSystemAnalysisEntrySubjectResponse, error)
	DeleteSystemAnalysisEntrySubject(context.Context, *DeleteSystemAnalysisEntrySubjectRequest) (*DeleteSystemAnalysisEntrySubjectResponse, error)
}

func (o operations) RegisterSystemAnalysis(api huma.API) {
	huma.Register(api, GetSystemAnalysis, o.GetSystemAnalysis)
	huma.Register(api, UpdateSystemAnalysis, o.UpdateSystemAnalysis)
	huma.Register(api, GetSystemAnalysisGraph, o.GetSystemAnalysisGraph)
	huma.Register(api, ListSystemAnalysisEntries, o.ListSystemAnalysisEntries)
	huma.Register(api, CreateSystemAnalysisEntry, o.CreateSystemAnalysisEntry)
	huma.Register(api, UpdateSystemAnalysisEntry, o.UpdateSystemAnalysisEntry)
	huma.Register(api, DeleteSystemAnalysisEntry, o.DeleteSystemAnalysisEntry)
	huma.Register(api, AddSystemAnalysisEntrySubject, o.AddSystemAnalysisEntrySubject)
	huma.Register(api, UpdateSystemAnalysisEntrySubject, o.UpdateSystemAnalysisEntrySubject)
	huma.Register(api, DeleteSystemAnalysisEntrySubject, o.DeleteSystemAnalysisEntrySubject)
}

type (
	SystemAnalysis struct {
		Id         uuid.UUID                `json:"id"`
		Attributes SystemAnalysisAttributes `json:"attributes"`
	}

	SystemAnalysisAttributes struct {
		ScopeEntityId   *uuid.UUID            `json:"scopeEntityId,omitempty"`
		SubjectEntityId *uuid.UUID            `json:"subjectEntityId,omitempty"`
		ReferenceTime   *time.Time            `json:"referenceTime,omitempty"`
		Entries         []SystemAnalysisEntry `json:"entries"`
	}

	SystemAnalysisEntry struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes SystemAnalysisEntryAttributes `json:"attributes"`
	}

	SystemAnalysisEntryAttributes struct {
		Kind       string                       `json:"kind" enum:"observation,context,decision,action,finding,recommendation"`
		OccurredAt *time.Time                   `json:"occurredAt,omitempty"`
		Sequence   int                          `json:"sequence"`
		Title      string                       `json:"title"`
		Body       string                       `json:"body,omitempty"`
		Properties map[string]any               `json:"properties"`
		Subjects   []SystemAnalysisEntrySubject `json:"subjects"`
	}

	SystemAnalysisEntrySubject struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemAnalysisEntrySubjectAttributes `json:"attributes"`
	}

	SystemAnalysisEntrySubjectAttributes struct {
		Role                    string                      `json:"role"`
		SubjectKind             string                      `json:"subjectKind" enum:"entity,relationship,evidence"`
		SubjectId               uuid.UUID                   `json:"subjectId"`
		KnowledgeEntityId       *uuid.UUID                  `json:"knowledgeEntityId,omitempty"`
		KnowledgeRelationshipId *uuid.UUID                  `json:"knowledgeRelationshipId,omitempty"`
		KnowledgeEvidenceId     *uuid.UUID                  `json:"knowledgeEvidenceId,omitempty"`
		KnowledgeEntity         *KnowledgeGraphEntity       `json:"knowledgeEntity,omitempty"`
		KnowledgeRelationship   *KnowledgeGraphRelationship `json:"knowledgeRelationship,omitempty"`
		KnowledgeEvidence       *KnowledgeGraphEvidence     `json:"knowledgeEvidence,omitempty"`
	}

	UpdateSystemAnalysisAttributes struct {
		ScopeEntityId   *uuid.UUID `json:"scopeEntityId,omitempty"`
		SubjectEntityId *uuid.UUID `json:"subjectEntityId,omitempty"`
		ReferenceTime   *time.Time `json:"referenceTime,omitempty"`
	}

	CreateSystemAnalysisEntryAttributes struct {
		Kind       string         `json:"kind" enum:"observation,context,decision,action,finding,recommendation"`
		OccurredAt *time.Time     `json:"occurredAt,omitempty"`
		Sequence   int            `json:"sequence"`
		Title      string         `json:"title"`
		Body       *string        `json:"body,omitempty"`
		Properties map[string]any `json:"properties,omitempty"`
	}

	UpdateSystemAnalysisEntryAttributes struct {
		Kind       *string        `json:"kind,omitempty" enum:"observation,context,decision,action,finding,recommendation"`
		OccurredAt *time.Time     `json:"occurredAt,omitempty"`
		Sequence   *int           `json:"sequence,omitempty"`
		Title      *string        `json:"title,omitempty"`
		Body       *string        `json:"body,omitempty"`
		Properties map[string]any `json:"properties,omitempty"`
	}

	SetSystemAnalysisEntrySubjectAttributes struct {
		Role                    string     `json:"role"`
		KnowledgeEntityId       *uuid.UUID `json:"knowledgeEntityId,omitempty"`
		KnowledgeRelationshipId *uuid.UUID `json:"knowledgeRelationshipId,omitempty"`
		KnowledgeEvidenceId     *uuid.UUID `json:"knowledgeEvidenceId,omitempty"`
	}

	UpdateSystemAnalysisEntrySubjectAttributes struct {
		Role string `json:"role"`
	}
)

func SystemAnalysisFromEnt(analysis *ent.SystemAnalysis, entries []SystemAnalysisEntry) SystemAnalysis {
	attrs := SystemAnalysisAttributes{
		ScopeEntityId:   analysis.ScopeEntityID,
		SubjectEntityId: analysis.SubjectEntityID,
		ReferenceTime:   analysis.ReferenceTime,
		Entries:         entries,
	}
	return SystemAnalysis{Id: analysis.ID, Attributes: attrs}
}

func SystemAnalysisWithEntriesFromEnt(analysis *ent.SystemAnalysis) SystemAnalysis {
	entries := ConvertSlice(analysis.Edges.Entries, SystemAnalysisEntryWithSubjectsFromEnt)
	return SystemAnalysisFromEnt(analysis, entries)
}

func SystemAnalysisEntryFromEnt(entry *ent.SystemAnalysisEntry, subjects []SystemAnalysisEntrySubject) SystemAnalysisEntry {
	properties := entry.Properties
	if properties == nil {
		properties = map[string]any{}
	}
	attrs := SystemAnalysisEntryAttributes{
		Kind:       entry.Kind.String(),
		OccurredAt: entry.OccurredAt,
		Sequence:   entry.Sequence,
		Title:      entry.Title,
		Body:       entry.Body,
		Properties: properties,
		Subjects:   subjects,
	}
	return SystemAnalysisEntry{Id: entry.ID, Attributes: attrs}
}

func SystemAnalysisEntryWithSubjectsFromEnt(entry *ent.SystemAnalysisEntry) SystemAnalysisEntry {
	subjects := ConvertSlice(entry.Edges.Subjects, SystemAnalysisEntrySubjectFromEnt)
	return SystemAnalysisEntryFromEnt(entry, subjects)
}

func SystemAnalysisEntrySubjectFromEnt(subject *ent.SystemAnalysisEntrySubject) SystemAnalysisEntrySubject {
	attrs := SystemAnalysisEntrySubjectAttributes{
		Role:                    subject.Role,
		KnowledgeEntityId:       subject.KnowledgeEntityID,
		KnowledgeRelationshipId: subject.KnowledgeRelationshipID,
		KnowledgeEvidenceId:     subject.KnowledgeEvidenceID,
	}

	if subject.KnowledgeEntityID != nil {
		attrs.SubjectKind = "entity"
		attrs.SubjectId = *subject.KnowledgeEntityID
		if subject.Edges.KnowledgeEntity != nil {
			converted := KnowledgeGraphEntityFromEnt(subject.Edges.KnowledgeEntity)
			attrs.KnowledgeEntity = &converted
		}
	}
	if subject.KnowledgeRelationshipID != nil {
		attrs.SubjectKind = "relationship"
		attrs.SubjectId = *subject.KnowledgeRelationshipID
		if subject.Edges.KnowledgeRelationship != nil {
			converted := KnowledgeGraphRelationshipFromEnt(subject.Edges.KnowledgeRelationship)
			attrs.KnowledgeRelationship = &converted
		}
	}
	if subject.KnowledgeEvidenceID != nil {
		attrs.SubjectKind = "evidence"
		attrs.SubjectId = *subject.KnowledgeEvidenceID
		if subject.Edges.KnowledgeEvidence != nil {
			attrs.KnowledgeEvidence = KnowledgeGraphEvidenceFromEnt(subject.Edges.KnowledgeEvidence)
		}
	}

	return SystemAnalysisEntrySubject{Id: subject.ID, Attributes: attrs}
}

var systemAnalysisTags = []string{"System Analysis"}

var GetSystemAnalysis = huma.Operation{
	OperationID: "get-system-analysis",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}",
	Summary:     "Get System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisRequest IdRequest
type GetSystemAnalysisResponse ItemResponse[SystemAnalysis]

var UpdateSystemAnalysis = huma.Operation{
	OperationID: "update-system-analysis",
	Method:      http.MethodPatch,
	Path:        "/system_analysis/{id}",
	Summary:     "Update System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisRequest IdRequestWithBody[UpdateSystemAnalysisAttributes]
type UpdateSystemAnalysisResponse ItemResponse[SystemAnalysis]

var GetSystemAnalysisGraph = huma.Operation{
	OperationID: "get-system-analysis-graph",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/graph",
	Summary:     "Get System Analysis Graph",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisGraphRequest struct {
	Id               uuid.UUID `path:"id"`
	Depth            int       `query:"depth" required:"false"`
	RelationshipKind []string  `query:"relationshipKind" required:"false"`
}
type GetSystemAnalysisGraphResponse ItemResponse[KnowledgeGraphView]

var ListSystemAnalysisEntries = huma.Operation{
	OperationID: "list-system-analysis-entries",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/entries",
	Summary:     "List System Analysis Entries",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type ListSystemAnalysisEntriesRequest ListIdRequest
type ListSystemAnalysisEntriesResponse ListResponse[SystemAnalysisEntry]

var CreateSystemAnalysisEntry = huma.Operation{
	OperationID: "create-system-analysis-entry",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/entries",
	Summary:     "Create System Analysis Entry",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type CreateSystemAnalysisEntryRequest IdRequestWithBody[CreateSystemAnalysisEntryAttributes]
type CreateSystemAnalysisEntryResponse ItemResponse[SystemAnalysisEntry]

var UpdateSystemAnalysisEntry = huma.Operation{
	OperationID: "update-system-analysis-entry",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_entries/{id}",
	Summary:     "Update System Analysis Entry",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisEntryRequest IdRequestWithBody[UpdateSystemAnalysisEntryAttributes]
type UpdateSystemAnalysisEntryResponse ItemResponse[SystemAnalysisEntry]

var DeleteSystemAnalysisEntry = huma.Operation{
	OperationID: "delete-system-analysis-entry",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_entries/{id}",
	Summary:     "Delete System Analysis Entry",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEntryRequest IdRequest
type DeleteSystemAnalysisEntryResponse EmptyResponse

var AddSystemAnalysisEntrySubject = huma.Operation{
	OperationID: "add-system-analysis-entry-subject",
	Method:      http.MethodPost,
	Path:        "/system_analysis_entries/{id}/subjects",
	Summary:     "Add System Analysis Entry Subject",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisEntrySubjectRequest IdRequestWithBody[SetSystemAnalysisEntrySubjectAttributes]
type AddSystemAnalysisEntrySubjectResponse ItemResponse[SystemAnalysisEntrySubject]

var UpdateSystemAnalysisEntrySubject = huma.Operation{
	OperationID: "update-system-analysis-entry-subject",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_entry_subjects/{id}",
	Summary:     "Update System Analysis Entry Subject",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisEntrySubjectRequest IdRequestWithBody[UpdateSystemAnalysisEntrySubjectAttributes]
type UpdateSystemAnalysisEntrySubjectResponse ItemResponse[SystemAnalysisEntrySubject]

var DeleteSystemAnalysisEntrySubject = huma.Operation{
	OperationID: "delete-system-analysis-entry-subject",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_entry_subjects/{id}",
	Summary:     "Delete System Analysis Entry Subject",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEntrySubjectRequest IdRequest
type DeleteSystemAnalysisEntrySubjectResponse EmptyResponse

package v1

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/openapi"
)

type (
	IncidentImpact struct {
		Id         uuid.UUID                `json:"id"`
		Attributes IncidentImpactAttributes `json:"attributes"`
	}

	IncidentImpactAttributes struct {
		KnowledgeEntity IncidentContextEntity `json:"knowledgeEntity"`
		Source          string                `json:"source,omitempty"`
		Note            string                `json:"note,omitempty"`
	}

	SetIncidentImpactsAttributes struct {
		Impacts []SetIncidentImpactAttributes `json:"impacts"`
	}

	SetIncidentImpactAttributes struct {
		KnowledgeEntityId *uuid.UUID `json:"knowledgeEntityId,omitempty"`
		Kind              string     `json:"kind,omitempty"`
		DisplayName       string     `json:"displayName,omitempty"`
		Description       string     `json:"description,omitempty"`
		Source            string     `json:"source,omitempty"`
		Note              string     `json:"note,omitempty"`
	}

	IncidentContextPack struct {
		IncidentId           uuid.UUID                        `json:"incidentId"`
		GeneratedAt          time.Time                        `json:"generatedAt"`
		ExplicitImpacts      []IncidentContextEntity          `json:"explicitImpacts" nullable:"false"`
		InferredImpacts      []IncidentContextEntity          `json:"inferredImpacts" nullable:"false"`
		ActiveAlerts         []IncidentContextAlert           `json:"activeAlerts" nullable:"false"`
		RecentEvidence       []IncidentContextEvidence        `json:"recentEvidence" nullable:"false"`
		RelatedIncidents     []IncidentContextRelatedIncident `json:"relatedIncidents" nullable:"false"`
		RetrievalLimitations []string                         `json:"retrievalLimitations" nullable:"false"`
	}

	IncidentContextEntity struct {
		Id          uuid.UUID                    `json:"id"`
		Kind        string                       `json:"kind"`
		DisplayName string                       `json:"displayName"`
		Description string                       `json:"description,omitempty"`
		Score       float64                      `json:"score"`
		Signals     []string                     `json:"signals" nullable:"false"`
		Evidence    []IncidentContextEvidenceRef `json:"evidence" nullable:"false"`
	}

	IncidentContextEvidenceRef struct {
		Kind        string    `json:"kind"`
		Id          uuid.UUID `json:"id"`
		Description string    `json:"description,omitempty"`
		ObservedAt  time.Time `json:"observedAt"`
	}

	IncidentContextAlert struct {
		Id                uuid.UUID                    `json:"id"`
		KnowledgeEntityId uuid.UUID                    `json:"knowledgeEntityId"`
		Title             string                       `json:"title"`
		Description       string                       `json:"description,omitempty"`
		ObservedAt        time.Time                    `json:"observedAt"`
		RelatedEntityIds  []uuid.UUID                  `json:"relatedEntityIds" nullable:"false"`
		Evidence          []IncidentContextEvidenceRef `json:"evidence" nullable:"false"`
	}

	IncidentContextEvidence struct {
		Id             uuid.UUID  `json:"id"`
		EventId        uuid.UUID  `json:"eventId"`
		EntityId       *uuid.UUID `json:"entityId,omitempty"`
		RelationshipId *uuid.UUID `json:"relationshipId,omitempty"`
		SubjectType    string     `json:"subjectType"`
		Assertion      string     `json:"assertion"`
		EvidenceKind   string     `json:"evidenceKind"`
		ObservedAt     time.Time  `json:"observedAt"`
		Description    string     `json:"description,omitempty"`
	}

	IncidentContextRelatedIncident struct {
		Id        uuid.UUID   `json:"id"`
		Slug      string      `json:"slug"`
		Title     string      `json:"title"`
		OpenedAt  time.Time   `json:"openedAt"`
		Signals   []string    `json:"signals" nullable:"false"`
		EntityIds []uuid.UUID `json:"entityIds" nullable:"false"`
	}
)

func IncidentImpactFromEnt(impact *ent.IncidentImpact) IncidentImpact {
	return IncidentImpact{
		Id: impact.ID,
		Attributes: IncidentImpactAttributes{
			KnowledgeEntity: IncidentContextEntityFromKnowledgeEntity(impact.Edges.KnowledgeEntity),
			Source:          impact.Source,
			Note:            impact.Note,
		},
	}
}

func IncidentContextEntityFromKnowledgeEntity(entity *ent.KnowledgeEntity) IncidentContextEntity {
	if entity == nil {
		return IncidentContextEntity{
			Signals:  []string{},
			Evidence: []IncidentContextEvidenceRef{},
		}
	}
	return IncidentContextEntity{
		Id:          entity.ID,
		Kind:        entity.Kind,
		DisplayName: entity.DisplayName,
		Description: entity.Description,
		Signals:     []string{},
		Evidence:    []IncidentContextEvidenceRef{},
	}
}

func IncidentContextPackFromDomain(pack *rez.IncidentContextPack) IncidentContextPack {
	if pack == nil {
		return IncidentContextPack{}
	}
	return IncidentContextPack{
		IncidentId:           pack.IncidentID,
		GeneratedAt:          pack.GeneratedAt,
		ExplicitImpacts:      convert(pack.ExplicitImpacts, IncidentContextEntityFromDomain),
		InferredImpacts:      convert(pack.InferredImpacts, IncidentContextEntityFromDomain),
		ActiveAlerts:         convert(pack.ActiveAlerts, IncidentContextAlertFromDomain),
		RecentEvidence:       convert(pack.RecentEvidence, IncidentContextEvidenceFromDomain),
		RelatedIncidents:     convert(pack.RelatedIncidents, IncidentContextRelatedIncidentFromDomain),
		RetrievalLimitations: pack.RetrievalLimitations,
	}
}

func IncidentContextEntityFromDomain(entity rez.IncidentContextEntity) IncidentContextEntity {
	return IncidentContextEntity{
		Id:          entity.ID,
		Kind:        entity.Kind,
		DisplayName: entity.DisplayName,
		Description: entity.Description,
		Score:       entity.Score,
		Signals:     entity.Signals,
		Evidence:    convert(entity.Evidence, IncidentContextEvidenceRefFromDomain),
	}
}

func IncidentContextEvidenceRefFromDomain(ref rez.IncidentContextEvidenceRef) IncidentContextEvidenceRef {
	return IncidentContextEvidenceRef{
		Kind:        ref.Kind,
		Id:          ref.ID,
		Description: ref.Description,
		ObservedAt:  ref.ObservedAt,
	}
}

func IncidentContextAlertFromDomain(alert rez.IncidentContextAlert) IncidentContextAlert {
	return IncidentContextAlert{
		Id:                alert.ID,
		KnowledgeEntityId: alert.KnowledgeEntityID,
		Title:             alert.Title,
		Description:       alert.Description,
		ObservedAt:        alert.ObservedAt,
		RelatedEntityIds:  alert.RelatedEntityIDs,
		Evidence:          convert(alert.Evidence, IncidentContextEvidenceRefFromDomain),
	}
}

func IncidentContextEvidenceFromDomain(evidence rez.IncidentContextEvidence) IncidentContextEvidence {
	return IncidentContextEvidence{
		Id:             evidence.ID,
		EventId:        evidence.EventID,
		EntityId:       evidence.EntityID,
		RelationshipId: evidence.RelationshipID,
		SubjectType:    evidence.SubjectType,
		Assertion:      evidence.Assertion,
		EvidenceKind:   evidence.EvidenceKind,
		ObservedAt:     evidence.ObservedAt,
		Description:    evidence.Description,
	}
}

func IncidentContextRelatedIncidentFromDomain(inc rez.IncidentContextRelatedIncident) IncidentContextRelatedIncident {
	return IncidentContextRelatedIncident{
		Id:        inc.ID,
		Slug:      inc.Slug,
		Title:     inc.Title,
		OpenedAt:  inc.OpenedAt,
		Signals:   inc.Signals,
		EntityIds: inc.EntityIDs,
	}
}

var ListIncidentImpacts = openapi.Operation{
	OperationID: "list-incident-impacts",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/impacts",
	Summary:     "List Incident Impacts",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type ListIncidentImpactsRequest = FlexibleIdRequest
type ListIncidentImpactsResponse ListResponse[IncidentImpact]

var SetIncidentImpacts = openapi.Operation{
	OperationID: "set-incident-impacts",
	Method:      http.MethodPut,
	Path:        "/incidents/{id}/impacts",
	Summary:     "Set Incident Impacts",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type SetIncidentImpactsRequest struct {
	FlexibleIdRequest
	RequestWithBodyAttributes[SetIncidentImpactsAttributes]
}
type SetIncidentImpactsResponse ListResponse[IncidentImpact]

var GetIncidentContextPack = openapi.Operation{
	OperationID: "get-incident-context-pack",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/context-pack",
	Summary:     "Get Incident Context Pack",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type GetIncidentContextPackRequest = FlexibleIdRequest
type GetIncidentContextPackResponse ItemResponse[IncidentContextPack]

var RequestIncidentContextPackAgentRun = huma.Operation{
	OperationID: "request-incident-context-pack-agent-run",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/context-pack/agent-runs",
	Summary:     "Request Incident Context Pack Agent Run",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type RequestIncidentContextPackAgentRunRequest = FlexibleIdRequest
type RequestIncidentContextPackAgentRunResponse ItemResponse[AgentRun]

package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ale "github.com/rezible/rezible/ent/alertepisode"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/situation"
	si "github.com/rezible/rezible/ent/situationinvestigation"
	sog "github.com/rezible/rezible/ent/situationobservationgroup"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	alertEpisodeLockNamespace              = "alert_episode"
	situationHazardAssessmentLockNamespace = "situation_hazard_assessment"
	situationKnowledgeEntityKind           = "situation"
	situationLockNamespace                 = "situation"
)

type SituationService struct {
	db        rez.Database
	jobs      rez.JobService
	agents    rez.AgentSessionService
	knowledge rez.KnowledgeGraphService
}

func NewSituationService(db rez.Database, jobs rez.JobService, agents rez.AgentSessionService, knowledge rez.KnowledgeGraphService) (*SituationService, error) {
	s := &SituationService{db: db, agents: agents, jobs: jobs, knowledge: knowledge}

	return s, nil
}

func (s *SituationService) GetMessageHandlers() []rez.MessageEventHandler {
	return []rez.MessageEventHandler{
		messages.NewEventHandler("db.SituationService.onAgentTurnUpdated", s.onAgentTurnUpdated),
	}
}

func (s *SituationService) ListSituations(ctx context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
	query := s.db.Client(ctx).Situation.Query().
		WithKnowledgeEntity().
		WithInvestigations(func(q *ent.SituationInvestigationQuery) {
			q.WithAgentSession()
		}).
		WithObservationGroups(func(q *ent.SituationObservationGroupQuery) {
			q.WithEvents()
			q.WithAlertEpisodes()
			q.Order(sog.ByID())
		}).
		WithIncidents().
		Order(situation.ByOpenedAt(params.GetOrder()), situation.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query = query.Where(situation.TitleContainsFold(search))
	}
	if params.Status != "" {
		query = query.Where(situation.StatusEQ(params.Status))
	}
	if params.OpenedAfter != nil {
		query = query.Where(situation.OpenedAtGTE(*params.OpenedAfter))
	}
	return ent.DoListQuery[ent.Situation, *ent.SituationQuery](ctx, query, params.ListParams)
}

func (s *SituationService) GetSituation(ctx context.Context, id uuid.UUID) (*ent.Situation, error) {
	return s.db.Client(ctx).Situation.Query().
		Where(situation.ID(id)).
		WithKnowledgeEntity().
		WithInvestigations(func(q *ent.SituationInvestigationQuery) { q.WithAgentSession() }).
		WithObservationGroups(func(q *ent.SituationObservationGroupQuery) {
			q.WithEvents().WithAlertEpisodes(func(eq *ent.AlertEpisodeQuery) {
				eq.WithAlertDefinition().Order(ale.ByStartedAt(), ale.ByID())
			}).Order(sog.ByID())
		}).
		WithIncidents().
		Only(ctx)
}

func (s *SituationService) CreateSituation(ctx context.Context, params rez.CreateSituationParams) (*ent.Situation, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: situation title is required", rez.ErrInvalidInput)
	}
	openedAt := params.OpenedAt
	if openedAt.IsZero() {
		openedAt = time.Now().UTC()
	}

	var result *ent.Situation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		situationId := uuid.New()

		knEntId, knEntErr := s.resolveSituationKnowledgeEntityId(ctx, situationId)
		if knEntErr != nil {
			return fmt.Errorf("resolve knowledge entity: %w", knEntErr)
		}

		createSituation := tx.Situation.Create().
			SetID(situationId).
			SetKnowledgeEntityID(knEntId).
			SetTitle(title).
			SetStatus(situation.StatusOpen).
			SetOpenedAt(openedAt)
		if summary := strings.TrimSpace(params.Summary); summary != "" {
			createSituation.SetSummary(summary)
		}
		if len(params.IncidentIDs) > 0 {
			createSituation.AddIncidentIDs(params.IncidentIDs...)
		}
		created, saveErr := createSituation.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create situation: %w", saveErr)
		}

		for _, groupParams := range params.ObservationGroups {
			groupTitle := strings.TrimSpace(groupParams.Title)
			if groupTitle == "" {
				return fmt.Errorf("observation group title is required")
			}
			createGroup := tx.SituationObservationGroup.Create().SetSituationID(created.ID).
				SetTitle(groupTitle).AddEventIDs(groupParams.NormalizedEventIDs...).
				AddAlertEpisodeIDs(groupParams.AlertEpisodeIDs...)
			if groupParams.Body != nil {
				createGroup.SetBody(*groupParams.Body)
			}
			if _, groupErr := createGroup.Save(ctx); groupErr != nil {
				return fmt.Errorf("create observation group: %w", groupErr)
			}
		}

		result = created.Unwrap()
		return nil
	})
}

func (s *SituationService) resolveSituationKnowledgeEntityId(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	situationEntityRef := &rez.KnowledgeEntityRef{
		Category:            kne.CategoryEvent,
		Kind:                situationKnowledgeEntityKind,
		ProviderResourceRef: projections.InternalEntityResourceRef(id),
	}
	alias, resErr := s.knowledge.ResolveInternalSubject(ctx, rez.KnowledgeSubjectRef{Entity: situationEntityRef})
	if resErr != nil || alias.EntityID == nil {
		return uuid.Nil, fmt.Errorf("create situation knowledge entity: %w", resErr)
	}
	return *alias.EntityID, nil
}

func (s *SituationService) CloseSituation(ctx context.Context, params rez.CloseSituationParams) (*ent.Situation, error) {
	if params.SituationID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}
	if !(params.Reason == situation.CloseReasonStabilized || params.Reason == situation.CloseReasonDismissed) {
		return nil, fmt.Errorf("%w: invalid situation close reason", rez.ErrInvalidInput)
	}

	var closed *ent.Situation
	return closed, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, params.SituationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		current, queryErr := tx.Situation.Get(ctx, params.SituationID)
		if queryErr != nil {
			return fmt.Errorf("get situation: %w", queryErr)
		}
		if current.Status == situation.StatusClosed {
			if current.CloseReason != nil && *current.CloseReason == params.Reason {
				closed = current.Unwrap()
				return nil
			}
			return fmt.Errorf("%w: situation is already closed with another reason", rez.ErrConflict)
		}

		update := tx.Situation.UpdateOneID(params.SituationID).
			SetStatus(situation.StatusClosed).
			SetCloseReason(params.Reason).
			SetClosedAt(time.Now().UTC())
		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return fmt.Errorf("close situation: %w", updateErr)
		}
		closed = updated.Unwrap()
		return nil
	})
}

type situationKnowledgeRelationshipEntity struct {
	id       uuid.UUID
	category kne.Category
	kind     string
	pred     knr.Predicate
	isTarget bool
}

func (s *SituationService) NotifySituationObservationGroupUpdated(ctx context.Context, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, id.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}
		updateRevision := tx.Situation.UpdateOneID(id).AddEvidenceRevision(1)
		if updateErr := updateRevision.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update observation revision: %w", updateErr)
		}
		investigationQuery := tx.SituationInvestigation.Query().Where(si.SituationID(id))
		investigationIDs, queryErr := investigationQuery.IDs(ctx)
		if queryErr != nil {
			return fmt.Errorf("query investigations: %w", queryErr)
		}
		if reconcileErr := s.requestReconcileInvestigations(ctx, investigationIDs...); reconcileErr != nil {
			return fmt.Errorf("request reconcile investigations: %w", reconcileErr)
		}
		return nil
	})
}

func (s *SituationService) ingestKnowledgeRelationship(ctx context.Context, sitId uuid.UUID, p situationKnowledgeRelationshipEntity) error {
	entityRef := rez.KnowledgeEntityRef{
		Category:            p.category,
		Kind:                p.kind,
		ProviderResourceRef: projections.InternalEntityResourceRef(p.id),
		LinkingAttributes:   projections.KnowledgeEntityLinkingAttributes{ID: p.id},
	}

	situationEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryEvent,
		Kind:                "situation",
		ProviderResourceRef: projections.InternalEntityResourceRef(sitId),
		LinkingAttributes:   projections.KnowledgeEntityLinkingAttributes{ID: sitId},
	}

	relationshipRef := &rez.KnowledgeRelationshipRef{
		ProviderResourceRef: projections.InternalRelationshipResourceRef(p.id, sitId),
		Source:              entityRef,
		Predicate:           p.pred,
		Target:              situationEntityRef,
	}
	if p.isTarget {
		relationshipRef.Source = situationEntityRef
		relationshipRef.Target = entityRef
	}

	subjRef := rez.KnowledgeSubjectRef{Relationship: relationshipRef}

	if _, relErr := s.knowledge.ResolveInternalSubject(ctx, subjRef); relErr != nil {
		return fmt.Errorf("ingest situation knowledge relationship: %w", relErr)
	}
	return nil
}

func (s *SituationService) AddIncidentToSituation(ctx context.Context, situationID, incidentID uuid.UUID) error {
	if situationID == uuid.Nil || incidentID == uuid.Nil {
		return fmt.Errorf("%w: situation and incident ids are required", rez.ErrInvalidInput)
	}

	addIncident := s.db.Client(ctx).Situation.UpdateOneID(situationID).AddIncidentIDs(incidentID)
	if updateErr := addIncident.Exec(ctx); updateErr != nil {
		return fmt.Errorf("add incident to situation: %w", updateErr)
	}
	return nil
}

func (s *SituationService) RemoveIncidentFromSituation(ctx context.Context, situationID, incidentID uuid.UUID) error {
	if situationID == uuid.Nil || incidentID == uuid.Nil {
		return fmt.Errorf("%w: situation and incident ids are required", rez.ErrInvalidInput)
	}

	removeIncident := s.db.Client(ctx).Situation.UpdateOneID(situationID).RemoveIncidentIDs(incidentID)
	if updateErr := removeIncident.Exec(ctx); updateErr != nil {
		return fmt.Errorf("remove incident from situation: %w", updateErr)
	}
	return nil
}

/*
func (s *SituationService) ingestSituationClassifiedAsHazard(ctx context.Context, situationID, systemHazardID uuid.UUID) error {
	client := s.db.Client(ctx)
	situation, situationErr := client.Situation.Get(ctx, situationID)
	if situationErr != nil {
		return fmt.Errorf("get situation: %w", situationErr)
	}
	hazard, hazardErr := client.SystemHazard.Get(ctx, systemHazardID)
	if hazardErr != nil {
		return fmt.Errorf("get system hazard: %w", hazardErr)
	}
	if hazard.KnowledgeEntityID == nil {
		return fmt.Errorf("%w: system hazard has no knowledge entity", rez.ErrConflict)
	}
	ref := situationClassifiedAsHazardRef(
		situationProjectionEndpoints{situationID: situation.ID, knowledgeEntityID: situation.KnowledgeEntityID},
		situationProjectionEndpoints{systemHazardID: hazard.ID, knowledgeEntityID: *hazard.KnowledgeEntityID},
	)
	return s.knowledge.IngestInternalSubject(ctx, ref)
}
*/

/*
func systemHazardEntityRef(hazard situationProjectionEndpoints) rez.KnowledgeEntityRef {
	return internalEntityRef("system-hazard/"+hazard.systemHazardID.String(), kne.CategoryConcern, systemHazardKnowledgeEntityKind, hazard.knowledgeEntityID)
}

func situationEntityRef(situation situationProjectionEndpoints) rez.KnowledgeEntityRef {
	return internalEntityRef("situation/"+situation.situationID.String(), kne.CategoryEvent, situationKnowledgeEntityKind, situation.knowledgeEntityID)
}

func incidentEntityRef(incident situationProjectionEndpoints) rez.KnowledgeEntityRef {
	return internalEntityRef("incident/"+incident.incidentID.String(), kne.CategoryEvent, incidentKnowledgeEntityKind, incident.knowledgeEntityID)
}

func situationHazardClassificationResourceRef(situationID, systemHazardID uuid.UUID) string {
	return "situation-hazard-classification/" + situationID.String() + "/" + systemHazardID.String()
}

func incidentRespondsToSituationResourceRef(incidentID, situationID uuid.UUID) string {
	return "incident-responds-to-situation/" + incidentID.String() + "/" + situationID.String()
}

func situationClassifiedAsHazardRef(situation, hazard situationProjectionEndpoints) rez.KnowledgeSubjectRef {
	return internalRelationshipRef("situation-hazard-classification/"+situation.situationID.String()+"/"+hazard.systemHazardID.String(), knr.PredicateClassifiedAs, situationEntityRef(situation), systemHazardEntityRef(hazard))
}

func incidentRespondsToSituationRef(incident, situation situationProjectionEndpoints) rez.KnowledgeSubjectRef {
	return internalRelationshipRef("incident-responds-to-situation/"+incident.incidentID.String()+"/"+situation.situationID.String(), knr.PredicateRespondsTo, incidentEntityRef(incident), situationEntityRef(situation))
}
*/

/*
func (s *SituationService) RebuildSituationProjection(ctx context.Context) error {
	client := s.db.Client(ctx)
	projectionAliasQuery := client.KnowledgeSubjectAlias.Query().
		Where(
			ksa.ProviderEQ(internalProvider),
			ksa.SubjectKindEQ(ksa.SubjectKindRelationship),
		)
	projectionAliases, aliasesErr := projectionAliasQuery.All(ctx)
	if aliasesErr != nil {
		return fmt.Errorf("load situation projection aliases: %w", aliasesErr)
	}
	for _, alias := range projectionAliases {
		ref := rez.ProviderResourceRef{
			Provider:          alias.Provider,
			ProviderNamespace: alias.ProviderNamespace,
			ResourceRef:       alias.ProviderResourceRef,
		}
		if removeErr := s.knowledge.RemoveInternalSubject(ctx, ref); removeErr != nil {
			return fmt.Errorf("remove situation projection %q: %w", alias.ProviderResourceRef, removeErr)
		}
	}

	assessments, assessmentsErr := client.SituationHazardAssessment.Query().
		Order(sha.ByAssessedAt(sql.OrderDesc()), sha.ByRevision(sql.OrderDesc()), sha.ByID(sql.OrderDesc())).
		All(ctx)
	if assessmentsErr != nil {
		return fmt.Errorf("load situation hazard assessments: %w", assessmentsErr)
	}
	latestStatus := make(map[string]sha.Status)
	latestConfirmed := make(map[string]uuid.UUID)
	for _, assessment := range assessments {
		key := assessment.SituationID.String() + "\x1f" + assessment.SystemHazardID.String()
		if _, exists := latestStatus[key]; exists {
			continue
		}
		latestStatus[key] = assessment.Status
		if assessment.Status == sha.StatusConfirmed {
			latestConfirmed[key] = assessment.SystemHazardID
		}
	}
	for key, systemHazardID := range latestConfirmed {
		situationID := uuid.MustParse(strings.Split(key, "\x1f")[0])
		if ingestErr := s.ingestSituationClassifiedAsHazard(ctx, situationID, systemHazardID); ingestErr != nil {
			return fmt.Errorf("rebuild situation hazard classification: %w", ingestErr)
		}
	}

	incidentsQuery := client.Incident.Query().WithSituations()
	incidents, incidentsErr := incidentsQuery.All(ctx)
	if incidentsErr != nil {
		return fmt.Errorf("load incident situations: %w", incidentsErr)
	}
	for _, incident := range incidents {
		situations, _ := incident.Edges.SituationsOrErr()
		for _, situation := range situations {
			if ingestErr := s.IngestIncidentRespondsToSituation(ctx, incident.ID, situation.ID); ingestErr != nil {
				return fmt.Errorf("rebuild incident situation link: %w", ingestErr)
			}
		}
	}
	return nil
}
*/

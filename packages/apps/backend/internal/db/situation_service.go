package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	sha "github.com/rezible/rezible/ent/situationhazardassessment"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"

	ale "github.com/rezible/rezible/ent/alertepisode"
	inc "github.com/rezible/rezible/ent/incident"
	nev "github.com/rezible/rezible/ent/normalizedevent"
	sit "github.com/rezible/rezible/ent/situation"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	sitog "github.com/rezible/rezible/ent/situationobservationgroup"
)

const (
	situationHazardAssessmentLockNamespace = "situation_hazard_assessment"
	situationLockNamespace                 = "situation"
	defaultSituationInvestigationQuestion  = "Explain what is happening in this situation."
)

type SituationService struct {
	db             rez.Database
	investigations rez.InvestigationService
	analyses       rez.SystemAnalysisService
	graph          rez.KnowledgeGraphQueryService
}

func NewSituationService(db rez.Database, investigations rez.InvestigationService, analyses rez.SystemAnalysisService, graph rez.KnowledgeGraphQueryService) (*SituationService, error) {
	return &SituationService{db: db, investigations: investigations, analyses: analyses, graph: graph}, nil
}

func (s *SituationService) ListSituations(ctx context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
	query := s.db.Client(ctx).Situation.Query().
		WithInvestigation().
		WithObservationGroups(func(q *ent.SituationObservationGroupQuery) {
			q.Order(sitog.ByID())
		}).
		Order(sit.ByOpenedAt(params.GetOrder()), sit.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query = query.Where(sit.TitleContainsFold(search))
	}
	if params.HasInvestigation != nil {
		pred := sit.HasInvestigation()
		if !*params.HasInvestigation {
			pred = sit.Not(pred)
		}
		query = query.Where(pred)
	}
	if params.Active != nil {
		if *params.Active {
			query = query.Where(sit.ClosedAtIsNil())
		} else {
			query = query.Where(sit.ClosedAtNotNil())
		}
	}
	if params.OpenedAfter != nil {
		query = query.Where(sit.OpenedAtGTE(*params.OpenedAfter))
	}
	return ent.DoListQuery[ent.Situation, *ent.SituationQuery](ctx, query, params.ListParams)
}

func (s *SituationService) GetSituation(ctx context.Context, id uuid.UUID) (*ent.Situation, error) {
	return s.db.Client(ctx).Situation.Query().
		Where(sit.ID(id)).
		WithInvestigation(func(q *ent.SituationInvestigationQuery) {
			q.WithInvestigation()
		}).
		WithObservationGroups(func(q *ent.SituationObservationGroupQuery) {
			q.WithEvents()
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
	groups, eventIDs, episodeIDs, normalizeErr := s.normalizeSituationObservationGroups(params.ObservationGroups)
	if normalizeErr != nil {
		return nil, normalizeErr
	}
	if len(eventIDs)+len(episodeIDs) == 0 {
		return nil, fmt.Errorf("%w: situation requires at least one normalized event or alert episode", rez.ErrInvalidInput)
	}
	incidentIDs := mapset.NewSet(params.IncidentIDs...).ToSlice()

	var result *ent.Situation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if validateErr := s.validateSituationSources(ctx, eventIDs, episodeIDs); validateErr != nil {
			return validateErr
		}
		if validateErr := s.validateSituationIncidents(ctx, incidentIDs); validateErr != nil {
			return validateErr
		}

		createSituation := tx.Situation.Create().
			SetTitle(title).
			SetOpenedAt(openedAt).
			AddIncidentIDs(incidentIDs...)
		if summary := strings.TrimSpace(params.Summary); summary != "" {
			createSituation.SetSummary(summary)
		}
		created, saveErr := createSituation.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create situation: %w", saveErr)
		}

		createGroups := tx.SituationObservationGroup.MapCreateBulk(groups, func(c *ent.SituationObservationGroupCreate, i int) {
			group := groups[i]
			c.SetSituationID(created.ID)
			c.SetTitle(group.Title)
			c.AddEventIDs(group.NormalizedEventIDs...)
			c.AddAlertEpisodeIDs(group.AlertEpisodeIDs...)
			c.SetNillableBody(group.Body)
		})
		if groupsErr := createGroups.Exec(ctx); groupsErr != nil {
			return fmt.Errorf("create situation observation groups: %w", groupsErr)
		}

		analysis, createAnalysisErr := s.analyses.SetSystemAnalysis(ctx, uuid.Nil, func(*ent.SystemAnalysisMutation) {})
		if createAnalysisErr != nil {
			return fmt.Errorf("create situation investigation analysis: %w", createAnalysisErr)
		}
		if prepareErr := s.prepareOrRefreshSituationAnalysis(ctx, analysis.ID, eventIDs, episodeIDs); prepareErr != nil {
			return fmt.Errorf("prepare situation investigation analysis: %w", prepareErr)
		}
		investigation, createInvestigationErr := s.investigations.CreateInvestigation(ctx, rez.CreateInvestigationParams{
			AnalysisID: analysis.ID,
			Query:      defaultSituationInvestigationQuestion,
		})
		if createInvestigationErr != nil {
			return fmt.Errorf("create situation investigation: %w", createInvestigationErr)
		}
		createLink := tx.SituationInvestigation.Create().
			SetSituationID(created.ID).
			SetInvestigationID(investigation.ID)
		if linkErr := createLink.Exec(ctx); linkErr != nil {
			return fmt.Errorf("link situation investigation: %w", linkErr)
		}

		querySituation := tx.Situation.Query().
			Where(sit.ID(created.ID)).
			WithInvestigation(func(q *ent.SituationInvestigationQuery) {
				q.WithInvestigation()
			}).
			WithObservationGroups().
			WithIncidents()
		loaded, loadErr := querySituation.Only(ctx)
		if loadErr != nil {
			return fmt.Errorf("load created situation: %w", loadErr)
		}
		result = loaded.Unwrap()
		return nil
	})
}

func (s *SituationService) AddSituationObservationGroup(ctx context.Context, situationID uuid.UUID, params rez.SituationObservationGroupParams) (*ent.SituationObservationGroup, error) {
	groups, eventIDs, episodeIDs, normalizeErr := s.normalizeSituationObservationGroups([]rez.SituationObservationGroupParams{params})
	if normalizeErr != nil {
		return nil, normalizeErr
	}
	if len(eventIDs)+len(episodeIDs) == 0 {
		return nil, fmt.Errorf("%w: observation group requires at least one normalized event or alert episode", rez.ErrInvalidInput)
	}

	var result *ent.SituationObservationGroup
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, situationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}
		querySituation := tx.Situation.Query().Where(sit.ID(situationID))
		if situationExists, queryErr := querySituation.Exist(ctx); queryErr != nil || !situationExists {
			if ent.IsNotFound(queryErr) {
				return fmt.Errorf("%w: situation not found", rez.ErrNotFound)
			}
			return fmt.Errorf("get situation: %w", queryErr)
		}
		if validateErr := s.validateSituationSources(ctx, eventIDs, episodeIDs); validateErr != nil {
			return validateErr
		}

		groupsQuery := tx.SituationObservationGroup.Query().
			Where(sitog.SituationID(situationID)).
			WithEvents().
			WithAlertEpisodes()
		existingGroups, queryGroupsErr := groupsQuery.All(ctx)
		if queryGroupsErr != nil {
			return fmt.Errorf("load existing situation evidence links: %w", queryGroupsErr)
		}

		linkedEvents := mapset.NewSet[uuid.UUID]()
		linkedEpisodes := mapset.NewSet[uuid.UUID]()
		for _, group := range existingGroups {
			for _, event := range group.Edges.Events {
				linkedEvents.Add(event.ID)
			}
			for _, episode := range group.Edges.AlertEpisodes {
				linkedEpisodes.Add(episode.ID)
			}
		}

		// TODO: use mapset functions for these loops

		newEventIDs := make([]uuid.UUID, 0, len(eventIDs))
		for _, eventID := range eventIDs {
			if !linkedEvents.Contains(eventID) {
				newEventIDs = append(newEventIDs, eventID)
			}
		}

		newEpisodeIDs := make([]uuid.UUID, 0, len(episodeIDs))
		for _, episodeID := range episodeIDs {
			if !linkedEpisodes.Contains(episodeID) {
				newEpisodeIDs = append(newEpisodeIDs, episodeID)
			}
		}
		if len(newEventIDs)+len(newEpisodeIDs) == 0 {
			return nil
		}

		group := groups[0]
		createGroup := tx.SituationObservationGroup.Create().
			SetSituationID(situationID).
			SetTitle(group.Title).
			AddEventIDs(newEventIDs...).
			AddAlertEpisodeIDs(newEpisodeIDs...).
			SetNillableBody(group.Body)
		created, saveErr := createGroup.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("create situation observation group: %w", saveErr)
		}
		result = created.Unwrap()

		queryInvLink := tx.SituationInvestigation.Query().
			Where(siti.SituationID(situationID)).
			WithInvestigation()
		investigationLink, queryLinkErr := queryInvLink.Only(ctx)
		if queryLinkErr != nil {
			return fmt.Errorf("load situation investigation: %w", queryLinkErr)
		}
		analysisID := investigationLink.Edges.Investigation.SystemAnalysisID
		if prepareErr := s.prepareOrRefreshSituationAnalysis(ctx, analysisID, newEventIDs, newEpisodeIDs); prepareErr != nil {
			return fmt.Errorf("refresh situation investigation analysis: %w", prepareErr)
		}
		for _, episodeID := range newEpisodeIDs {
			revParams := rez.RecordInvestigationEvidenceRevisionParams{
				InvestigationID: investigationLink.InvestigationID,
				CallerKey:       "situation:" + situationID.String() + ":episode:" + episodeID.String(),
				Explanation:     "Alert episode " + episodeID.String() + " was added to this situation.",
			}
			if _, revisionErr := s.investigations.RecordInvestigationEvidenceRevision(ctx, revParams); revisionErr != nil {
				return fmt.Errorf("record investigation evidence revision for alert episode: %w", revisionErr)
			}
		}
		return nil
	})
}

func (s *SituationService) normalizeSituationObservationGroups(input []rez.SituationObservationGroupParams) ([]rez.SituationObservationGroupParams, []uuid.UUID, []uuid.UUID, error) {
	groups := make([]rez.SituationObservationGroupParams, len(input))
	eventIDs := make([]uuid.UUID, 0)
	episodeIDs := make([]uuid.UUID, 0)
	seenEvents := mapset.NewSet[uuid.UUID]()
	seenEpisodes := mapset.NewSet[uuid.UUID]()
	for i, inputGroup := range input {
		group := rez.SituationObservationGroupParams{
			Title: strings.TrimSpace(inputGroup.Title),
			Body:  inputGroup.Body,
		}
		for _, eventID := range inputGroup.NormalizedEventIDs {
			if eventID != uuid.Nil && seenEvents.Add(eventID) {
				group.NormalizedEventIDs = append(group.NormalizedEventIDs, eventID)
				eventIDs = append(eventIDs, eventID)
			}
		}
		for _, episodeID := range inputGroup.AlertEpisodeIDs {
			if episodeID != uuid.Nil && seenEpisodes.Add(episodeID) {
				group.AlertEpisodeIDs = append(group.AlertEpisodeIDs, episodeID)
				episodeIDs = append(episodeIDs, episodeID)
			}
		}
		groups[i] = group
	}
	return groups, eventIDs, episodeIDs, nil
}

func (s *SituationService) validateSituationSources(ctx context.Context, eventIDs, episodeIDs []uuid.UUID) error {
	if len(eventIDs) > 0 {
		queryEvents := s.db.Client(ctx).NormalizedEvent.Query().
			Where(nev.IDIn(eventIDs...))
		count, queryErr := queryEvents.Count(ctx)
		if queryErr != nil {
			return fmt.Errorf("validate normalized event access: %w", queryErr)
		}
		if count != len(eventIDs) {
			return fmt.Errorf("%w: normalized event not found", rez.ErrNotFound)
		}
	}
	if len(episodeIDs) > 0 {
		queryEpisodes := s.db.Client(ctx).AlertEpisode.Query().
			Where(ale.IDIn(episodeIDs...))
		count, queryErr := queryEpisodes.Count(ctx)
		if queryErr != nil {
			return fmt.Errorf("validate alert episode access: %w", queryErr)
		}
		if count != len(episodeIDs) {
			return fmt.Errorf("%w: alert episode not found", rez.ErrNotFound)
		}
	}
	return nil
}

func (s *SituationService) validateSituationIncidents(ctx context.Context, incidentIDs []uuid.UUID) error {
	if len(incidentIDs) == 0 {
		return nil
	}
	queryIncidents := s.db.Client(ctx).Incident.Query().
		Where(inc.IDIn(incidentIDs...))
	count, queryErr := queryIncidents.Count(ctx)
	if queryErr != nil {
		return fmt.Errorf("validate situation incident access: %w", queryErr)
	}
	if count != len(incidentIDs) {
		return fmt.Errorf("%w: incident not found", rez.ErrNotFound)
	}
	return nil
}

func (s *SituationService) RefreshSituationEpisodeAnalysis(ctx context.Context, situationID, episodeID uuid.UUID) error {
	if situationID == uuid.Nil || episodeID == uuid.Nil {
		return fmt.Errorf("%w: situation and alert episode IDs are required", rez.ErrInvalidInput)
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, situationID.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		queryObservationGroup := tx.SituationObservationGroup.Query().
			Where(sitog.SituationID(situationID), sitog.HasAlertEpisodesWith(ale.ID(episodeID)))
		linked, queryLinkedErr := queryObservationGroup.Exist(ctx)
		if queryLinkedErr != nil {
			return fmt.Errorf("check alert episode situation link: %w", queryLinkedErr)
		}
		if !linked {
			return nil
		}

		queryLink := tx.SituationInvestigation.Query().
			Where(siti.SituationID(situationID)).
			WithInvestigation()
		link, queryLinkErr := queryLink.Only(ctx)
		if queryLinkErr != nil {
			return fmt.Errorf("load situation investigation: %w", queryLinkErr)
		}

		prepErr := s.prepareOrRefreshSituationAnalysis(ctx, link.Edges.Investigation.SystemAnalysisID, nil, []uuid.UUID{episodeID})
		if prepErr != nil {
			return fmt.Errorf("prepare or refresh: %w", prepErr)
		}
		return nil
	})
}

func (s *SituationService) CloseSituation(ctx context.Context, id uuid.UUID, reason sit.CloseReason) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}
	if reason != sit.CloseReasonStabilized && reason != sit.CloseReasonDismissed {
		return fmt.Errorf("%w: invalid situation close reason", rez.ErrInvalidInput)
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, id.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}
		current, queryErr := tx.Situation.Get(ctx, id)
		if queryErr != nil {
			return fmt.Errorf("get situation: %w", queryErr)
		}
		if current.ClosedAt != nil {
			if current.CloseReason != nil && *current.CloseReason == reason {
				return nil
			}
			return fmt.Errorf("%w: situation is already closed with another reason", rez.ErrConflict)
		}
		update := current.Update().
			SetCloseReason(reason).
			SetClosedAt(time.Now().UTC())
		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("close situation: %w", updateErr)
		}
		return nil
	})
}

func (s *SituationService) AddIncidentToSituation(ctx context.Context, situationID, incidentID uuid.UUID) error {
	if situationID == uuid.Nil || incidentID == uuid.Nil {
		return fmt.Errorf("%w: situation and incident IDs are required", rez.ErrInvalidInput)
	}
	if validateErr := s.validateSituationIncidents(ctx, []uuid.UUID{incidentID}); validateErr != nil {
		return fmt.Errorf("validate incident access: %w", validateErr)
	}
	return s.db.Client(ctx).Situation.UpdateOneID(situationID).
		AddIncidentIDs(incidentID).
		Exec(ctx)
}

func (s *SituationService) RemoveIncidentFromSituation(ctx context.Context, situationID, incidentID uuid.UUID) error {
	if situationID == uuid.Nil || incidentID == uuid.Nil {
		return fmt.Errorf("%w: situation and incident IDs are required", rez.ErrInvalidInput)
	}
	if validateErr := s.validateSituationIncidents(ctx, []uuid.UUID{incidentID}); validateErr != nil {
		return fmt.Errorf("validate incident access: %w", validateErr)
	}
	return s.db.Client(ctx).Situation.UpdateOneID(situationID).
		RemoveIncidentIDs(incidentID).
		Exec(ctx)
}

func (s *SituationService) ListSituationHazardAssessments(ctx context.Context, params rez.ListSituationHazardAssessmentsParams) (*ent.ListResult[ent.SituationHazardAssessment], error) {
	order := params.GetOrder()
	query := s.db.Client(ctx).SituationHazardAssessment.Query().
		Order(sha.ByAssessedAt(order), sha.ByRevision(order), sha.ByID(order))
	if params.SituationID != uuid.Nil {
		query = query.Where(sha.SituationID(params.SituationID))
	}
	if params.SystemHazardID != uuid.Nil {
		query = query.Where(sha.SystemHazardID(params.SystemHazardID))
	}
	return ent.DoListQuery[ent.SituationHazardAssessment, *ent.SituationHazardAssessmentQuery](ctx, query, params.ListParams)
}

var validSituationHazardAssessmentStatus = mapset.NewSet(sha.StatusSuspected, sha.StatusConfirmed, sha.StatusDisproven)

func (s *SituationService) AddSituationHazardAssessment(ctx context.Context, params rez.AddSituationHazardAssessmentParams) (*ent.SituationHazardAssessment, error) {
	if params.SituationID == uuid.Nil || params.SystemHazardID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation and system hazard IDs are required", rez.ErrInvalidInput)
	}
	if (params.UserID == nil) == (params.AgentTurnID == nil) {
		return nil, fmt.Errorf("%w: exactly one assessor is required", rez.ErrInvalidInput)
	}
	if !validSituationHazardAssessmentStatus.Contains(params.Status) {
		return nil, fmt.Errorf("%w: invalid situation hazard assessment status", rez.ErrInvalidInput)
	}
	summary := strings.TrimSpace(params.Summary)
	if summary == "" {
		return nil, fmt.Errorf("%w: assessment summary is required", rez.ErrInvalidInput)
	}
	assessedAt := params.AssessedAt
	if assessedAt.IsZero() {
		assessedAt = time.Now().UTC()
	}

	var assessment *ent.SituationHazardAssessment
	return assessment, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		pairKey := params.SituationID.String() + "\x1f" + params.SystemHazardID.String()
		if lockErr := s.db.AcquireTxLocks(ctx, situationHazardAssessmentLockNamespace, pairKey); lockErr != nil {
			return fmt.Errorf("lock situation hazard pair: %w", lockErr)
		}
		latestQuery := tx.SituationHazardAssessment.Query().
			Where(sha.SituationID(params.SituationID), sha.SystemHazardID(params.SystemHazardID)).
			Order(sha.ByRevision(sql.OrderDesc()))
		latest, latestErr := latestQuery.First(ctx)
		nextRevision := 1
		if latestErr == nil {
			nextRevision = latest.Revision + 1
		} else if !ent.IsNotFound(latestErr) {
			return fmt.Errorf("get latest situation hazard assessment: %w", latestErr)
		}
		createAssessment := tx.SituationHazardAssessment.Create().
			SetSituationID(params.SituationID).
			SetSystemHazardID(params.SystemHazardID).
			SetRevision(nextRevision).
			SetStatus(params.Status).
			SetSummary(summary).
			SetAssessedAt(assessedAt).
			SetNillableUserID(params.UserID).
			SetNillableAgentTurnID(params.AgentTurnID)
		created, saveErr := createAssessment.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("add situation hazard assessment: %w", saveErr)
		}
		assessment = created
		return nil
	})
}

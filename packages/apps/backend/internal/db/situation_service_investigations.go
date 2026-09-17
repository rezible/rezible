package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	agt "github.com/rezible/rezible/ent/agentturn"
	ale "github.com/rezible/rezible/ent/alertepisode"
	"github.com/rezible/rezible/ent/investigation"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/situation"
	sha "github.com/rezible/rezible/ent/situationhazardassessment"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	sog "github.com/rezible/rezible/ent/situationobservationgroup"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
)

func (s *SituationService) CreateSituationInvestigation(ctx context.Context, params rez.CreateSituationInvestigationParams) (*ent.SituationInvestigation, error) {
	situationId := params.SituationID
	if situationId == uuid.Nil {
		return nil, fmt.Errorf("%w: situation id is required", rez.ErrInvalidInput)
	}

	agentInput := rezai.InvestigationAgentSessionInput{}
	if params.Query != nil {
		if cleaned := strings.TrimSpace(*params.Query); cleaned != "" {
			agentInput.Query = new(cleaned)
		}
	}

	var result *ent.SituationInvestigation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, situationId.String()); lockErr != nil {
			return fmt.Errorf("lock situation: %w", lockErr)
		}

		existingSitInv, queryExistingSitInvErr := s.LookupSituationInvestigation(ctx, siti.SituationID(situationId))
		if queryExistingSitInvErr != nil && !ent.IsNotFound(queryExistingSitInvErr) {
			return fmt.Errorf("lookup existing situation investigation: %w", queryExistingSitInvErr)
		} else if existingSitInv != nil {
			result = existingSitInv.Unwrap()
			return nil
		}

		sit, querySituationErr := tx.Situation.Get(ctx, situationId)
		if querySituationErr != nil {
			return fmt.Errorf("get situation: %w", querySituationErr)
		}

		investigationParams := rez.CreateInvestigationParams{
			Query:           agentInput.Query,
			SubjectEntityID: &sit.KnowledgeEntityID,
		}
		inv, createInvErr := s.investigations.CreateInvestigation(ctx, investigationParams)
		if createInvErr != nil {
			return fmt.Errorf("create investigation: %w", createInvErr)
		}

		createLink := tx.SituationInvestigation.Create().
			SetSituation(sit).
			SetInvestigation(inv)
		if createLinkErr := createLink.Exec(ctx); createLinkErr != nil {
			return fmt.Errorf("link situation investigation: %w", createLinkErr)
		}

		sitInv, lookupSitInvErr := s.LookupSituationInvestigation(ctx, siti.SituationID(situationId))
		if lookupSitInvErr != nil {
			return fmt.Errorf("lookup situation investigation: %w", lookupSitInvErr)
		}

		if queueJobErr := s.requestBumpSituationInvestigation(ctx, sitInv.ID); queueJobErr != nil {
			return fmt.Errorf("queue investigation job: %w", queueJobErr)
		}

		result = sitInv.Unwrap()
		return nil
	})
}

func (s *SituationService) requestBumpSituationInvestigation(ctx context.Context, id uuid.UUID) error {
	args := jobs.BumpSituationInvestigation{SituationInvestigationID: id}
	if _, insertErr := s.jobs.Insert(ctx, args, nil); insertErr != nil {
		return fmt.Errorf("insert investigation reconciliation job: %w", insertErr)
	}
	return nil
}

func (s *SituationService) LookupSituationInvestigation(ctx context.Context, pred predicate.SituationInvestigation) (*ent.SituationInvestigation, error) {
	return s.db.Client(ctx).SituationInvestigation.Query().
		Where(pred).
		WithSituation().
		WithRequestedTurn().
		WithInvestigation().
		Only(ctx)
}

func (s *SituationService) ListSituationHazardAssessments(ctx context.Context, params rez.ListSituationHazardAssessmentsParams) (*ent.ListResult[ent.SituationHazardAssessment], error) {
	order := params.GetOrder()
	query := s.db.Client(ctx).SituationHazardAssessment.Query().
		Order(sha.ByAssessedAt(order), sha.ByRevision(order), sha.ByID(order))
	if params.SituationID != uuid.Nil {
		query.Where(sha.SituationID(params.SituationID))
	}
	if params.SystemHazardID != uuid.Nil {
		query.Where(sha.SystemHazardID(params.SystemHazardID))
	}
	return ent.DoListQuery[ent.SituationHazardAssessment, *ent.SituationHazardAssessmentQuery](ctx, query, params.ListParams)
}

var validSituationHazardAssessmentStatus = mapset.NewSet(sha.StatusSuspected, sha.StatusConfirmed, sha.StatusDisproven)

func (s *SituationService) AddSituationHazardAssessment(ctx context.Context, params rez.AddSituationHazardAssessmentParams) (*ent.SituationHazardAssessment, error) {
	if params.SituationID == uuid.Nil || params.SystemHazardID == uuid.Nil {
		return nil, fmt.Errorf("%w: situation and system hazard ids are required", rez.ErrInvalidInput)
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

	situationId := params.SituationID

	var assessment *ent.SituationHazardAssessment
	return assessment, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		pairKey := situationId.String() + "\x1f" + params.SystemHazardID.String()
		if lockErr := s.db.AcquireTxLocks(ctx, situationHazardAssessmentLockNamespace, pairKey); lockErr != nil {
			return fmt.Errorf("lock situation hazard pair: %w", lockErr)
		}

		nextRevision := 1
		latestQuery := tx.SituationHazardAssessment.Query().
			Where(sha.SituationID(situationId), sha.SystemHazardID(params.SystemHazardID)).
			Order(sha.ByRevision(sql.OrderDesc()))
		latest, latestErr := latestQuery.First(ctx)
		if latestErr == nil {
			nextRevision = latest.Revision + 1
		} else if !ent.IsNotFound(latestErr) {
			return fmt.Errorf("get latest situation hazard assessment: %w", latestErr)
		}

		createAssessment := tx.SituationHazardAssessment.Create().
			SetSituationID(situationId).
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
		assessment = created.Unwrap()

		hazardRelEnt := &situationKnowledgeRelationshipEntity{
			id:       params.SystemHazardID,
			category: kne.CategoryConcern,
			kind:     "system_hazard",
			pred:     knr.PredicateClassifiedAs,
			isTarget: true,
		}
		if params.Status == sha.StatusConfirmed {
			if relErr := s.ingestKnowledgeRelationship(ctx, situationId, *hazardRelEnt); relErr != nil {
				return fmt.Errorf("ingest situation hazard assessment: %w", relErr)
			}
		} else {
			// TODO: remove knowledge relationship ?
		}
		return nil
	})
}

func (s *SituationService) SetSituationInvestigationReport(ctx context.Context, params rez.SetSituationInvestigationReportParams) (*ent.SituationInvestigation, error) {
	var result *ent.SituationInvestigation
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		sitInv, lookupByTurnErr := s.LookupSituationInvestigation(ctx, siti.RequestedTurnID(params.AgentTurnID))
		if lookupByTurnErr != nil {
			if ent.IsNotFound(lookupByTurnErr) {
				return fmt.Errorf("%w: stale investigation turn", rez.ErrConflict)
			}
			return fmt.Errorf("lookup investigation for turn: %w", lookupByTurnErr)
		}

		if situLockErr := s.db.AcquireTxLocks(ctx, situationLockNamespace, sitInv.SituationID.String()); situLockErr != nil {
			return situLockErr
		}

		sessId := sitInv.Edges.Investigation.AgentSessionID
		if agentTurnLockErr := acquireAgentSessionTurnLock(ctx, s.db, sessId); agentTurnLockErr != nil {
			return agentTurnLockErr
		}

		var sitInvErr error
		sitInv, sitInvErr = s.LookupSituationInvestigation(ctx, siti.ID(sitInv.ID))
		if sitInvErr != nil {
			return fmt.Errorf("lookup investigation: %w", sitInvErr)
		}
		if sitInv.RequestedTurnID == nil || *sitInv.RequestedTurnID != params.AgentTurnID {
			return fmt.Errorf("%w: stale investigation turn", rez.ErrConflict)
		}

		if sitInv.CompletedRevision == sitInv.RequestedRevision {
			result = sitInv.Unwrap()
			return nil
		}

		inv, invErr := sitInv.Edges.InvestigationOrErr()
		if invErr != nil {
			return fmt.Errorf("investigation: %w", invErr)
		}

		turn, turnErr := sitInv.Edges.RequestedTurnOrErr()
		if turnErr != nil {
			return fmt.Errorf("requested turn: %w", turnErr)
		}

		if turn.AgentSessionID != inv.AgentSessionID || turn.Status != agt.StatusRunning {
			return fmt.Errorf("%w: report turn must be running", rez.ErrConflict)
		}
		params.Report.Text = strings.TrimSpace(params.Report.Text)
		if params.Report.Text == "" {
			return fmt.Errorf("%w: report text is required", rez.ErrInvalidInput)
		}

		createReport := tx.InvestigationReport.Create().
			SetInvestigationID(inv.ID).
			SetAgentTurnID(turn.ID).
			SetText(params.Report.Text).
			SetLikelyCause(params.Report.LikelyCause).
			SetBestNextStep(params.Report.BestNextStep)
		if params.Report.Limitations != nil {
			createReport.SetLimitations(params.Report.Limitations)
		}
		if params.Report.RecommendedActions != nil {
			createReport.SetRecommendedActions(params.Report.RecommendedActions)
		}
		if params.Report.SuggestedChecks != nil {
			createReport.SetSuggestedChecks(params.Report.SuggestedChecks)
		}
		// TODO: only allow 1 report?
		//upsert := createReport.OnConflict().UpdateNewValues()
		upsert := createReport
		if saveReportErr := upsert.Exec(ctx); saveReportErr != nil {
			return fmt.Errorf("save investigation report: %w", saveReportErr)
		}

		updateInvRevision := sitInv.Update().
			SetCompletedRevision(sitInv.RequestedRevision)
		if updateInvRevisionErr := updateInvRevision.Exec(ctx); updateInvRevisionErr != nil {
			return fmt.Errorf("update investigation revision: %w", updateInvRevisionErr)
		}

		for _, a := range params.Assessments {
			addAssessmentParams := rez.AddSituationHazardAssessmentParams{
				SituationID:    sitInv.SituationID,
				SystemHazardID: a.SystemHazardID,
				Status:         a.Status,
				Summary:        a.Summary,
				AgentTurnID:    &turn.ID,
			}
			if _, addAssmtErr := s.AddSituationHazardAssessment(ctx, addAssessmentParams); addAssmtErr != nil {
				return fmt.Errorf("add situation hazard assessment: %w", addAssmtErr)
			}
		}

		sitInv, sitInvErr = s.LookupSituationInvestigation(ctx, siti.ID(sitInv.ID))
		if sitInvErr != nil || sitInv == nil {
			return fmt.Errorf("lookup investigation: %w", sitInvErr)
		}

		result = sitInv.Unwrap()

		return nil
	})
}

func (s *SituationService) onAgentTurnUpdated(ctx context.Context, event *rezai.AgentTurnUpdated) error {
	if event.Status != agt.StatusCompleted && event.Status != agt.StatusFailed && event.Status != agt.StatusAborted {
		return nil
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		querySessionInvestigation := tx.Investigation.Query().
			Where(investigation.AgentSessionID(event.AgentSessionId)).
			WithSituations(func(lq *ent.SituationInvestigationQuery) {
				lq.WithSituation()
			})
		inv, queryInvErr := querySessionInvestigation.Only(ctx)
		if queryInvErr != nil {
			if ent.IsNotFound(queryInvErr) {
				return nil
			}
			return fmt.Errorf("lookup investigation: %w", queryInvErr)
		}
		for _, sitInv := range inv.Edges.Situations {
			if sitInv.Edges.Situation != nil && sitInv.Edges.Situation.EvidenceRevision > sitInv.RequestedRevision {
				return s.requestBumpSituationInvestigation(ctx, sitInv.ID)
			}
		}
		return nil
	})
}

func NewReconcileSituationInvestigationWorker(db rez.Database, s rez.SituationService, agents rez.AgentSessionService) *ReconcileSituationInvestigationWorker {
	return &ReconcileSituationInvestigationWorker{db: db, situations: s, agents: agents}
}

type ReconcileSituationInvestigationWorker struct {
	jobs.WorkerDefaults[jobs.BumpSituationInvestigation]
	db         rez.Database
	situations rez.SituationService
	agents     rez.AgentSessionService
}

func (w *ReconcileSituationInvestigationWorker) Work(ctx context.Context, job *jobs.Job[jobs.BumpSituationInvestigation]) error {
	investigationID := job.Args.SituationInvestigationID
	return w.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		link, queryInvestigationErr := tx.SituationInvestigation.Get(ctx, investigationID)
		if queryInvestigationErr != nil {
			if ent.IsNotFound(queryInvestigationErr) {
				return nil
			}
			return fmt.Errorf("lookup investigation: %w", queryInvestigationErr)
		}

		if situationLockErr := w.db.AcquireTxLocks(ctx, situationLockNamespace, link.SituationID.String()); situationLockErr != nil {
			return fmt.Errorf("lock situation before reconciliation: %w", situationLockErr)
		}
		link, rereadErr := tx.SituationInvestigation.Get(ctx, link.ID)
		if rereadErr != nil {
			return fmt.Errorf("reread situation investigation: %w", rereadErr)
		}

		inv, investigationErr := link.QueryInvestigation().WithAgentSession().Only(ctx)
		if investigationErr != nil {
			return fmt.Errorf("lookup linked investigation: %w", investigationErr)
		}
		querySituation := link.QuerySituation().
			Where(situation.EvidenceRevisionGT(link.CompletedRevision))
		sit, querySituationErr := querySituation.Only(ctx)
		if querySituationErr != nil {
			if ent.IsNotFound(querySituationErr) {
				return nil
			}
			return fmt.Errorf("get requested situation: %w", querySituationErr)
		}

		if sessionLockErr := acquireAgentSessionTurnLock(ctx, w.db, inv.AgentSessionID); sessionLockErr != nil {
			return fmt.Errorf("lock investigation session after situation: %w", sessionLockErr)
		}

		// TODO: AgentSessionService.GetSessionActiveTurn
		activeTurnQuery := tx.AgentTurn.Query().
			Where(agt.AgentSessionID(inv.AgentSessionID), agt.StatusIn(agt.StatusQueued, agt.StatusRunning))
		active, queryActiveTurnErr := activeTurnQuery.Exist(ctx)
		if queryActiveTurnErr != nil {
			return fmt.Errorf("lookup active agent turn: %w", queryActiveTurnErr)
		} else if active {
			return nil
		}

		if link.RequestedTurnID != nil && link.RequestedRevision == sit.EvidenceRevision {
			reqTurn, lookupRequestedTurnErr := tx.AgentTurn.Get(ctx, *link.RequestedTurnID)
			if lookupRequestedTurnErr != nil {
				return fmt.Errorf("lookup requested turn: %w", lookupRequestedTurnErr)
			}
			if reqTurn.Status == agt.StatusFailed || reqTurn.Status == agt.StatusAborted {
				return nil
			}
		}

		var sessionInput rezai.InvestigationAgentSessionInput
		if unmarshalErr := json.Unmarshal(inv.Edges.AgentSession.Input, &sessionInput); unmarshalErr != nil {
			return fmt.Errorf("decode investigation session input: %w", unmarshalErr)
		}
		query := ""
		if sessionInput.Query != nil {
			query = *sessionInput.Query
		}
		turnInput, inputErr := w.makeTurnInput(ctx, sit, query)
		if inputErr != nil {
			return fmt.Errorf("make turn input: %w", inputErr)
		}

		reqTurnParams := &rez.RequestAgentTurnParams{Input: turnInput}
		turn, requestTurnErr := w.agents.RequestAgentTurn(ctx, inv.AgentSessionID, reqTurnParams)
		if requestTurnErr != nil {
			return fmt.Errorf("request investigation agent turn: %w", requestTurnErr)
		}
		update := link.Update().
			SetRequestedTurnID(turn.ID).
			SetRequestedRevision(sit.EvidenceRevision)
		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update investigation: %w", updateErr)
		}

		return nil
	})
}

func (w *ReconcileSituationInvestigationWorker) makeTurnInput(ctx context.Context, sit *ent.Situation, query string) (*rez.AiAgentTurnInput, error) {
	episodesQuery := w.db.Client(ctx).SituationObservationGroup.Query().
		Where(sog.SituationID(sit.ID)).
		QueryAlertEpisodes().
		WithAlertDefinition().
		WithInstances(func(q *ent.AlertInstanceQuery) {
			q.WithEvent()
		}).
		Order(ale.ByStartedAt(), ale.ByID())
	episodes, queryEpisodesErr := episodesQuery.All(ctx)
	if queryEpisodesErr != nil {
		return nil, fmt.Errorf("load investigation evidence: %w", queryEpisodesErr)
	}

	input := rezai.InvestigationAgentTurnInput{
		Situation:     sit,
		AlertEpisodes: episodes,
		Query:         query,
	}

	return input.MakeTurnInput()
}

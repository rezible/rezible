package db

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/jobs"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/predicate"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	sha "github.com/rezible/rezible/ent/situationhazardassessment"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	sitog "github.com/rezible/rezible/ent/situationobservationgroup"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/situations"
)

const defaultSituationInvestigationQuestion = "Explain what is happening in this situation."

// SituationSignalService tells situations that signals changed. Signal sources depend on it, so they
// need not depend on situations.
type SituationSignalService struct {
	jobs rez.JobService
}

var _ rez.SituationSignalService = (*SituationSignalService)(nil)

func NewSituationSignalService(js rez.JobService) *SituationSignalService {
	return &SituationSignalService{jobs: js}
}

func (s *SituationSignalService) NotifySignalChanged(ctx context.Context, signalEntityID uuid.UUID) error {
	args := jobs.ProcessSituationSignal{SignalEntityID: signalEntityID}
	if _, insertErr := s.jobs.Insert(ctx, args, nil); insertErr != nil {
		return fmt.Errorf("insert process situation signal job: %w", insertErr)
	}
	return nil
}

type SituationService struct {
	db             rez.Database
	clock          rez.Clock
	jobs           rez.JobService
	investigations rez.InvestigationService
	analyses       rez.SystemAnalysisService
	graph          rez.KnowledgeGraphQueryService
	judge          rezai.JudgeSituationCandidateWorkflow
	sources        map[situations.SignalKind]situations.SituationSignalSource
}

func NewSituationService(db rez.Database, clock rez.Clock, js rez.JobService, investigations rez.InvestigationService, analyses rez.SystemAnalysisService, graph rez.KnowledgeGraphQueryService, judge rezai.JudgeSituationCandidateWorkflow, sources ...situations.SituationSignalSource) (*SituationService, error) {
	s := &SituationService{
		db:             db,
		clock:          clock,
		jobs:           js,
		investigations: investigations,
		analyses:       analyses,
		graph:          graph,
		judge:          judge,
		sources:        make(map[situations.SignalKind]situations.SituationSignalSource, len(sources)),
	}
	for _, source := range sources {
		s.sources[source.SituationSignalKind()] = source
	}
	return s, nil
}

func (s *SituationService) withSituationMembership(query *ent.SituationQuery) *ent.SituationQuery {
	return query.
		WithInvestigation().
		WithActions(func(q *ent.SituationActionQuery) {
			q.Where(sitact.ActionEQ(sitact.ActionRaised)).Order(sitact.ByID())
		}).
		WithObservationGroups(func(q *ent.SituationObservationGroupQuery) {
			q.Order(sitog.ByCreatedAt(), sitog.ByID())
			q.WithSignals(func(q *ent.SituationSignalQuery) {
				q.Order(sitsig.ByAttachedAt(), sitsig.ByID())
			})
		}).
		WithEntities(func(q *ent.SituationEntityQuery) {
			q.WithKnowledgeEntity()
		}).
		WithLatestJudgment(func(q *ent.SituationJudgmentQuery) {
			q.Select(
				sitjudg.FieldJudgedAt,
				sitjudg.FieldOutcome,
				sitjudg.FieldDecision,
				sitjudg.FieldReasons,
				sitjudg.FieldCitedReasons,
				sitjudg.FieldExplanation,
				sitjudg.FieldJudge,
			)
		})
}

func (s *SituationService) ListSituations(ctx context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
	query := s.withSituationMembership(s.db.Client(ctx).Situation.Query()).
		Order(sit.ByOpenedAt(params.GetOrder()), sit.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(sit.TitleContainsFold(search))
	}
	if len(params.Stages) > 0 {
		stagePredicates := make([]predicate.Situation, 0, len(params.Stages))
		for _, stage := range params.Stages {
			switch stage {
			case rez.SituationStageCandidate:
				stagePredicates = append(stagePredicates, sit.And(sit.ClosedAtIsNil(), sit.RaisedAtIsNil()))
			case rez.SituationStageRaised:
				stagePredicates = append(stagePredicates, sit.And(sit.ClosedAtIsNil(), sit.RaisedAtNotNil()))
			case rez.SituationStageClosed:
				stagePredicates = append(stagePredicates, sit.ClosedAtNotNil())
			default:
				return nil, fmt.Errorf("%w: unknown situation stage %q", errs.ErrInvalidInput, stage)
			}
		}
		query.Where(sit.Or(stagePredicates...))
	}
	if params.Muted != nil {
		if *params.Muted {
			query.Where(sit.MutedAtNotNil())
		} else {
			query.Where(sit.MutedAtIsNil())
		}
	}
	if params.OpenedAfter != nil {
		query.Where(sit.OpenedAtGTE(*params.OpenedAfter))
	}
	return ent.DoListQuery[ent.Situation, *ent.SituationQuery](ctx, query, params.ListParams)
}

func (s *SituationService) GetSituation(ctx context.Context, id uuid.UUID) (*ent.Situation, error) {
	query := s.withSituationMembership(s.db.Client(ctx).Situation.Query()).
		Where(sit.ID(id)).
		WithLinks(func(q *ent.SituationLinkQuery) {
			q.WithLinkedSituation()
		}).
		WithIncidents().
		WithActions(func(q *ent.SituationActionQuery) {
			q.Order(sitact.ByID())
		})
	return query.Only(ctx)
}

// ListSituationJudgments returns complete stored snapshots, newest first even when times tie.
func (s *SituationService) ListSituationJudgments(ctx context.Context, params rez.ListSituationJudgmentsParams) (*ent.ListResult[ent.SituationJudgment], error) {
	query := s.db.Client(ctx).SituationJudgment.Query().
		Where(sitjudg.SituationID(params.SituationID)).
		Order(sitjudg.ByJudgedAt(sql.OrderDesc()), sitjudg.ByID(sql.OrderDesc()))
	return ent.DoListQuery[ent.SituationJudgment, *ent.SituationJudgmentQuery](ctx, query, params.ListParams)
}

func (s *SituationService) CreateSituation(ctx context.Context, params rez.CreateSituationParams) (*ent.Situation, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: situation title is required", errs.ErrInvalidInput)
	}
	groups, signalIDs, normalizeErr := s.normalizeObservationGroups(params.ObservationGroups)
	if normalizeErr != nil {
		return nil, normalizeErr
	}
	if !slices.Contains(signalIDs, params.SeedEntityID) {
		return nil, fmt.Errorf("%w: the seed must be one of the situation's signals", errs.ErrInvalidInput)
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Situation, error) {
		now := s.clock.Now()
		signals, loadErr := s.loadSignals(ctx, signalIDs, situations.LoadSignalsOptions{})
		if loadErr != nil {
			return nil, loadErr
		}
		createSituation := tx.Situation.Create().
			SetTitle(title).
			SetOpenedAt(s.earliestStart(signals)).
			SetSeedEntityID(params.SeedEntityID).
			SetCreatedAt(now).
			SetUpdatedAt(now)
		if summary := strings.TrimSpace(params.Summary); summary != "" {
			createSituation.SetSummary(summary)
		}
		created, createErr := createSituation.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create situation: %w", createErr)
		}

		for _, group := range groups {
			createGroup := tx.SituationObservationGroup.Create().
				SetSituationID(created.ID).
				SetTitle(group.Title).
				SetNillableBody(group.Body)
			createdGroup, groupErr := createGroup.Save(ctx)
			if groupErr != nil {
				return nil, fmt.Errorf("create situation observation group: %w", groupErr)
			}
			for _, signalID := range group.SignalEntityIDs {
				membership := situationMembership{groupID: createdGroup.ID, matchKind: sitsig.MatchKindManual}
				if signalID == params.SeedEntityID {
					membership.matchKind = sitsig.MatchKindSeed
				}
				memberSignals := []situations.Signal{signals[signalID]}
				if memberErr := s.createMemberships(ctx, created.ID, membership, memberSignals, now); memberErr != nil {
					return nil, memberErr
				}
			}
		}
		if entitiesErr := s.recomputeSituationEntities(ctx, created.ID); entitiesErr != nil {
			return nil, entitiesErr
		}
		if params.Raise != nil {
			if raiseErr := s.raiseLocked(ctx, created, *params.Raise, now); raiseErr != nil {
				return nil, raiseErr
			}
		}
		if evaluateErr := s.requestEvaluation(ctx, created.ID); evaluateErr != nil {
			return nil, evaluateErr
		}
		return s.GetSituation(ctx, created.ID)
	})
}

// normalizeObservationGroups trims group titles and keeps each signal in the first group naming it.
func (s *SituationService) normalizeObservationGroups(input []rez.SituationObservationGroupParams) ([]rez.SituationObservationGroupParams, []uuid.UUID, error) {
	groups := make([]rez.SituationObservationGroupParams, 0, len(input))
	var signalIDs []uuid.UUID
	seen := mapset.NewSet[uuid.UUID]()
	for _, inputGroup := range input {
		group := rez.SituationObservationGroupParams{Title: strings.TrimSpace(inputGroup.Title), Body: inputGroup.Body}
		if group.Title == "" {
			return nil, nil, fmt.Errorf("%w: observation group title is required", errs.ErrInvalidInput)
		}
		for _, signalID := range inputGroup.SignalEntityIDs {
			if signalID != uuid.Nil && seen.Add(signalID) {
				group.SignalEntityIDs = append(group.SignalEntityIDs, signalID)
				signalIDs = append(signalIDs, signalID)
			}
		}
		if len(group.SignalEntityIDs) > 0 {
			groups = append(groups, group)
		}
	}
	return groups, signalIDs, nil
}

// loadSignals loads each signal through the source for its knowledge entity's kind. An unknown entity or
// a kind without a source is invalid input.
func (s *SituationService) loadSignals(ctx context.Context, entityIDs []uuid.UUID, opts situations.LoadSignalsOptions) (map[uuid.UUID]situations.Signal, error) {
	queryEntities := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.IDIn(entityIDs...)).
		Select(kne.FieldID, kne.FieldKind)
	entities, queryErr := queryEntities.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query signal entities: %w", queryErr)
	}
	idsByKind := make(map[situations.SignalKind][]uuid.UUID)
	for _, entity := range entities {
		kind := situations.SignalKind(entity.Kind)
		idsByKind[kind] = append(idsByKind[kind], entity.ID)
	}
	signals := make(map[uuid.UUID]situations.Signal, len(entityIDs))
	for kind, ids := range idsByKind {
		source, supported := s.sources[kind]
		if !supported {
			return nil, fmt.Errorf("%w: unsupported signal kind %q", errs.ErrInvalidInput, kind)
		}
		loaded, loadErr := source.LoadSignals(ctx, ids, opts)
		if loadErr != nil {
			return nil, fmt.Errorf("load %s signals: %w", kind, loadErr)
		}
		for _, signal := range loaded {
			signals[signal.EntityID] = signal
		}
	}
	if len(signals) != len(entityIDs) {
		return nil, fmt.Errorf("%w: signal not found", errs.ErrInvalidInput)
	}
	return signals, nil
}

func (s *SituationService) earliestStart(signals map[uuid.UUID]situations.Signal) time.Time {
	var earliest time.Time
	for _, signal := range signals {
		if earliest.IsZero() || signal.StartedAt.Before(earliest) {
			earliest = signal.StartedAt
		}
	}
	return earliest
}

// lockSituations loads the situations with row locks, in ID order so concurrent writers lock them in the
// same order. Every situation write decides on the state it reads here.
func (s *SituationService) lockSituations(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]*ent.Situation, error) {
	lockQuery := s.db.Client(ctx).Situation.Query().
		Where(sit.IDIn(ids...)).
		Order(sit.ByID()).
		ForUpdate()
	locked, lockErr := lockQuery.All(ctx)
	if lockErr != nil {
		return nil, fmt.Errorf("lock situations: %w", lockErr)
	}
	byID := make(map[uuid.UUID]*ent.Situation, len(locked))
	for _, situation := range locked {
		byID[situation.ID] = situation
	}
	for _, id := range ids {
		if byID[id] == nil {
			return nil, fmt.Errorf("%w: situation not found", errs.ErrNotFound)
		}
	}
	return byID, nil
}

// recordAction appends a lifecycle decision, attributed to the acting user when there is one.
func (s *SituationService) recordAction(ctx context.Context, situationID uuid.UUID, action sitact.Action, reason string, now time.Time) error {
	createAction := s.db.Client(ctx).SituationAction.Create().
		SetSituationID(situationID).
		SetAction(action).
		SetReason(strings.TrimSpace(reason)).
		SetAt(now)
	if userID, hasUser := execution.GetContext(ctx).UserID(); hasUser {
		createAction.SetUserID(userID)
	}
	if createErr := createAction.Exec(ctx); createErr != nil {
		return fmt.Errorf("record situation %s action: %w", action, createErr)
	}
	return nil
}

// ListSituationEventIDs lists the events of the situation's signals, as their sources report them.
func (s *SituationService) ListSituationEventIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	memberIDs, membersErr := s.memberEntityIDs(ctx, id)
	if membersErr != nil || len(memberIDs) == 0 {
		return nil, membersErr
	}
	signals, loadErr := s.loadSignals(ctx, memberIDs, situations.LoadSignalsOptions{})
	if loadErr != nil {
		return nil, loadErr
	}
	eventIDs := mapset.NewSet[uuid.UUID]()
	for _, signal := range signals {
		eventIDs.Append(signal.EventIDs...)
	}
	return eventIDs.ToSlice(), nil
}

func (s *SituationService) memberEntityIDs(ctx context.Context, situationID uuid.UUID) ([]uuid.UUID, error) {
	queryMembers := s.db.Client(ctx).SituationSignal.Query().
		Where(sitsig.SituationID(situationID)).
		Select(sitsig.FieldKnowledgeEntityID)
	members, queryErr := queryMembers.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query situation signals: %w", queryErr)
	}
	ids := make([]uuid.UUID, len(members))
	for i, member := range members {
		ids[i] = member.KnowledgeEntityID
	}
	return ids, nil
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
		return nil, fmt.Errorf("%w: situation and system hazard IDs are required", errs.ErrInvalidInput)
	}
	if (params.UserID == nil) == (params.AgentTurnID == nil) {
		return nil, fmt.Errorf("%w: exactly one assessor is required", errs.ErrInvalidInput)
	}
	if !validSituationHazardAssessmentStatus.Contains(params.Status) {
		return nil, fmt.Errorf("%w: invalid situation hazard assessment status", errs.ErrInvalidInput)
	}
	summary := strings.TrimSpace(params.Summary)
	if summary == "" {
		return nil, fmt.Errorf("%w: assessment summary is required", errs.ErrInvalidInput)
	}
	assessedAt := params.AssessedAt
	if assessedAt.IsZero() {
		assessedAt = s.clock.Now()
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.SituationHazardAssessment, error) {
		// The situation's row lock serializes revision numbering for its hazards.
		if _, lockErr := s.lockSituations(ctx, params.SituationID); lockErr != nil {
			return nil, lockErr
		}
		latestQuery := tx.SituationHazardAssessment.Query().
			Where(sha.SituationID(params.SituationID), sha.SystemHazardID(params.SystemHazardID)).
			Order(sha.ByRevision(sql.OrderDesc()))
		latest, latestErr := latestQuery.First(ctx)
		nextRevision := 1
		if latestErr == nil {
			nextRevision = latest.Revision + 1
		} else if !ent.IsNotFound(latestErr) {
			return nil, fmt.Errorf("get latest situation hazard assessment: %w", latestErr)
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
		return createAssessment.Save(ctx)
	})
}

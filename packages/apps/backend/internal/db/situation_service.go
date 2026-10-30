package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/jobs"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/sequencecounter"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	sitnot "github.com/rezible/rezible/ent/situationnotice"
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	sitog "github.com/rezible/rezible/ent/situationobservationgroup"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
)

const defaultSituationInvestigationQuestion = "Explain what is happening in this situation."

// SituationObservationService tells situations that observations changed. Observation sources depend on
// it, so they need not depend on situations.
type SituationObservationService struct {
	jobs rez.JobService
}

var _ rez.SituationObservationService = (*SituationObservationService)(nil)

func NewSituationObservationService(js rez.JobService) *SituationObservationService {
	return &SituationObservationService{jobs: js}
}

func (s *SituationObservationService) NotifyObservationChanged(ctx context.Context, ref rez.SituationObservationRef) error {
	args := jobs.ProcessSituationObservation{Ref: ref}
	if _, insertErr := s.jobs.Insert(ctx, args, nil); insertErr != nil {
		return fmt.Errorf("insert process situation observation job: %w", insertErr)
	}
	return nil
}

type SituationService struct {
	cfg            rez.SituationsConfig
	db             rez.Database
	clock          rez.Clock
	jobs           rez.JobService
	msgs           rez.MessageQueue
	investigations rez.InvestigationService
	analyses       rez.SystemAnalysisService
	graph          rez.KnowledgeGraphQueryService
	judge          rezai.JudgeSituationCandidateWorkflow
	sources        map[schematypes.SituationObservationKind]rez.SituationObservationSource
}

func NewSituationService(cfg rez.SituationsConfig, db rez.Database, clock rez.Clock, js rez.JobService, msgs rez.MessageQueue, investigations rez.InvestigationService, analyses rez.SystemAnalysisService, graph rez.KnowledgeGraphQueryService, judge rezai.JudgeSituationCandidateWorkflow, sources ...rez.SituationObservationSource) (*SituationService, error) {
	if judge == nil {
		return nil, fmt.Errorf("situation judge workflow is required")
	}
	s := &SituationService{
		cfg:            cfg,
		db:             db,
		clock:          clock,
		jobs:           js,
		msgs:           msgs,
		investigations: investigations,
		analyses:       analyses,
		graph:          graph,
		judge:          judge,
		sources:        make(map[schematypes.SituationObservationKind]rez.SituationObservationSource, len(sources)),
	}
	for _, source := range sources {
		s.sources[source.SituationObservationKind()] = source
	}
	return s, nil
}

func (s *SituationService) withSituationMembership(query *ent.SituationQuery) *ent.SituationQuery {
	return query.
		WithActions(func(q *ent.SituationActionQuery) {
			q.Where(sitact.ActionEQ(sitact.ActionRaised)).Order(sitact.ByID())
		}).
		WithObservationGroups(func(q *ent.SituationObservationGroupQuery) {
			q.Order(sitog.ByCreatedAt(), sitog.ByID())
		}).
		WithObservations(func(q *ent.SituationObservationQuery) {
			q.Order(sitobs.ByAttachedAt(), sitobs.ByID())
		}).
		WithEntities(func(q *ent.SituationEntityQuery) {
			q.WithKnowledgeEntity()
		})
}

// latestJudgment returns the situation's most recent judgment, or nil if it has none.
func (s *SituationService) latestJudgment(ctx context.Context, situationID uuid.UUID) (*ent.SituationJudgment, error) {
	queryLatest := s.db.Client(ctx).SituationJudgment.Query().
		Where(sitjudg.SituationID(situationID)).
		Order(sitjudg.ByJudgedAt(sql.OrderDesc()), sitjudg.ByID(sql.OrderDesc()))
	latest, queryErr := queryLatest.First(ctx)
	if ent.IsNotFound(queryErr) {
		return nil, nil
	}
	if queryErr != nil {
		return nil, fmt.Errorf("query latest situation judgment: %w", queryErr)
	}
	return latest, nil
}

func (s *SituationService) ListSituations(ctx context.Context, params rez.ListSituationsParams) (*ent.ListResult[ent.Situation], error) {
	query := s.withSituationMembership(s.db.Client(ctx).Situation.Query()).
		Order(sit.ByOpenedAt(params.GetOrder()), sit.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		matches := []predicate.Situation{sit.TitleContainsFold(search)}
		if number, isReference := ent.ParseSituationReference(search); isReference {
			matches = append(matches, sit.Number(number))
		}
		query.Where(sit.Or(matches...))
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
	if params.InvestigationID != uuid.Nil {
		query.Where(sit.InvestigationID(params.InvestigationID))
	}
	return ent.DoListQuery[ent.Situation, *ent.SituationQuery](ctx, query, params.ListParams)
}

func (s *SituationService) GetSituation(ctx context.Context, pred predicate.Situation) (*ent.Situation, error) {
	query := s.withSituationMembership(s.db.Client(ctx).Situation.Query()).
		Where(pred).
		WithIncidents().
		WithActions(func(q *ent.SituationActionQuery) {
			q.Order(sitact.ByID())
		}).
		WithNotices(func(q *ent.SituationNoticeQuery) {
			q.Order(sitnot.ByCreatedAt(), sitnot.ByID())
		}).
		WithJudgments(func(q *ent.SituationJudgmentQuery) {
			// The limit applies across parents, so only one situation is read: Only fails on several.
			q.Order(sitjudg.ByJudgedAt(sql.OrderDesc()), sitjudg.ByID(sql.OrderDesc())).Limit(1)
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
	ungrouped, groups, refs, normalizeErr := s.normalizeObservations(params)
	if normalizeErr != nil {
		return nil, normalizeErr
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf("%w: a situation needs an observation", errs.ErrInvalidInput)
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Situation, error) {
		now := s.clock.Now()
		observations, loadErr := s.LoadObservations(ctx, refs, rez.LoadSituationObservationsOptions{})
		if loadErr != nil {
			return nil, loadErr
		}
		number, numberErr := tx.SequenceCounter.Next(ctx, sequencecounter.KindSituation)
		if numberErr != nil {
			return nil, fmt.Errorf("allocate situation number: %w", numberErr)
		}
		createSituation := tx.Situation.Create().
			SetNumber(number).
			SetTitle(title).
			SetOpenedAt(s.earliestStart(observations)).
			SetCreatedAt(now).
			SetUpdatedAt(now)
		if summary := strings.TrimSpace(params.Summary); summary != "" {
			createSituation.SetSummary(summary)
		}
		created, createErr := createSituation.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create situation: %w", createErr)
		}

		if observeErr := s.createObservations(ctx, created.ID, nil, ungrouped, "", now); observeErr != nil {
			return nil, observeErr
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
			if observeErr := s.createObservations(ctx, created.ID, &createdGroup.ID, group.Observations, "", now); observeErr != nil {
				return nil, observeErr
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
		return s.GetSituation(ctx, sit.ID(created.ID))
	})
}

// normalizeObservations trims group titles and keeps each observation in the first place naming it,
// ungrouped first. It returns the ungrouped observations, the nonempty groups, and every observation.
func (s *SituationService) normalizeObservations(params rez.CreateSituationParams) ([]rez.SituationObservationRef, []rez.SituationObservationGroupParams, []rez.SituationObservationRef, error) {
	seen := mapset.NewSet[rez.SituationObservationRef]()
	var all []rez.SituationObservationRef
	keep := func(refs []rez.SituationObservationRef) []rez.SituationObservationRef {
		var kept []rez.SituationObservationRef
		for _, ref := range refs {
			if ref.Kind() != "" && seen.Add(ref) {
				kept = append(kept, ref)
				all = append(all, ref)
			}
		}
		return kept
	}
	ungrouped := keep(params.Observations)
	groups := make([]rez.SituationObservationGroupParams, 0, len(params.ObservationGroups))
	for _, inputGroup := range params.ObservationGroups {
		group := rez.SituationObservationGroupParams{Title: strings.TrimSpace(inputGroup.Title), Body: inputGroup.Body}
		if group.Title == "" {
			return nil, nil, nil, fmt.Errorf("%w: observation group title is required", errs.ErrInvalidInput)
		}
		if group.Observations = keep(inputGroup.Observations); len(group.Observations) > 0 {
			groups = append(groups, group)
		}
	}
	return ungrouped, groups, all, nil
}

func (s *SituationService) earliestStart(observations map[rez.SituationObservationRef]rez.SituationObservation) time.Time {
	var earliest time.Time
	for _, observation := range observations {
		if earliest.IsZero() || observation.StartedAt.Before(earliest) {
			earliest = observation.StartedAt
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

// recordAction appends a lifecycle decision, attributed to the acting user when there is one, and says that
// the situation changed. Every lifecycle decision passes through here.
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
	if publishErr := s.msgs.Publish(ctx, rez.EventOnSituationUpdated{SituationID: situationID}); publishErr != nil {
		return fmt.Errorf("publish situation updated: %w", publishErr)
	}
	return nil
}

// ListSituationEventIDs lists the events of the situation's observations, as their sources report them.
func (s *SituationService) ListSituationEventIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	refs, refsErr := s.observedRefs(ctx, id)
	if refsErr != nil || len(refs) == 0 {
		return nil, refsErr
	}
	observations, loadErr := s.LoadObservations(ctx, refs, rez.LoadSituationObservationsOptions{})
	if loadErr != nil {
		return nil, loadErr
	}
	eventIDs := mapset.NewSet[uuid.UUID]()
	for _, observation := range observations {
		eventIDs.Append(observation.EventIDs...)
	}
	return eventIDs.ToSlice(), nil
}

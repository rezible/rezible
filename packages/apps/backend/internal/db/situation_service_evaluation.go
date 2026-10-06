package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	inc "github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	sitjudg "github.com/rezible/rezible/ent/situationjudgment"
	sitsig "github.com/rezible/rezible/ent/situationsignal"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/situations"
)

func NewEvaluateSituationWorker(service *SituationService) jobs.WorkerDefinition {
	return jobs.DefineWorkerFunc(func(ctx context.Context, args jobs.EvaluateSituation) error {
		return service.EvaluateSituation(ctx, args.SituationID)
	})
}

// requestEvaluation inserts an immediate evaluation of the situation in the caller's transaction.
func (s *SituationService) requestEvaluation(ctx context.Context, situationID uuid.UUID) error {
	args := jobs.EvaluateSituation{SituationID: situationID}
	if _, insertErr := s.jobs.Insert(ctx, args, nil); insertErr != nil {
		return fmt.Errorf("insert situation evaluation job: %w", insertErr)
	}
	return nil
}

// scheduleEvaluation inserts an evaluation of the situation due at the deadline, if there is one.
func (s *SituationService) scheduleEvaluation(ctx context.Context, situationID uuid.UUID, deadline *time.Time) error {
	if deadline == nil {
		return nil
	}
	dueAt := deadline.UTC()
	args := jobs.EvaluateSituation{SituationID: situationID, DueAt: &dueAt}
	if _, insertErr := s.jobs.Insert(ctx, args, &river.InsertOpts{ScheduledAt: dueAt}); insertErr != nil {
		return fmt.Errorf("schedule situation evaluation: %w", insertErr)
	}
	return nil
}

// EvaluateSituation brings an unclosed situation up to date as of one processing time, under its row lock:
// it closes the situation if the closure rule allows, refreshes its investigation for changed evidence,
// judges an unmuted candidate, and schedules the next deadline. Repeating it changes nothing.
//
// When the model judge is to decide a candidate, evaluation runs in three phases and holds no transaction or
// lock while the model runs: capture under the row lock, judge outside any transaction, then apply under the
// lock only if the facts are unchanged.
func (s *SituationService) EvaluateSituation(ctx context.Context, id uuid.UUID) error {
	captured, evaluateErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*candidateSnapshot, error) {
		return s.evaluateLocked(ctx, id)
	})
	if evaluateErr != nil || captured == nil {
		return evaluateErr
	}
	judgment, judgeErr := s.judgeByModel(ctx, captured)
	if judgeErr != nil {
		return judgeErr
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		return s.applyModelJudgment(ctx, captured, judgment)
	})
}

// evaluateLocked evaluates the situation in the caller's transaction. It returns the snapshot of a candidate
// the model judge is to decide, leaving its judgment and the next deadline to the later phases.
func (s *SituationService) evaluateLocked(ctx context.Context, id uuid.UUID) (*candidateSnapshot, error) {
	locked, lockErr := s.lockSituations(ctx, id)
	if lockErr != nil {
		return nil, lockErr
	}
	situation := locked[id]
	if situation.ClosedAt != nil {
		return nil, nil
	}
	now := s.clock.Now()
	input, members, inputErr := s.loadEvaluationInput(ctx, situation, now)
	if inputErr != nil {
		return nil, inputErr
	}
	if closeableAt := input.CloseableAt(); closeableAt != nil && !now.Before(*closeableAt) {
		return nil, s.closeLocked(ctx, situation, input.CloseReason(), "", *closeableAt, now)
	}
	if refreshErr := s.refreshInvestigation(ctx, situation, members, input); refreshErr != nil {
		return nil, refreshErr
	}
	if !input.Raised && !input.Muted {
		snapshot, snapshotErr := s.snapshotCandidate(ctx, situation, input)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		// While collecting, only a hard reason is judged, and a latest judgment of the same facts stands.
		collecting := input.Collecting() && snapshot.outcome != sitjudg.OutcomeRaise
		if !collecting && !input.JudgmentStands(snapshot.fingerprint) {
			if snapshot.input.Context != nil {
				return snapshot, nil
			}
			judgment := snapshot.reasons.JudgeByRules()
			if recordErr := s.recordJudgment(ctx, situation, snapshot, judgment, now); recordErr != nil {
				return nil, recordErr
			}
			input.Record(judgment, snapshot.fingerprint)
		}
	}
	return nil, s.scheduleEvaluation(ctx, id, input.NextDeadline())
}

// loadEvaluationInput loads the situation's members with their signals, including history, its linked
// incidents and, for an unmuted candidate, its latest judgment.
func (s *SituationService) loadEvaluationInput(ctx context.Context, situation *ent.Situation, now time.Time) (situations.EvaluationInput, []*ent.SituationSignal, error) {
	input := situations.EvaluationInput{
		Now:       now,
		OpenedAt:  situation.OpenedAt,
		CreatedAt: situation.CreatedAt,
		Raised:    situation.RaisedAt != nil,
		Muted:     situation.MutedAt != nil,
		HoldUntil: situation.HoldUntil,
	}
	client := s.db.Client(ctx)
	queryMembers := client.SituationSignal.Query().
		Where(sitsig.SituationID(situation.ID))
	members, membersErr := queryMembers.All(ctx)
	if membersErr != nil {
		return input, nil, fmt.Errorf("query situation signals: %w", membersErr)
	}
	memberIDs := make([]uuid.UUID, len(members))
	for i, member := range members {
		memberIDs[i] = member.KnowledgeEntityID
	}
	signals, loadErr := s.loadSignals(ctx, memberIDs, situations.LoadSignalsOptions{History: true})
	if loadErr != nil {
		return input, nil, loadErr
	}
	for _, member := range members {
		evaluated := situations.Member{
			Signal:     signals[member.KnowledgeEntityID],
			AttachedAt: member.AttachedAt,
			MatchKind:  member.MatchKind,
		}
		input.Members = append(input.Members, evaluated)
	}

	queryIncidents := client.Situation.QueryIncidents(situation).
		Select(inc.FieldID, inc.FieldTitle, inc.FieldResponseState, inc.FieldResolvedAt)
	incidents, incidentsErr := queryIncidents.All(ctx)
	if incidentsErr != nil {
		return input, nil, fmt.Errorf("query situation incidents: %w", incidentsErr)
	}
	for _, incident := range incidents {
		linked := situations.LinkedIncident{
			ID:            incident.ID,
			Title:         incident.Title,
			ResponseState: incident.ResponseState,
			ResolvedAt:    incident.ResolvedAt,
		}
		input.Incidents = append(input.Incidents, linked)
	}
	if !input.Raised && !input.Muted && situation.LatestJudgmentID != nil {
		queryLatest := client.SituationJudgment.Query().
			Where(sitjudg.ID(*situation.LatestJudgmentID)).
			Select(sitjudg.FieldJudgedAt, sitjudg.FieldJudge, sitjudg.FieldFingerprint)
		latest, latestErr := queryLatest.Only(ctx)
		if latestErr != nil {
			return input, nil, fmt.Errorf("query latest situation judgment: %w", latestErr)
		}
		input.Latest = &situations.LatestJudgment{
			JudgedAt:    latest.JudgedAt,
			Judge:       latest.Judge,
			Fingerprint: latest.Fingerprint,
		}
	}
	return input, members, nil
}

// refreshInvestigation refreshes the investigation of an unmuted situation for the members whose signal
// changed since it last saw them, and records that it saw them. A raised situation also records one evidence
// revision for the change. A muted situation is left alone, and catches up once after an unmute.
func (s *SituationService) refreshInvestigation(ctx context.Context, situation *ent.Situation, members []*ent.SituationSignal, input situations.EvaluationInput) error {
	if input.Muted {
		return nil
	}
	queryInvestigation := s.db.Client(ctx).SituationInvestigation.Query().
		Where(siti.SituationID(situation.ID)).
		WithInvestigation()
	link, queryErr := queryInvestigation.Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil
		}
		return fmt.Errorf("query situation investigation: %w", queryErr)
	}

	revisions := make(map[uuid.UUID]int, len(members))
	for _, member := range members {
		revisions[member.KnowledgeEntityID] = member.ObservedRevision
	}
	var changed []situations.Signal
	var keyParts, titles []string
	var changedEventIDs []uuid.UUID
	for _, evaluated := range input.Members {
		signal := evaluated.Signal
		if revisions[signal.EntityID] == signal.Revision {
			continue
		}
		changed = append(changed, signal)
		changedEventIDs = append(changedEventIDs, signal.EventIDs...)
		keyParts = append(keyParts, fmt.Sprintf("%s@%d", signal.EntityID, signal.Revision))
		titles = append(titles, signal.Title)
	}
	if len(changed) == 0 {
		return nil
	}
	analysisID := link.Edges.Investigation.SystemAnalysisID
	if refreshErr := s.prepareOrRefreshSituationAnalysis(ctx, analysisID, changedEventIDs); refreshErr != nil {
		return fmt.Errorf("refresh situation investigation analysis: %w", refreshErr)
	}
	if input.Raised {
		slices.Sort(keyParts)
		digest := sha256.Sum256([]byte(strings.Join(keyParts, ",")))
		revisionParams := rez.RecordInvestigationEvidenceRevisionParams{
			InvestigationID: link.InvestigationID,
			CallerKey:       "situation:" + situation.ID.String() + ":evidence:" + hex.EncodeToString(digest[:16]),
			Explanation:     "Evidence changed: " + strings.Join(titles, ", ") + ".",
		}
		if _, revisionErr := s.investigations.RecordInvestigationEvidenceRevision(ctx, revisionParams); revisionErr != nil {
			return fmt.Errorf("record investigation evidence revision: %w", revisionErr)
		}
	}
	return s.observeRevisions(ctx, situation.ID, changed)
}

// observeRevisions records that the situation's investigation saw the signals' current revisions.
func (s *SituationService) observeRevisions(ctx context.Context, situationID uuid.UUID, signals []situations.Signal) error {
	for _, signal := range signals {
		observeSignal := s.db.Client(ctx).SituationSignal.Update().
			Where(sitsig.SituationID(situationID), sitsig.KnowledgeEntityID(signal.EntityID)).
			SetObservedRevision(signal.Revision)
		if observeErr := observeSignal.Exec(ctx); observeErr != nil {
			return fmt.Errorf("observe situation signal revision: %w", observeErr)
		}
	}
	return nil
}

// candidateSnapshot is an unmuted candidate's decision inputs at one processing time.
type candidateSnapshot struct {
	situationID uuid.UUID
	input       situations.EvaluationInput
	facts       schematypes.SituationFacts
	reasons     situations.Reasons
	outcome     sitjudg.Outcome
	fingerprint string
}

// snapshotCandidate builds the candidate's facts, reasons, outcome and fingerprint. Only when the model judge
// is to decide, after collection, does it load the judge's context into the facts and their fingerprint.
func (s *SituationService) snapshotCandidate(ctx context.Context, situation *ent.Situation, input situations.EvaluationInput) (*candidateSnapshot, error) {
	outcomes, outcomesErr := s.priorOutcomes(ctx, situation.ID, input)
	if outcomesErr != nil {
		return nil, outcomesErr
	}
	input.PriorOutcomes = outcomes
	reasons := situations.CheckReasons(input.Facts())
	outcome := reasons.Outcome()
	if s.judge != nil && outcome == sitjudg.OutcomeNeedsDecision && !input.Collecting() {
		judgeContext, contextErr := s.loadJudgeContext(ctx, situation.ID, input)
		if contextErr != nil {
			return nil, contextErr
		}
		input.Context = judgeContext
	}
	facts := input.Facts()
	fingerprint, fingerprintErr := reasons.Fingerprint(facts)
	if fingerprintErr != nil {
		return nil, fingerprintErr
	}
	snapshot := &candidateSnapshot{
		situationID: situation.ID,
		input:       input,
		facts:       facts,
		reasons:     reasons,
		outcome:     outcome,
		fingerprint: fingerprint,
	}
	return snapshot, nil
}

// recordJudgment records the judgment of the snapshot's facts at the processing time, makes it the latest,
// and raises the candidate automatically when it decides to.
func (s *SituationService) recordJudgment(ctx context.Context, situation *ent.Situation, snapshot *candidateSnapshot, judgment situations.Judgment, now time.Time) error {
	client := s.db.Client(ctx)
	createJudgment := client.SituationJudgment.Create().
		SetSituationID(situation.ID).
		SetJudgedAt(now).
		SetOutcome(snapshot.outcome).
		SetDecision(judgment.Decision).
		SetReasons(snapshot.reasons).
		SetCitedReasons(judgment.Cited).
		SetFacts(snapshot.facts).
		SetExplanation(judgment.Explanation).
		SetJudge(judgment.Judge).
		SetFingerprint(snapshot.fingerprint)
	created, createErr := createJudgment.Save(ctx)
	if createErr != nil {
		return fmt.Errorf("record situation judgment: %w", createErr)
	}
	setLatest := client.Situation.UpdateOneID(situation.ID).
		SetLatestJudgmentID(created.ID)
	if setErr := setLatest.Exec(ctx); setErr != nil {
		return fmt.Errorf("set latest situation judgment: %w", setErr)
	}
	if judgment.Decision != sitjudg.DecisionRaise {
		return nil
	}
	return s.raiseAutomatically(ctx, situation, judgment.Cited, now)
}

func (s *SituationService) aggregateCountDistinctSituations(sel *sql.Selector) string {
	return sql.As(sql.Count(sql.Distinct(sel.C(sitsig.FieldSituationID))), "count")
}

// priorOutcomes counts, for each source of the members' signals, the other situations opened within
// PastIncidentLookback that hold a signal of it: raised, muted as not noteworthy, and linked to an incident.
// The database counts each outcome, grouped by source.
func (s *SituationService) priorOutcomes(ctx context.Context, situationID uuid.UUID, input situations.EvaluationInput) (map[uuid.UUID]schematypes.SituationPriorOutcomeFacts, error) {
	sources := mapset.NewSet[uuid.UUID]()
	for _, member := range input.Members {
		if source := member.Signal.SourceEntityID; source != nil {
			sources.Add(*source)
		}
	}
	outcomes := make(map[uuid.UUID]schematypes.SituationPriorOutcomeFacts, sources.Cardinality())
	if sources.IsEmpty() {
		return outcomes, nil
	}
	sourceIDs := sources.ToSlice()
	cutoff := input.Now.Add(-situations.PastIncidentLookback)
	client := s.db.Client(ctx)

	type matchingSituationSourceRow struct {
		SourceEntityID uuid.UUID `json:"source_entity_id"`
		Count          int       `json:"count"`
	}

	pIsSourceEntitySignal := sitsig.SourceEntityIDIn(sourceIDs...)
	// countSituationsBySource counts, for each of the sources, the matching situations that hold a signal of it.
	// A situation holding several signals of one source counts once for it.
	countSituationsBySource := func(matches predicate.Situation) (map[uuid.UUID]int, error) {
		pHasMatchingSituation := sitsig.HasSituationWith(sit.IDNEQ(situationID), sit.OpenedAtGTE(cutoff), matches)
		query := client.SituationSignal.Query().
			Where(pIsSourceEntitySignal, pHasMatchingSituation).
			GroupBy(sitsig.FieldSourceEntityID).
			Aggregate(s.aggregateCountDistinctSituations)

		var rows []matchingSituationSourceRow
		if scanErr := query.Scan(ctx, &rows); scanErr != nil {
			return nil, scanErr
		}

		counts := make(map[uuid.UUID]int, len(rows))
		for _, row := range rows {
			counts[row.SourceEntityID] = row.Count
		}
		return counts, nil
	}

	raised, raisedErr := countSituationsBySource(sit.RaisedAtNotNil())
	if raisedErr != nil {
		return nil, fmt.Errorf("count previously raised situations: %w", raisedErr)
	}

	muted, mutedErr := countSituationsBySource(sit.MuteReasonEQ(sit.MuteReasonNotNoteworthy))
	if mutedErr != nil {
		return nil, fmt.Errorf("count previously muted situations: %w", mutedErr)
	}

	incidentLinked, incidentErr := countSituationsBySource(sit.HasIncidents())
	if incidentErr != nil {
		return nil, fmt.Errorf("count incident-linked situations: %w", incidentErr)
	}

	for _, sourceID := range sourceIDs {
		outcomes[sourceID] = schematypes.SituationPriorOutcomeFacts{
			Raised:             raised[sourceID],
			MutedNotNoteworthy: muted[sourceID],
			IncidentLinked:     incidentLinked[sourceID],
		}
	}
	return outcomes, nil
}

// raiseAutomatically raises the candidate with no user, citing its reasons. It starts the investigation
// unless the tenant's automatic raises (raised actions with no user, other than incident links) within
// AutoInvestigationWindow have reached AutoInvestigationLimit; then a person can start it. Concurrent
// evaluations may overshoot the limit by the number running at once.
func (s *SituationService) raiseAutomatically(ctx context.Context, situation *ent.Situation, cited []schematypes.SituationRaiseReason, now time.Time) error {
	countRaises := s.db.Client(ctx).SituationAction.Query().
		Where(
			sitact.ActionEQ(sitact.ActionRaised),
			sitact.UserIDIsNil(),
			sitact.ReasonNEQ(situationIncidentRaiseReason),
			sitact.AtGT(now.Add(-situations.AutoInvestigationWindow)),
		)
	raises, countErr := countRaises.Count(ctx)
	if countErr != nil {
		return fmt.Errorf("count automatic situation raises: %w", countErr)
	}
	reasons := make([]string, len(cited))
	for i, reason := range cited {
		reasons[i] = string(reason)
	}
	params := rez.RaiseSituationParams{
		StartInvestigation: raises < situations.AutoInvestigationLimit,
		Reason:             strings.Join(reasons, ", "),
	}
	return s.raiseLocked(ctx, situation, params, now)
}

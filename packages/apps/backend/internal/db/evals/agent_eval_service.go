package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	at "github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/pkg/execution"
)

// scenarioEventSource is the provider event source of every scenario event.
const scenarioEventSource = "scenario"

// AgentEvalService is the base agent harness: it runs a scenario's repetitions against the running
// application, each in a new tenant, with events through the real pipeline and turns on the real workers, and
// hands what is particular to the agent to its harness.
type AgentEvalService struct {
	db           rez.Database
	clock        rez.Clock
	auth         rez.AuthSessionService
	integrations rez.IntegrationService
	ingester     rez.ProviderEventIngester
	projection   rez.EventProjectionService
	harnesses    map[string]AgentHarness
}

func NewAgentEvalService(
	db rez.Database,
	clock rez.Clock,
	auth rez.AuthSessionService,
	integrations rez.IntegrationService,
	ingester rez.ProviderEventIngester,
	projection rez.EventProjectionService,
	situations rez.SituationService,
	investigations rez.InvestigationService,
) *AgentEvalService {
	return &AgentEvalService{
		db:           db,
		clock:        clock,
		auth:         auth,
		integrations: integrations,
		ingester:     ingester,
		projection:   projection,
		harnesses: newAgentHarnesses(&agentHarnessDeps{
			db:             db,
			situations:     situations,
			investigations: investigations,
		}),
	}
}

// Run runs the named scenario runs times, one repetition after another. It fails only when the scenario does
// not load; a repetition's failures are recorded on it.
func (s *AgentEvalService) Run(ctx context.Context, name string, runs int) (*RunResult, error) {
	scenario, loadErr := loadScenario(name, s.clock.Now(), s.harnesses)
	if loadErr != nil {
		return nil, loadErr
	}
	result := &RunResult{Target: Target{Name: name, Kind: TargetKindAgent, Description: scenario.Description}}
	for index := 1; index <= runs; index++ {
		run := &agentRun{service: s, name: name, index: index, reportedTurns: mapset.NewSet[uuid.UUID]()}
		result.Repetitions = append(result.Repetitions, run.execute(ctx))
	}
	return result, nil
}

// agentRun is one repetition of a scenario, in its own tenant.
type agentRun struct {
	service *AgentEvalService
	name    string
	index   int

	repetition Repetition
	harness    AgentHarness
	// provider and namespace are the evaluation installation's, which every scenario event is ingested under.
	provider  string
	namespace string
	sessionID uuid.UUID
	// lastSequence is the newest turn recorded; reportedTurns are the turns recorded as turn_completed checks.
	lastSequence  int
	reportedTurns mapset.Set[uuid.UUID]
}

func (r *agentRun) execute(ctx context.Context) Repetition {
	r.repetition = Repetition{Index: r.index, Checks: []Check{}, Usage: []ModelUsage{}}
	attrs := []attribute.KeyValue{
		attribute.String("eval_target", r.name),
		attribute.Int("eval_repetition", r.index),
	}
	// The repetition's function never fails: its failures are recorded on the repetition.
	_ = execution.Do(ctx, "eval.repetition", func(ctx context.Context) error {
		if spanContext := trace.SpanContextFromContext(ctx); spanContext.HasTraceID() {
			r.repetition.TraceID = spanContext.TraceID().String()
		}
		started := time.Now()
		tenantCtx, runErr := r.run(ctx)
		r.repetition.Error = runErr
		if tenantCtx != nil && r.sessionID != uuid.Nil {
			if collectErr := r.collect(tenantCtx); collectErr != nil && r.repetition.Error == nil {
				r.repetition.Error = &RunError{Step: "collect", Message: collectErr.Error()}
			}
		}
		r.repetition.DurationMilliseconds = time.Since(started).Milliseconds()
		return nil
	}, attrs...)
	r.repetition.Status = repetitionStatus(r.repetition)
	return r.repetition
}

// run runs the scenario's steps in order and stops at the first that fails. It returns the tenant's context
// once there is one, so what the run left can be collected.
func (r *agentRun) run(ctx context.Context) (context.Context, *RunError) {
	scenario, loadErr := loadScenario(r.name, r.service.clock.Now(), r.service.harnesses)
	if loadErr != nil {
		return nil, &RunError{Step: "load", Message: loadErr.Error()}
	}
	r.harness = r.service.harnesses[scenario.Agent]
	tenantCtx, identityErr := r.enterTenant(ctx)
	if identityErr != nil {
		return nil, &RunError{Step: "identity", Message: identityErr.Error()}
	}
	ctx = tenantCtx

	install := func(ctx context.Context) error {
		return r.install(ctx, scenario)
	}
	if stepErr := r.step(ctx, "install", install); stepErr != nil {
		return ctx, stepErr
	}
	setup := func(ctx context.Context) error {
		return r.setup(ctx, scenario)
	}
	if stepErr := r.step(ctx, "setup", setup); stepErr != nil {
		return ctx, stepErr
	}
	checkIDs := scenario.PhaseCheckIDs()
	for index, phase := range scenario.Phases {
		runPhase := func(ctx context.Context) error {
			return r.runPhase(ctx, index, phase, checkIDs[index])
		}
		if stepErr := r.step(ctx, fmt.Sprintf("phase %d", index+1), runPhase); stepErr != nil {
			return ctx, stepErr
		}
	}
	return ctx, nil
}

// step runs fn as a traced unit of work, and names the step in its error.
func (r *agentRun) step(ctx context.Context, name string, fn func(context.Context) error) *RunError {
	if stepErr := execution.Do(ctx, "eval.step", fn, attribute.String("eval_step", name)); stepErr != nil {
		return &RunError{Step: name, Message: stepErr.Error()}
	}
	return nil
}

// enterTenant creates the repetition's organization and user, and works as that user.
func (r *agentRun) enterTenant(ctx context.Context) (context.Context, error) {
	key := uuid.NewString()
	providerSession := &rez.UserAuthProviderSession{
		User: ent.User{
			Email:          fmt.Sprintf("eval-%s@eval.rezible.com", key),
			Name:           "Evaluation",
			AuthProviderID: "eval-user-" + key,
		},
		Org: ent.Organization{
			Name:           fmt.Sprintf("Evaluation %s %d", r.name, r.index),
			AuthProviderID: "eval-org-" + key,
		},
		ExpiresAt: r.service.clock.Now().Add(24 * time.Hour),
	}
	// Creating a tenant is work across tenants.
	sess, sessErr := r.service.auth.CreateFromUserAuthResponse(execution.NewSystemContext(ctx), providerSession)
	if sessErr != nil {
		return nil, fmt.Errorf("create evaluation tenant: %w", sessErr)
	}
	return execution.NewUserContext(ctx, sess), nil
}

// install installs the evaluation integration with the scenario's logs.
func (r *agentRun) install(ctx context.Context, scenario *Scenario) error {
	config, encodeErr := json.Marshal(integrationConfig{Logs: scenario.Logs})
	if encodeErr != nil {
		return fmt.Errorf("encode installation config: %w", encodeErr)
	}
	installed, installErr := r.service.integrations.InstallNew(ctx, EvaluationProviderName, config)
	if installErr != nil {
		return fmt.Errorf("install the evaluation integration: %w", installErr)
	}
	r.provider = installed.Integration().Provider
	r.namespace = installed.Integration().ID.String()
	return nil
}

// setup ingests the setup events, recording the signals each produces instead of letting detection place
// them, then has the harness start the session and waits for its first turn.
func (r *agentRun) setup(ctx context.Context, scenario *Scenario) error {
	signals := make(map[string][]uuid.UUID, len(scenario.Events))
	for _, event := range scenario.Events {
		setupCtx, recorder := withSetupEvent(ctx)
		if ingestErr := r.ingest(setupCtx, event); ingestErr != nil {
			return ingestErr
		}
		signals[event.Ref] = recorder.signals
	}
	sessionID, setupErr := r.harness.Setup(ctx, scenario.Setup, signals)
	if setupErr != nil {
		return setupErr
	}
	r.sessionID = sessionID
	return r.waitForAgent(ctx)
}

func (r *agentRun) runPhase(ctx context.Context, index int, phase ScenarioPhase, checkIDs []string) error {
	for _, event := range phase.Events {
		if ingestErr := r.ingest(ctx, event); ingestErr != nil {
			return ingestErr
		}
	}
	if phase.Ask != "" {
		if askErr := r.harness.Ask(ctx, r.sessionID, phase.Ask, fmt.Sprintf("eval-phase-%d", index+1)); askErr != nil {
			return askErr
		}
	}
	if phase.acts() {
		if waitErr := r.waitForAgent(ctx); waitErr != nil {
			return waitErr
		}
	}
	records, recordsErr := r.harness.Records(ctx, r.sessionID)
	if recordsErr != nil {
		return fmt.Errorf("load records: %w", recordsErr)
	}
	for checkIndex, check := range phase.Check {
		r.repetition.Checks = append(r.repetition.Checks, check.Evaluate(checkIDs[checkIndex], records))
	}
	return nil
}

// waitForAgent waits for a turn after the last recorded one, then records the turns that ended.
func (r *agentRun) waitForAgent(ctx context.Context) error {
	if waitErr := r.harness.WaitForReady(ctx, r.sessionID, r.lastSequence); waitErr != nil {
		return waitErr
	}
	return r.recordTurns(ctx)
}

// ingest passes the event through the evaluation integration and projects it. Projection skips a kind it has
// no projector for, so such an event is refused here instead of vanishing.
func (r *agentRun) ingest(ctx context.Context, event ScenarioEvent) error {
	if _, projected := r.service.projection.GetEventProjectorFunc(&ent.NormalizedEvent{Kind: event.Kind}); !projected {
		return fmt.Errorf("event %s: no projection for kind %q", event.Ref, event.Kind)
	}
	encoded, encodeErr := json.Marshal(event)
	if encodeErr != nil {
		return fmt.Errorf("encode event %s: %w", event.Ref, encodeErr)
	}
	providerEvent := rez.ProviderEvent{
		Provider:            r.provider,
		ProviderNamespace:   r.namespace,
		ProviderEventSource: scenarioEventSource,
		ProviderEventRef:    event.Ref,
		Attributes:          encoded,
		ReceivedAt:          r.service.clock.Now(),
	}
	eventIDs, processErr := r.service.ingester.ProcessProviderEvent(ctx, providerEvent)
	if processErr != nil {
		return fmt.Errorf("process event %s: %w", event.Ref, processErr)
	}
	for _, eventID := range eventIDs {
		if projectErr := r.service.ingester.ProjectNormalizedEvent(ctx, eventID); projectErr != nil {
			return fmt.Errorf("project event %s: %w", event.Ref, projectErr)
		}
	}
	return nil
}

// recordTurns records a turn_completed check for each turn that ended since the last call. A turn that failed
// or was aborted fails the step with its error, after its check is recorded.
func (r *agentRun) recordTurns(ctx context.Context) error {
	queryTurns := r.service.db.Client(ctx).AgentTurn.Query().
		Where(at.AgentSessionID(r.sessionID)).
		Order(ent.Asc(at.FieldSequence))
	turns, turnsErr := queryTurns.All(ctx)
	if turnsErr != nil {
		return fmt.Errorf("query turns: %w", turnsErr)
	}
	var failure error
	for _, turn := range turns {
		if r.reportedTurns.Contains(turn.ID) || turn.Status == at.StatusQueued || turn.Status == at.StatusRunning {
			continue
		}
		r.reportedTurns.Add(turn.ID)
		r.lastSequence = max(r.lastSequence, turn.Sequence)
		check := Check{
			ID:       fmt.Sprintf("%s:%d", CheckTurnCompleted, turn.Sequence),
			Passed:   turn.Status == at.StatusCompleted && turn.FinishReason == "stop",
			Expected: "completed, finishing with stop",
			Observed: fmt.Sprintf("%s, finishing with %q", turn.Status, turn.FinishReason),
		}
		check.Summary = fmt.Sprintf("Turn %d %s.", turn.Sequence, check.Observed)
		r.repetition.Checks = append(r.repetition.Checks, check)
		if turn.Status != at.StatusCompleted && failure == nil {
			message := "no error recorded"
			if turn.Error != nil {
				message = *turn.Error
			}
			failure = fmt.Errorf("turn %d %s: %s", turn.Sequence, turn.Status, message)
		}
	}
	return failure
}

// collect records the session's turns, the tokens they used by model, and what the agent produced.
func (r *agentRun) collect(ctx context.Context) error {
	queryTurns := r.service.db.Client(ctx).AgentTurn.Query().
		Where(at.AgentSessionID(r.sessionID)).
		Order(ent.Asc(at.FieldSequence)).
		WithProviderReads()
	turns, turnsErr := queryTurns.All(ctx)
	if turnsErr != nil {
		return fmt.Errorf("query turns: %w", turnsErr)
	}
	var usage UsageTotals
	for _, turn := range turns {
		recorded := Turn{
			Sequence:          turn.Sequence,
			Status:            turn.Status.String(),
			FinishReason:      turn.FinishReason,
			Model:             turn.Model,
			InputTokens:       turn.InputTokens,
			CachedInputTokens: turn.CachedInputTokens,
			OutputTokens:      turn.OutputTokens,
			Reads:             len(turn.Edges.ProviderReads),
		}
		if turn.FinishedAt != nil {
			recorded.DurationMilliseconds = turn.FinishedAt.Sub(turn.CreatedAt).Milliseconds()
		}
		r.repetition.Turns = append(r.repetition.Turns, recorded)
		usage.Add(turn.Model, turn.InputTokens, turn.CachedInputTokens, turn.OutputTokens)
	}
	r.repetition.Usage = usage.List()

	records, recordsErr := r.harness.Records(ctx, r.sessionID)
	if recordsErr != nil {
		return fmt.Errorf("load outputs: %w", recordsErr)
	}
	r.repetition.Outputs = r.harness.Outputs(records)
	return nil
}

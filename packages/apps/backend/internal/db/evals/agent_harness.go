package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	apr "github.com/rezible/rezible/ent/agentproviderread"
	at "github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
)

const (
	// readyBound is how long WaitForReady waits for the agent.
	readyBound    = 10 * time.Minute
	readyInterval = time.Second
)

// AgentHarness is what one agent adds to the base agent harness: how its session starts, how it is asked a
// question, and what its checks read. A specific harness embeds agentHarnessBase for what every agent's
// session shares.
type AgentHarness interface {
	// CheckKinds are the agent's own check kinds, beside the generic ones.
	CheckKinds() []string
	// ValidateSetup checks the scenario's setup without running anything. setupEventRefs are the refs of the
	// events ingested before it.
	ValidateSetup(setup json.RawMessage, setupEventRefs mapset.Set[string]) error
	// Setup prepares the agent's environment and starts its session. signals holds the signal entities each
	// setup event produced, by the event's ref.
	Setup(ctx context.Context, setup json.RawMessage, signals map[string][]uuid.UUID) (uuid.UUID, error)
	// Ask puts a person's question to the session. key is unique within the session.
	Ask(ctx context.Context, sessionID uuid.UUID, text, key string) error
	// WaitForReady blocks until the session has a turn after the given sequence and none queued or running.
	WaitForReady(ctx context.Context, sessionID uuid.UUID, after int) error
	// Records loads what the agent's checks read.
	Records(ctx context.Context, sessionID uuid.UUID) (Records, error)
	// Outputs is what the agent produced, for people to read in a result.
	Outputs(records Records) any
}

type agentHarnessDeps struct {
	db             rez.Database
	situations     rez.SituationService
	investigations rez.InvestigationService
}

// newAgentHarnesses returns the harness of each agent scenarios can evaluate, by agent name. Validating a
// scenario needs no dependencies.
func newAgentHarnesses(deps *agentHarnessDeps) map[string]AgentHarness {
	return map[string]AgentHarness{
		rezai.InvestigationAgent.Name: newInvestigationHarness(deps),
	}
}

// agentHarnessBase is what every agent's session has: turns and provider reads.
type agentHarnessBase struct {
	db rez.Database
}

func (b agentHarnessBase) WaitForReady(ctx context.Context, sessionID uuid.UUID, after int) error {
	deadline := time.Now().Add(readyBound)
	for {
		waitingFor, checkErr := b.waitingFor(ctx, sessionID, after)
		if checkErr != nil {
			return checkErr
		}
		if waitingFor == "" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("not ready after %s: %s", readyBound, waitingFor)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readyInterval):
		}
	}
}

// waitingFor says what the session is still waiting for, or nothing when it is ready.
func (b agentHarnessBase) waitingFor(ctx context.Context, sessionID uuid.UUID, after int) (string, error) {
	queryTurns := b.db.Client(ctx).AgentTurn.Query().
		Where(at.AgentSessionID(sessionID), at.SequenceGT(after)).
		Order(ent.Asc(at.FieldSequence))
	turns, queryErr := queryTurns.All(ctx)
	if queryErr != nil {
		return "", fmt.Errorf("query turns: %w", queryErr)
	}
	if len(turns) == 0 {
		return fmt.Sprintf("a turn after turn %d", after), nil
	}
	for _, turn := range turns {
		if turn.Status == at.StatusQueued || turn.Status == at.StatusRunning {
			return fmt.Sprintf("turn %d, %s", turn.Sequence, turn.Status), nil
		}
	}
	return "", nil
}

// loadReads reads the session's provider reads, by ID, with the read service's name.
func (b agentHarnessBase) loadReads(ctx context.Context, sessionID uuid.UUID) (map[uuid.UUID]RecordedRead, error) {
	queryReads := b.db.Client(ctx).AgentProviderRead.Query().
		Where(apr.AgentSessionID(sessionID)).
		WithServiceEntity()
	reads, readsErr := queryReads.All(ctx)
	if readsErr != nil {
		return nil, fmt.Errorf("query provider reads: %w", readsErr)
	}
	recorded := make(map[uuid.UUID]RecordedRead, len(reads))
	for _, read := range reads {
		service := ""
		if entity := read.Edges.ServiceEntity; entity != nil {
			service = entity.State.DisplayName
		}
		recorded[read.ID] = RecordedRead{Tool: read.ToolName, Service: service, Outcome: read.Status.String()}
	}
	return recorded, nil
}

package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/db"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type AgentRegistrySuite struct {
	test.Suite
}

func TestAgentsRegistrySuite(t *testing.T) {
	suite.Run(t, &AgentRegistrySuite{Suite: test.NewSuite()})
}

func (s *AgentRegistrySuite) makeRegistry() *AiService {
	snapshots, snapshotsErr := db.NewAgentRunSnapshotService(s.Database())
	s.Require().NoError(snapshotsErr)
	return NewAiService(s.T().Context(), s.Config(), snapshots)
}

func (s *AgentRegistrySuite) makeAgentRun(name string, inputState rezai.AgentState) *ent.AiAgentRun {
	s.Require().NoError(inputState.Validate())
	input, inputErr := json.Marshal(inputState)
	s.Require().NoError(inputErr)
	ctx := s.SeedTenantContext()
	create := s.Database().Client(ctx).AiAgentRun.Create().
		SetAgentName(name).
		SetInput(input).
		SetOwnerUserID(s.SeedUser.ID)
	run, saveErr := create.Save(ctx)
	s.Require().NoError(saveErr)
	return run
}

func (s *AgentRegistrySuite) TestAlertInvestigationAgent() {
	s.SeedTestEntities()

	ctx := s.SeedTenantContext()
	reg := s.makeRegistry()

	agentName := rezai.AlertInvestigationAgent.Name

	var alert *ent.Alert
	var run *ent.AiAgentRun
	makeEntitiesTx := func(ctx context.Context, tx *ent.Client) error {
		createAlert := tx.Alert.Create().
			SetTitle("foo")
		txAlert, saveAlertErr := createAlert.Save(ctx)
		if saveAlertErr != nil {
			return saveAlertErr
		}
		alert = txAlert.Unwrap()

		inputState := rezai.AlertInvestigationState{AlertID: alert.ID}
		input, inputErr := json.Marshal(inputState)
		if inputErr != nil {
			return inputErr
		}

		createRun := tx.AiAgentRun.Create().
			SetAgentName(agentName).
			SetInput(input).
			SetOwnerUserID(s.SeedUser.ID)
		txRun, saveRunErr := createRun.Save(ctx)
		if saveRunErr != nil {
			return saveRunErr
		}
		run = txRun.Unwrap()

		return nil
	}
	s.Require().NoError(s.Database().WithTx(ctx, makeEntitiesTx))

	alerts := mocks.NewMockAlertService(s.T())
	alerts.EXPECT().GetAlert(mock.Anything, mock.Anything).Return(alert, nil)

	aia := &AlertInvestigationAgent{alerts: alerts}
	RegisterAgent(reg, aia)

	a, invErr := reg.GetAgentRunInvoker(run)
	s.Require().NoError(invErr)

	snapshotId, invokeErr := a.Start(ctx)
	s.Require().NoError(invokeErr)

	snapshot, snapshotErr := s.Client(ctx).AiAgentRunSnapshot.Get(ctx, snapshotId)
	s.Require().NoError(snapshotErr)
	s.Require().NotNil(snapshot)
	//s.Require().Len(snapshot.State.Messages, 2)
}

type testAgentState struct {
	foo string
}

func (s testAgentState) Validate() error {
	return nil
}

func (s *AgentRegistrySuite) TestSimpleWorkflowAgent() {
	s.SeedTestEntities()

	ctx := s.SeedTenantContext()
	reg := s.makeRegistry()
	agentName := "test_agent"
	ta := &testAgent[testAgentState]{
		def:      rezai.AgentDefinition[testAgentState]{Name: agentName},
		fakeCall: true,
	}
	RegisterAgent(reg, ta)

	state := testAgentState{foo: "bar!"}
	run := s.makeAgentRun(agentName, state)

	a, invErr := reg.GetAgentRunInvoker(run)
	s.Require().NoError(invErr)

	snapshotId, runErr := a.Start(ctx)
	s.Require().NoError(runErr)
	s.Require().NotEmpty(snapshotId)
	//snapshot, snapshotErr := store.GetSnapshot(s.T().Context(), snapshotId.String())
	//s.Require().NoError(snapshotErr)
	//s.Require().NotNil(snapshot.State)
	//s.Require().Len(snapshot.State.Messages, 2)
}

type testAgent[S rezai.AgentState] struct {
	def      rezai.AgentDefinition[S]
	custom   *S
	customFn func(S) S
	fakeCall bool
}

func makeTestAgent[S rezai.AgentState](def rezai.AgentDefinition[S]) *testAgent[S] {
	return &testAgent[S]{def: def}
}

func (t *testAgent[S]) makeInitialState(input []byte) (*ai.Message, *aix.SessionState[S], error) {
	s := &aix.SessionState[S]{
		Messages: []*ai.Message{ai.NewUserTextMessage("hello world")},
	}
	if t.custom != nil {
		s.Custom = *t.custom
	}
	return ai.NewSystemTextMessage("foo bar"), s, nil
}

func (t *testAgent[S]) definition() rezai.AgentDefinition[S] {
	return t.def
}

func (t *testAgent[S]) run(g *genkit.Genkit) aix.AgentFunc[S] {
	return func(ctx context.Context, resp aix.Responder, sess *aix.SessionRunner[S]) (*aix.AgentResult, error) {
		if t.customFn != nil {
			sess.UpdateCustom(t.customFn)
		}

		runSessTurnFn := func(ctx context.Context, input *aix.AgentInput) (*aix.TurnResult, error) {
			var finishReason aix.AgentFinishReason
			var msg *ai.Message

			if t.fakeCall {
				finishReason = aix.AgentFinishReasonStop
				msg = ai.NewModelTextMessage("response")
			} else {
				gen, genErr := genkit.Generate(ctx, g,
					ai.WithModelName("googleai/gemini-flash-latest"),
					ai.WithSystem("You are a concise assistant."),
					ai.WithMessages(sess.Messages()...),
				)
				if genErr != nil {
					return nil, fmt.Errorf("generate err: %w", genErr)
				}
				msg = gen.Message
				finishReason = aix.AgentFinishReason(gen.FinishReason)
			}
			sess.AddMessages(msg)

			return &aix.TurnResult{FinishReason: finishReason}, nil
		}

		if turnErr := sess.Run(ctx, runSessTurnFn); turnErr != nil {
			return nil, fmt.Errorf("run turn: %w", turnErr)
		}

		return sess.Result(), nil
	}
}

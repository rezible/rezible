package genkit

import (
	"context"
	"fmt"
	"testing"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/pkg/agents"
	"github.com/rezible/rezible/test"
)

type AgentRegistrySuite struct {
	test.Suite
}

func TestAgentsRegistrySuite(t *testing.T) {
	suite.Run(t, &AgentRegistrySuite{Suite: test.NewSuite()})
}

func (s *AgentRegistrySuite) makeRegistry() *AgentRegistry {
	snapshots, snapshotsErr := db.NewAgentRunSnapshotService(s.Database())
	s.Require().NoError(snapshotsErr)
	return NewAgentRegistry(s.T().Context(), s.Config(), snapshots)
}

func (s *AgentRegistrySuite) makeAgentRun(workflow string, input []byte) *ent.AgentRun {
	ctx := s.SeedTenantContext()
	create := s.Database().Client(ctx).AgentRun.Create().
		SetWorkflow(workflow).
		SetInput(input).
		SetOwnerUserID(s.SeedUser.ID).
		SetTriggerKind(agentrun.TriggerKindManual)
	run, saveErr := create.Save(ctx)
	s.Require().NoError(saveErr)
	return run
}

func (s *AgentRegistrySuite) TestAlertInvestigationAgent() {
	s.SeedTestEntities()

	ctx := s.SeedTenantContext()
	reg := s.makeRegistry()

	workflowName := agents.WorkflowAlertInvestigation.Name()

	var alert *ent.Alert
	var run *ent.AgentRun
	makeEntitiesTx := func(ctx context.Context, tx *ent.Client) error {
		createAlert := tx.Alert.Create().
			SetTitle("foo")
		txAlert, saveAlertErr := createAlert.Save(ctx)
		if saveAlertErr != nil {
			return saveAlertErr
		}
		alert = txAlert.Unwrap()

		createRun := tx.AgentRun.Create().
			SetWorkflow(workflowName).
			SetInput([]byte("{}")).
			SetOwnerUserID(s.SeedUser.ID).
			SetTriggerKind(agentrun.TriggerKindManual)
		txRun, saveRunErr := createRun.Save(ctx)
		if saveRunErr != nil {
			return saveRunErr
		}

		createSubject := tx.AgentRunSubject.Create().
			SetAgentRunID(txRun.ID).
			SetDomainEntityID(alert.ID).
			SetSubjectKind("alert")
		if subjErr := createSubject.Exec(ctx); subjErr != nil {
			return subjErr
		}

		queryRun := tx.AgentRun.Query().
			Where(agentrun.ID(txRun.ID)).
			WithSubjects()
		subjRun, queryRunErr := queryRun.Only(ctx)
		if queryRunErr != nil {
			return queryRunErr
		}
		run = subjRun.Unwrap()

		return nil
	}
	s.Require().NoError(s.Database().WithTx(ctx, makeEntitiesTx))

	alerts := mocks.NewMockAlertService(s.T())
	alerts.EXPECT().GetAlert(mock.Anything, mock.Anything).Return(alert, nil)

	//store := localstore.NewInMemorySessionStore[agents.AlertInvestigationState]()

	aia := &AlertInvestigationAgent{alerts: alerts}
	RegisterWorkflowAgent(reg, aia)

	a, ok := reg.Get(workflowName)
	s.Require().True(ok)

	snapshotId, invokeErr := a.Invoke(ctx, run, ai.NewSystemTextMessage("look into this"))
	s.Require().NoError(invokeErr)

	snapshot, snapshotErr := s.Client(ctx).AgentRunSnapshot.Get(ctx, snapshotId)
	s.Require().NoError(snapshotErr)
	s.Require().NotNil(snapshot)
	//s.Require().Len(snapshot.State.Messages, 2)
}

func (s *AgentRegistrySuite) TestSimpleWorkflowAgent() {
	s.SeedTestEntities()

	ctx := s.SeedTenantContext()
	reg := s.makeRegistry()
	type testAgentState struct {
		foo string
	}
	type testAgentOutput struct {
		bar string
	}
	customFoo := "bar!"
	//store := localstore.NewInMemorySessionStore[testAgentState]()
	ta := &testWorkflowAgent[testAgentState, testAgentOutput]{
		name:     "testWorkflow",
		custom:   &testAgentState{foo: customFoo},
		fakeCall: true,
	}
	RegisterWorkflowAgent(reg, ta)

	a, ok := reg.Get(ta.workflow().Name())
	s.Require().True(ok)

	run := s.makeAgentRun(ta.workflow().Name(), []byte("{}"))
	snapshot, runErr := a.Invoke(ctx, run, ai.NewSystemTextMessage("look into this"))
	s.Require().NoError(runErr)
	s.Require().NotNil(snapshot)
	//snapshot, snapshotErr := store.GetSnapshot(s.T().Context(), snapshotId.String())
	//s.Require().NoError(snapshotErr)
	//s.Require().NotNil(snapshot.State)
	//s.Require().Len(snapshot.State.Messages, 2)
}

type testWorkflowAgent[S any, O any] struct {
	name     string
	custom   *S
	customFn func(S) S
	fakeCall bool
}

func (t *testWorkflowAgent[S, O]) validateInput(bytes []byte) error {
	return nil
}

func (t *testWorkflowAgent[S, O]) makeInitialState(run *ent.AgentRun) (*aix.SessionState[S], error) {
	s := &aix.SessionState[S]{
		Messages: []*ai.Message{ai.NewUserTextMessage("hello world")},
	}
	if t.custom != nil {
		s.Custom = *t.custom
	}
	return s, nil
}

func (t *testWorkflowAgent[S, O]) workflow() agents.Workflow[S, O] {
	return agents.NewWorkflow[S, O](t.name, "a workflow just for testing")
}

func (t *testWorkflowAgent[S, O]) agentFunc(g *genkit.Genkit) aix.AgentFunc[S] {
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

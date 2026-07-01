package genkit

import (
	"context"
	"fmt"
	"testing"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
	"github.com/firebase/genkit/go/genkit"

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
	return NewAgentRegistry(s.T().Context(), s.Config())
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
	store := localstore.NewInMemorySessionStore[testAgentState]()
	ta := &testWorkflowAgent[testAgentState, testAgentOutput]{
		name:     "testWorkflow",
		custom:   &testAgentState{foo: customFoo},
		fakeCall: true,
	}
	reg.Register(wrapWorkflowAgent(reg.gk, store, ta))

	a, ok := reg.Get(ta.workflow().Name())
	s.Require().True(ok)

	run := s.makeAgentRun(ta.workflow().Name(), []byte("{}"))
	snapshotId, runErr := a.Run(ctx, run, nil)
	s.Require().NoError(runErr)

	snapshot, snapshotErr := store.GetSnapshot(s.T().Context(), snapshotId.String())
	s.Require().NoError(snapshotErr)
	s.Require().NotNil(snapshot.State)
	s.Require().Len(snapshot.State.Messages, 2)

	snap2, snap2Err := store.GetLatestSnapshot(s.T().Context(), run.ID.String())
	s.Require().NoError(snap2Err)
	s.Require().NotNil(snap2.State)
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

func (t *testWorkflowAgent[S, O]) makeInitial(run *ent.AgentRun) (*ai.Message, *S, error) {
	return ai.NewUserTextMessage("hello world"), t.custom, nil
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

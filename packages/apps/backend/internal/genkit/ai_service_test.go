package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	aar "github.com/rezible/rezible/ent/aiagentrun"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/db"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type AiServiceSuite struct {
	test.Suite
}

func TestAiServiceSuite(t *testing.T) {

	suite.Run(t, &AiServiceSuite{Suite: test.NewSuite()})
}

func (s *AiServiceSuite) checkSkip(name string) {
	if os.Getenv("AI_TESTS_ALL") != "true" && os.Getenv("AI_TESTS_"+name) != "true" {
		s.T().Skipf("Skipping live AI test '%s'", name)
	}
}

func (s *AiServiceSuite) makeService(opts ...AiServiceOption) *AiService {
	msgs := mocks.NewMockMessageService(s.T())
	msgs.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	snapshots, snapshotsErr := db.NewAiAgentSnapshotService(s.Database(), msgs)
	s.Require().NoError(snapshotsErr)

	kg, kgErr := db.NewKnowledgeGraphService(s.Database())
	s.Require().NoError(kgErr)

	svc := NewAiService(s.Config(), snapshots, kg)
	s.Require().NoError(svc.Init(s.T().Context(), opts...))

	return svc
}

func (s *AiServiceSuite) makeAgentRun(name string, inputState rezai.SessionState) *ent.AiAgentRun {
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

func (s *AiServiceSuite) TestAlertInvestigation() {
	s.checkSkip("alert_investigation")

	s.SeedTestEntities()

	ctx := s.SeedTenantContext()

	agentName := rezai.AlertsAgent.Name

	createAlert := s.Client(ctx).Alert.Create().
		SetTitle("foo")
	fooAlert, saveAlertErr := createAlert.Save(ctx)
	s.Require().NoError(saveAlertErr)

	input, inputErr := json.Marshal(rezai.AlertAgentInput{AlertID: fooAlert.ID})
	s.Require().NoError(inputErr)

	createRun := s.Client(ctx).AiAgentRun.Create().
		SetAgentName(agentName).
		SetInput(input).
		SetOwnerUserID(s.SeedUser.ID)
	run, saveRunErr := createRun.Save(ctx)
	s.Require().NoError(saveRunErr)

	alerts := mocks.NewMockAlertService(s.T())
	alerts.EXPECT().GetAlert(mock.Anything, mock.Anything).Return(fooAlert, nil)

	aia := &AlertsAgent{alerts: alerts}
	reg := s.makeService(WithAgent(aia))

	a, invErr := reg.GetAgentRunner(run)
	s.Require().NoError(invErr)

	_, invokeErr := a.Invoke(ctx, nil, ai.NewUserTextMessage("investigate this"), nil)
	s.Require().NoError(invokeErr)
}

type (
	testAgentInput struct{}
	testAgentState struct {
		foo string
	}
	testAgentOutput struct {
		Greeting string `json:"greeting"`
	}
)

func (i testAgentInput) Validate() error {
	return nil
}

func (o testAgentOutput) Validate() error {
	if len(o.Greeting) == 0 {
		return fmt.Errorf("empty greeting")
	}
	return nil
}

func makeTestAgent[S rezai.SessionState](userMessage string) *testAgent[S] {
	taDef := testAgentDef[S]{
		Name:         "test_agent",
		Description:  "A simple agent",
		SystemPrompt: "You are an ai agent that follow user instructions exactly. Keep output concise",
	}
	return &testAgent[S]{
		def:         taDef,
		userMessage: userMessage,
	}
}

func (s *AiServiceSuite) TestSimpleGreetingAgent() {
	s.checkSkip("simple_greeting")

	s.SeedTestEntities()

	ctx := s.SeedTenantContext()

	initialState := testAgentState{foo: "bar"}
	ta := makeTestAgent[testAgentState]("Write a one-word greeting to result, then reply with a simple 'done'.")
	reg := s.makeService(WithAgent(ta))

	ar := s.makeAgentRun(ta.def.Name, initialState)
	runId := ar.ID

	a, invErr := reg.GetAgentRunner(ar)
	s.Require().NoError(invErr)

	s.T().Logf("Starting test agent run (id %s)", runId.String())
	snapshotId, startErr := a.Invoke(ctx, nil, ai.NewUserTextMessage(ta.userMessage), nil)
	s.Require().NoError(startErr)
	s.Require().NotEmpty(snapshotId)

	queryRun := s.Client(ctx).AiAgentRun.Query().
		Where(aar.ID(runId)).
		WithSnapshots()
	run, runErr := queryRun.Only(ctx)
	s.Require().NoError(runErr)

	s.Require().NotEmpty(run.Edges.Snapshots)
	for _, snap := range run.Edges.Snapshots {
		s.T().Logf("Snapshot ID: %s (%s, %s)\n", snap.ID, snap.Status.String(), snap.FinishReason)
		s.Require().NotNil(snap.State)
		var snapState *aix.SessionState[testAgentState]
		s.Require().NoError(json.Unmarshal(*snap.State, &snapState))
		s.Require().NotEmpty(snapState.Messages)
		for _, m := range snapState.Messages {
			if len(strings.TrimSpace(m.Text())) > 0 {
				s.T().Logf("\t[%s]: %s", m.Role, m.Text())
			} else {
				s.T().Logf("\t[%s]: ", m.Role)
				for _, p := range m.Content {
					var line string
					if p.IsToolRequest() {
						line = fmt.Sprintf("[tool request] %s: %+v", p.ToolRequest.Name, p.ToolRequest.Input)
					} else if p.IsToolResponse() {
						line = fmt.Sprintf("[tool response] %s: %+v", p.ToolResponse.Name, p.ToolResponse.Output)
					}
					s.T().Logf("\t\t%s", line)
				}
			}
		}

		s.T().Log("Outputs:")
		outputs, outputsErr := ta.def.GetSnapshotOutputs(snap)
		s.Require().NoError(outputsErr)
		s.Require().NotEmpty(outputs)
		for i, o := range outputs {
			s.T().Logf("\t[%d]: %+v", i, o)
		}
	}
}

type (
	testAgentDef[S rezai.SessionState] = rezai.AgentDefinition[testAgentInput, S, testAgentOutput]

	testAgent[S rezai.SessionState] struct {
		def         testAgentDef[S]
		custom      *S
		customFn    func(S) S
		userMessage string
		fakeCall    bool
	}
)

func (t *testAgent[S]) definition() testAgentDef[S] {
	return t.def
}

func (t *testAgent[S]) makeInitialUserMessage(ctx context.Context, input testAgentInput) (*ai.Message, error) {
	return ai.NewUserTextMessage(t.userMessage), nil
}

func (t *testAgent[S]) transformState(ctx context.Context, state *aix.SessionState[S]) (*aix.SessionState[S], error) {
	return state, nil
}

func (t *testAgent[S]) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

//func (t *testAgent[S]) run(ctx context.Context, resp aix.Responder, sess *aix.SessionRunner[S]) (*aix.AgentResult, error) {
//	if t.customFn != nil {
//		sess.UpdateCustom(t.customFn)
//	}
//
//	runSessTurnFn := func(ctx context.Context, input *aix.AgentInput) (*aix.TurnResult, error) {
//		var finishReason aix.AgentFinishReason
//		var msg *ai.Message
//
//		if t.fakeCall {
//			finishReason = aix.AgentFinishReasonStop
//			msg = ai.NewModelTextMessage("response")
//		} else {
//			gen, genErr := genkit.Generate(ctx, genkit.FromContext(ctx),
//				ai.WithModelName("googleai/gemini-flash-latest"),
//				ai.WithSystem("You are a concise assistant."),
//				ai.WithMessages(sess.Messages()...),
//			)
//			if genErr != nil {
//				return nil, fmt.Errorf("generate err: %w", genErr)
//			}
//			msg = gen.Message
//			finishReason = aix.AgentFinishReason(gen.FinishReason)
//		}
//		sess.AddMessages(msg)
//
//		return &aix.TurnResult{FinishReason: finishReason}, nil
//	}
//
//	if turnErr := sess.Run(ctx, runSessTurnFn); turnErr != nil {
//		return nil, fmt.Errorf("run turn: %w", turnErr)
//	}
//
//	return sess.Result(), nil
//}

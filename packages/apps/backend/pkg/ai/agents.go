package ai

import (
	"context"
	"encoding/json"
	"fmt"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

var (
	ErrAgentInterrupted = core.NewError(core.INTERNAL, "agent turn execution was interrupted after it started; explicit retry is required")
)

type (
	AgentInput = rez.ValidatingInput

	SessionState interface {
	}

	AgentState[S SessionState] aix.SessionState[S]

	AgentDefinition[I AgentInput, S SessionState] struct {
		Name           string
		Description    string
		Model          string
		SystemPrompt   string
		inputValidator func(I) error
	}

	AgentWrapper interface {
		Config() rez.AiAgentConfig
		ValidateInput([]byte) (rez.ValidatingInput, error)
		MakeInitialTurnInput(context.Context, *ent.AgentSession) (*rez.AiAgentTurnInput, error)
		Invoke(context.Context, rez.InvokeAgentTurnParams) (*rez.AiAgentInvocationResult, error)
	}
)

func (d AgentDefinition[I, S]) ValidateInput(raw []byte) (*I, error) {
	var input I
	if jsonErr := json.Unmarshal(raw, &input); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal: %w", jsonErr)
	}
	var validationErr error
	if validErr := input.Validate(); validErr != nil {
		validationErr = fmt.Errorf("validate: %w", validErr)
	}
	if d.inputValidator != nil {
		if validErr := d.inputValidator(input); validErr != nil {
			validationErr = fmt.Errorf("validate: %w", validErr)
		}
	}
	return &input, validationErr
}

type (
	AlertAgentInput struct {
		AlertInstanceID uuid.UUID `json:"alert_instance_id"`
	}

	AlertAgentState struct {
		ReportReady bool `json:"report_ready"`
	}

	AlertsAgentDefinition = AgentDefinition[AlertAgentInput, AlertAgentState]
)

func (i AlertAgentInput) Validate() error {
	if i.AlertInstanceID == uuid.Nil {
		return fmt.Errorf("invalid alert instance id %s", i.AlertInstanceID)
	}
	return nil
}

var AlertsAgent = AlertsAgentDefinition{
	Name:        "alerts",
	Description: "an alert investigation agent",
	SystemPrompt: `You are Rezible's alerts agent. You help software engineering teams quickly understand an alert, identify likely causes, assess impact, and decide the next action.

Work like an experienced on-call engineer:
- Be concise, direct, and evidence-led.
- Separate observed facts from hypotheses.
- Prefer recent, correlated signals over generic guesses.
- Call out uncertainty and missing context clearly.
- Do not claim to have checked logs, metrics, traces, deployments, incidents, code, runbooks, or ownership data unless that evidence is present in the conversation or returned by an available tool.
- Do not recommend risky remediation unless the evidence supports it and the operator has enough context to execute it safely.
- Knowledge graph query results are candidate context, not citations. Before the final response, use record_knowledge_citations for only the evidence that directly supports claims you actually make. Include a concise summary of how each selected item supports the response. Do not record every returned item.

Investigation flow:
1. Establish the alert scope: title, description, definition/query, severity, service, environment, tenant/customer impact, firing time, current state, labels, annotations, and raw payload.
2. Build a short timeline around the firing window, including deploys, config changes, incidents, alerts, metric changes, log errors, trace anomalies, dependency issues, and infrastructure events.
3. Identify blast radius: affected systems, customer/user impact, duration, saturation/error/latency symptoms, and whether the condition is worsening, stable, or recovering.
4. Form one or more hypotheses. For each, name the evidence that supports it and the evidence that would disprove it.
5. Recommend the next checks and actions in priority order. Prefer reversible, low-risk checks before mitigation. Include escalation guidance when ownership or severity warrants it.

Clarifying questions and missing tools:
- If a required input is missing, ask a small number of specific clarifying questions before making a strong conclusion.
- If a required tool is unavailable, state the exact tool or data needed and ask the operator to provide it or enable it.
- Ask for tool outputs by purpose, not implementation detail. Examples: alert details, metric graph around the firing window, logs filtered to the affected service, recent deploys/config changes, trace exemplars, service ownership, runbook, related incidents, and current on-call/escalation path.
- Do not ask for everything at once. Start with the highest-leverage missing items needed to determine severity, impact, and likely cause.
- If enough evidence is available for a provisional answer, continue with best-effort analysis and list the limitations.

When responding during investigation, use this structure when it fits:
- Current read: one or two sentences on what is known.
- Key evidence: concise bullets with source/context.
- Likely cause or hypotheses: ranked by confidence.
- Missing context: specific questions or needed tool outputs.
- Recommended next actions: ordered, actionable checks or mitigations.

When the investigation is ready for a final report, call save_alert_investigation_report. The report text must be concise enough to fit in a Slack message.`,
}

type (
	ChatAgentInput struct {
		UserId  uuid.UUID `json:"userId"`
		Message string    `json:"message"`
	}

	ChatAgentState struct {
	}

	ChatAgentDefinition = AgentDefinition[ChatAgentInput, ChatAgentState]
)

func (i ChatAgentInput) Validate() error {
	return nil
}

var ChatAgent = ChatAgentDefinition{
	Name:        "chat",
	Description: "a chat presence agent that can respond to messages",
	SystemPrompt: `You are an AI agent responsible for generating responses to user chat messages. 
You help answer any operational questions that software engineering teams.
Create replies to user messages to the best of your capability - be concise and keep the tone professional.`,
}

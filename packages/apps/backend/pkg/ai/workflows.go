package ai

import (
	"context"
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
)

type (
	WorkflowWrapper interface {
		Config() rez.AiAgentConfig
		Run(context.Context, rez.AiWorkflowInput) (rez.AiWorkflowOutput, error)
	}

	WorkflowRunner[I rez.AiWorkflowInput, O rez.AiWorkflowOutput] interface {
		Run(context.Context, I) (*O, error)
	}

	WorkflowDefinition[I rez.AiWorkflowInput, O rez.AiWorkflowOutput] struct {
		Name         string
		Description  string
		Model        string
		SystemPrompt string
		Prompt       func(I) string
	}

	workflowDefinitionRunner[I rez.AiWorkflowInput, O rez.AiWorkflowOutput] struct {
		runner rez.AiWorkflowRunner
	}
)

func (r *workflowDefinitionRunner[I, O]) Run(ctx context.Context, input I) (*O, error) {
	runOutput, runErr := r.runner.Run(ctx, input)
	if runErr != nil {
		return nil, fmt.Errorf("run: %w", runErr)
	}
	output, ok := runOutput.(O)
	if !ok {
		return nil, fmt.Errorf("invalid output type %T", output)
	}
	return &output, nil
}

func GetWorkflowRunner[I rez.AiWorkflowInput, O rez.AiWorkflowOutput](s rez.AiService, d WorkflowDefinition[I, O]) (WorkflowRunner[I, O], error) {
	runner, runnerErr := s.GetWorkflowRunner(d.Name)
	if runnerErr != nil {
		return nil, runnerErr
	}
	return &workflowDefinitionRunner[I, O]{runner: runner}, nil
}

type (
	ClassifyAgentThreadResponseInput struct {
		PreviousMessages []string `json:"previousMessages"`
		UserMessage      string   `json:"message"`
	}

	ClassifyAgentThreadResponseOutput struct {
		ShouldReply bool `json:"should_reply"`
	}

	ClassifyAgentThreadResponseWorkflowRunner = WorkflowRunner[ClassifyAgentThreadResponseInput, ClassifyAgentThreadResponseOutput]
)

func (i ClassifyAgentThreadResponseInput) Validate() error {
	if len(i.UserMessage) == 0 {
		return fmt.Errorf("no user message")
	}
	return nil
}

var ClassifyAgentThreadResponseWorkflow = WorkflowDefinition[ClassifyAgentThreadResponseInput, ClassifyAgentThreadResponseOutput]{
	Name:        "classify_agent_thread_response",
	Description: "Classify whether a chat thread message needs a reply from the agent.",
	SystemPrompt: `You are a simple classifier to determine whether an incoming message in a slack thread is directed at the rezible agent, and requires a reply.

Return should_reply=true if the latest human message is clearly an question or instruction directed at the rezible agent.

For messages that are not OBVIOUSLY directed at the rezible agent, or messages that do not need a reply (such as acknowledgements, thanks, human-to-human discussion, etc), return should_reply=false.

If in doubt, err on the side of caution (set should_reply=false) - users can directly tag the agent to avoid ambiguity.`,
	Prompt: func(input ClassifyAgentThreadResponseInput) string {
		var b strings.Builder
		if len(input.PreviousMessages) > 0 {
			b.WriteString("For context, the ")
			if len(input.PreviousMessages) == 1 {
				b.WriteString("previous message")
			} else {
				b.WriteString(fmt.Sprintf("previous %d messages", len(input.PreviousMessages)))
			}
			b.WriteString(":\n")
			for _, msg := range input.PreviousMessages {
				b.WriteString(fmt.Sprintf(" %s\n", strings.TrimSpace(msg)))
			}
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("Message to classify: [%s]", strings.TrimSpace(input.UserMessage)))
		return b.String()
	},
}

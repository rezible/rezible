package genkit

import (
	"context"
	"fmt"
	"log/slog"

	gkai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	gk "github.com/firebase/genkit/go/genkit"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func WithWorkflow[I rez.AiWorkflowInput, O rez.AiWorkflowOutput](def rezai.WorkflowDefinition[I, O]) AiServiceOption {
	return AiServiceOption{
		kind: AiServiceOptionKindWorkflow,
		optFn: func(s *AiService) error {
			runner, runnerErr := makeWorkflowWrapper(s, def)
			if runnerErr != nil || runner == nil {
				return fmt.Errorf("workflow runner: %w", runnerErr)
			}
			s.workflowWrappers[def.Name] = runner
			return nil
		},
	}
}

type (
	workflowWrapper[I rez.AiWorkflowInput, O rez.AiWorkflowOutput] struct {
		def  rezai.WorkflowDefinition[I, O]
		flow *core.Flow[I, O, struct{}]
	}
)

func buildWorkflowOpts[I rez.AiWorkflowInput, O rez.AiWorkflowOutput](svc *AiService, def rezai.WorkflowDefinition[I, O], input I) []gkai.GenerateOption {
	opts := make([]gkai.GenerateOption, 1, 4)
	modelOpt := gkai.WithModel(svc.getDefaultModel())
	if def.Model != "" {
		modelOpt = gkai.WithModelName(def.Model)
	}
	opts[0] = modelOpt
	if def.SystemPrompt != "" {
		opts = append(opts, gkai.WithSystem(def.SystemPrompt))
		//} else if def.SystemPromptFn != nil {
		//	opts = append(opts, gkai.WithSystemFn(def.SystemPromptFn))
	} else {
		// TODO: no system prompt??
	}
	if def.Prompt != nil {
		opts = append(opts, gkai.WithPrompt(def.Prompt(input)))
	}
	return opts
}

func makeWorkflowWrapper[I rez.AiWorkflowInput, O rez.AiWorkflowOutput](svc *AiService, def rezai.WorkflowDefinition[I, O]) (*workflowWrapper[I, O], error) {
	flow := gk.DefineFlow(svc.gk, def.Name, func(ctx context.Context, input I) (O, error) {
		var output O
		if inputErr := input.Validate(); inputErr != nil {
			return output, fmt.Errorf("input: %w", inputErr)
		}

		return gk.Run(ctx, "generate", func() (O, error) {
			out, resp, genErr := gk.GenerateData[O](ctx, svc.gk, buildWorkflowOpts(svc, def, input)...)
			if genErr != nil {
				return output, genErr
			} else if out == nil {
				return output, fmt.Errorf("workflow generated nil output")
			} else if out != nil {
				output = *out
			}
			if resp != nil {
				slog.Debug("model generate response", "resp", resp)
			}
			return output, nil
		})
	})
	return &workflowWrapper[I, O]{flow: flow}, nil
}

func (r *workflowWrapper[I, O]) Config() rez.AiAgentConfig {
	return rez.AiAgentConfig{Name: r.def.Name}
}

func (r *workflowWrapper[I, O]) Run(ctx context.Context, input rez.AiWorkflowInput) (rez.AiWorkflowOutput, error) {
	i, ok := input.(I)
	if !ok {
		return nil, fmt.Errorf("invalid input")
	}
	out, runErr := r.flow.Run(ctx, i)
	if runErr != nil {
		return nil, fmt.Errorf("run: %w", runErr)
	}
	return out, nil
}

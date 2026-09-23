package genkit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	gkai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	gk "github.com/firebase/genkit/go/genkit"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type WorkflowBuilder struct {
	runtime *AiRuntime
	runner  rez.AiWorkflowRunner

	mu            sync.Mutex
	workflowNames map[string]struct{}
}

func NewWorkflowBuilder(runtime *AiRuntime, runner rez.AiWorkflowRunner) *WorkflowBuilder {
	return &WorkflowBuilder{
		runtime:       runtime,
		runner:        runner,
		workflowNames: make(map[string]struct{}),
	}
}

func (b *WorkflowBuilder) DefineWorkflow[I, O any](name string, run func(context.Context, I) (O, error)) (rez.AiWorkflow[I, O], error) {
	if b == nil || b.runtime == nil {
		return nil, fmt.Errorf("AI runtime is required")
	}
	if b.runtime.gk == nil {
		return nil, fmt.Errorf("AI runtime is not initialized")
	}
	if b.runner == nil {
		return nil, fmt.Errorf("AI workflow runner is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if run == nil {
		return nil, fmt.Errorf("workflow callback is required")
	}

	b.mu.Lock()
	if _, exists := b.workflowNames[name]; exists {
		b.mu.Unlock()
		return nil, fmt.Errorf("workflow %q is already defined", name)
	}
	b.workflowNames[name] = struct{}{}
	b.mu.Unlock()

	flow := gk.DefineFlow(b.runtime.gk, name, func(ctx context.Context, input I) (output O, executeErr error) {
		executeErr = b.runner.ExecuteWorkflow(ctx, name, func(ctx context.Context) error {
			output, executeErr = run(ctx, input)
			return executeErr
		})
		return output, executeErr
	})
	return &typedWorkflow[I, O]{flow: flow}, nil
}

func (b *WorkflowBuilder) DefinePromptWorkflow[I rez.ValidatingInput, O any](def rezai.AiPromptWorkflowDefinition[I, O]) (rez.AiWorkflow[I, O], error) {
	if strings.TrimSpace(def.Name) == "" {
		return nil, fmt.Errorf("workflow definition name is required")
	}
	if def.Prompt == nil {
		return nil, fmt.Errorf("workflow definition prompt is required")
	}

	return b.DefineWorkflow(def.Name, func(ctx context.Context, input I) (O, error) {
		var output O
		if inputErr := input.Validate(); inputErr != nil {
			return output, fmt.Errorf("input: %w", inputErr)
		}

		generated, response, generateErr := gk.GenerateData[O](ctx, b.runtime.gk, b.buildPromptWorkflowOpts(def, input)...)
		if generateErr != nil {
			return output, generateErr
		}
		if generated == nil {
			return output, fmt.Errorf("workflow generated nil output")
		}
		if response != nil {
			slog.Debug("model generate response", "resp", response)
		}
		return *generated, nil
	})
}

func (b *WorkflowBuilder) buildPromptWorkflowOpts[I rez.ValidatingInput, O any](def rezai.AiPromptWorkflowDefinition[I, O], input I) []gkai.GenerateOption {
	var opts []gkai.GenerateOption
	if def.Model != "" {
		opts = append(opts, gkai.WithModelName(def.Model))
	}
	if def.SystemPrompt != "" {
		opts = append(opts, gkai.WithSystem(def.SystemPrompt))
	}
	opts = append(opts, gkai.WithPrompt(def.Prompt(input)))
	return opts
}

type typedWorkflow[I, O any] struct {
	flow *core.Flow[I, O, struct{}]
}

func (w *typedWorkflow[I, O]) Run(ctx context.Context, input I) (O, error) {
	return w.flow.Run(ctx, input)
}

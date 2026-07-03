package genkit

import (
	"context"
	"encoding/json"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	WorkflowInvokerFunc func() rez.AiWorkflowInvoker

	workflowRunner[I rezai.WorkflowInput, S any, O rezai.WorkflowOutput] interface {
		definition() rezai.WorkflowDefinition[I, S, O]
		run(context.Context, I) (O, error)
	}
)

type exampleWorkflowRunner struct{}

func (w *exampleWorkflowRunner) run(ctx context.Context, input rezai.ExampleWorkflowInput) (rezai.ExampleWorkflowOutput, error) {
	return rezai.ExampleWorkflowOutput{}, nil
}

func makeWorkflowInvokerFunc[I rezai.WorkflowInput, S any, O rezai.WorkflowOutput](g *genkit.Genkit, store aix.SessionStore[S], r workflowRunner[I, S, O]) WorkflowInvokerFunc {
	name := r.definition().Name
	flow := genkit.DefineFlow[I, O](g, name, r.run)
	return func() rez.AiWorkflowInvoker {
		return newWorkflowInvoker[I, any](flow)
	}
}

type workflowInvoker[I rezai.WorkflowInput, State any, O rezai.WorkflowOutput, StreamChunk any] struct {
	flow            *core.Flow[I, O, StreamChunk]
	streamCallbacks []core.StreamCallback[json.RawMessage]
	input           json.RawMessage
}

func newWorkflowInvoker[I rezai.WorkflowInput, S any, O rezai.WorkflowOutput, C any](flow *core.Flow[I, O, C]) *workflowInvoker[I, S, O, C] {
	return &workflowInvoker[I, S, O, C]{flow: flow}
}

func (i workflowInvoker[I, S, O, C]) AddStreamCallback(cb func(ctx context.Context, msg json.RawMessage) error) {
	i.streamCallbacks = append(i.streamCallbacks, cb)
}

func (i workflowInvoker[I, S, O, C]) Run(ctx context.Context) (json.RawMessage, error) {
	return i.flow.RunJSON(ctx, i.input, func(ctx context.Context, msg json.RawMessage) error {
		for _, cb := range i.streamCallbacks {
			if cbErr := cb(ctx, msg); cbErr != nil {
				return cbErr
			}
		}
		return nil
	})
}

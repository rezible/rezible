package genkit

import (
	"context"
	"fmt"
	"log/slog"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	WorkflowInvokerFunc func() rez.AiWorkflowInvoker

	workflowRunner[I rezai.WorkflowInput, O rezai.WorkflowOutput, S rezai.WorkflowState] interface {
		definition() rezai.WorkflowDefinition[I, O, S]
		run(context.Context, I, core.StreamCallback[S]) (O, error)
	}
)

func makeWorkflowInvokerFunc[I rezai.WorkflowInput, O rezai.WorkflowOutput, S rezai.WorkflowState](g *genkit.Genkit, store aix.SessionStore[S], r workflowRunner[I, O, S]) WorkflowInvokerFunc {
	name := r.definition().Name
	flow := genkit.DefineStreamingFlow[I, O, S](g, name, r.run)
	return func() rez.AiWorkflowInvoker {
		return &workflowInvoker[I, O, S]{flow: flow}
	}
}

type workflowInvoker[I rezai.WorkflowInput, O rezai.WorkflowOutput, S rezai.WorkflowState] struct {
	flow            *core.Flow[I, O, S]
	streamCallbacks []core.StreamCallback[S]
}

func (i workflowInvoker[I, O, S]) AddStreamCallback(cb func(ctx context.Context, stream S) error) {
	i.streamCallbacks = append(i.streamCallbacks, cb)
}

func (i workflowInvoker[I, O, S]) Run(ctx context.Context, inp any) (any, error) {
	var output *O
	input, ok := inp.(I)
	if !ok {
		return nil, fmt.Errorf("invalid input value")
	}
	for res, err := range i.flow.Stream(ctx, input) {
		if err != nil {
			return nil, err
		}
		if res.Done {
			output = &res.Output
			break
		}
		for _, cb := range i.streamCallbacks {
			if cbErr := cb(ctx, res.Stream); cbErr != nil {
				slog.Warn("stream callback error", "error", cbErr)
			}
		}
	}
	return output, nil
}

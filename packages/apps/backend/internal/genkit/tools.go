package genkit

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type Tool[Input any, Output any] interface {
	Name() string
	Description() string
	ToolFunc(context.Context, Input) (Output, error)
	ToolOpts() []ai.ToolOption
}

func WithTool[Input any, Output any](t Tool[Input, Output]) AiServiceOption {
	return func(s *AiService) error {
		ref := genkitx.DefineTool(s.gk, t.Name(), t.Description(), t.ToolFunc, t.ToolOpts()...)
		s.toolRefs = append(s.toolRefs, ref)
		return nil
	}
}

func (s *AiService) getRequiredToolRefs(names []string) ([]ai.ToolRef, error) {
	var refs []ai.ToolRef
	nameSet := mapset.NewSet(names...)
	for _, tool := range s.toolRefs {
		if !nameSet.Contains(tool.Name()) {
			return nil, fmt.Errorf("tool %s not found in tools %v", tool.Name(), names)
		}
		refs = append(refs, tool)
	}
	return refs, nil
}

func makeWriteOutputArtifactTool[O rezai.AgentOutput](partFn func(O) (*ai.Part, error)) *aix.Tool[O, AgentOutputToolResult] {
	artifactName := "output"
	writeOutputArtifactFn := func(ctx context.Context, output O) error {
		as := aix.ArtifactStoreFromContext(ctx)
		if as == nil {
			return fmt.Errorf("no artifact store found in context")
		}
		part, partErr := partFn(output)
		if partErr != nil {
			return fmt.Errorf("encoding part: %w", partErr)
		}
		outputArtifact := &aix.Artifact{Name: artifactName, Parts: []*ai.Part{part}}
		for _, art := range as.Artifacts() {
			if art.Name == artifactName {
				outputArtifact.Parts = append(art.Parts, outputArtifact.Parts...)
				outputArtifact.Metadata = art.Metadata
				break
			}
		}
		as.AddArtifacts(outputArtifact)
		return nil
	}
	return aix.NewTool(
		"write_output",
		"Writes outputs of an agent run. For example a chat message response.",
		func(ctx context.Context, output O) (AgentOutputToolResult, error) {
			status := "Output artifact saved successfully"
			if writeErr := writeOutputArtifactFn(ctx, output); writeErr != nil {
				status = fmt.Sprintf("Error writing result: %s", writeErr.Error())
			}
			return AgentOutputToolResult{Status: status}, nil
		},
	)
}

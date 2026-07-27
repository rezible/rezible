package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type toolCallDisplayLabelMiddleware struct{}

func (m *toolCallDisplayLabelMiddleware) Name() string {
	return "toolcall_display_label"
}

func (m *toolCallDisplayLabelMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapTool: func(ctx context.Context, params *ai.ToolParams, next ai.ToolNext) (*ai.MultipartToolResponse, error) {
			// TODO: wrap tool call with user-facing display text
			return next(ctx, params)
		},
	}, nil
}

type knowledgeGraphMiddleware[I rezai.AgentInput, S rezai.SessionState] struct {
	runner    agentRunner[I, S]
	knowledge rez.KnowledgeGraphService
}

func newKnowledgeGraphMiddleware[I rezai.AgentInput, S rezai.SessionState](knowledge rez.KnowledgeGraphService, runner agentRunner[I, S]) *knowledgeGraphMiddleware[I, S] {
	return &knowledgeGraphMiddleware[I, S]{knowledge: knowledge, runner: runner}
}

func (m *knowledgeGraphMiddleware[I, S]) Name() string {
	return "knowledge_graph"
}

func (m *knowledgeGraphMiddleware[I, S]) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		Tools: []ai.Tool{
			m.makeQueryTool(),
		},
		WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {

			return next(ctx, params)
		},
	}, nil
}

func (m *knowledgeGraphMiddleware[I, S]) makeQueryTool() ai.Tool {
	return aix.NewTool(
		rezai.QueryKnowledgeGraphTool.Name(),
		rezai.QueryKnowledgeGraphTool.Description(),
		func(ctx context.Context, input rezai.QueryKnowledgeGraphInput) (rezai.QueryKnowledgeGraphOutput, error) {
			output, queryErr := m.query(ctx, input)
			if queryErr != nil {
				return rezai.QueryKnowledgeGraphOutput{}, queryErr
			}
			return output, nil
		},
	)
}

func (m *knowledgeGraphMiddleware[I, S]) query(ctx context.Context, input rezai.QueryKnowledgeGraphInput) (rezai.QueryKnowledgeGraphOutput, error) {
	entityID, parseErr := uuid.Parse(input.EntityID)
	if parseErr != nil {
		return rezai.QueryKnowledgeGraphOutput{}, fmt.Errorf("invalid knowledge graph entity ID %q: %w", input.EntityID, parseErr)
	}
	viewParams := rez.GetKnowledgeGraphViewParams{EntityID: entityID, Depth: input.Depth}
	view, viewErr := m.knowledge.GetView(ctx, viewParams)
	if viewErr != nil {
		return rezai.QueryKnowledgeGraphOutput{}, fmt.Errorf("get knowledge graph view: %w", viewErr)
	}

	output := rezai.QueryKnowledgeGraphOutput{
		RootEntityID:  entityID.String(),
		Truncated:     view.Truncated,
		Entities:      make([]rezai.KnowledgeGraphToolEntity, 0, len(view.Entities)),
		Relationships: make([]rezai.KnowledgeGraphToolRelationship, 0, len(view.Relationships)),
		Evidence:      []rezai.KnowledgeGraphToolEvidence{},
	}

	for _, entity := range view.Entities {
		outputEntity := rezai.KnowledgeGraphToolEntity{
			ID:   entity.ID.String(),
			Kind: entity.Kind,
		}
		if currEv := entity.LatestEvidence(); currEv != nil {
			outputEntity.DisplayName = currEv.SubjectState.DisplayName
			outputEntity.Description = currEv.SubjectState.Description
			outputEntity.Properties = currEv.SubjectState.Properties
		}
		output.Entities = append(output.Entities, outputEntity)
	}
	for _, relationship := range view.Relationships {
		outputRelationship := rezai.KnowledgeGraphToolRelationship{
			ID:       relationship.ID.String(),
			Kind:     relationship.Kind,
			SourceID: relationship.SourceEntityID.String(),
			TargetID: relationship.TargetEntityID.String(),
		}
		if currEv := relationship.LatestEvidence(); currEv != nil {
			outputRelationship.Description = currEv.SubjectState.Description
			outputRelationship.Properties = currEv.SubjectState.Properties
		}
		output.Relationships = append(output.Relationships, outputRelationship)
	}
	return output, nil
}

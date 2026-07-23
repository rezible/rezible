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
	view, viewErr := m.knowledge.GetView(ctx, entityID, rez.GetKnowledgeGraphViewParams{Depth: input.Depth})
	if viewErr != nil {
		return rezai.QueryKnowledgeGraphOutput{}, fmt.Errorf("get knowledge graph view: %w", viewErr)
	}

	output := rezai.QueryKnowledgeGraphOutput{
		RootEntityID:  entityID.String(),
		Truncated:     view.Truncated,
		Warnings:      view.Warnings,
		Entities:      make([]rezai.KnowledgeGraphToolEntity, 0, len(view.Entities)),
		Relationships: make([]rezai.KnowledgeGraphToolRelationship, 0, len(view.Relationships)),
		Evidence:      make([]rezai.KnowledgeGraphToolEvidence, len(view.Evidence)),
	}

	for _, entity := range view.Entities {
		output.Entities = append(output.Entities, rezai.KnowledgeGraphToolEntity{
			ID:          entity.ID.String(),
			Kind:        entity.Kind,
			DisplayName: entity.DisplayName,
			Description: entity.Description,
			Properties:  entity.LiveProperties,
		})
	}
	for _, relationship := range view.Relationships {
		output.Relationships = append(output.Relationships, rezai.KnowledgeGraphToolRelationship{
			ID:          relationship.ID.String(),
			Kind:        relationship.Kind,
			SourceID:    relationship.SourceEntityID.String(),
			TargetID:    relationship.TargetEntityID.String(),
			Description: relationship.Description,
			Properties:  relationship.Properties,
		})
	}

	for i, evidence := range view.Evidence {
		item := rezai.KnowledgeGraphToolEvidence{
			ID:           evidence.ID.String(),
			EventID:      evidence.EventID.String(),
			Assertion:    evidence.Assertion,
			EvidenceKind: evidence.EvidenceKind.String(),
			EffectiveAt:  evidence.EffectiveAt,
			Properties:   evidence.Properties,
		}
		if event, edgeErr := evidence.Edges.EventOrErr(); edgeErr == nil {
			item.Provider = event.Provider
			item.ProviderSource = event.ProviderSource
		}
		if alias, edgeErr := evidence.Edges.AliasOrErr(); edgeErr == nil {
			if alias.EntityID != uuid.Nil {
				item.EntityID = alias.EntityID.String()
			}
			if alias.RelationshipID != uuid.Nil {
				item.RelationshipID = alias.RelationshipID.String()
			}
		}
		output.Evidence[i] = item
	}
	return output, nil
}

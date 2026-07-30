package genkit

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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
			m.makeRecordCitationTool(),
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
			outputEntity.State = currEv.SubjectState
			output.Evidence = append(output.Evidence, knowledgeEvidenceToolOutput(currEv, entity.ID.String(), ""))
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
			outputRelationship.State = currEv.SubjectState
			output.Evidence = append(output.Evidence, knowledgeEvidenceToolOutput(currEv, "", relationship.ID.String()))
		}
		output.Relationships = append(output.Relationships, outputRelationship)
	}
	return output, nil
}

func knowledgeEvidenceToolOutput(ev *ent.KnowledgeEvidence, entityID string, relationshipID string) rezai.KnowledgeGraphToolEvidence {
	output := rezai.KnowledgeGraphToolEvidence{
		ID:             ev.ID.String(),
		Assertion:      ev.Assertion,
		EvidenceKind:   ev.Kind.String(),
		EffectiveAt:    ev.EffectiveAt,
		Properties:     ev.SubjectState.Properties,
		EntityID:       entityID,
		RelationshipID: relationshipID,
	}
	return output
}

type evidenceCitationCustomArtifactPart map[string]any

func (a evidenceCitationCustomArtifactPart) Citation() (*rez.AgentKnowledgeCitation, error) {
	stringId, exists := a["evidence_id"]
	if !exists {
		return nil, fmt.Errorf("evidence ID not found in citation custom artifact")
	}
	id, idErr := uuid.Parse(stringId.(string))
	if idErr != nil {
		return nil, fmt.Errorf("evidence ID not found in citation custom artifact: %w", idErr)
	}
	summary, summaryExists := a["summary"].(string)
	if !summaryExists {
		return nil, fmt.Errorf("summary not found in citation custom artifact")
	}
	return &rez.AgentKnowledgeCitation{
		EvidenceID: id,
		Summary:    summary,
	}, nil
}

func getAgentKnowledgeCitations(artifacts []*aix.Artifact) ([]rez.AgentKnowledgeCitation, error) {
	var citations []rez.AgentKnowledgeCitation
	for _, a := range artifacts {
		if a.Name == "citations" {
			for _, p := range a.Parts {
				c, cErr := evidenceCitationCustomArtifactPart(p.Custom).Citation()
				if cErr != nil {
					return nil, cErr
				}
				citations = append(citations, *c)
			}
			break
		}
	}
	return citations, nil
}

func (m *knowledgeGraphMiddleware[I, S]) makeRecordCitationTool() ai.Tool {
	return aix.NewTool(
		rezai.RecordKnowledgeCitationsTool.Name(),
		rezai.RecordKnowledgeCitationsTool.Description(),
		func(ctx context.Context, input rezai.RecordKnowledgeCitationsInput) (rezai.RecordKnowledgeCitationsOutput, error) {
			as := aix.ArtifactStoreFromContext(ctx)

			citationsArtifact := &aix.Artifact{Name: "citations"}
			for _, a := range as.Artifacts() {
				if a.Name == citationsArtifact.Name {
					citationsArtifact = a
					break
				}
			}
			recorded := 0
			for _, citation := range input.Citations {
				evidenceID, parseErr := uuid.Parse(citation.EvidenceID)
				if parseErr != nil {
					return rezai.RecordKnowledgeCitationsOutput{}, fmt.Errorf("invalid evidence ID %q: %w", citation.EvidenceID, parseErr)
				}
				summary := strings.TrimSpace(citation.Summary)
				if summary == "" {
					return rezai.RecordKnowledgeCitationsOutput{}, fmt.Errorf("%w: citation summary is required", rez.ErrInvalidInput)
				}
				if _, evidenceErr := m.knowledge.GetEvidence(ctx, evidenceID); evidenceErr != nil {
					return rezai.RecordKnowledgeCitationsOutput{}, fmt.Errorf("knowledge evidence %s: %w", evidenceID, evidenceErr)
				}
				part := evidenceCitationCustomArtifactPart{
					"evidence_id": evidenceID,
					"summary":     summary,
				}
				citationsArtifact.Parts = append(citationsArtifact.Parts, ai.NewCustomPart(part))
				recorded++
			}
			as.AddArtifacts(citationsArtifact)
			return rezai.RecordKnowledgeCitationsOutput{Recorded: recorded}, nil
		},
	)
}

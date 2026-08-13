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

type (
	AgentDetails struct {
		Name string
	}
	AgentMiddlewareConstructorFn = func(AgentDetails) ai.Middleware
)

type agentDebugMiddleware struct{}

func (m *agentDebugMiddleware) Name() string {
	return "agent_debug"
}

func (m *agentDebugMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
			resp, respErr := next(ctx, params)
			//pretty.Println("model response", resp, "error", respErr)
			return resp, respErr
		},
	}, nil
}

type toolCallDisplayLabelMiddleware struct{}

func (m *toolCallDisplayLabelMiddleware) Name() string {
	return "toolcall_display_label"
}

func (m *toolCallDisplayLabelMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapTool: func(ctx context.Context, params *ai.ToolParams, next ai.ToolNext) (*ai.MultipartToolResponse, error) {
			// TODO: wrap tool call params to add user-facing display text field
			return next(ctx, params)
		},
	}, nil
}

func WithIntegrationToolsMiddleware(integrations rez.IntegrationService) AgentMiddlewareConstructorFn {
	return func(d AgentDetails) ai.Middleware {
		return newIntegrationToolsMiddleware(d.Name, integrations)
	}
}

type integrationToolsMiddleware struct {
	agentName    string
	integrations rez.IntegrationService
}

func newIntegrationToolsMiddleware(agentName string, integrations rez.IntegrationService) *integrationToolsMiddleware {
	return &integrationToolsMiddleware{agentName: agentName, integrations: integrations}
}

func (m *integrationToolsMiddleware) Name() string {
	return "integration_tools"
}

func (m *integrationToolsMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	params := rez.GetAvailableAgentToolsParams{AgentName: m.agentName}
	tools, toolsErr := m.integrations.GetAvailableAgentTools(ctx, params)
	if toolsErr != nil {
		return nil, fmt.Errorf("get available integration agent tools: %w", toolsErr)
	}
	return &ai.Hooks{Tools: tools}, nil
}

type knowledgeGraphMiddleware struct {
	knowledge rez.KnowledgeGraphService
}

func newKnowledgeGraphMiddleware(knowledge rez.KnowledgeGraphService) *knowledgeGraphMiddleware {
	return &knowledgeGraphMiddleware{knowledge: knowledge}
}

func (m *knowledgeGraphMiddleware) Name() string {
	return "knowledge_graph"
}

func (m *knowledgeGraphMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	tools := []ai.Tool{
		m.makeQueryTool(),
		m.makeRecordCitationTool(),
	}
	return &ai.Hooks{Tools: tools}, nil
}

func (m *knowledgeGraphMiddleware) makeQueryTool() ai.Tool {
	return aix.NewTool(
		rezai.QueryKnowledgeGraphTool.Name(),
		rezai.QueryKnowledgeGraphTool.Description(),
		func(ctx context.Context, input rezai.QueryKnowledgeGraphInput) (rezai.QueryKnowledgeGraphOutput, error) {
			output, queryErr := m.query(ctx, input)
			if queryErr != nil || output == nil {
				return rezai.QueryKnowledgeGraphOutput{}, queryErr
			}
			return *output, nil
		},
	)
}

func (m *knowledgeGraphMiddleware) query(ctx context.Context, input rezai.QueryKnowledgeGraphInput) (*rezai.QueryKnowledgeGraphOutput, error) {
	entityID, parseErr := uuid.Parse(input.EntityID)
	if parseErr != nil {
		return nil, fmt.Errorf("invalid knowledge graph entity ID %q: %w", input.EntityID, parseErr)
	}
	viewParams := rez.GetKnowledgeGraphViewParams{EntityID: entityID, Depth: input.Depth}
	view, viewErr := m.knowledge.GetView(ctx, viewParams)
	if viewErr != nil {
		return nil, fmt.Errorf("get knowledge graph view: %w", viewErr)
	}

	output := &rezai.QueryKnowledgeGraphOutput{
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
			e := knowledgeEvidenceToolOutput(currEv)
			e.EntityID = entity.ID.String()
			output.Evidence = append(output.Evidence, e)
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
			ev := knowledgeEvidenceToolOutput(currEv)
			ev.RelationshipID = relationship.ID.String()
			output.Evidence = append(output.Evidence, ev)
		}
		output.Relationships = append(output.Relationships, outputRelationship)
	}

	return output, nil
}

func knowledgeEvidenceToolOutput(ev *ent.KnowledgeEvidence) rezai.KnowledgeGraphToolEvidence {
	return rezai.KnowledgeGraphToolEvidence{
		ID:           ev.ID.String(),
		Assertion:    ev.Assertion,
		EvidenceKind: ev.Kind.String(),
		EffectiveAt:  ev.EffectiveAt,
		Properties:   ev.SubjectState.Properties,
	}
}

type evidenceCitationCustomArtifactPart map[string]any

func (a evidenceCitationCustomArtifactPart) Citation() (*rez.AiAgentKnowledgeCitation, error) {
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
	return &rez.AiAgentKnowledgeCitation{
		EvidenceID: id,
		Summary:    summary,
	}, nil
}

func (m *knowledgeGraphMiddleware) makeRecordCitationTool() ai.Tool {
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

func getAgentKnowledgeCitations(artifacts []*aix.Artifact) ([]rez.AiAgentKnowledgeCitation, error) {
	var citations []rez.AiAgentKnowledgeCitation
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

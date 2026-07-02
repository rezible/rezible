package genkit

import (
	"context"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	rez "github.com/rezible/rezible"
)

type AgentRegistry struct {
	cfg       rez.AiConfig
	gk        *genkit.Genkit
	snapshots rez.AgentRunSnapshotService
	agents    map[string]rez.Agent
}

func NewAgentRegistry(ctx context.Context, cfg rez.Config, snapshots rez.AgentRunSnapshotService) *AgentRegistry {
	gkOpts := []genkit.GenkitOption{
		genkit.WithExperimental(),
	}
	if cfg.AI.Gemini.Enabled {
		gkOpts = append(gkOpts, genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: cfg.AI.Gemini.APIKey,
		}))
	}
	return &AgentRegistry{
		cfg:       cfg.AI,
		gk:        genkit.Init(ctx, gkOpts...),
		snapshots: snapshots,
		agents:    make(map[string]rez.Agent),
	}
}

func (r *AgentRegistry) Register(a rez.Agent) {
	r.agents[a.Workflow()] = a
}

func RegisterAgent[S any](r *AgentRegistry, wa agent[S]) {
	r.Register(wrapAgent(r.gk, makeAgentSessionStore[S](r.snapshots), wa))
}

func (r *AgentRegistry) RegisterMultiple(was ...agent[any]) {
	for _, wa := range was {
		RegisterAgent[any](r, wa)
	}
}

func (r *AgentRegistry) Get(workflow string) (rez.Agent, bool) {
	a, ok := r.agents[workflow]
	return a, ok
}

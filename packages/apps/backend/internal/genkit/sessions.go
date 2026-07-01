package genkit

import (
	"context"

	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
)

type agentSessionStore[S any] struct {
	agents rez.AgentService
}

func makeAgentSessionStore[S any](agents rez.AgentService) *agentSessionStore[S] {
	return &agentSessionStore[S]{agents: agents}
}

func (s *agentSessionStore[S]) GetSnapshot(ctx context.Context, snapshotID string) (*aix.SessionSnapshot[S], error) {
	//TODO implement me
	panic("implement me")
}

func (s *agentSessionStore[S]) GetLatestSnapshot(ctx context.Context, sessionID string) (*aix.SessionSnapshot[S], error) {
	//TODO implement me
	panic("implement me")
}

func (s *agentSessionStore[S]) SaveSnapshot(ctx context.Context, snapshotID string, setFn func(existing *aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error)) (*aix.SessionSnapshot[S], error) {
	//TODO implement me
	panic("implement me")
}

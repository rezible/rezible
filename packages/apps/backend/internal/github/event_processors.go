package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/go-github/v84/github"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/internal/projections"
)

const zeroSHA = "0000000000000000000000000000000000000000"

func lookupTenantIntegration(ctx context.Context, integrations rez.IntegrationsService, orgName string) (*ConfiguredIntegration, error) {
	params := rez.ListIntegrationsParams{
		Names:        []string{integrationName},
		ConfigValues: map[string]any{"org": orgName},
	}
	intgs, listErr := integrations.ListConfigured(execution.SystemContext(ctx), params)
	if listErr != nil {
		if ent.IsNotFound(listErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing configured integrations: %w", listErr)
	}
	for _, intg := range intgs {
		if ci, ok := intg.(*ConfiguredIntegration); ok {
			return ci, nil
		}
	}
	return nil, nil
}

type pushEventProcessor struct {
	services *rez.Services
}

func (p *pushEventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var event github.PushEvent
	if err := json.Unmarshal(prov.Payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal push event: %w", err)
	}

	if event.GetAfter() == zeroSHA {
		return nil, nil
	}

	orgName := event.GetRepo().GetOwner().GetLogin()
	ci, lookupErr := lookupTenantIntegration(ctx, p.services.Integrations, orgName)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if ci == nil {
		slog.WarnContext(ctx, "received github push event with no configured integration", "org", orgName)
		return nil, nil
	}

	var occurredAt time.Time
	if hc := event.GetHeadCommit(); hc != nil {
		occurredAt = hc.GetTimestamp().Time
	}

	result := &ent.NormalizedEvent{
		TenantID:          ci.intg.TenantID,
		Provider:          integrationName,
		ProviderSource:    "push",
		Kind:              ne.KindChangeEventObserved,
		SubjectKind:       "change_event",
		ProviderEventRef:  event.GetAfter(),
		SubjectRef:        fmt.Sprintf("github:%s:%s", event.GetRepo().GetFullName(), event.GetAfter()),
		OccurredAt:        occurredAt,
		ProcessingVersion: "github.change-event-observed.v1",
		DedupeKey:         prov.DedupeKey,
		Attributes: projections.ChangeEventObservedAttributes{
			RepositoryExternalRef: event.GetRepo().GetFullName(),
			DisplayName:           event.GetRef(),
		}.Encode(),
	}

	return ent.NormalizedEvents{result}, nil
}

type pullRequestEventProcessor struct {
	services *rez.Services
}

func (p *pullRequestEventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(prov.Payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal pull_request event: %w", err)
	}

	orgName := event.GetRepo().GetOwner().GetLogin()
	ci, lookupErr := lookupTenantIntegration(ctx, p.services.Integrations, orgName)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if ci == nil {
		slog.WarnContext(ctx, "received github pull_request event with no configured integration", "org", orgName)
		return nil, nil
	}

	pr := event.GetPullRequest()
	prNum := pr.GetNumber()

	result := &ent.NormalizedEvent{
		TenantID:          ci.intg.TenantID,
		Provider:          integrationName,
		ProviderSource:    "pull_request",
		Kind:              ne.KindChangeEventObserved,
		SubjectKind:       "change_event",
		ProviderEventRef:  fmt.Sprintf("pr:%d", prNum),
		SubjectRef:        fmt.Sprintf("github:%s:pr:%d", event.GetRepo().GetFullName(), prNum),
		OccurredAt:        pr.GetCreatedAt().Time,
		ProcessingVersion: "github.change-event-observed.v1",
		DedupeKey:         prov.DedupeKey,
		Attributes: projections.ChangeEventObservedAttributes{
			RepositoryExternalRef: event.GetRepo().GetFullName(),
			DisplayName:           pr.GetTitle(),
		}.Encode(),
	}

	return ent.NormalizedEvents{result}, nil
}

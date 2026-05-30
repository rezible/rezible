package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/integrations/projections"
)

const zeroSHA = "0000000000000000000000000000000000000000"

func (i *Integration) MakeProviderEventProcessor() rez.ProviderEventProcessor {
	return &eventProcessor{}
}

type eventProcessor struct {
}

const (
	sourcePushEvent    = "push"
	sourcePullEvent    = "pull_request"
	sourceRepositories = "repositories"
)

func (p *eventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	switch prov.ProviderSource {
	case sourcePushEvent:
		return p.processPushEvent(prov)
	case sourcePullEvent:
		return p.processPullRequest(prov)
	case sourceRepositories:
		return p.processRepoObserved(prov)
	}
	return nil, fmt.Errorf("unknown provider source: %s", prov.ProviderSource)
}

func (p *eventProcessor) processPushEvent(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var event github.PushEvent
	if err := json.Unmarshal(prov.Payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal push event: %w", err)
	}

	if event.GetAfter() == zeroSHA {
		return nil, nil
	}

	var occurredAt time.Time
	if hc := event.GetHeadCommit(); hc != nil {
		occurredAt = hc.GetTimestamp().Time
	}

	ProviderSubjectRef := prov.ProviderSubjectRef
	if ProviderSubjectRef == "" {
		ProviderSubjectRef = fmt.Sprintf("github:%s:%s", event.GetRepo().GetFullName(), event.GetAfter())
	}

	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: event.GetRepo().GetFullName(),
		DisplayName:           event.GetRef(),
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode change event observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourcePushEvent,
		ProviderEventRef:   prov.ProviderEventRef,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindCodeChange.String(),
		ProviderSubjectRef: ProviderSubjectRef,
		OccurredAt:         occurredAt,
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processPullRequest(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(prov.Payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal pull_request event: %w", err)
	}

	pr := event.GetPullRequest()
	prNum := pr.GetNumber()

	ProviderSubjectRef := prov.ProviderSubjectRef
	if ProviderSubjectRef == "" {
		ProviderSubjectRef = fmt.Sprintf("github:%s:pr:%d", event.GetRepo().GetFullName(), prNum)
	}

	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: event.GetRepo().GetFullName(),
		DisplayName:           pr.GetTitle(),
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode change event observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourcePullEvent,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindCodeChange.String(),
		ProviderEventRef:   prov.ProviderEventRef,
		ProviderSubjectRef: ProviderSubjectRef,
		OccurredAt:         pr.GetCreatedAt().Time,
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processRepoObserved(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload githubRepositoryObservedPayload
	if err := json.Unmarshal(prov.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal repository observed event: %w", err)
	}
	if payload.FullName == "" {
		return nil, fmt.Errorf("repository observed payload missing full_name")
	}

	occurredAt := payload.UpdatedAt
	if occurredAt.IsZero() {
		occurredAt = payload.CreatedAt
	}
	if occurredAt.IsZero() {
		occurredAt = prov.ReceivedAt
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	repositoryRef := payload.FullName

	attrs := projections.CodeForgeSubjectAttributes{
		DisplayName: repositoryRef,
		URL:         payload.HTMLURL,
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode repository observed attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:           integrationName,
		ProviderSource:     sourceRepositories,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindCodeForge.String(),
		ProviderEventRef:   prov.ProviderEventRef,
		ProviderSubjectRef: repositoryRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         prov.ReceivedAt,
		Attributes:         encodedAttrs,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

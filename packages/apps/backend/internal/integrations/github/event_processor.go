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
	"github.com/rezible/rezible/pkg/projections"
)

const zeroSHA = "0000000000000000000000000000000000000000"

func (i *Integration) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	p := &eventProcessor{event: &prov}
	return p.process()
}

type eventProcessor struct {
	event *rez.ProviderEvent
}

const (
	sourcePushEvent    = "push"
	sourcePullEvent    = "pull_request"
	sourceRepositories = "repositories"
)

func (p *eventProcessor) process() (ent.NormalizedEvents, error) {
	switch p.event.ProviderSource {
	case sourcePushEvent:
		return p.processPushEvent()
	case sourcePullEvent:
		return p.processPullRequest()
	case sourceRepositories:
		return p.processRepoObserved()
	default:
		return nil, fmt.Errorf("unknown provider source: %s", p.event.ProviderSource)
	}
}

func (p *eventProcessor) processPushEvent() (ent.NormalizedEvents, error) {
	var event github.PushEvent
	if err := json.Unmarshal(p.event.Payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal push event: %w", err)
	}

	if event.GetAfter() == zeroSHA {
		return nil, nil
	}

	var occurredAt time.Time
	if hc := event.GetHeadCommit(); hc != nil {
		occurredAt = hc.GetTimestamp().Time
	}

	ProviderSubjectRef := p.event.ProviderSubjectRef
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
		ProviderEventRef:   p.event.ProviderEventRef,
		ActivityKind:       ne.ActivityKindObserved,
		SubjectKind:        projections.SubjectKindCodeChange.String(),
		ProviderSubjectRef: ProviderSubjectRef,
		OccurredAt:         occurredAt,
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processPullRequest() (ent.NormalizedEvents, error) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(p.event.Payload, &event); err != nil {
		return nil, fmt.Errorf("unmarshal pull_request event: %w", err)
	}

	pr := event.GetPullRequest()
	prNum := pr.GetNumber()

	ProviderSubjectRef := p.event.ProviderSubjectRef
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
		ProviderEventRef:   p.event.ProviderEventRef,
		ProviderSubjectRef: ProviderSubjectRef,
		OccurredAt:         pr.GetCreatedAt().Time,
		Attributes:         encodedAttrs,
	}

	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processRepoObserved() (ent.NormalizedEvents, error) {
	var payload githubRepositoryObservedPayload
	if err := json.Unmarshal(p.event.Payload, &payload); err != nil {
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
		occurredAt = p.event.ReceivedAt
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
		ProviderEventRef:   p.event.ProviderEventRef,
		ProviderSubjectRef: repositoryRef,
		OccurredAt:         occurredAt,
		ReceivedAt:         p.event.ReceivedAt,
		Attributes:         encodedAttrs,
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

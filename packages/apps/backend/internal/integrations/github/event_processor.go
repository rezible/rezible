package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/rezible/rezible/integrations/eventprojections"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
)

const zeroSHA = "0000000000000000000000000000000000000000"

func (i *integration) MakeProviderEventProcessor() rez.ProviderEventProcessor {
	return &eventProcessor{services: i.services}
}

type eventProcessor struct {
	services *rez.Services
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

	subjectRef := prov.SubjectRef
	if subjectRef == "" {
		subjectRef = fmt.Sprintf("github:%s:%s", event.GetRepo().GetFullName(), event.GetAfter())
	}

	attrs := eventprojections.ChangeEventObservedAttributes{
		RepositoryExternalRef: event.GetRepo().GetFullName(),
		DisplayName:           event.GetRef(),
	}
	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourcePushEvent,
		ProviderEventRef: prov.ProviderEventRef,
		Kind:             ne.KindChangeEventObserved,
		SubjectKind:      "change_event",
		SubjectRef:       subjectRef,
		OccurredAt:       occurredAt,
		Attributes:       attrs.Encode(),
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

	subjectRef := prov.SubjectRef
	if subjectRef == "" {
		subjectRef = fmt.Sprintf("github:%s:pr:%d", event.GetRepo().GetFullName(), prNum)
	}

	attrs := eventprojections.ChangeEventObservedAttributes{
		RepositoryExternalRef: event.GetRepo().GetFullName(),
		DisplayName:           pr.GetTitle(),
	}
	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourcePullEvent,
		Kind:             ne.KindChangeEventObserved,
		SubjectKind:      "change_event",
		ProviderEventRef: prov.ProviderEventRef,
		SubjectRef:       subjectRef,
		OccurredAt:       pr.GetCreatedAt().Time,
		Attributes:       attrs.Encode(),
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

	attrs := eventprojections.RepositoryObservedAttributes{
		DisplayName: repositoryRef,
		URL:         payload.HTMLURL,
	}
	result := &ent.NormalizedEvent{
		Provider:         integrationName,
		ProviderSource:   sourceRepositories,
		Kind:             ne.KindRepositoryObserved,
		SubjectKind:      "repository",
		ProviderEventRef: prov.ProviderEventRef,
		SubjectRef:       repositoryRef,
		OccurredAt:       occurredAt,
		ReceivedAt:       prov.ReceivedAt,
		Attributes:       attrs.Encode(),
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

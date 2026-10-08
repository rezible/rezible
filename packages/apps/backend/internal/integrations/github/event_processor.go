package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

const zeroSHA = "0000000000000000000000000000000000000000"

// EventProcessor normalizes GitHub provider events. It has no dependencies.
type EventProcessor struct{}

func (p EventProcessor) ProcessProviderEvent(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return (&eventProcessor{event: &prov}).process()
}

type eventProcessor struct {
	event *rez.ProviderEvent
}

func (p *eventProcessor) process() (ent.NormalizedEvents, error) {
	switch p.event.ProviderEventSource {
	case sourcePushEvent:
		return p.processPushEvent()
	case sourcePullEvent:
		return p.processPullRequest()
	case sourceRepositories:
		return p.processRepoObserved()
	default:
		return nil, fmt.Errorf("unknown provider event source: %s", p.event.ProviderEventSource)
	}
}

func (p *eventProcessor) processPushEvent() (ent.NormalizedEvents, error) {
	var event github.PushEvent
	if err := json.Unmarshal(p.event.Attributes, &event); err != nil {
		return nil, fmt.Errorf("unmarshal push event: %w", err)
	}
	if event.GetAfter() == zeroSHA {
		return nil, nil
	}
	repository := event.GetRepo()
	if repository == nil || repository.GetID() == 0 {
		return nil, fmt.Errorf("push event missing repository id")
	}
	occurredAt := p.event.ReceivedAt
	if headCommit := event.GetHeadCommit(); headCommit != nil && !headCommit.GetTimestamp().Time.IsZero() {
		occurredAt = headCommit.GetTimestamp().Time
	}
	repositoryObservation := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: p.event.ProviderNamespace,
			ResourceRef:       strconv.FormatInt(repository.GetID(), 10),
		},
		Category:          kne.CategoryCode,
		Kind:              "repository",
		DisplayName:       repository.GetFullName(),
		Properties:        map[string]any{"url": repository.GetHTMLURL()},
		LinkingAttributes: projections.RepositoryLinkingAttributes(repository.GetFullName()),
	}
	attrs := projections.CodeChangeEventAttributes{
		Repository:  repositoryObservation,
		DisplayName: event.GetRef(),
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode change event attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            ProviderName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: fmt.Sprintf("change:%d:%s", repository.GetID(), event.GetAfter()),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindCodeChange,
		OccurredAt:          occurredAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

// processPullRequest records merges only: a pull request opened, edited or closed without merging is not a
// change.
func (p *eventProcessor) processPullRequest() (ent.NormalizedEvents, error) {
	var event github.PullRequestEvent
	if unmarshalErr := json.Unmarshal(p.event.Attributes, &event); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal pull request event: %w", unmarshalErr)
	}
	repository := event.GetRepo()
	pullRequest := event.GetPullRequest()
	if repository == nil || repository.GetID() == 0 || pullRequest == nil {
		return nil, fmt.Errorf("pull request event missing repository or pull request")
	}
	if event.GetAction() != "closed" || !pullRequest.GetMerged() {
		return nil, nil
	}
	mergedAt := pullRequest.GetMergedAt().Time
	mergeCommitSha := strings.ToLower(pullRequest.GetMergeCommitSHA())
	if mergedAt.IsZero() || mergeCommitSha == "" {
		return nil, fmt.Errorf("merged pull request missing merge time or merge commit")
	}
	baseRef := pullRequest.GetBase().GetRef()
	repositoryObservation := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: p.event.ProviderNamespace,
			ResourceRef:       strconv.FormatInt(repository.GetID(), 10),
		},
		Category:          kne.CategoryCode,
		Kind:              "repository",
		DisplayName:       repository.GetFullName(),
		Properties:        map[string]any{"url": repository.GetHTMLURL()},
		LinkingAttributes: projections.RepositoryLinkingAttributes(repository.GetFullName()),
	}
	attrs := projections.CodeChangeEventAttributes{
		Repository:  repositoryObservation,
		DisplayName: pullRequest.GetTitle(),
		Merge: &projections.CodeChangeMerge{
			MergedAt:          mergedAt,
			MergeCommitSha:    mergeCommitSha,
			BaseRef:           baseRef,
			IntoDefaultBranch: baseRef != "" && baseRef == repository.GetDefaultBranch(),
			Number:            pullRequest.GetNumber(),
			URL:               pullRequest.GetHTMLURL(),
			AuthorLogin:       pullRequest.GetUser().GetLogin(),
		},
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode pull request attributes: %w", encodeErr)
	}
	result := &ent.NormalizedEvent{
		Provider:            ProviderName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: fmt.Sprintf("change:%d:pr:%d", repository.GetID(), pullRequest.GetNumber()),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindCodeChange,
		OccurredAt:          mergedAt,
		ReceivedAt:          p.event.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

func (p *eventProcessor) processRepoObserved() (ent.NormalizedEvents, error) {
	var payload githubRepositoryObservedPayload
	if err := json.Unmarshal(p.event.Attributes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal repository observed event: %w", err)
	}
	if payload.FullName == "" || payload.ID == 0 {
		return nil, fmt.Errorf("repository observed payload missing repository identity")
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
	attrs := projections.CodeForgeEventAttributes{
		DisplayName:       payload.FullName,
		URL:               payload.HTMLURL,
		LinkingAttributes: projections.RepositoryLinkingAttributes(payload.FullName),
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode repository observed attributes: %w", encodeErr)
	}
	receivedAt := p.event.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}
	result := &ent.NormalizedEvent{
		Provider:            ProviderName,
		ProviderNamespace:   p.event.ProviderNamespace,
		ProviderResourceRef: strconv.FormatInt(payload.ID, 10),
		ProviderEventSource: p.event.ProviderEventSource,
		ProviderEventRef:    p.event.ProviderEventRef,
		Kind:                projections.KindCodeForge,
		OccurredAt:          occurredAt,
		ReceivedAt:          receivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/rezible/rezible/integrations/eventprojections"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"
)

const zeroSHA = "0000000000000000000000000000000000000000"

func lookupTenantIntegration(ctx context.Context, integrations rez.IntegrationsService, installationID int64) (*ConfiguredIntegration, error) {
	if installationID == 0 {
		return nil, nil
	}
	params := rez.ListIntegrationsParams{
		Providers:    []string{integrationName},
		ExternalRefs: []string{strconv.FormatInt(installationID, 10)},
	}
	intgs, listErr := integrations.ListConfigured(execution.NewSystemContext(ctx), params)
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

	installationID := event.GetInstallation().GetID()
	ci, lookupErr := lookupTenantIntegration(ctx, p.services.Integrations, installationID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if ci == nil {
		slog.WarnContext(ctx, "received github push event with no configured integration", "installationId", installationID)
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

	result := &ent.NormalizedEvent{
		Kind:                     ne.KindChangeEventObserved,
		TenantID:                 ci.intg.TenantID,
		Provider:                 integrationName,
		ProviderSource:           "push",
		SubjectKind:              "change_event",
		ProviderEventRef:         event.GetAfter(),
		ProviderEventDeliveryRef: prov.ProviderDeliveryRef,
		SubjectRef:               subjectRef,
		OccurredAt:               occurredAt,
		ProcessingVersion:        "github.change-event-observed.v1",
		Attributes: eventprojections.ChangeEventObservedAttributes{
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

	installationID := event.GetInstallation().GetID()
	ci, lookupErr := lookupTenantIntegration(ctx, p.services.Integrations, installationID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if ci == nil {
		slog.WarnContext(ctx, "received github pull_request event with no configured integration", "installationId", installationID)
		return nil, nil
	}

	pr := event.GetPullRequest()
	prNum := pr.GetNumber()

	subjectRef := prov.SubjectRef
	if subjectRef == "" {
		subjectRef = fmt.Sprintf("github:%s:pr:%d", event.GetRepo().GetFullName(), prNum)
	}

	result := &ent.NormalizedEvent{
		TenantID:                 ci.intg.TenantID,
		Provider:                 integrationName,
		ProviderSource:           "pull_request",
		Kind:                     ne.KindChangeEventObserved,
		SubjectKind:              "change_event",
		ProviderEventRef:         fmt.Sprintf("pr:%d", prNum),
		SubjectRef:               subjectRef,
		OccurredAt:               pr.GetCreatedAt().Time,
		ProcessingVersion:        "github.change-event-observed.v1",
		ProviderEventDeliveryRef: prov.ProviderDeliveryRef,
		Attributes: eventprojections.ChangeEventObservedAttributes{
			RepositoryExternalRef: event.GetRepo().GetFullName(),
			DisplayName:           pr.GetTitle(),
		}.Encode(),
	}

	return ent.NormalizedEvents{result}, nil
}

type repositoryObservedProcessor struct {
	services *rez.Services
}

func (p *repositoryObservedProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var payload githubRepositoryObservedPayload
	if err := json.Unmarshal(prov.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal repository observed event: %w", err)
	}
	if payload.FullName == "" {
		return nil, fmt.Errorf("repository observed payload missing full_name")
	}

	ci, lookupErr := lookupTenantIntegration(ctx, p.services.Integrations, payload.InstallationID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if ci == nil {
		slog.WarnContext(ctx, "received github repository event with no configured integration", "installationId", payload.InstallationID)
		return nil, nil
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

	eventRefID := repositoryRef
	if payload.ID != 0 {
		eventRefID = fmt.Sprintf("%d", payload.ID)
	}
	result := &ent.NormalizedEvent{
		TenantID:                 ci.intg.TenantID,
		Provider:                 integrationName,
		ProviderSource:           "repositories",
		Kind:                     ne.KindRepositoryObserved,
		SubjectKind:              "repository",
		ProviderEventRef:         "repository:" + eventRefID,
		ProviderEventDeliveryRef: prov.ProviderDeliveryRef,
		SubjectRef:               repositoryRef,
		OccurredAt:               occurredAt,
		ReceivedAt:               prov.ReceivedAt,
		ProcessingVersion:        "github.repository-observed.v1",
		Attributes: eventprojections.RepositoryObservedAttributes{
			DisplayName: repositoryRef,
			URL:         payload.HTMLURL,
		}.Encode(),
	}
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = occurredAt
	}

	return ent.NormalizedEvents{result}, nil
}

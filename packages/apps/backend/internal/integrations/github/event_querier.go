package github

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"sort"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
)

func (i *Integration) MakeProviderEventQuerier(cfg rez.IntegrationsConfigGithub, intg *ent.Integration) (rez.IntegrationEventQuerier, error) {
	ii := i.newInstalledIntegration(intg)
	client, clientErr := newAppClient(cfg, ii)
	if clientErr != nil {
		return nil, fmt.Errorf("app client: %w", clientErr)
	}
	icfg, icfgErr := ii.config()
	if icfgErr != nil {
		return nil, fmt.Errorf("config: %w", icfgErr)
	}
	return &eventQuerier{ii: ii, cfg: icfg, client: client}, nil
}

type eventQuerier struct {
	ii     *InstalledIntegration
	cfg    *installationConfig
	client *githubClient
}

func (q *eventQuerier) Integration() *ent.Integration {
	return q.ii.intg
}

func (q *eventQuerier) PullEvents(ctx context.Context, cursors map[string]string) iter.Seq2[*rez.IntegrationEventQueryResult, error] {
	return func(yield func(*rez.IntegrationEventQueryResult, error) bool) {
		if reposCursor, ok := integrations.GetSourceQueryCursor(cursors, sourceRepositories); ok {
			for ev, evErr := range q.pullRepositoryEvents(ctx, reposCursor) {
				yield(ev, evErr)
			}
		}
	}
}

func (q *eventQuerier) pullRepositoryEvents(ctx context.Context, cursorAfter string) iter.Seq2[*rez.IntegrationEventQueryResult, error] {
	return func(yield func(*rez.IntegrationEventQueryResult, error) bool) {
		repos, listErr := q.client.ListRepositories(ctx)
		if listErr != nil {
			yield(nil, listErr)
			return
		}
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].GetFullName() < repos[j].GetFullName()
		})

		for _, repo := range repos {
			if repo == nil || repo.GetFullName() == "" {
				continue
			}
			cursor := repo.GetFullName()
			if cursorAfter != "" && cursor <= cursorAfter {
				continue
			}

			payload := githubRepositoryObservedPayload{
				InstallationID: q.cfg.InstallationID,
				ID:             repo.GetID(),
				FullName:       repo.GetFullName(),
				HTMLURL:        repo.GetHTMLURL(),
				UpdatedAt:      repo.GetUpdatedAt().Time,
				CreatedAt:      repo.GetCreatedAt().Time,
			}
			body, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				if !yield(nil, fmt.Errorf("marshal repository observation: %w", marshalErr)) {
					return
				}
				continue
			}

			receivedAt := payload.UpdatedAt
			if receivedAt.IsZero() {
				receivedAt = payload.CreatedAt
			}
			if receivedAt.IsZero() {
				receivedAt = time.Now().UTC()
			}

			deliveryRefID := payload.FullName
			if payload.ID != 0 {
				deliveryRefID = fmt.Sprintf("%d", payload.ID)
			}

			res := &rez.IntegrationEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:           integrationName,
					ProviderSource:     sourceRepositories,
					ProviderSubjectRef: "github:" + payload.FullName,
					ProviderEventRef:   fmt.Sprintf("github:repositories:%s:%s", deliveryRefID, receivedAt.Format(time.RFC3339Nano)),
					ReceivedAt:         receivedAt,
					Payload:            body,
					ContentType:        "application/json",
				},
				SourceCursorAfter: new(cursor),
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}

type githubRepositoryObservedPayload struct {
	InstallationID int64     `json:"installation_id"`
	ID             int64     `json:"id"`
	FullName       string    `json:"full_name"`
	HTMLURL        string    `json:"html_url,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

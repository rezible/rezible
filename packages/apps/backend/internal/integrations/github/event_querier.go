package github

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"sort"
	"strconv"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func (i *Integration) MakeProviderEventQuerier(cfg rez.IntegrationsConfigGithub, intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	ii, iiErr := i.newInstalledIntegration(intg)
	if iiErr != nil {
		return nil, fmt.Errorf("integration installation: %w", iiErr)
	}
	client, clientErr := newAppClient(cfg, ii)
	if clientErr != nil {
		return nil, fmt.Errorf("app client: %w", clientErr)
	}
	return &eventQuerier{ii: ii, client: client}, nil
}

type eventQuerier struct {
	ii     *InstalledIntegration
	client *githubClient
}

func (q *eventQuerier) Integration() *ent.Integration {
	return q.ii.intg
}

func (q *eventQuerier) QueryProviderEvents(ctx context.Context, cursors rez.ProviderEventSourceCursors) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if reposCursor, ok := cursors.GetForSource(sourceRepositories); ok {
			for ev, evErr := range q.pullRepositoryEvents(ctx, reposCursor) {
				yield(ev, evErr)
			}
		}
	}
}

func (q *eventQuerier) pullRepositoryEvents(ctx context.Context, cursorAfter string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
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
				InstallationID: q.ii.config.InstallationID,
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

			res := &rez.ProviderEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:            providerName,
					ProviderNamespace:   strconv.FormatInt(q.ii.config.AccountID, 10),
					ProviderEventSource: sourceRepositories,
					ProviderEventRef:    fmt.Sprintf("github:repositories:%s:%s", deliveryRefID, receivedAt.Format(time.RFC3339Nano)),
					ReceivedAt:          receivedAt,
					Attributes:          body,
				},
				ProviderEventSourceCursorAfter: new(cursor),
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

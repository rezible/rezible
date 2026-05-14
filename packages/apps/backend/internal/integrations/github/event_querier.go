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
)

func (i *integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	ci := newConfiguredIntegration(i.services, intg)
	client, clientErr := newClient(ci)
	if clientErr != nil {
		return nil, clientErr
	}
	return &eventQuerier{ci: ci, client: client}, nil
}

type eventQuerier struct {
	ci     *ConfiguredIntegration
	client *githubClient
}

func (q *eventQuerier) Provider() string {
	return integrationName
}

func (q *eventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if reposCursor, ok := req.SourceCursors[sourceRepositories]; ok || len(req.SourceCursors) == 0 {
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
				InstallationID: q.ci.installationID(),
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
					Provider:         integrationName,
					ProviderSource:   sourceRepositories,
					SubjectRef:       "github:" + payload.FullName,
					ProviderEventRef: fmt.Sprintf("github:repositories:%s:%s", deliveryRefID, receivedAt.Format(time.RFC3339Nano)),
					ReceivedAt:       receivedAt,
					Payload:          body,
					ContentType:      "application/json",
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

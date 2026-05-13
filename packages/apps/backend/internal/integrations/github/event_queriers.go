package github

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"sort"
	"time"

	rez "github.com/rezible/rezible"
)

type repositoryEventQuerier struct {
	ci     *ConfiguredIntegration
	client *githubClient
}

func (q *repositoryEventQuerier) Provider() string {
	return integrationName
}

func (q *repositoryEventQuerier) ProviderSource() string {
	return "repositories"
}

func (q *repositoryEventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[rez.ProviderEventQueryResult, error] {
	return func(yield func(rez.ProviderEventQueryResult, error) bool) {
		repos, listErr := q.client.ListRepositories(ctx)
		if listErr != nil {
			yield(rez.ProviderEventQueryResult{}, listErr)
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
			if req.CursorAfter != "" && cursor <= req.CursorAfter {
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
				if !yield(rez.ProviderEventQueryResult{}, fmt.Errorf("marshal repository observation: %w", marshalErr)) {
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
			if !yield(rez.ProviderEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:            integrationName,
					ProviderSource:      "repositories",
					SubjectRef:          "github:" + payload.FullName,
					ProviderDeliveryRef: fmt.Sprintf("github:repositories:%s:%s", deliveryRefID, receivedAt.Format(time.RFC3339Nano)),
					ReceivedAt:          receivedAt,
					Payload:             body,
					ContentType:         "application/json",
					RequestMetadata: map[string]string{
						"integration_id": q.ci.ID().String(),
					},
				},
				CursorAfter: new(cursor),
			}, nil) {
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

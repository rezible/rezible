package github

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/projections"
)

type systemComponentsProvider struct {
	client   *githubClient
	services *rez.Services
}

var _ rez.SystemComponentsDataProvider = (*systemComponentsProvider)(nil)

func (p *systemComponentsProvider) SystemComponentDataMapping() *ent.SystemComponent {
	return &ent.SystemComponent{}
}

func (p *systemComponentsProvider) PullSystemComponents(ctx context.Context) iter.Seq2[*ent.SystemComponent, error] {
	return func(yield func(*ent.SystemComponent, error) bool) {
		repos, err := p.client.ListRepositories(ctx)
		if err != nil {
			yield(nil, fmt.Errorf("list repositories: %w", err))
			return
		}

		for _, repo := range repos {
			component := &ent.SystemComponent{
				ExternalID: repo.GetFullName(),
				Name:       repo.GetName(),
			}
			if !yield(component, nil) {
				return
			}

			attrs := projections.RepositoryObservedAttributes{
				DisplayName: repo.GetName(),
				URL:         repo.GetHTMLURL(),
			}.Encode()
			payload, jsonErr := json.Marshal(attrs)
			if jsonErr != nil {
				yield(nil, fmt.Errorf("marshal repository event payload: %w", jsonErr))
				return
			}

			pe := rez.ProviderEvent{
				Provider:  integrationName,
				Source:    "api/repositories",
				DedupeKey: "repo:" + repo.GetFullName(),
				Payload:   payload,
			}
			if ingestErr := p.services.ProviderEvents.Ingest(ctx, pe); ingestErr != nil {
				yield(nil, fmt.Errorf("ingest repository event: %w", ingestErr))
				return
			}
		}
	}
}

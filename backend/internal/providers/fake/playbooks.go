package fakeprovider

import (
	"context"
	"iter"

	"github.com/rezible/rezible/ent"
)

type PlaybookDataProvider struct {
}

type PlaybookDataProviderConfig struct {
}

func NewPlaybookDataProvider(cfg PlaybookDataProviderConfig) (*PlaybookDataProvider, error) {
	return &PlaybookDataProvider{}, nil
}

var (
	playbookDataMapping = &ent.Playbook{}
)

func (p *PlaybookDataProvider) TeamDataMapping() *ent.Playbook {
	return playbookDataMapping
}

var fakePlaybook = &ent.Playbook{
	Title:      "Example Playbook",
	ProviderID: "example-playbook",
	Content:    []byte("<p>hello</p>"),
}

func (p *PlaybookDataProvider) PullPlaybooks(ctx context.Context) iter.Seq2[*ent.Playbook, error] {
	return func(yield func(*ent.Playbook, error) bool) {
		yield(fakePlaybook, nil)
	}
}

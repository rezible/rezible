package postgres

import (
	"context"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type PlaybookService struct {
	db   *ent.Client
	prov rez.PlaybookDataProvider
}

func NewPlaybookService(db *ent.Client, prov rez.PlaybookDataProvider) (*PlaybookService, error) {
	s := &PlaybookService{
		db:   db,
		prov: prov,
	}

	return s, nil
}

func (s *PlaybookService) ListPlaybooks(ctx context.Context, params *rez.ListPlaybooksParams) ([]*ent.Playbook, int, error) {
	// TODO: don't fake
	fakePlaybook := &ent.Playbook{
		ID:         uuid.New(),
		Title:      "Example Playbook",
		ProviderID: "example",
		Content:    nil,
		Edges:      ent.PlaybookEdges{},
	}
	return []*ent.Playbook{fakePlaybook}, 1, nil
}

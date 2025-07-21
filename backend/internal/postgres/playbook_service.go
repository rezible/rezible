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

// TODO: don't fake
var fakePlaybook = &ent.Playbook{
	ID:         uuid.New(),
	Title:      "Example Playbook",
	ProviderID: "example",
	Content:    nil,
	Edges:      ent.PlaybookEdges{},
}

func (s *PlaybookService) ListPlaybooks(ctx context.Context, params *rez.ListPlaybooksParams) ([]*ent.Playbook, int, error) {
	return []*ent.Playbook{fakePlaybook}, 1, nil
}

func (s *PlaybookService) GetPlaybook(ctx context.Context, id uuid.UUID) (*ent.Playbook, error) {
	return fakePlaybook, nil
}

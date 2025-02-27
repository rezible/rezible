package postgres

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type SystemComponentsService struct {
	db     *ent.Client
	loader rez.ProviderLoader
}

func NewSystemComponentsService(db *ent.Client, pl rez.ProviderLoader) (*SystemComponentsService, error) {
	s := &SystemComponentsService{
		db:     db,
		loader: pl,
	}

	return s, nil
}

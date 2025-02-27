package postgres

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type SystemAnalysisService struct {
	db     *ent.Client
	loader rez.ProviderLoader
}

func NewSystemAnalysisService(db *ent.Client, pl rez.ProviderLoader) (*SystemAnalysisService, error) {
	s := &SystemAnalysisService{
		db:     db,
		loader: pl,
	}

	return s, nil
}

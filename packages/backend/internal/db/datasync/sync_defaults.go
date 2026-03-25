package datasync

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
)

func createDefaultIncidentSeverities(db *ent.Client) []*ent.IncidentSeverityCreate {
	return []*ent.IncidentSeverityCreate{
		db.IncidentSeverity.Create().SetName("P1").SetRank(1),
	}
}

func createDefaultIncidentTypes(db *ent.Client) []*ent.IncidentTypeCreate {
	return []*ent.IncidentTypeCreate{
		db.IncidentType.Create().SetName("Regular"),
	}
}

func syncRequiredDefaultData(ctx context.Context, db *ent.Client) error {
	sevs, sevsErr := db.IncidentSeverity.Query().All(ctx)
	if sevsErr != nil {
		return fmt.Errorf("query severities: %w", sevsErr)
	}

	if len(sevs) == 0 {
		log.Debug().Msg("creating default incident severities")

		builders := createDefaultIncidentSeverities(db)
		if createErr := db.IncidentSeverity.CreateBulk(builders...).Exec(ctx); createErr != nil {
			return fmt.Errorf("incident severities: %w", createErr)
		}
	}

	types, typesErr := db.IncidentType.Query().All(ctx)
	if typesErr != nil {
		return fmt.Errorf("query incident types: %w", typesErr)
	}

	if len(types) == 0 {
		log.Debug().Msg("creating default incident types")

		builders := createDefaultIncidentTypes(db)
		if createErr := db.IncidentType.CreateBulk(builders...).Exec(ctx); createErr != nil {
			return fmt.Errorf("incident types: %w", createErr)
		}
	}

	return nil
}

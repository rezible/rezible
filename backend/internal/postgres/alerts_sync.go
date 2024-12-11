package postgres

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
	"time"
)

type alertsDataSyncer struct {
	db       *ent.Client
	users    rez.UserService
	provider rez.AlertsDataProvider

	mutations []ent.Mutation
}

func newAlertsDataSyncer(db *ent.Client, users rez.UserService, provider rez.AlertsDataProvider) *alertsDataSyncer {
	return &alertsDataSyncer{db: db, users: users, provider: provider}
}

func (as *alertsDataSyncer) syncProviderData(ctx context.Context) error {
	start := time.Now().Add(-time.Hour * 24)
	end := time.Now()

	log.Debug().Msg("syncing alerts data")
	for instance, pullErr := range as.provider.PullAlertInstancesBetweenDates(ctx, start, end) {
		if pullErr != nil {
			return fmt.Errorf("pull: %w", pullErr)
		}

		fmt.Printf("\tinstance:%+v\n", instance)
	}

	return nil
}

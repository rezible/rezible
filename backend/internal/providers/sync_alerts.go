package providers

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
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

func (as *alertsDataSyncer) SyncProviderData(ctx context.Context) error {

	log.Debug().Msg("syncing alerts data")
	/*
		start := time.Now().Add(-time.Hour * 24)
		for instance, pullErr := range as.provider.PullAlertInstancesBetweenDates(ctx, start, end) {
			if pullErr != nil {
				return fmt.Errorf("pull: %w", pullErr)
			}

			fmt.Printf("\tinstance:%+v\n", instance)
		}
		end := time.Now()
	*/

	return nil
}

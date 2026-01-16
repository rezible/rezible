package google

import (
	"context"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/rezible/rezible/ent"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/meet/v2"
	"google.golang.org/api/option"
)

const integrationName = "google"

type IntegrationConfig struct {
	ServiceAccount         string
	ServiceCredentialsJson []byte
}

func NewClient(ctx context.Context, intg *ent.Integration) error {
	var cfg IntegrationConfig
	if cfgErr := mapstructure.Decode(intg, &cfg); cfgErr != nil {
		return fmt.Errorf("failed to decode integration config: %w", cfgErr)
	}

	credsOpt := option.WithAuthCredentialsJSON(option.ServiceAccount, cfg.ServiceCredentialsJson)
	_, calErr := calendar.NewService(ctx, credsOpt)
	if calErr != nil {
		return fmt.Errorf("failed to create calendar service: %w", calErr)
	}
	_, meetErr := meet.NewService(ctx, credsOpt)
	if meetErr != nil {
		return fmt.Errorf("failed to create meet service: %w", meetErr)
	}

	return nil
}

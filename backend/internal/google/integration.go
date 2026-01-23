package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/meet/v2"
	"google.golang.org/api/option"
)

const integrationName = "google"

type integration struct{}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{}
	return intg, nil
}

func (d *integration) Name() string {
	return integrationName
}

func (d *integration) Enabled() bool {
	// TODO: check config
	return true
}

func (d *integration) EventListeners() map[string]rez.EventListener {
	return nil
}

func (d *integration) WebhookHandlers() map[string]http.Handler {
	return nil
}

func (d *integration) SupportedDataKinds() []string {
	return []string{"video_conferencing"}
}

func (d *integration) OAuthConfigRequired() bool {
	return false
}

func (d *integration) ValidateConfig(cfg json.RawMessage) (bool, error) {

	return true, nil
}

type IntegrationConfig struct {
	ServiceAccount         string
	ServiceCredentialsJson []byte
}

func NewClient(ctx context.Context, intg *ent.Integration) error {
	var cfg IntegrationConfig
	if cfgErr := json.Unmarshal(intg.Config, &cfg); cfgErr != nil {
		return fmt.Errorf("failed to decode *integration config: %w", cfgErr)
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

package dataproviders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rs/zerolog/log"
)

// TODO: move this

type (
	providerTenantConfig struct {
		OrgName       string                      `json:"organization_name"`
		ConfigEntries []providerTenantConfigEntry `json:"configs"`
	}

	providerTenantConfigEntry struct {
		Type       providerconfig.ProviderType `json:"type"`
		ProviderID string                      `json:"provider_id"`
		Disabled   bool                        `json:"disabled"`
		Config     json.RawMessage             `json:"config"`
	}
)

func loadTenantProviderConfig(ctx context.Context, client *ent.Client, cfg *providerTenantConfig) {
	org, orgErr := client.Organization.Query().Where(organization.Name(cfg.OrgName)).Only(ctx)
	if orgErr != nil {
		if !ent.IsNotFound(orgErr) {
			log.Fatal().Err(orgErr).Msg("failed to load organization")
		}
		if ent.IsNotFound(orgErr) {
			tnt := client.Tenant.Create().SaveX(ctx)
			org = client.Organization.Create().
				SetName(cfg.OrgName).
				SetProviderID(cfg.OrgName).
				SetTenantID(tnt.ID).
				SaveX(ctx)
		}
	}
	ctx = access.TenantSystemContext(ctx, org.TenantID)
	loadConfigTxFn := func(tx *ent.Tx) error {
		for _, c := range cfg.ConfigEntries {
			log.Info().
				Str("name", c.ProviderID).
				Str("type", string(c.Type)).
				Msg("loading provider")

			upsert := tx.ProviderConfig.Create().
				SetProviderID(c.ProviderID).
				SetProviderType(c.Type).
				SetConfig(c.Config).
				SetEnabled(!c.Disabled).
				SetUpdatedAt(time.Now()).
				OnConflictColumns(providerconfig.FieldProviderID, providerconfig.FieldProviderType).
				UpdateConfig().
				UpdateUpdatedAt()

			if upsertErr := upsert.Exec(ctx); upsertErr != nil {
				return fmt.Errorf("upserting (%s %s): %w", string(c.Type), c.ProviderID, upsertErr)
			}
		}
		return nil
	}

	if txErr := ent.WithTx(ctx, client, loadConfigTxFn); txErr != nil {
		log.Fatal().Err(txErr).Msg("failed to load provider config")
	}
}

func LoadDevConfig(ctx context.Context, client *ent.Client) {
	// TODO: allow specifying file name
	fileName := ".dev_provider_configs.json"
	f, openErr := os.Open(fileName)
	if openErr != nil {
		log.Fatal().Err(openErr).Str("fileName", fileName).Msg("failed to open")
	}
	defer f.Close()
	fileContents, readErr := io.ReadAll(f)
	if readErr != nil {
		log.Fatal().Err(readErr).Str("fileName", fileName).Msg("failed to read")
	}

	var cfg providerTenantConfig
	if cfgErr := json.Unmarshal(fileContents, &cfg); cfgErr != nil {
		log.Fatal().Err(cfgErr).Msg("failed to unmarshal")
	}

	loadTenantProviderConfig(ctx, client, &cfg)
}

// TODO: use fake oncall provider
func grafanaOncallProviderConfig() providerTenantConfigEntry {
	apiEndpoint := rez.Config.GetString("GRAFANA_ONCALL_API_ENDPOINT")
	apiToken := rez.Config.GetString("GRAFANA_ONCALL_API_TOKEN")
	grafanaOncallRawConfig := fmt.Sprintf(`{"api_endpoint":"%s","api_token":"%s"}`, apiEndpoint, apiToken)
	return providerTenantConfigEntry{
		Type:       providerconfig.ProviderTypeOncall,
		ProviderID: "grafana",
		Config:     []byte(grafanaOncallRawConfig),
	}
}

func LoadFakeConfig(ctx context.Context, client *ent.Client) {
	fakeProviderConfigEntry := func(t providerconfig.ProviderType) providerTenantConfigEntry {
		return providerTenantConfigEntry{Type: t, ProviderID: "fake", Config: []byte("{}")}
	}

	cfg := &providerTenantConfig{
		OrgName: "Rezible Test",
		ConfigEntries: []providerTenantConfigEntry{
			grafanaOncallProviderConfig(),
			fakeProviderConfigEntry(providerconfig.ProviderTypeIncidents),
			fakeProviderConfigEntry(providerconfig.ProviderTypeAlerts),
			fakeProviderConfigEntry(providerconfig.ProviderTypeTickets),
			fakeProviderConfigEntry(providerconfig.ProviderTypePlaybooks),
			fakeProviderConfigEntry(providerconfig.ProviderTypeSystemComponents),
		},
	}
	loadTenantProviderConfig(ctx, client, cfg)
}

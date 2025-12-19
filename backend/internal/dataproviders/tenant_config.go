package dataproviders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/organization"
)

// TODO: move this

type (
	providerTenantConfig struct {
		OrgName       string                      `json:"organization_name"`
		ConfigEntries []providerTenantConfigEntry `json:"configs"`
	}

	providerTenantConfigEntry struct {
		Type     integration.IntegrationType `json:"type"`
		Name     string                      `json:"provider_id"`
		Disabled bool                        `json:"disabled"`
		Config   json.RawMessage             `json:"config"`
	}
)

func loadTenantProviderConfig(ctx context.Context, client *ent.Client, cfg *providerTenantConfig) error {
	org, orgErr := client.Organization.Query().Where(organization.Name(cfg.OrgName)).Only(ctx)
	if orgErr != nil {
		if !ent.IsNotFound(orgErr) {
			return fmt.Errorf("query organization: %w", orgErr)
		}
		tnt, tntErr := client.Tenant.Create().Save(ctx)
		if tntErr != nil {
			return fmt.Errorf("create tenant: %w", tntErr)
		}
		org, orgErr = client.Organization.Create().
			SetName(cfg.OrgName).
			SetExternalID(cfg.OrgName).
			SetTenantID(tnt.ID).
			Save(ctx)
		if orgErr != nil {
			return fmt.Errorf("create organization: %w", orgErr)
		}
	}
	ctx = access.TenantSystemContext(ctx, org.TenantID)
	loadConfigTxFn := func(tx *ent.Tx) error {
		for _, c := range cfg.ConfigEntries {
			log.Info().
				Str("name", c.Name).
				Str("type", string(c.Type)).
				Msg("loading provider")

			upsert := tx.Integration.Create().
				SetName(c.Name).
				SetIntegrationType(c.Type).
				SetConfig(c.Config).
				SetEnabled(!c.Disabled).
				SetUpdatedAt(time.Now()).
				OnConflictColumns(integration.FieldName, integration.FieldIntegrationType).
				UpdateConfig().
				UpdateUpdatedAt()

			if upsertErr := upsert.Exec(ctx); upsertErr != nil {
				return fmt.Errorf("upserting (%s %s): %w", string(c.Type), c.Name, upsertErr)
			}
		}
		return nil
	}

	if txErr := ent.WithTx(ctx, client, loadConfigTxFn); txErr != nil {
		return fmt.Errorf("tx: %w", txErr)
	}
	return nil
}

func LoadTenantConfig(ctx context.Context, client *ent.Client, fileName string) error {
	f, openErr := os.Open(fileName)
	if openErr != nil {
		return fmt.Errorf("open file: %w", openErr)
	}
	defer f.Close()
	fileContents, readErr := io.ReadAll(f)
	if readErr != nil {
		return fmt.Errorf("read file: %w", readErr)
	}

	var cfg providerTenantConfig
	if cfgErr := json.Unmarshal(fileContents, &cfg); cfgErr != nil {
		return fmt.Errorf("unmarshal file: %w", cfgErr)
	}

	if loadErr := loadTenantProviderConfig(ctx, client, &cfg); loadErr != nil {
		return fmt.Errorf("loading config: %w", loadErr)
	}
	return nil
}

// TODO: use fake oncall provider
func grafanaOncallProviderConfig() providerTenantConfigEntry {
	apiEndpoint := rez.Config.GetString("GRAFANA_ONCALL_API_ENDPOINT")
	apiToken := rez.Config.GetString("GRAFANA_ONCALL_API_TOKEN")
	grafanaOncallRawConfig := fmt.Sprintf(`{"api_endpoint":"%s","api_token":"%s"}`, apiEndpoint, apiToken)
	return providerTenantConfigEntry{
		Type:   integration.IntegrationTypeOncall,
		Name:   "grafana",
		Config: []byte(grafanaOncallRawConfig),
	}
}

func LoadFakeConfig(ctx context.Context, client *ent.Client) error {
	fakeProviderConfigEntry := func(t integration.IntegrationType) providerTenantConfigEntry {
		return providerTenantConfigEntry{Type: t, Name: "fake", Config: []byte("{}")}
	}

	cfg := &providerTenantConfig{
		OrgName: "Test Organization",
		ConfigEntries: []providerTenantConfigEntry{
			grafanaOncallProviderConfig(),
			fakeProviderConfigEntry(integration.IntegrationTypeIncidents),
			fakeProviderConfigEntry(integration.IntegrationTypeAlerts),
			fakeProviderConfigEntry(integration.IntegrationTypeTickets),
			fakeProviderConfigEntry(integration.IntegrationTypePlaybooks),
			fakeProviderConfigEntry(integration.IntegrationTypeSystemComponents),
		},
	}
	if loadErr := loadTenantProviderConfig(ctx, client, cfg); loadErr != nil {
		return fmt.Errorf("loading fake config: %w", loadErr)
	}
	return nil
}

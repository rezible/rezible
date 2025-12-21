package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/organization"
)

type (
	orgConfig struct {
		OrgName      string           `json:"organization_name"`
		Integrations []orgIntegration `json:"configs"`
	}

	orgIntegration struct {
		Name     string          `json:"provider_id"`
		Disabled bool            `json:"disabled"`
		Config   json.RawMessage `json:"config"`
	}
)

func (cfg *orgConfig) load(ctx context.Context, client *ent.Client) error {
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
		for _, intg := range cfg.Integrations {
			log.Info().Str("name", intg.Name).Msg("loading integration")

			upsert := tx.Integration.Create().
				SetName(intg.Name).
				SetConfig(intg.Config).
				SetEnabled(!intg.Disabled).
				SetUpdatedAt(time.Now()).
				OnConflictColumns(integration.FieldName).
				UpdateConfig().
				UpdateUpdatedAt()

			if upsertErr := upsert.Exec(ctx); upsertErr != nil {
				return fmt.Errorf("upserting (%s): %w", intg.Name, upsertErr)
			}
		}
		return nil
	}

	if txErr := ent.WithTx(ctx, client, loadConfigTxFn); txErr != nil {
		return fmt.Errorf("tx: %w", txErr)
	}
	return nil
}

func LoadOrganization(ctx context.Context, client *ent.Client, fileName string) error {
	f, openErr := os.Open(fileName)
	if openErr != nil {
		return fmt.Errorf("open file: %w", openErr)
	}
	defer f.Close()
	fileContents, readErr := io.ReadAll(f)
	if readErr != nil {
		return fmt.Errorf("read file: %w", readErr)
	}

	var cfg *orgConfig
	if cfgErr := json.Unmarshal(fileContents, cfg); cfgErr != nil || cfg == nil {
		return fmt.Errorf("unmarshal file: %w", cfgErr)
	}

	if loadErr := cfg.load(ctx, client); loadErr != nil {
		return fmt.Errorf("loading config: %w", loadErr)
	}
	return nil
}

func LoadDevOrganization(ctx context.Context, client *ent.Client) error {
	cfg := &orgConfig{
		OrgName:      "Test Organization",
		Integrations: []orgIntegration{{Name: "fake", Config: []byte("{}")}},
	}
	if loadErr := cfg.load(ctx, client); loadErr != nil {
		return fmt.Errorf("loading fake config: %w", loadErr)
	}
	return nil
}

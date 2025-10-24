package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/jobs"
	"github.com/rezible/rezible/openapi"
)

func makeCommand(name string, desc string, cmdFn func(ctx context.Context, opts *Options) error) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: desc,
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, o *Options) {
			log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
			ctx := access.SystemContext(cmd.Context())
			if cmdErr := cmdFn(ctx, o); cmdErr != nil {
				log.Fatal().Err(cmdErr).Str("cmd", name).Msg("command failed")
			}
		}),
	}
}

func printSpecCmd(ctx context.Context, opts *Options) error {
	spec, yamlErr := yaml.Marshal(openapi.MakeApi(&api.Handler{}, "").OpenAPI())
	if yamlErr != nil {
		return yamlErr
	}
	fmt.Println(string(spec))
	return nil
}

func migrateCmd(ctx context.Context, opts *Options) error {
	if dbErr := postgres.RunMigrations(ctx, opts.DatabaseUrl); dbErr != nil {
		return fmt.Errorf("failed to run postgres migrations: %w", dbErr)
	}

	return nil
}

func withDatabaseClient(ctx context.Context, opts *Options, fn func(client *ent.Client) error) error {
	dbc, dbErr := postgres.Open(ctx, opts.DatabaseUrl)
	if dbErr != nil {
		return fmt.Errorf("failed to open database: %w", dbErr)
	}
	defer func() {
		if closeErr := dbc.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close database connection")
		}
	}()

	return fn(dbc.Client())
}

func syncCmd(ctx context.Context, opts *Options) error {
	return withDatabaseClient(ctx, opts, func(client *ent.Client) error {
		args := jobs.SyncProviderData{Hard: true}
		cfgs, cfgsErr := db.NewProviderConfigService(client)
		if cfgsErr != nil {
			return cfgsErr
		}
		syncSvc := datasync.NewProviderSyncService(client, providers.NewProviderLoader(cfgs))
		return syncSvc.SyncProviderData(ctx, args)
	})
}

func seedCmd(ctx context.Context, opts *Options) error {
	return withDatabaseClient(ctx, opts, func(client *ent.Client) error {
		return seedDatabase(ctx, client)
	})
}

func seedDatabase(ctx context.Context, client *ent.Client) error {
	return nil
}

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

func loadDevConfigCmd(ctx context.Context, opts *Options) error {
	return withDatabaseClient(ctx, opts, func(client *ent.Client) error {
		// TODO: allow specifying file name
		f, openErr := os.Open(".dev_provider_configs.json")
		if openErr != nil {
			return fmt.Errorf("opening file: %w", openErr)
		}
		defer f.Close()
		fileContents, readErr := io.ReadAll(f)
		if readErr != nil {
			return fmt.Errorf("reading file: %w", readErr)
		}

		var cfg providerTenantConfig
		if cfgErr := json.Unmarshal(fileContents, &cfg); cfgErr != nil {
			return fmt.Errorf("unmarshalling file: %w", cfgErr)
		}

		return loadTenantProviderConfig(ctx, client, &cfg)
	})
}

func loadFakeConfigCmd(ctx context.Context, opts *Options) error {
	fakeProviderConfigEntry := func(t providerconfig.ProviderType) providerTenantConfigEntry {
		return providerTenantConfigEntry{Type: t, ProviderID: "fake", Config: []byte("{}")}
	}
	return withDatabaseClient(ctx, opts, func(client *ent.Client) error {
		// TODO: use fake oncall provider
		grafanaOncallRawConfig := fmt.Sprintf(`{"api_endpoint":"%s","api_token":"%s"}`,
			os.Getenv("GRAFANA_ONCALL_API_ENDPOINT"),
			os.Getenv("GRAFANA_ONCALL_API_TOKEN"))
		cfg := &providerTenantConfig{
			OrgName: "Rezible Test",
			ConfigEntries: []providerTenantConfigEntry{
				{
					Type:       providerconfig.ProviderTypeOncall,
					ProviderID: "grafana",
					Config:     []byte(grafanaOncallRawConfig),
				},
				fakeProviderConfigEntry(providerconfig.ProviderTypeIncidents),
				fakeProviderConfigEntry(providerconfig.ProviderTypeAlerts),
				fakeProviderConfigEntry(providerconfig.ProviderTypeTickets),
				fakeProviderConfigEntry(providerconfig.ProviderTypePlaybooks),
				fakeProviderConfigEntry(providerconfig.ProviderTypeSystemComponents),
			},
		}
		return loadTenantProviderConfig(ctx, client, cfg)
	})
}

func loadTenantProviderConfig(ctx context.Context, client *ent.Client, cfg *providerTenantConfig) error {
	org, orgErr := client.Organization.Query().Where(organization.Name(cfg.OrgName)).Only(ctx)
	if orgErr != nil {
		if !ent.IsNotFound(orgErr) {
			return fmt.Errorf("querying org %q: %w", cfg.OrgName, orgErr)
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
	return ent.WithTx(ctx, client, func(tx *ent.Tx) error {
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
	})
}

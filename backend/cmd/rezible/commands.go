package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var printSpecCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Print the OpenAPI spec",
	Run: func(cmd *cobra.Command, args []string) {
		spec, yamlErr := yaml.Marshal(openapi.MakeApi(&api.Handler{}, "").OpenAPI())
		if yamlErr != nil {
			log.Fatal().Err(yamlErr).Msg("failed to marshal OpenAPI spec")
		}
		fmt.Println(string(spec))
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := access.SystemContext(cmd.Context())
		if err := postgres.RunMigrations(ctx, viper.GetString("db_url")); err != nil {
			log.Fatal().Err(err).Msg("failed to run database migrations")
		}
	},
}

func withDatabaseClient(fn func(ctx context.Context, client *ent.Client)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		dbUrl, flagErr := cmd.Flags().GetString("db_url")
		if flagErr != nil {
			log.Fatal().Err(flagErr).Msg("failed to get database url")
		}

		ctx := cmd.Context()
		dbc, dbErr := postgres.Open(ctx, dbUrl)
		if dbErr != nil {
			log.Fatal().Err(dbErr).Msg("failed to open database")
		}
		defer func() {
			if closeErr := dbc.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("failed to close database connection")
			}
		}()

		fn(ctx, dbc.Client())
	}
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run sync",
	Run: withDatabaseClient(func(ctx context.Context, client *ent.Client) {
		args := jobs.SyncProviderData{Hard: true}
		cfgs, cfgsErr := db.NewProviderConfigService(client)
		if cfgsErr != nil {
			log.Fatal().Err(cfgsErr).Msg("failed to load provider configs")
		}
		syncSvc := datasync.NewProviderSyncService(client, providers.NewProviderLoader(cfgs))
		if syncErr := syncSvc.SyncProviderData(ctx, args); syncErr != nil {
			log.Fatal().Err(syncErr).Msg("failed to sync provider data")
		}
	}),
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database",
	Run: withDatabaseClient(func(ctx context.Context, client *ent.Client) {

	}),
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

var loadDevConfigCmd = &cobra.Command{
	Use:   "load-dev-config",
	Short: "Load dev config",
	Run:   withDatabaseClient(loadDevConfig),
}

func loadDevConfig(ctx context.Context, client *ent.Client) {
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

var loadFakeConfigCmd = &cobra.Command{
	Use:   "load-fake-config",
	Short: "Load fake config",
	Run:   withDatabaseClient(loadFakeConfig),
}

func loadFakeConfig(ctx context.Context, client *ent.Client) {
	fakeProviderConfigEntry := func(t providerconfig.ProviderType) providerTenantConfigEntry {
		return providerTenantConfigEntry{Type: t, ProviderID: "fake", Config: []byte("{}")}
	}
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
	loadTenantProviderConfig(ctx, client, cfg)
}

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

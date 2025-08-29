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
	"github.com/rezible/rezible/ent/tenant"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/datasync"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/internal/river"
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

func withDatabase(ctx context.Context, opts *Options, fn func(db *postgres.Database) error) error {
	db, dbErr := postgres.Open(ctx, opts.DatabaseUrl)
	if dbErr != nil {
		return fmt.Errorf("failed to open database: %w", dbErr)
	}
	defer db.Close()

	return fn(db)
}

func migrateCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		if dbErr := db.RunMigrations(ctx); dbErr != nil {
			return fmt.Errorf("failed to run ent migrations: %w", dbErr)
		}

		if riverErr := river.RunMigrations(ctx, db.Pool); riverErr != nil {
			return fmt.Errorf("failed to run river migrations: %w", riverErr)
		}

		return nil
	})
}

func syncCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		args := jobs.SyncProviderData{Hard: true}
		dbc := db.Client()
		svc := datasync.NewProviderSyncService(dbc, providers.NewProviderLoader(dbc.ProviderConfig))
		return svc.SyncProviderData(ctx, args)
	})
}

func seedCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		return seedDatabase(ctx, db)
	})
}

func seedDatabase(ctx context.Context, db *postgres.Database) error {
	return nil
}

type (
	providerTenantConfig struct {
		TenantName    string                      `json:"tenant_name"`
		ConfigEntries []providerTenantConfigEntry `json:"configs"`
	}

	providerTenantConfigEntry struct {
		Type         providerconfig.ProviderType `json:"type"`
		ProviderName string                      `json:"provider_name"`
		Disabled     bool                        `json:"disabled"`
		Config       json.RawMessage             `json:"config"`
	}
)

func loadDevConfigCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
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

		return loadTenantProviderConfig(ctx, db.Client(), &cfg)
	})
}

func loadFakeConfigCmd(ctx context.Context, opts *Options) error {
	fakeProviderConfigEntry := func(t providerconfig.ProviderType) providerTenantConfigEntry {
		return providerTenantConfigEntry{Type: t, ProviderName: "fake", Config: []byte("{}")}
	}
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		// TODO: use fake oncall provider
		grafanaOncallRawConfig := fmt.Sprintf(`{"api_endpoint":"%s","api_token":"%s"}`,
			os.Getenv("GRAFANA_ONCALL_API_ENDPOINT"),
			os.Getenv("GRAFANA_ONCALL_API_TOKEN"))
		cfg := &providerTenantConfig{
			TenantName: "Rezible Test",
			ConfigEntries: []providerTenantConfigEntry{
				{
					Type:         providerconfig.ProviderTypeOncall,
					ProviderName: "grafana",
					Config:       []byte(grafanaOncallRawConfig),
				},
				fakeProviderConfigEntry(providerconfig.ProviderTypeIncidents),
				fakeProviderConfigEntry(providerconfig.ProviderTypeAlerts),
				fakeProviderConfigEntry(providerconfig.ProviderTypeTickets),
				fakeProviderConfigEntry(providerconfig.ProviderTypePlaybooks),
				fakeProviderConfigEntry(providerconfig.ProviderTypeSystemComponents),
			},
		}
		return loadTenantProviderConfig(ctx, db.Client(), cfg)
	})
}

func loadTenantProviderConfig(ctx context.Context, client *ent.Client, cfg *providerTenantConfig) error {
	tenantName := cfg.TenantName
	tnt, tenantErr := client.Tenant.Query().Where(tenant.Name(tenantName)).Only(ctx)
	if ent.IsNotFound(tenantErr) {
		create := client.Tenant.Create().
			SetName(tenantName).
			SetProviderID(tenantName)

		tnt, tenantErr = create.Save(ctx)
	}
	if tenantErr != nil {
		return fmt.Errorf("querying tenant %q: %w", tenantName, tenantErr)
	}
	ctx = access.TenantSystemContext(ctx, tnt.ID)

	return ent.WithTx(ctx, client, func(tx *ent.Tx) error {
		for _, c := range cfg.ConfigEntries {
			log.Info().
				Str("name", c.ProviderName).
				Str("type", string(c.Type)).
				Msg("loading provider")

			upsert := tx.ProviderConfig.Create().
				SetProviderName(c.ProviderName).
				SetProviderType(c.Type).
				SetProviderConfig(c.Config).
				SetEnabled(!c.Disabled).
				SetUpdatedAt(time.Now()).
				OnConflictColumns(providerconfig.FieldProviderName, providerconfig.FieldProviderType).
				UpdateProviderConfig().
				UpdateUpdatedAt()

			if upsertErr := upsert.Exec(ctx); upsertErr != nil {
				return fmt.Errorf("upserting (%s %s): %w", string(c.Type), c.ProviderName, upsertErr)
			}
		}
		return nil
	})
}

package postgres

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"strings"
	"text/template"

	"github.com/jackc/pgx/v5"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/migrations"
)

type BootstrapConfig struct {
	DexPassword          string `koanf:"dex_password"`
	RezAdminPassword     string `koanf:"rez_admin_password"`
	RezAppPassword       string `koanf:"rez_app_password"`
	RezDocumentsPassword string `koanf:"rez_documents_password"`
}

var bootstrapQueryTemplate = template.Must(template.New("").Parse(migrations.BootstrapQueryTemplate))

func RunBootstrap(ctx context.Context, connString string) error {
	var cfg BootstrapConfig
	if cfgErr := rez.Config.Unmarshal("bootstrap", &cfg); cfgErr != nil {
		return fmt.Errorf("config: %w", cfgErr)
	}

	bootstrapQueryBuff := &bytes.Buffer{}
	if tplErr := bootstrapQueryTemplate.Execute(bootstrapQueryBuff, &cfg); tplErr != nil {
		return fmt.Errorf("failed to execute bootstrap query template: %w", tplErr)
	}

	connCfg, connCfgErr := pgx.ParseConfig(connString)
	if connCfgErr != nil {
		return fmt.Errorf("failed to parse connection string: %w", connCfgErr)
	}

	parts := strings.SplitN(bootstrapQueryBuff.String(), `\connect rezible`, 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected \\connect rezible in setup script")
	}

	// --- Part 1: run against postgres (admin) DB ---
	adminCfg := connCfg.Copy()
	postgresDbStatements := parts[0]
	if execErr := execStatements(ctx, adminCfg, postgresDbStatements); execErr != nil {
		return fmt.Errorf("admin setup failed: %w", execErr)
	}

	// --- Part 2: run against rezible DB ---
	rezCfg := connCfg.Copy()
	rezCfg.Database = "rezible"
	rezibleDbStatements := parts[1]
	if err := execStatements(ctx, rezCfg, rezibleDbStatements); err != nil {
		return fmt.Errorf("rezible setup failed: %w", err)
	}
	return nil
}

func execStatements(ctx context.Context, connCfg *pgx.ConnConfig, sql string) error {
	conn, err := pgx.ConnectConfig(ctx, connCfg)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer conn.Close(ctx)
	for stmt := range iterStatements(sql) {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed executing statement:\n%s\n\nerror: %w", stmt, err)
		}
	}
	return nil
}

func iterStatements(sql string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, stmt := range strings.Split(sql, ";") {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if !yield(stmt) {
				return
			}
		}
	}
}

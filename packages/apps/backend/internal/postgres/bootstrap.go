package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"

	"github.com/jackc/pgx/v5"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rs/zerolog/log"
)

type bootstrapParams struct {
	RezibleDatabaseName string

	AppSchemaName   string
	RiverSchemaName string

	AppRoleName     string
	AppRolePassword string

	DocumentsRoleName     string
	DocumentsRolePassword string
}

func (p bootstrapParams) validate() error {
	ensureNotEmpty := func(name string, s string) error {
		if len(s) == 0 {
			return fmt.Errorf("%s is empty", name)
		}
		return nil
	}
	return errors.Join(
		ensureNotEmpty("RezibleDatabaseName", p.RezibleDatabaseName),
		ensureNotEmpty("AppSchemaName", p.AppSchemaName),
		ensureNotEmpty("RiverSchemaName", p.RiverSchemaName),
		ensureNotEmpty("AppRoleName", p.AppRoleName),
		ensureNotEmpty("AppRolePassword", p.AppRolePassword),
		ensureNotEmpty("DocumentsRoleName", p.DocumentsRoleName),
		ensureNotEmpty("DocumentsRolePassword", p.DocumentsRolePassword),
	)
}

const bootstrapTemplate = `
CREATE USER {{ .AppRoleName }} WITH LOGIN PASSWORD '{{ .AppRolePassword }}';
CREATE USER {{ .DocumentsRoleName }} WITH LOGIN PASSWORD '{{ .DocumentsRolePassword }}';

REVOKE ALL ON DATABASE {{ .RezibleDatabaseName }} FROM PUBLIC;

GRANT CONNECT ON DATABASE {{ .RezibleDatabaseName }} TO {{ .AppRoleName }};
GRANT CONNECT ON DATABASE {{ .RezibleDatabaseName }} TO {{ .DocumentsRoleName }};

CREATE SCHEMA IF NOT EXISTS {{ .RiverSchemaName }};
CREATE SCHEMA IF NOT EXISTS {{ .AppSchemaName }};

GRANT USAGE ON SCHEMA {{ .AppSchemaName }} TO {{ .AppRoleName }};
GRANT USAGE ON SCHEMA {{ .AppSchemaName }} TO {{ .DocumentsRoleName }};
GRANT USAGE ON SCHEMA {{ .RiverSchemaName }} TO {{ .AppRoleName }};

ALTER DEFAULT PRIVILEGES IN SCHEMA {{ .AppSchemaName }}
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO {{ .AppRoleName }};
ALTER DEFAULT PRIVILEGES IN SCHEMA {{ .AppSchemaName }}
    GRANT USAGE, SELECT ON SEQUENCES TO {{ .AppRoleName }};

ALTER DEFAULT PRIVILEGES IN SCHEMA {{ .RiverSchemaName }}
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO {{ .AppRoleName }};
ALTER DEFAULT PRIVILEGES IN SCHEMA {{ .RiverSchemaName }}
    GRANT USAGE, SELECT ON SEQUENCES TO {{ .AppRoleName }};

ALTER ROLE {{ .AppRoleName }} SET search_path TO {{ .AppSchemaName }}, {{ .RiverSchemaName }};
ALTER ROLE {{ .DocumentsRoleName }} SET search_path TO {{ .AppSchemaName }};`

var bootstrapQueryTemplate = template.Must(template.New("").Parse(bootstrapTemplate))

func renderBootstrapQuery(cfg Config) (string, error) {
	params := bootstrapParams{
		RezibleDatabaseName:   cfg.Database,
		AppSchemaName:         SchemaName,
		RiverSchemaName:       river.SchemaName,
		AppRoleName:           cfg.AppRole.Name,
		AppRolePassword:       cfg.AppRole.Password,
		DocumentsRoleName:     cfg.DocumentsRole.Name,
		DocumentsRolePassword: cfg.DocumentsRole.Password,
	}
	if validErr := params.validate(); validErr != nil {
		return "", fmt.Errorf("bootstrap query params: %w", validErr)
	}
	buff := &bytes.Buffer{}
	if tplErr := bootstrapQueryTemplate.Execute(buff, &params); tplErr != nil {
		return "", fmt.Errorf("execute bootstrap query template: %w", tplErr)
	}
	return buff.String(), nil
}

func BootstrapDatabase(ctx context.Context) error {
	cfg, cfgErr := LoadConfig()
	if cfgErr != nil || cfg.AdminRole.Name == "" {
		return fmt.Errorf("postgres migrations config error: %w", cfgErr)
	}
	cfg.Pool = nil

	conn, connErr := pgx.Connect(ctx, cfg.getDsn(cfg.AdminRole))
	if connErr != nil {
		return fmt.Errorf("pgx connect: %w", connErr)
	}
	defer func(c *pgx.Conn) {
		if c != nil {
			if closeErr := c.Close(ctx); closeErr != nil {
				log.Error().Err(closeErr).Msgf("failed to close connection")
			}
		}
	}(conn)

	query, queryErr := renderBootstrapQuery(cfg)
	if queryErr != nil {
		return fmt.Errorf("rendering bootstrap query: %w", queryErr)
	}

	if _, execErr := conn.Exec(ctx, query); execErr != nil {
		return fmt.Errorf("executing bootstrap query: %w", execErr)
	}

	return nil
}

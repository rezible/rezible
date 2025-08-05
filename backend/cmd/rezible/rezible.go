package main

import (
	"context"
	"github.com/danielgtaylor/huma/v2/humacli"
)

type Options struct {
	Mode                  string `doc:"App Mode" default:"debug"`
	Host                  string `doc:"Hostname to listen on." default:"localhost"`
	Port                  string `doc:"Port to listen on." short:"p" default:"8888"`
	StopTimeoutSeconds    int    `doc:"Timeout in seconds to wait before stopping" default:"30"`
	DocumentServerAddress string `doc:"Document server address" name:"document_server_address" default:"localhost:8889"`
	DatabaseUrl           string `doc:"Database connection url" name:"db_url"`
	AuthSessionSecretKey  string `doc:"Auth session secret key" name:"auth_session_secret_key"`
	PosthogApiKey         string `doc:"Posthog api key" name:"posthog_api_key"`
}

func main() {
	onOptionsParsed := func(hooks humacli.Hooks, opts *Options) {
		s := newRezibleServer(opts)
		hooks.OnStart(s.Start)
		hooks.OnStop(s.Stop)
	}
	cli := humacli.New(onOptionsParsed)

	addCliCommand := func(name string, desc string, cmdFn func(ctx context.Context, opts *Options) error) {
		cli.Root().AddCommand(makeCommand(name, desc, cmdFn))
	}
	addCliCommand("openapi", "Print the OpenAPI spec", printSpecCmd)
	addCliCommand("migrate", "Run database migrations", migrateCmd)
	addCliCommand("seed", "Seed the database", seedCmd)
	addCliCommand("load-configs", "Load provider configs file into database", loadConfigCmd)
	addCliCommand("sync", "Sync the data providers", syncCmd)

	cli.Run()
}

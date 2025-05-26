package main

import (
	"context"

	"github.com/danielgtaylor/huma/v2/humacli"
)

func main() {
	onOptionsParsed := func(hooks humacli.Hooks, opts *Options) {
		rezSrv := newRezServer(opts)
		hooks.OnStart(rezSrv.Start)
		hooks.OnStop(rezSrv.Stop)
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

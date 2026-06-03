package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"
)

type appRunner struct {
	i do.Injector
}

func makeServiceRunner() *appRunner {
	return &appRunner{i: do.New()}
}

func (r *appRunner) setupContext(ctx context.Context) (context.Context, error) {
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	cfg, cfgErr := koanf.LoadConfig(ctx, koanf.Options{LoadEnvironment: true})
	if cfgErr != nil {
		return nil, fmt.Errorf("load config: %w", cfgErr)
	}
	do.ProvideValue(r.i, *cfg)
	declareServices(ctx, r.i)

	return ctx, nil
}

func (r *appRunner) printConfig() error {
	fmt.Printf("%+v\n", do.MustInvoke[rez.Config](r.i))
	return nil
}

func (r *appRunner) runServer(ctx context.Context) error {
	return runService[*http.Server](ctx, r)
}

func (r *appRunner) runSchemaMigration(ctx context.Context, direction string) error {
	ms := do.MustInvoke[*postgres.MigrationService](r.i)
	return ms.Run(ctx, direction)
}

func (r *appRunner) createSchemaMigration(ctx context.Context, name string) error {
	ms := do.MustInvoke[*postgres.MigrationService](r.i)
	return ms.CreateSchemaMigration(ctx, name)
}

func (r *appRunner) updateMigrationChecksumFile() error {
	return postgres.UpdateMigrationsChecksum()
}

func (r *appRunner) printOpenApiSpec(asJson bool) error {
	spec, specErr := oapiv1.GetSpec(asJson)
	if specErr != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", specErr)
	}
	fmt.Printf("%s", spec)
	return nil
}

type startable interface {
	Start(context.Context) error
}

func runService[Entrypoint startable](ctx context.Context, r *appRunner) error {
	if initErr := registerIntegrations(r.i); initErr != nil {
		return fmt.Errorf("failed to initialize services: %w", initErr)
	}
	// invoke entrypoint service to load required service dependencies
	es, srvErr := do.Invoke[Entrypoint](r.i)
	if srvErr != nil {
		return fmt.Errorf("failed to initialize %T: %v", es, srvErr)
	}
	var services []startable
	for _, desc := range r.i.ListInvokedServices() {
		s, invErr := do.InvokeNamed[any](r.i, desc.Service)
		if invErr != nil {
			return fmt.Errorf("failed to invoke: %v", invErr)
		}
		if svc, ok := s.(startable); ok {
			services = append(services, svc)
		}
	}
	return r.start(ctx, services)
}

func (r *appRunner) start(ctx context.Context, services []startable) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx).WithFirstError()
		for _, l := range services {
			slog.Info("Starting " + strings.TrimLeft(fmt.Sprintf("%T", l), "*"))
			p.Go(l.Start)
		}
		errChan <- p.Wait()
	}()
	slog.Info("=== Starting Services ===")
	var servicesErr error
	select {
	case <-ctx.Done():
		servicesErr = ctx.Err()
	case poolErr := <-errChan:
		servicesErr = poolErr
	}
	slog.Info("=== Stopping Services ===")
	if servicesErr != nil && !errors.Is(servicesErr, context.Canceled) {
		return fmt.Errorf("run services: %s", servicesErr.Error())
	}
	return nil
}

func (r *appRunner) shutdown(baseCtx context.Context) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(baseCtx), 5*time.Second)
	defer cancel()
	shutdown := r.i.ShutdownWithContext(ctx)
	var shutdownErr error
	for sd, sErr := range shutdown.Errors {
		if !errors.Is(sErr, context.Canceled) {
			fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
			shutdownErr = errors.Join(shutdownErr, sErr)
		}
	}
	return shutdownErr
}

package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations/projections"
	apiv1 "github.com/rezible/rezible/internal/http/api/v1"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/telemetry"
)

type Server struct {
	cfg ServerConfig
}

type ServerConfig struct {
	StopTimeout time.Duration `cfg:"stop_timeout"`
}

func runServer(ctx context.Context, i do.Injector) error {
	s := &Server{
		cfg: ServerConfig{
			StopTimeout: 5 * time.Second,
		},
	}
	if cfgErr := rez.Config.Unmarshal("server", &s.cfg); cfgErr != nil {
		return fmt.Errorf("failed to get server config: %w", cfgErr)
	}

	if setupErr := s.setup(ctx, i); setupErr != nil {
		return fmt.Errorf("setup: %s", setupErr)
	}

	var serveErr error
	if startErr := s.start(ctx); startErr != nil && !errors.Is(startErr, context.Canceled) {
		serveErr = fmt.Errorf("start: %w", startErr)
	}
	if stopErr := s.stop(ctx); stopErr != nil {
		serveErr = errors.Join(fmt.Errorf("failed to stop server: %w", stopErr), serveErr)
	}

	return serveErr
}

func (s *Server) start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx)
		for _, l := range s.listeners {
			p.Go(l.Start)
		}
		errChan <- p.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case waitErr := <-errChan:
		return fmt.Errorf("server error: %w", waitErr)
	}
}

func (s *Server) stop(ctx context.Context) error {
	timeoutCtx, cancelStopCtx := context.WithTimeout(ctx, s.cfg.StopTimeout)
	defer cancelStopCtx()

	var err error
	for name, l := range s.listeners {
		if listenerErr := l.Stop(timeoutCtx); listenerErr != nil && !errors.Is(listenerErr, context.Canceled) {
			err = errors.Join(err, fmt.Errorf("stopping %s: %w", name, listenerErr))
		}
	}

	if s.db != nil {
		s.db.Close()
	}

	if telErr := telemetry.Shutdown(timeoutCtx); telErr != nil {
		err = errors.Join(err, fmt.Errorf("shutdown telemetry: %w", telErr))
	}

	return err
}

func (s *Server) setup(ctx context.Context, i do.Injector) error {
	do.ProvideValue(i, integrations.NewPackageRegistry())
	do.ProvideValue(i, projections.NewEventProjectionHandlerRegistry())

	s.setupPostgres(ctx, i)

	do.Provide(i, func(inj do.Injector) (rez.MessageService, error) {
		return watermill.NewMessageService()
	})

	db.Package(i)

	intgListeners, intgPkgsErr := ipr.SetupPackages(ctx, svcs)
	if intgPkgsErr != nil {
		return fmt.Errorf("integrations.Setup: %w", intgPkgsErr)
	}
	for name, el := range intgListeners {
		s.listeners[name] = el
	}

	auth, authErr := oidc.NewAuthSessionService(s.services.Organizations, s.services.Users)
	if authErr != nil {
		return fmt.Errorf("oidc.NewAuthSessionService: %w", authErr)
	}

	apiv1Handler := apiv1.NewHandler(s.services)

	srv, srvErr := http.NewServer(auth, apiv1Handler, ipr.GetWebhookHandlers())
	if srvErr != nil {
		return fmt.Errorf("http.NewServer: %w", srvErr)
	}
	s.listeners["http_server"] = srv

	return nil
}

func (s *Server) setupPostgres(ctx context.Context, i do.Injector) {
	do.Provide(i, func(inj do.Injector) (*postgres.DatabaseClient, error) {
		return postgres.NewDatabaseClient(ctx)
	})

	do.Provide(i, func(inj do.Injector) (rez.DatabaseClient, error) {
		return do.Invoke[*postgres.DatabaseClient](inj)
	})

	do.Provide(i, func(inj do.Injector) (*ent.Client, error) {
		dbClient, clientErr := do.Invoke[rez.DatabaseClient](inj)
		if clientErr != nil {
			return nil, clientErr
		}
		return dbClient.Client(), nil
	})

	do.Provide(i, func(inj do.Injector) (*river.JobService, error) {
		return river.NewJobService(do.MustInvoke[*postgres.DatabaseClient](inj).Pool())
	})

	do.Provide(i, func(inj do.Injector) (rez.JobsService, error) {
		return do.Invoke[*river.JobService](inj)
	})
}

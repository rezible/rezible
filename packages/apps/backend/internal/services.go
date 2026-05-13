package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sourcegraph/conc/pool"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/telemetry"
)

type Server struct {
	cfg ServerConfig

	services  *rez.Services
	listeners map[string]rez.EventListener
	db        rez.DatabaseClient
}

type ServerConfig struct {
	StopTimeout time.Duration `cfg:"stop_timeout"`
}

func NewServer(ctx context.Context) (*Server, error) {
	s := &Server{
		listeners: make(map[string]rez.EventListener),
		cfg: ServerConfig{
			StopTimeout: 5 * time.Second,
		},
	}
	if cfgErr := rez.Config.Unmarshal("server", &s.cfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to get server config: %w", cfgErr)
	}
	return s, nil
}

func (s *Server) RunServe(ctx context.Context) error {
	if setupErr := s.setup(ctx); setupErr != nil {
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

func (s *Server) setup(ctx context.Context) error {
	if telemetryErr := telemetry.Init(ctx); telemetryErr != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", telemetryErr)
	}

	services, servicesErr := s.setupServices(ctx)
	if servicesErr != nil {
		return fmt.Errorf("setup services: %w", servicesErr)
	}
	s.services = services

	if integrationsErr := integrations.Setup(ctx, services); integrationsErr != nil {
		return fmt.Errorf("integrations.Setup: %w", integrationsErr)
	}
	for name, el := range integrations.GetEventListeners() {
		s.listeners[name] = el
	}

	db.RegisterEventProcessors()

	srv, srvErr := http.NewServer(ctx, services)
	if srvErr != nil {
		return fmt.Errorf("http.NewServer: %w", srvErr)
	}
	s.listeners["http_server"] = srv

	return nil
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

func (s *Server) setupServices(ctx context.Context) (*rez.Services, error) {
	svcs := &rez.Services{}
	var svcErr error

	svcs.Database, svcErr = postgres.NewDatabaseClient(ctx)
	if svcErr != nil {
		return nil, fmt.Errorf("postgres.NewDatabaseClient: %w", svcErr)
	}
	s.db = svcs.Database

	svcs.Jobs, svcErr = river.NewJobService(ctx, svcs.Database.(*postgres.DatabaseClient).Pool())
	if svcErr != nil {
		return nil, fmt.Errorf("river.NewJobService: %w", svcErr)
	}
	s.listeners["river_job_service"] = svcs.Jobs

	svcs.Messages, svcErr = watermill.NewMessageService(ctx)
	if svcErr != nil {
		return nil, fmt.Errorf("watermill.NewMessageService: %w", svcErr)
	}
	s.listeners["watermill_message_service"] = svcs.Messages.(*watermill.MessageService)

	svcs.Integrations, svcErr = db.NewIntegrationsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewIntegrationsService: %w", svcErr)
	}

	svcs.ProviderEvents = db.NewProviderEventService(ctx, svcs)

	svcs.Organizations, svcErr = db.NewOrganizationsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewOrganizationsService: %w", svcErr)
	}

	svcs.Users, svcErr = db.NewUserService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewUserService: %w", svcErr)
	}

	svcs.Teams, svcErr = db.NewTeamService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewTeamService: %w", svcErr)
	}

	svcs.Events, svcErr = db.NewEventsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewEventsService: %w", svcErr)
	}

	svcs.EventAnnotations, svcErr = db.NewEventAnnotationsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewEventAnnotationsService: %w", svcErr)
	}

	svcs.Incidents, svcErr = db.NewIncidentService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewIncidentService: %w", svcErr)
	}

	svcs.OncallRosters, svcErr = db.NewOncallRostersService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewOncallRostersService: %w", svcErr)
	}

	svcs.Topology, svcErr = db.NewSystemTopologyService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewTopologyService: %w", svcErr)
	}

	svcs.OncallShifts, svcErr = db.NewOncallShiftsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewOncallShiftsService: %w", svcErr)
	}

	svcs.OncallMetrics, svcErr = db.NewOncallMetricsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewOncallMetricsService: %w", svcErr)
	}

	svcs.Debriefs, svcErr = db.NewDebriefService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewDebriefService: %w", svcErr)
	}

	svcs.Retros, svcErr = db.NewRetrospectiveService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewRetrospectiveService: %w", svcErr)
	}

	svcs.Alerts, svcErr = db.NewAlertService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewAlertService: %w", svcErr)
	}

	svcs.Playbooks, svcErr = db.NewPlaybookService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewPlaybookService: %w", svcErr)
	}

	svcs.Documents, svcErr = db.NewDocumentsService(svcs)
	if svcErr != nil {
		return nil, fmt.Errorf("db.NewDocumentsService: %w", svcErr)
	}

	return svcs, nil
}

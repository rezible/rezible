package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rezible/rezible/integrations/eventprojections"
	"github.com/sourcegraph/conc/pool"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/adk"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/prosemirror"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/telemetry"
)

type Server struct {
	cfg ServerConfig

	listeners map[string]rez.EventListener
	db        rez.DatabaseClient
}

type ServerConfig struct {
	StopTimeout time.Duration `koanf:"stop_timeout"`
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
	setupErr := s.setup(ctx)
	if setupErr != nil {
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

	if integrationsErr := integrations.Setup(ctx, services); integrationsErr != nil {
		return fmt.Errorf("integrations.Setup: %w", integrationsErr)
	}
	for name, el := range integrations.GetEventListeners() {
		s.listeners[name] = el
	}

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
	pgDb, pgErr := postgres.NewDatabaseClient(ctx)
	if pgErr != nil {
		return nil, fmt.Errorf("postgres.NewDatabaseClient: %w", pgErr)
	}
	s.db = pgDb

	msgs, msgsErr := watermill.NewMessageService(ctx)
	if msgsErr != nil {
		return nil, fmt.Errorf("watermill.NewMessageService: %w", msgsErr)
	}
	s.listeners["watermill_message_service"] = msgs

	jobSvc, jobSvcErr := river.NewJobService(ctx, pgDb.Pool())
	if jobSvcErr != nil {
		return nil, fmt.Errorf("river.NewJobService: %w", jobSvcErr)
	}
	s.listeners["river_job_service"] = jobSvc

	dbc := s.db.Client()

	eventprojections.RegisterHandler("knowledge", db.KnowledgeEntityEventProjectionHandler)

	orgs, orgsErr := db.NewOrganizationsService(dbc, jobSvc)
	if orgsErr != nil {
		return nil, fmt.Errorf("db.NewOrganizationsService: %w", orgsErr)
	}

	users, usersErr := db.NewUserService(dbc, orgs)
	if usersErr != nil {
		return nil, fmt.Errorf("db.NewUserService: %w", usersErr)
	}

	teams, teamsErr := db.NewTeamService(dbc)
	if teamsErr != nil {
		return nil, fmt.Errorf("db.NewTeamService: %w", teamsErr)
	}

	intgs, intgsErr := db.NewIntegrationsService(dbc, jobSvc)
	if intgsErr != nil {
		return nil, fmt.Errorf("db.NewIntegrationsService: %w", intgsErr)
	}

	provEvents := db.NewProviderEventService(ctx, dbc, jobSvc, intgs)

	agents, agentsErr := adk.NewAgentService()
	if agentsErr != nil {
		return nil, fmt.Errorf("adk.NewAgentService: %w", agentsErr)
	}

	events, eventsErr := db.NewEventsService(dbc, users)
	if eventsErr != nil {
		return nil, fmt.Errorf("db.NewEventsService: %w", eventsErr)
	}

	annos, annosErr := db.NewEventAnnotationsService(dbc, events)
	if annosErr != nil {
		return nil, fmt.Errorf("db.NewEventAnnotationsService: %w", annosErr)
	}

	_, nodesErr := prosemirror.NewNodeService()
	if nodesErr != nil {
		return nil, fmt.Errorf("prosemirror.NewNodeService: %w", nodesErr)
	}

	incidents, incidentsErr := db.NewIncidentService(dbc, jobSvc, msgs, users)
	if incidentsErr != nil {
		return nil, fmt.Errorf("db.NewIncidentService: %w", incidentsErr)
	}

	rosters, rostersErr := db.NewOncallRostersService(dbc, jobSvc)
	if rostersErr != nil {
		return nil, fmt.Errorf("db.NewOncallRostersService: %w", rostersErr)
	}

	topology, topologyErr := db.NewSystemTopologyService(dbc)
	if topologyErr != nil {
		return nil, fmt.Errorf("db.NewTopologyService: %w", topologyErr)
	}

	shifts, shiftsErr := db.NewOncallShiftsService(dbc, jobSvc, intgs)
	if shiftsErr != nil {
		return nil, fmt.Errorf("db.NewOncallShiftsService: %w", shiftsErr)
	}

	oncallMetrics, oncallMetricsErr := db.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return nil, fmt.Errorf("db.NewOncallMetricsService: %w", oncallMetricsErr)
	}

	debriefs, debriefsErr := db.NewDebriefService(dbc, jobSvc, agents)
	if debriefsErr != nil {
		return nil, fmt.Errorf("db.NewDebriefService: %w", debriefsErr)
	}

	retros, retrosErr := db.NewRetrospectiveService(dbc, msgs, incidents)
	if retrosErr != nil {
		return nil, fmt.Errorf("db.NewRetrospectiveService: %w", retrosErr)
	}

	alerts, alertsErr := db.NewAlertService(dbc)
	if alertsErr != nil {
		return nil, fmt.Errorf("db.NewAlertService: %w", alertsErr)
	}

	playbooks, playbooksErr := db.NewPlaybookService(dbc)
	if playbooksErr != nil {
		return nil, fmt.Errorf("db.NewPlaybookService: %w", playbooksErr)
	}

	docs, docsErr := db.NewDocumentsService(dbc, teams)
	if docsErr != nil {
		return nil, fmt.Errorf("db.NewDocumentsService: %w", docsErr)
	}

	return &rez.Services{
		Database:       pgDb,
		Jobs:           jobSvc,
		ProviderEvents: provEvents,
		Messages:       msgs,
		//Knowledge:        knowledge,
		Topology:         topology,
		Organizations:    orgs,
		Integrations:     intgs,
		Users:            users,
		Teams:            teams,
		Incidents:        incidents,
		Debriefs:         debriefs,
		OncallRosters:    rosters,
		OncallShifts:     shifts,
		OncallMetrics:    oncallMetrics,
		Events:           events,
		EventAnnotations: annos,
		Documents:        docs,
		Retros:           retros,
		Alerts:           alerts,
		Playbooks:        playbooks,
	}, nil
}

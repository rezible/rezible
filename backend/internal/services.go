package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/apiv1"
	"github.com/rezible/rezible/internal/http"
	"github.com/sourcegraph/conc/pool"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/oidc"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/prosemirror"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/rezible/rezible/jobs"
)

type Config struct {
	StopTimeout time.Duration `koanf:"stop_timeout"`
}

type Server struct {
	listeners map[string]rez.EventListener
	database  rez.Database
	cfg       Config
}

func NewServer(ctx context.Context) (*Server, error) {
	s := &Server{
		listeners: make(map[string]rez.EventListener),
		cfg: Config{
			StopTimeout: 5 * time.Second,
		},
	}
	if cfgErr := rez.Config.Unmarshal("server", &s.cfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to get server config: %w", cfgErr)
	}
	return s, nil
}

func (s *Server) RunDataSync(ctx context.Context, args jobs.SyncIntegrationsData) error {
	if setupErr := s.setup(ctx); setupErr != nil {
		return fmt.Errorf("setup: %s", setupErr)
	}
	return datasync.NewSyncerService(s.database.Client()).SyncIntegrationsData(access.SystemContext(ctx), args)
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
	if stopErr := s.stop(); stopErr != nil {
		serveErr = errors.Join(fmt.Errorf("failed to stop server: %w", stopErr), serveErr)
	}

	return serveErr
}

func (s *Server) setup(ctx context.Context) error {
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

	srv, srvErr := http.NewServer(services.Auth, apiv1.NewHandler(services, s.database.Client()))
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

func (s *Server) stop() error {
	timeoutCtx, cancelStopCtx := context.WithTimeout(context.Background(), s.cfg.StopTimeout)
	defer cancelStopCtx()

	var err error
	for name, l := range s.listeners {
		if listenerErr := l.Stop(timeoutCtx); listenerErr != nil && !errors.Is(listenerErr, context.Canceled) {
			err = errors.Join(err, fmt.Errorf("stopping %s: %w", name, listenerErr))
		}
	}

	if s.database != nil {
		s.database.Close()
	}

	return err
}

func (s *Server) makeJobService() (rez.JobsService, error) {
	if pgDb, ok := s.database.(*postgres.DatabaseClient); ok {
		jobSvc, jobSvcErr := river.NewJobService(pgDb.Pool())
		if jobSvcErr != nil {
			return nil, fmt.Errorf("river.NewJobService: %w", jobSvcErr)
		}
		return jobSvc, nil
	}
	return nil, fmt.Errorf("unable to create job service")
}

func (s *Server) setupServices(ctx context.Context) (*rez.Services, error) {
	pgClient, pgErr := postgres.NewDatabasePoolClient(ctx)
	if pgErr != nil {
		return nil, fmt.Errorf("postgres.NewDatabasePoolClient: %w", pgErr)
	}
	s.database = pgClient
	dbc := s.database.Client()

	jobSvc, jobSvcErr := s.makeJobService()
	if jobSvcErr != nil {
		return nil, fmt.Errorf("job service: %w", jobSvcErr)
	}
	s.listeners["job_service"] = jobSvc

	msgs, msgsErr := watermill.NewMessageService()
	if msgsErr != nil {
		return nil, fmt.Errorf("watermill.NewMessageService: %w", msgsErr)
	}
	s.listeners["message_service"] = msgs

	orgs, orgsErr := db.NewOrganizationsService(dbc, jobSvc)
	if orgsErr != nil {
		return nil, fmt.Errorf("postgres.NewOrganizationsService: %w", orgsErr)
	}

	users, usersErr := db.NewUserService(dbc, orgs)
	if usersErr != nil {
		return nil, fmt.Errorf("postgres.NewUserService: %w", usersErr)
	}

	auth, authErr := oidc.NewAuthService(ctx, orgs, users)
	if authErr != nil {
		return nil, fmt.Errorf("dex.NewAuthService: %w", authErr)
	}

	intgs, intgsErr := db.NewIntegrationsService(dbc, jobSvc, auth)
	if intgsErr != nil {
		return nil, fmt.Errorf("db.NewIntegrationsService: %w", intgsErr)
	}

	events, eventsErr := db.NewEventsService(dbc, users)
	if eventsErr != nil {
		return nil, fmt.Errorf("postgres.NewEventsService: %w", eventsErr)
	}

	annos, annosErr := db.NewEventAnnotationsService(dbc, events)
	if annosErr != nil {
		return nil, fmt.Errorf("postgres.NewEventAnnotationsService: %w", annosErr)
	}

	teams, teamsErr := db.NewTeamService(dbc)
	if teamsErr != nil {
		return nil, fmt.Errorf("postgres.NewTeamService: %w", teamsErr)
	}

	_, nodesErr := prosemirror.NewNodeService()
	if nodesErr != nil {
		return nil, fmt.Errorf("prosemirror.NewNodeService: %w", nodesErr)
	}

	incidents, incidentsErr := db.NewIncidentService(dbc, jobSvc, msgs, users)
	if incidentsErr != nil {
		return nil, fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	rosters, rostersErr := db.NewOncallRostersService(dbc, jobSvc)
	if rostersErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallRostersService: %w", rostersErr)
	}

	components, componentsErr := db.NewSystemComponentsService(dbc)
	if componentsErr != nil {
		return nil, fmt.Errorf("postgres.NewSystemComponentsService: %w", componentsErr)
	}

	shifts, shiftsErr := db.NewOncallShiftsService(dbc, jobSvc, intgs)
	if shiftsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallShiftsService: %w", shiftsErr)
	}

	oncallMetrics, oncallMetricsErr := db.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallMetricsService: %w", oncallMetricsErr)
	}

	ai, aiErr := eino.NewAiAgentService(ctx)
	if aiErr != nil {
		return nil, fmt.Errorf("eino.NewAiAgentService: %w", aiErr)
	}

	debriefs, debriefsErr := db.NewDebriefService(dbc, jobSvc, ai)
	if debriefsErr != nil {
		return nil, fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

	retros, retrosErr := db.NewRetrospectiveService(dbc, msgs, incidents)
	if retrosErr != nil {
		return nil, fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	alerts, alertsErr := db.NewAlertService(dbc)
	if alertsErr != nil {
		return nil, fmt.Errorf("postgres.NewAlertService: %w", alertsErr)
	}

	playbooks, playbooksErr := db.NewPlaybookService(dbc)
	if playbooksErr != nil {
		return nil, fmt.Errorf("postgres.NewPlaybookService: %w", playbooksErr)
	}

	docs, docsErr := db.NewDocumentsService(dbc, auth, teams)
	if docsErr != nil {
		return nil, fmt.Errorf("db.NewDocumentsService: %w", docsErr)
	}

	return &rez.Services{
		Jobs:             jobSvc,
		Messages:         msgs,
		Auth:             auth,
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
		Components:       components,
		Alerts:           alerts,
		Playbooks:        playbooks,
	}, nil
}

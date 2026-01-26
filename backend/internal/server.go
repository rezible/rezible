package internal

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/jobs"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/apiv1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/prosemirror"
	"github.com/rezible/rezible/internal/watermill"
)

func RunAutoMigrations(ctx context.Context) error {
	return postgres.RunAutoMigrations(ctx)
}

func OpenDatabase(ctx context.Context) (rez.Database, error) {
	dbc, dbcErr := postgres.NewDatabaseClient(ctx)
	if dbcErr != nil {
		return nil, fmt.Errorf("postgres.NewDatabaseClient: %w", dbcErr)
	}
	return dbc, nil
}

func RunServer(ctx context.Context) error {
	ctx = access.AnonymousContext(ctx)

	s := newServer()

	if setupErr := s.setup(ctx); setupErr != nil {
		return fmt.Errorf("setup: %s", setupErr)
	}

	runErr := s.start(ctx)
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		runErr = nil
	}

	if stopErr := s.stop(); stopErr != nil {
		log.Error().Err(stopErr).Msg("Failed to stop server")
	}

	return runErr
}

type Server struct {
	listeners map[string]rez.EventListener
}

func newServer() *Server {
	return &Server{
		listeners: make(map[string]rez.EventListener),
	}
}

func (s *Server) addEventListener(name string, l rez.EventListener) {
	s.listeners[name] = l
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
	timeout := rez.Config.GetDurationOr("stop_timeout", time.Second*10)
	timeoutCtx, cancelStopCtx := context.WithTimeout(context.Background(), timeout)
	defer cancelStopCtx()

	var err error
	for name, l := range s.listeners {
		if listenerErr := l.Stop(timeoutCtx); listenerErr != nil && !errors.Is(listenerErr, context.Canceled) {
			err = errors.Join(err, fmt.Errorf("stopping %s: %w", name, listenerErr))
		}
	}

	return err
}

func (s *Server) setup(ctx context.Context) error {
	conn, dbcErr := OpenDatabase(ctx)
	if dbcErr != nil {
		return dbcErr
	}
	s.addEventListener("database", db.NewListener(conn))

	client := conn.Client()

	svcs, svcsErr := s.makeServices(ctx, conn, client)
	if svcsErr != nil {
		return fmt.Errorf("services: %w", svcsErr)
	}

	// TODO: this shouldn't need the db client
	apiv1Handler := apiv1.NewHandler(client, svcs)

	srv := http.NewServer(svcs.Auth)
	srv.MountOpenApiV1(apiv1Handler)
	srv.MountMCP(eino.NewMCPHandler(svcs.Auth))

	frontendFS, feFSErr := http.GetEmbeddedFrontendFiles()
	if feFSErr != nil {
		return fmt.Errorf("failed to get embedded frontend files: %w", feFSErr)
	}
	srv.MountStaticFrontend(frontendFS)
	s.addEventListener("http_server", srv)

	if intgsErr := integrations.Setup(ctx, svcs); intgsErr != nil {
		return fmt.Errorf("integrations: %w", intgsErr)
	}

	for _, p := range integrations.GetEnabled() {
		if elIntegration, ok := p.(rez.IntegrationWithEventListeners); ok {
			for name, l := range elIntegration.EventListeners() {
				s.addEventListener(name, l)
			}
		}
		if whIntegration, ok := p.(rez.IntegrationWithWebhookHandlers); ok {
			for whPrefix, h := range whIntegration.WebhookHandlers() {
				srv.AddWebhookHandler(path.Join(p.Name(), whPrefix), h)
			}
		}
	}

	return nil
}

func (s *Server) makeJobService(dbc rez.Database) (rez.JobsService, error) {
	pgDb, ok := dbc.(*postgres.DatabaseClient)
	if !ok {
		return nil, errors.New("non-postgres db client with river job service")
	}
	jobSvc, jobSvcErr := river.NewJobService(pgDb.Pool())
	if jobSvcErr != nil {
		return nil, fmt.Errorf("river.NewJobService: %w", jobSvcErr)
	}
	s.addEventListener("job_service", jobSvc)
	return jobSvc, nil
}

func (s *Server) makeMessageService() (rez.MessageService, error) {
	msgs, msgsErr := watermill.NewMessageService()
	if msgsErr != nil {
		return nil, fmt.Errorf("watermill.NewMessageService: %w", msgsErr)
	}
	s.addEventListener("message_service", msgs)
	return msgs, nil
}

func (s *Server) makeServices(ctx context.Context, dbConn rez.Database, dbc *ent.Client) (*rez.Services, error) {
	jobSvc, jobSvcErr := s.makeJobService(dbConn)
	if jobSvcErr != nil {
		return nil, fmt.Errorf("job service: %w", jobSvcErr)
	}

	msgs, msgsErr := s.makeMessageService()
	if msgsErr != nil {
		return nil, fmt.Errorf("watermill.NewMessageService: %w", msgsErr)
	}

	intgs, intgsErr := db.NewIntegrationsService(dbc, jobSvc)
	if intgsErr != nil {
		return nil, fmt.Errorf("db.NewIntegrationsService: %w", intgsErr)
	}

	orgs, orgsErr := db.NewOrganizationsService(dbc, jobSvc)
	if orgsErr != nil {
		return nil, fmt.Errorf("postgres.NewOrganizationsService: %w", orgsErr)
	}

	users, usersErr := db.NewUserService(dbc, orgs)
	if usersErr != nil {
		return nil, fmt.Errorf("postgres.NewUserService: %w", usersErr)
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

	ai, aiErr := eino.NewAiAgentService(ctx)
	if aiErr != nil {
		return nil, fmt.Errorf("eino.NewAiAgentService: %w", aiErr)
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

	shifts, shiftsErr := db.NewOncallShiftsService(dbc, jobSvc)
	if shiftsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallShiftsService: %w", shiftsErr)
	}
	jobs.RegisterPeriodicJob(jobs.ScanOncallShiftsPeriodicJob)
	jobs.RegisterWorkerFunc(shifts.HandlePeriodicScanShifts)
	jobs.RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverReminderSent)
	jobs.RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverSent)

	oncallMetrics, oncallMetricsErr := db.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallMetricsService: %w", oncallMetricsErr)
	}
	jobs.RegisterWorkerFunc(oncallMetrics.HandleGenerateShiftMetrics)

	debriefs, debriefsErr := db.NewDebriefService(dbc, jobSvc, ai)
	if debriefsErr != nil {
		return nil, fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}
	jobs.RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	jobs.RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	jobs.RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)

	retros, retrosErr := db.NewRetrospectiveService(dbc)
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

	auth, authErr := http.NewAuthSessionService(ctx, orgs, users)
	if authErr != nil {
		return nil, fmt.Errorf("http.NewAuthSessionService: %w", authErr)
	}

	docs, docsErr := db.NewDocumentsService(dbc, auth, users)
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

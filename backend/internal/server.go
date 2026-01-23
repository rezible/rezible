package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rezible/rezible/jobs"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/apiv1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/prosemirror"
	"github.com/rezible/rezible/internal/slack"
	"github.com/rezible/rezible/internal/watermill"
)

func RunAutoMigrations(ctx context.Context) error {
	return postgres.RunAutoMigrations(ctx)
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

func (s *Server) addListener(name string, l rez.EventListener) {
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
	dbConn, dbConnErr := s.OpenDatabase(ctx)
	if dbConnErr != nil {
		return fmt.Errorf("database: %w", dbConnErr)
	}

	dbc := dbConn.Client()

	jobSvc, jobSvcErr := s.makeJobService(dbConn)
	if jobSvcErr != nil {
		return fmt.Errorf("job service: %w", jobSvcErr)
	}

	msgs, msgsErr := s.makeMessageService()
	if msgsErr != nil {
		return fmt.Errorf("watermill.NewMessageService: %w", msgsErr)
	}

	syncer := datasync.NewSyncer(dbc)
	jobs.RegisterPeriodicJob(jobs.SyncAllTenantIntegrationsDataPeriodicJob)
	jobs.RegisterWorkerFunc(syncer.SyncIntegrationsData)

	intgs, intgsErr := db.NewIntegrationsService(dbc, jobSvc, syncer)
	if intgsErr != nil {
		return fmt.Errorf("db.NewIntegrationsService: %w", intgsErr)
	}

	orgs, orgsErr := db.NewOrganizationsService(dbc, jobSvc)
	if orgsErr != nil {
		return fmt.Errorf("postgres.NewOrganizationsService: %w", orgsErr)
	}

	users, usersErr := db.NewUserService(dbc, orgs)
	if usersErr != nil {
		return fmt.Errorf("postgres.NewUserService: %w", usersErr)
	}

	auth, authErr := http.NewAuthSessionService(ctx, orgs, users)
	if authErr != nil {
		return fmt.Errorf("http.NewAuthSessionService: %w", authErr)
	}

	events, eventsErr := db.NewEventsService(dbc, users)
	if eventsErr != nil {
		return fmt.Errorf("postgres.NewEventsService: %w", eventsErr)
	}

	annos, annosErr := db.NewEventAnnotationsService(dbc, events)
	if annosErr != nil {
		return fmt.Errorf("postgres.NewEventAnnotationsService: %w", annosErr)
	}

	_, teamsErr := db.NewTeamService(dbc)
	if teamsErr != nil {
		return fmt.Errorf("postgres.NewTeamService: %w", teamsErr)
	}

	_, nodesErr := prosemirror.NewNodeService()
	if nodesErr != nil {
		return fmt.Errorf("prosemirror.NewNodeService: %w", nodesErr)
	}

	ai, aiErr := eino.NewAiAgentService(ctx)
	if aiErr != nil {
		return fmt.Errorf("eino.NewAiAgentService: %w", aiErr)
	}

	incidents, incidentsErr := db.NewIncidentService(dbc, jobSvc, msgs, users)
	if incidentsErr != nil {
		return fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	rosters, rostersErr := db.NewOncallRostersService(dbc, jobSvc)
	if rostersErr != nil {
		return fmt.Errorf("postgres.NewOncallRostersService: %w", rostersErr)
	}

	components, componentsErr := db.NewSystemComponentsService(dbc)
	if componentsErr != nil {
		return fmt.Errorf("postgres.NewSystemComponentsService: %w", componentsErr)
	}

	chat, chatErr := slack.NewChatService(jobSvc, msgs, intgs, users, incidents, annos, components)
	if chatErr != nil {
		return fmt.Errorf("postgres.NewChatService: %w", chatErr)
	}

	shifts, shiftsErr := db.NewOncallShiftsService(dbc, jobSvc)
	if shiftsErr != nil {
		return fmt.Errorf("postgres.NewOncallShiftsService: %w", shiftsErr)
	}
	jobs.RegisterPeriodicJob(jobs.ScanOncallShiftsPeriodicJob)
	jobs.RegisterWorkerFunc(shifts.HandlePeriodicScanShifts)
	jobs.RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverReminderSent)
	jobs.RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverSent)

	oncallMetrics, oncallMetricsErr := db.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return fmt.Errorf("postgres.NewOncallMetricsService: %w", oncallMetricsErr)
	}
	jobs.RegisterWorkerFunc(oncallMetrics.HandleGenerateShiftMetrics)

	debriefs, debriefsErr := db.NewDebriefService(dbc, jobSvc, ai)
	if debriefsErr != nil {
		return fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}
	jobs.RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	jobs.RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	jobs.RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)

	retros, retrosErr := db.NewRetrospectiveService(dbc)
	if retrosErr != nil {
		return fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	alerts, alertsErr := db.NewAlertService(dbc)
	if alertsErr != nil {
		return fmt.Errorf("postgres.NewAlertService: %w", alertsErr)
	}

	playbooks, playbooksErr := db.NewPlaybookService(dbc)
	if playbooksErr != nil {
		return fmt.Errorf("postgres.NewPlaybookService: %w", playbooksErr)
	}

	docs, docsErr := db.NewDocumentsService(dbc, auth, users)
	if docsErr != nil {
		return fmt.Errorf("db.NewDocumentsService: %w", docsErr)
	}

	apiv1Handler := apiv1.NewHandler(
		dbc,
		auth,
		orgs,
		intgs,
		users,
		incidents,
		debriefs,
		rosters,
		shifts,
		oncallMetrics,
		events,
		annos,
		docs,
		retros,
		components,
		alerts,
		playbooks,
	)

	srv, srvErr := s.makeHttpServer(auth, apiv1Handler)
	if srvErr != nil {
		return fmt.Errorf("http server: %w", srvErr)
	}

	if chatEventsErr := s.setupChatEventListener(chat, srv); chatEventsErr != nil {
		return fmt.Errorf("chatEvents: %w", chatEventsErr)
	}

	return nil
}

func OpenDatabase(ctx context.Context) (rez.Database, error) {
	dbc, dbcErr := postgres.NewDatabaseClient(ctx)
	if dbcErr != nil {
		return nil, fmt.Errorf("postgres.NewDatabaseClient: %w", dbcErr)
	}
	return dbc, nil
}

func (s *Server) OpenDatabase(ctx context.Context) (rez.Database, error) {
	dbc, dbcErr := OpenDatabase(ctx)
	if dbcErr != nil {
		return nil, dbcErr
	}
	s.addListener("database", db.NewListener(dbc))
	return dbc, nil
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
	s.addListener("job_service", jobSvc)
	return jobSvc, nil
}

func (s *Server) makeMessageService() (rez.MessageService, error) {
	msgs, msgsErr := watermill.NewMessageService()
	if msgsErr != nil {
		return nil, fmt.Errorf("watermill.NewMessageService: %w", msgsErr)
	}
	s.addListener("message_service", msgs)
	return msgs, nil
}

func (s *Server) makeHttpServer(auth rez.AuthService, oapiHandler oapiv1.Handler) (*http.Server, error) {
	srv := http.NewServer(auth)
	srv.MountOpenApiV1(oapiHandler)
	srv.MountMCP(eino.NewMCPHandler(auth))

	frontendFS, feFSErr := http.GetEmbeddedFrontendFiles()
	if feFSErr != nil {
		return nil, fmt.Errorf("failed to get embedded frontend files: %w", feFSErr)
	}
	srv.MountStaticFrontend(frontendFS)
	s.addListener("http_server", srv)

	return srv, nil
}

func (s *Server) setupChatEventListener(chat rez.ChatService, srv *http.Server) error {
	if chat.EnableEventListener() {
		sml, listenerErr := chat.MakeEventListener()
		if listenerErr != nil {
			return fmt.Errorf("chat.MakeEventListener: %w", listenerErr)
		}
		s.addListener("chat_events", sml)
	} else if slackChatSvc, ok := chat.(*slack.ChatService); ok {
		webhooks, whErr := slack.NewWebhookEventListener(slackChatSvc)
		if whErr != nil {
			return fmt.Errorf("slack.NewWebhookEventListener: %w", whErr)
		}
		srv.AddWebhookPathHandler("/slack", webhooks.Handler())
		srv.AddWebhookPathHandler("/foo", webhooks.Handler())
	}
	return nil
}

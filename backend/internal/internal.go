package internal

import (
	"context"
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/dataproviders"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/hocuspocus"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/slack"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
)

func RunAutoMigrations(ctx context.Context) error {
	return postgres.RunAutoMigrations(ctx)
}

func OpenPostgresDatabase(ctx context.Context) (rez.Database, error) {
	// TODO: allow different db?
	return postgres.NewDatabaseClient(ctx)
}

func RunServer(ctx context.Context) error {
	ctx = access.AnonymousContext(ctx)

	listeners, setupErr := setupListeners(ctx)
	if setupErr != nil {
		return fmt.Errorf("setup: %s", setupErr)
	}

	runErr := runListeners(ctx, listeners)

	if stopErr := stopListeners(listeners); stopErr != nil {
		log.Error().Err(stopErr).Msg("Failed to stop rezible server")
	}

	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		return runErr
	}
	return nil
}

type listener interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func runListeners(ctx context.Context, listeners map[string]listener) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx)
		for _, l := range listeners {
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

func stopListeners(listeners map[string]listener) error {
	timeout := rez.Config.ServerStopTimeout()
	timeoutCtx, cancelStopCtx := context.WithTimeout(context.Background(), timeout)
	defer cancelStopCtx()

	var err error
	for name, l := range listeners {
		if listenerErr := l.Stop(timeoutCtx); listenerErr != nil && !errors.Is(listenerErr, context.Canceled) {
			err = errors.Join(err, fmt.Errorf("stopping %s: %w", name, listenerErr))
		}
	}

	return err
}

func setupListeners(ctx context.Context) (map[string]listener, error) {
	listeners := make(map[string]listener)

	dbConn, dbConnErr := OpenPostgresDatabase(ctx)
	if dbConnErr != nil {
		return nil, fmt.Errorf("postgres.NewDatabaseClient: %w", dbConnErr)
	}
	listeners["database"] = dbListener{dbc: dbConn}

	pgDb, ok := dbConn.(*postgres.DatabaseClient)
	if !ok {
		return nil, errors.New("non-postgres db client with river job service")
	}
	jobSvc, jobSvcErr := river.NewJobService(pgDb.Pool())
	if jobSvcErr != nil {
		return nil, fmt.Errorf("river.NewJobService: %w", jobSvcErr)
	}
	listeners["job_service"] = jobSvc

	dbc := dbConn.Client()

	pc, pcErr := db.NewProviderConfigService(dbc)
	if pcErr != nil {
		return nil, fmt.Errorf("db.NewProviderConfigService: %w", pcErr)
	}

	orgs, orgsErr := db.NewOrganizationsService(dbc)
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

	_, teamsErr := db.NewTeamService(dbc)
	if teamsErr != nil {
		return nil, fmt.Errorf("postgres.NewTeamService: %w", teamsErr)
	}

	lms, lmsErr := eino.NewLanguageModelService(ctx)
	if lmsErr != nil {
		return nil, fmt.Errorf("eino.NewLanguageModelService: %w", lmsErr)
	}

	incidents, incidentsErr := db.NewIncidentService(dbc, jobSvc, users)
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

	chat, chatErr := slack.NewChatService(jobSvc, users, incidents, annos, components)
	if chatErr != nil {
		return nil, fmt.Errorf("postgres.NewChatService: %w", chatErr)
	}

	shifts, shiftsErr := db.NewOncallShiftsService(dbc, jobSvc, chat)
	if shiftsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallShiftsService: %w", shiftsErr)
	}

	oncallMetrics, oncallMetricsErr := db.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallMetricsService: %w", oncallMetricsErr)
	}

	debriefs, debriefsErr := db.NewDebriefService(dbc, jobSvc, lms)
	if debriefsErr != nil {
		return nil, fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

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

	// TODO: just move this into api
	docs, docsErr := hocuspocus.NewDocumentsService(dbc, auth, users)
	if docsErr != nil {
		return nil, fmt.Errorf("hocuspocus.NewDocumentsService: %w", docsErr)
	}

	apiHandler := api.NewHandler(dbc, auth, orgs, pc, users, incidents, debriefs, rosters, shifts, oncallMetrics, events, annos, docs, retros, components, alerts, playbooks)

	srv := http.NewServer(auth)
	srv.MountOpenApi(apiHandler)
	srv.MountDocuments(docs)
	srv.MountMCP(eino.NewMCPHandler(auth))

	frontendFS, feFSErr := http.GetEmbeddedFrontendFiles()
	if feFSErr != nil {
		return nil, fmt.Errorf("failed to get embedded frontend files: %w", feFSErr)
	}
	srv.MountStaticFrontend(frontendFS)
	listeners["httpServer"] = srv

	if chat.EnableEventListener() {
		sml, listenerErr := chat.MakeEventListener()
		if listenerErr != nil {
			return nil, fmt.Errorf("slack.NewSocketModeEventListener: %w", listenerErr)
		}
		listeners["chat events"] = sml
	} else {
		webhooks, whErr := slack.NewWebhookEventHandler(chat)
		if whErr != nil {
			return nil, fmt.Errorf("slack.NewWebhookEventListener: %w", whErr)
		}
		srv.AddWebhookPathHandler("/slack", webhooks.Handler())
	}

	syncSvc := datasync.NewProviderSyncService(dbc, dataproviders.NewProviderLoader(pc))
	river.RegisterJobWorkers(chat, syncSvc, shifts, oncallMetrics, debriefs)

	return listeners, nil
}

type dbListener struct {
	dbc rez.Database
}

func (l dbListener) Start(ctx context.Context) error {
	return nil
}

func (l dbListener) Stop(ctx context.Context) error {
	// log.Debug().Msg("Closing database connection")
	return l.dbc.Close()
}

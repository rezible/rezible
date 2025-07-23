package main

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/datasync"
	"github.com/rezible/rezible/internal/prosemirror"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/internal/river"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
)

type rezServer struct {
	opts *Options

	db         *postgres.Database
	jobs       *river.JobService
	httpServer *http.Server
}

func newRezibleServer(opts *Options) *rezServer {
	if opts.Mode == "PROD" {
		rez.DebugMode = false
	} else {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return &rezServer{opts: opts}
}

func (s *rezServer) Start(ctx context.Context) {
	if setupErr := s.setup(ctx); setupErr != nil {
		s.Stop(ctx)
		log.Fatal().Err(setupErr).Msg("failed to setup rezible server")
	}

	if startErr := s.start(ctx); startErr != nil {
		log.Fatal().Err(startErr).Msg("rezServer.start")
	}
}

func (s *rezServer) setup(ctx context.Context) error {
	db, poolErr := postgres.Open(ctx, s.opts.DatabaseUrl)
	if poolErr != nil {
		return fmt.Errorf("failed to open db: %w", poolErr)
	}
	s.db = db

	jobs, jobsErr := river.NewJobService(db.Pool)
	if jobsErr != nil {
		return fmt.Errorf("failed to create job service: %w", jobsErr)
	}
	s.jobs = jobs

	dbc := db.Client()

	frontendFiles, feFilesErr := http.GetEmbeddedFrontendFiles()
	if feFilesErr != nil {
		return fmt.Errorf("failed to make embedded frontend server: %w", feFilesErr)
	}

	pl := providers.NewProviderLoader(dbc.ProviderConfig)

	provs, provsErr := pl.LoadProviders(ctx)
	if provsErr != nil {
		return fmt.Errorf("failed to load providers: %w", provsErr)
	}

	sc := datasync.NewSyncController(dbc, pl)
	if syncErr := sc.RegisterPeriodicSyncJob(jobs); syncErr != nil {
		return fmt.Errorf("datasyncer.SyncController.RegisterPeriodicSyncJob: %w", syncErr)
	}
	//syncer := providers.NewProviderDataSyncer(dbc, pl)
	//if syncErr := syncer.RegisterPeriodicSyncJob(jobs, time.Hour); syncErr != nil {
	//	return fmt.Errorf("failed to register data sync job: %w", syncErr)
	//}

	users, usersErr := postgres.NewUserService(dbc)
	if usersErr != nil {
		return fmt.Errorf("postgres.NewUserService: %w", usersErr)
	}

	auth, authErr := http.NewAuthSessionService(ctx, users, provs.AuthSession, s.opts.AuthSessionSecretKey)
	if authErr != nil {
		return fmt.Errorf("http.NewAuthSessionService: %w", authErr)
	}

	chat, chatErr := postgres.NewChatService(dbc, jobs, provs.Chat)
	if chatErr != nil {
		return fmt.Errorf("postgres.NewChatService: %w", chatErr)
	}

	_, teamsErr := postgres.NewTeamService(dbc)
	if teamsErr != nil {
		return fmt.Errorf("postgres.NewTeamService: %w", teamsErr)
	}

	lms, lmsErr := eino.NewLanguageModelService(ctx, provs.LanguageModel)
	if lmsErr != nil {
		return fmt.Errorf("eino.NewLanguageModelService: %w", lmsErr)
	}

	docs, docsErr := prosemirror.NewDocumentsService(s.opts.DocumentServerAddress, users)
	if docsErr != nil {
		return fmt.Errorf("prosemirror.NewDocumentsService: %w", docsErr)
	}

	incidents, incidentsErr := postgres.NewIncidentService(ctx, dbc, jobs, lms, chat, users)
	if incidentsErr != nil {
		return fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	oncall, oncallErr := postgres.NewOncallService(ctx, dbc, jobs, docs, chat, users, incidents)
	if oncallErr != nil {
		return fmt.Errorf("postgres.NewOncallService: %w", oncallErr)
	}

	oncallEvents, eventsErr := postgres.NewOncallEventsService(ctx, dbc, users, oncall, incidents)
	if eventsErr != nil {
		return fmt.Errorf("postgres.NewOncallEventsService: %w", eventsErr)
	}

	debriefs, debriefsErr := postgres.NewDebriefService(dbc, jobs, lms, chat)
	if debriefsErr != nil {
		return fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

	retros, retrosErr := postgres.NewRetrospectiveService(dbc)
	if retrosErr != nil {
		return fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	components, componentsErr := postgres.NewSystemComponentsService(dbc)
	if componentsErr != nil {
		return fmt.Errorf("postgres.NewSystemComponentsService: %w", componentsErr)
	}

	alerts, alertsErr := postgres.NewAlertService(dbc, jobs, provs.AlertsData)
	if alertsErr != nil {
		return fmt.Errorf("postgres.NewAlertService: %w", alertsErr)
	}

	playbooks, playbooksErr := postgres.NewPlaybookService(dbc, provs.PlaybooksData)
	if playbooksErr != nil {
		return fmt.Errorf("postgres.NewPlaybookService: %w", playbooksErr)
	}

	provs.Chat.SetMessageContextProvider(rez.ChatMessageContextProvider{
		LookupChatUserFn:         users.GetByChatId,
		AnnotateMessageFn:        oncallEvents.UpdateAnnotation,
		LookupChatMessageEventFn: oncallEvents.GetProviderEvent,
	})
	provs.Chat.SetEventHandler(chat)

	webhookHandler := pl.WebhookHandler()
	apiHandler := api.NewHandler(dbc, auth, users, incidents, debriefs, oncall, oncallEvents, docs, retros, components, alerts, playbooks)
	mcpHandler := eino.NewMCPHandler(auth)

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	s.httpServer = http.NewServer(listenAddr, auth, apiHandler, frontendFiles, webhookHandler, mcpHandler)

	return nil
}

func (s *rezServer) start(ctx context.Context) error {
	if jobsErr := s.jobs.Start(ctx); jobsErr != nil {
		return fmt.Errorf("failed to start background jobs client: %w", jobsErr)
	}

	if serverErr := s.httpServer.Start(ctx); serverErr != nil {
		return fmt.Errorf("http Server error: %w", serverErr)
	}

	return nil
}

func (s *rezServer) Stop(ctx context.Context) {
	if s.httpServer != nil {
		if dbErr := s.httpServer.Stop(ctx); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to stop http server")
		}
	}

	if s.jobs != nil {
		if dbErr := s.jobs.Stop(ctx); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to stop jobs client")
		}
	}

	if s.db != nil {
		if dbErr := s.db.Close(); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to close db")
		}
	}
}

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/ai"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/prosemirror"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/internal/river"
)

type rezServer struct {
	opts *Options

	db         *postgres.Database
	jobs       *river.JobService
	httpServer *http.Server
}

func newRezServer(opts *Options) *rezServer {
	if opts.Mode == "PROD" {
		rez.DebugMode = false
	} else {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return &rezServer{opts: opts}
}

func (s *rezServer) Start() {
	ctx := context.Background()

	if setupErr := s.setup(ctx); setupErr != nil {
		s.Stop()
		log.Fatal().Err(setupErr).Msg("failed to setup rezible server")
	}

	if jobsErr := s.jobs.Start(ctx); jobsErr != nil {
		log.Fatal().Err(jobsErr).Msg("failed to start background jobs client")
	}

	if serverErr := s.httpServer.Start(ctx); serverErr != nil {
		log.Fatal().Err(serverErr).Msg("server error")
	}
}

func (s *rezServer) setup(ctx context.Context) error {
	db, poolErr := postgres.Open(ctx, s.opts.DatabaseUrl)
	if poolErr != nil {
		return fmt.Errorf("failed to open db: %w", poolErr)
	}
	s.db = db

	jobSvc, jobsErr := river.NewJobService(db.Pool)
	if jobsErr != nil {
		return fmt.Errorf("failed to create job service: %w", jobsErr)
	}
	s.jobs = jobSvc

	srv, srvErr := s.setupServices(ctx, db.Client(), jobSvc)
	if srvErr != nil {
		return fmt.Errorf("failed to setup http server: %w", srvErr)
	}
	s.httpServer = srv

	return nil
}

func (s *rezServer) setupServices(ctx context.Context, dbc *ent.Client, j rez.JobsService) (*http.Server, error) {
	pl := providers.NewProviderLoader(dbc.ProviderConfig)
	provs, provsErr := pl.LoadProviders(ctx)
	if provsErr != nil {
		return nil, fmt.Errorf("failed to load providers: %w", provsErr)
	}

	chat := provs.Chat

	users, usersErr := postgres.NewUserService(dbc)
	if usersErr != nil {
		return nil, fmt.Errorf("postgres.UserService: %w", usersErr)
	}

	_, teamsErr := postgres.NewTeamService(dbc)
	if teamsErr != nil {
		return nil, fmt.Errorf("postgres.TeamService: %w", teamsErr)
	}

	llm, llmErr := ai.NewLanguageModelService(ctx, provs.AiModel)
	if llmErr != nil {
		return nil, fmt.Errorf("failed to create language model service: %w", llmErr)
	}

	docs, docsErr := prosemirror.NewDocumentsService(s.opts.DocumentServerAddress, users)
	if docsErr != nil {
		return nil, fmt.Errorf("failed to create document service: %w", docsErr)
	}

	incidents, incidentsErr := postgres.NewIncidentService(ctx, dbc, j, llm, chat, users)
	if incidentsErr != nil {
		return nil, fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	oncall, oncallErr := postgres.NewOncallService(ctx, dbc, j, docs, chat, users, incidents)
	if oncallErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallService: %w", oncallErr)
	}

	oncallEvents, eventsErr := postgres.NewOncallEventsService(ctx, dbc, users, oncall, incidents)
	if eventsErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallEventsService: %w", eventsErr)
	}

	chat.SetMessageAnnotator(oncallEvents)

	debriefs, debriefsErr := postgres.NewDebriefService(dbc, j, llm, chat)
	if debriefsErr != nil {
		return nil, fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

	retros, retrosErr := postgres.NewRetrospectiveService(dbc)
	if retrosErr != nil {
		return nil, fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	components, cmpsErr := postgres.NewSystemComponentsService(dbc)
	if cmpsErr != nil {
		return nil, fmt.Errorf("postgres.NewSystemComponentsService: %w", cmpsErr)
	}

	auth, authErr := http.NewAuthSessionService(ctx, users, provs.AuthSession, s.opts.AuthSessionSecretKey)
	if authErr != nil {
		return nil, fmt.Errorf("http auth service: %w", authErr)
	}

	apiHandler := api.NewHandler(dbc, auth, users, incidents, debriefs, oncall, oncallEvents, docs, retros, components)
	webhookHandler := pl.WebhookHandler()
	mcpHandler := ai.NewMCPHandler()

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	httpServer, httpErr := http.NewServer(listenAddr, auth, apiHandler, webhookHandler, mcpHandler)
	if httpErr != nil {
		return nil, fmt.Errorf("http.NewServer: %w", httpErr)
	}

	syncer := providers.NewDataSyncer(dbc, pl)
	if syncErr := syncer.RegisterPeriodicSyncJob(j, time.Hour); syncErr != nil {
		return nil, fmt.Errorf("failed to register data sync job: %w", syncErr)
	}

	return httpServer, nil
}

func (s *rezServer) Stop() {
	timeout := time.Duration(s.opts.StopTimeoutSeconds) * time.Second
	timeoutCtx, cancelStopCtx := context.WithTimeout(context.Background(), timeout)
	defer cancelStopCtx()

	if s.httpServer != nil {
		if dbErr := s.httpServer.Stop(timeoutCtx); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to stop http server")
		}
	}

	if s.jobs != nil {
		if dbErr := s.jobs.Stop(timeoutCtx); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to stop jobs client")
		}
	}

	if s.db != nil {
		if dbErr := s.db.Close(); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to close db")
		}
	}
}

package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/twohundreds/rezible/internal/api"
	"github.com/twohundreds/rezible/internal/documents"
	"github.com/twohundreds/rezible/internal/http"
	"github.com/twohundreds/rezible/internal/langchain"
	"github.com/twohundreds/rezible/internal/postgres"
	"github.com/twohundreds/rezible/internal/providers"
	"github.com/twohundreds/rezible/jobs"
)

type Options struct {
	Mode                  string `doc:"App Mode" default:"debug"`
	Host                  string `doc:"Hostname to listen on." default:"localhost"`
	Port                  string `doc:"Port to listen on." short:"p" default:"8888"`
	StopTimeoutSeconds    int    `doc:"Timeout in seconds to wait before stopping" default:"30"`
	DocumentServerAddress string `doc:"Document server address" name:"document_server_address" default:"localhost:8889"`
	DatabaseUrl           string `doc:"Database connection url" name:"db_url"`
	AuthSessionSecretKey  string `doc:"Auth session secret key" name:"auth_session_secret_key"`
}

type rezServer struct {
	opts *Options

	db         *postgres.Database
	jobClient  *jobs.BackgroundJobClient
	httpServer *http.Server
}

func (s *rezServer) Start() {
	ctx := context.Background()

	if setupErr := s.setup(ctx); setupErr != nil {
		log.Fatal().Err(setupErr).Msg("failed to setup rezible server")
	}

	if jobsErr := s.jobClient.Start(ctx); jobsErr != nil {
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
	s.jobClient = jobs.NewBackgroundJobClient(&log.Logger, db.Pool)

	return s.setupServices(ctx)
}

func (s *rezServer) setupServices(ctx context.Context) error {
	dbc := s.db.Client()
	pl := providers.NewProviderLoader(dbc.ProviderConfig)

	users, usersErr := postgres.NewUserService(dbc, pl)
	if usersErr != nil {
		return fmt.Errorf("postgres.UserService: %w", usersErr)
	}

	chat, chatErr := postgres.NewChatService(ctx, pl, users)
	if chatErr != nil {
		return fmt.Errorf("failed to create chat service: %w", chatErr)
	}

	ai, aiErr := langchain.NewAiService(ctx, pl)
	if aiErr != nil {
		return fmt.Errorf("failed to create AI service: %w", aiErr)
	}

	docs, docsErr := documents.NewService(s.opts.DocumentServerAddress, users)
	if docsErr != nil {
		return fmt.Errorf("failed to create document service: %w", docsErr)
	}

	incidents, incidentsErr := postgres.NewIncidentService(ctx, dbc, s.jobClient, pl, ai, chat, users)
	if incidentsErr != nil {
		return fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	oncall, handoverErr := postgres.NewOncallService(ctx, dbc, s.jobClient, pl, docs, chat, users, incidents)
	if handoverErr != nil {
		return fmt.Errorf("postgres.NewOncallHandoverService: %w", handoverErr)
	}

	retros, retrosErr := postgres.NewRetrospectiveService(dbc)
	if retrosErr != nil {
		return fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	alerts, alertsErr := postgres.NewAlertsService(ctx, dbc, s.jobClient, pl, users)
	if alertsErr != nil {
		return fmt.Errorf("postgres.NewAlertsService: %w", alertsErr)
	}

	auth, authErr := http.NewAuthService(ctx, users, pl, s.opts.AuthSessionSecretKey)
	if authErr != nil {
		return fmt.Errorf("http auth service: %w", authErr)
	}

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	oapiHandler := api.NewHandler(dbc, auth, users, incidents, oncall, alerts, docs, retros)
	httpServer, httpErr := http.NewServer(listenAddr, pl, auth, oapiHandler)
	if httpErr != nil {
		return fmt.Errorf("http.NewServer: %w", httpErr)
	}
	s.httpServer = httpServer

	syncJobErr := jobs.RegisterProviderDataSyncJob(s.jobClient, users, incidents, oncall, alerts)
	if syncJobErr != nil {
		return fmt.Errorf("jobs.RegisterProviderDataSyncJob: %w", syncJobErr)
	}

	return nil
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

	if s.jobClient != nil {
		if dbErr := s.jobClient.Stop(timeoutCtx); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to stop jobs client")
		}
	}

	if s.db != nil {
		if dbErr := s.db.Close(); dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to close db")
		}
	}
}

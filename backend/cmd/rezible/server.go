package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/documents"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/langchain"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/internal/river"
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
	jobs       *river.JobService
	httpServer *http.Server
}

func (s *rezServer) Start() {
	ctx := context.Background()

	if setupErr := s.setup(ctx); setupErr != nil {
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

	return s.setupServices(ctx)
}

func (s *rezServer) setupServices(ctx context.Context) error {
	dbc := s.db.Client()

	s.jobs = river.NewJobService(s.db.Pool)
	pl := providers.NewProviderLoader(dbc.ProviderConfig)

	users, usersErr := postgres.NewUserService(dbc, pl)
	if usersErr != nil {
		return fmt.Errorf("postgres.UserService: %w", usersErr)
	}

	chat, chatErr := documents.NewChatService(ctx, pl, users)
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

	incidents, incidentsErr := postgres.NewIncidentService(ctx, dbc, s.jobs, pl, ai, chat, users)
	if incidentsErr != nil {
		return fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	oncall, handoverErr := postgres.NewOncallService(ctx, dbc, s.jobs, pl, docs, chat, users, incidents)
	if handoverErr != nil {
		return fmt.Errorf("postgres.NewOncallHandoverService: %w", handoverErr)
	}

	debriefs, debriefsErr := postgres.NewDebriefService(dbc, s.jobs, ai, chat)
	if debriefsErr != nil {
		return fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

	retros, retrosErr := postgres.NewRetrospectiveService(dbc)
	if retrosErr != nil {
		return fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	alerts, alertsErr := postgres.NewAlertsService(ctx, dbc, s.jobs, pl, users)
	if alertsErr != nil {
		return fmt.Errorf("postgres.NewAlertsService: %w", alertsErr)
	}

	auth, authErr := http.NewAuthService(ctx, users, pl, s.opts.AuthSessionSecretKey)
	if authErr != nil {
		return fmt.Errorf("http auth service: %w", authErr)
	}

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	apiHandler := api.NewHandler(dbc, auth, users, incidents, debriefs, oncall, alerts, docs, retros)

	httpServer, httpErr := http.NewServer(listenAddr, pl, auth, apiHandler.MakeAdapter())
	if httpErr != nil {
		return fmt.Errorf("http.NewServer: %w", httpErr)
	}
	s.httpServer = httpServer

	jobsErr := s.jobs.RegisterWorkers(users, incidents, oncall, alerts, debriefs)
	if jobsErr != nil {
		return fmt.Errorf("jobs.RegisterWorkers: %w", jobsErr)
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

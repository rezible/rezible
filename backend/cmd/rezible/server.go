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
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/langchain"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/prosemirror"
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

	ai, aiErr := langchain.NewAiService(ctx, provs.AiModel)
	if aiErr != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", aiErr)
	}

	docs, docsErr := prosemirror.NewDocumentsService(s.opts.DocumentServerAddress, users)
	if docsErr != nil {
		return nil, fmt.Errorf("failed to create document service: %w", docsErr)
	}

	incidents, incidentsErr := postgres.NewIncidentService(ctx, dbc, j, ai, chat, users)
	if incidentsErr != nil {
		return nil, fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	oncall, handoverErr := postgres.NewOncallService(ctx, dbc, j, docs, chat, users, incidents)
	if handoverErr != nil {
		return nil, fmt.Errorf("postgres.NewOncallHandoverService: %w", handoverErr)
	}

	provs.Chat.SetAnnotationSupporter(oncall)

	debriefs, debriefsErr := postgres.NewDebriefService(dbc, j, ai, chat)
	if debriefsErr != nil {
		return nil, fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

	retros, retrosErr := postgres.NewRetrospectiveService(dbc)
	if retrosErr != nil {
		return nil, fmt.Errorf("postgres.NewRetrospectiveService: %w", retrosErr)
	}

	componentsSvc, cmpsErr := postgres.NewSystemComponentsService(dbc)
	if cmpsErr != nil {
		return nil, fmt.Errorf("postgres.NewSystemComponentsService: %w", cmpsErr)
	}

	alerts, alertsErr := postgres.NewAlertsService(ctx, dbc, users)
	if alertsErr != nil {
		return nil, fmt.Errorf("postgres.NewAlertsService: %w", alertsErr)
	}

	auth, authErr := http.NewAuthSessionService(ctx, users, provs.AuthSession, s.opts.AuthSessionSecretKey)
	if authErr != nil {
		return nil, fmt.Errorf("http auth service: %w", authErr)
	}

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	apiHandler := api.NewHandler(dbc, auth, users, incidents, debriefs, oncall, alerts, docs, retros, componentsSvc)

	httpServer, httpErr := http.NewServer(listenAddr, auth, apiHandler, pl.WebhookHandler())
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

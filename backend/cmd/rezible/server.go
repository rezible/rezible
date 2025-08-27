package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/goth"
	"github.com/rezible/rezible/internal/hocuspocus"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/datasync"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/internal/river"
	"github.com/rezible/rezible/internal/saml"
	"github.com/rezible/rezible/internal/slack"
	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"
)

type rezServer struct {
	opts *Options

	db         *postgres.Database
	jobs       *river.JobService
	httpServer *http.Server
}

func newRezibleServer(opts *Options) *rezServer {
	rez.DebugMode = opts.DebugMode
	if opts.DebugMode {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return &rezServer{opts: opts}
}

func (s *rezServer) Start() {
	if setupErr := s.setup(); setupErr != nil {
		s.Stop()
		log.Fatal().Err(setupErr).Msg("failed to setup rezible server")
	}

	if startErr := s.start(); startErr != nil {
		log.Fatal().Err(startErr).Msg("rezServer.start")
	}
}

func (s *rezServer) setup() error {
	ctx := access.SystemContext(context.Background())

	db, poolErr := postgres.Open(ctx, s.opts.DatabaseUrl)
	if poolErr != nil {
		return fmt.Errorf("failed to open db: %w", poolErr)
	}
	s.db = db

	jobSvc, jobSvcErr := river.NewJobService(db.Pool)
	if jobSvcErr != nil {
		return fmt.Errorf("failed to create job service: %w", jobSvcErr)
	}
	s.jobs = jobSvc

	dbc := db.Client()

	frontendFiles, feFilesErr := http.GetEmbeddedFrontendFiles()
	if feFilesErr != nil {
		return fmt.Errorf("failed to make embedded frontend server: %w", feFilesErr)
	}

	pl := providers.NewProviderLoader(dbc.ProviderConfig)

	users, usersErr := postgres.NewUserService(dbc)
	if usersErr != nil {
		return fmt.Errorf("postgres.NewUserService: %w", usersErr)
	}

	auth, authErr := s.makeAuthService(ctx, users)
	if authErr != nil {
		return fmt.Errorf("http.NewAuthService: %w", authErr)
	}

	chat, chatErr := slack.NewChatService(users)
	if chatErr != nil {
		return fmt.Errorf("postgres.NewChatService: %w", chatErr)
	}

	_, teamsErr := postgres.NewTeamService(dbc)
	if teamsErr != nil {
		return fmt.Errorf("postgres.NewTeamService: %w", teamsErr)
	}

	lms, lmsErr := eino.NewLanguageModelService(ctx)
	if lmsErr != nil {
		return fmt.Errorf("eino.NewLanguageModelService: %w", lmsErr)
	}

	docs, docsErr := hocuspocus.NewDocumentsService(s.opts.DocumentServerAddress, dbc, auth, users)
	if docsErr != nil {
		return fmt.Errorf("prosemirror.NewDocumentsService: %w", docsErr)
	}

	incidents, incidentsErr := postgres.NewIncidentService(dbc, jobSvc, lms, chat, users)
	if incidentsErr != nil {
		return fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	oncall, oncallErr := postgres.NewOncallService(dbc, jobSvc, docs, chat, users, incidents)
	if oncallErr != nil {
		return fmt.Errorf("postgres.NewOncallService: %w", oncallErr)
	}

	oncallEvents, eventsErr := postgres.NewOncallEventsService(dbc, users, oncall, incidents)
	if eventsErr != nil {
		return fmt.Errorf("postgres.NewOncallEventsService: %w", eventsErr)
	}

	debriefs, debriefsErr := postgres.NewDebriefService(dbc, jobSvc, lms, chat)
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

	alerts, alertsErr := postgres.NewAlertService(dbc)
	if alertsErr != nil {
		return fmt.Errorf("postgres.NewAlertService: %w", alertsErr)
	}

	playbooks, playbooksErr := postgres.NewPlaybookService(dbc, pl)
	if playbooksErr != nil {
		return fmt.Errorf("postgres.NewPlaybookService: %w", playbooksErr)
	}

	chat.SetOncallEventsService(oncallEvents)

	apiHandler := api.NewHandler(dbc, auth, users, incidents, debriefs, oncall, oncallEvents, docs, retros, components, alerts, playbooks)
	documentsHandler := docs.Handler()
	webhookHandler := http.NewWebhooksHandler(chat)
	mcpHandler := eino.NewMCPHandler(auth)

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	s.httpServer = http.NewServer(listenAddr, auth, users, frontendFiles, apiHandler, documentsHandler, webhookHandler, mcpHandler)

	syncSvc := datasync.NewProviderSyncService(dbc, pl)
	if jobsErr := s.registerJobs(syncSvc, oncall, debriefs); jobsErr != nil {
		return fmt.Errorf("registering jobs: %w", jobsErr)
	}

	return nil
}

func authProviderEnabled(name string) bool {
	return strings.ToLower(os.Getenv("AUTH_ENABLE_"+strings.ToUpper(name))) == "true"
}

func (s *rezServer) makeAuthService(ctx context.Context, users rez.UserService) (rez.AuthService, error) {
	var provs []rez.AuthSessionProvider

	if authProviderEnabled("saml") {
		samlProv, spErr := saml.NewAuthSessionProvider(ctx)
		if spErr != nil {
			return nil, fmt.Errorf("saml.NewAuthSessionProvider: %w", spErr)
		}
		provs = append(provs, samlProv)
	}

	if authProviderEnabled("github") {
		ghProv, ghErr := goth.NewGithubProvider()
		if ghErr != nil {
			return nil, fmt.Errorf("goth.NewGithubProvider: %w", ghErr)
		}
		provs = append(provs, ghProv)
	}

	if authProviderEnabled("google_oidc") {
		googleProv, googleErr := goth.NewGoogleOIDCProvider()
		if googleErr != nil {
			return nil, fmt.Errorf("goth.NewGoogleOIDCProvider: %w", googleErr)
		}
		provs = append(provs, googleProv)
	}

	secretKey := os.Getenv("AUTH_SESSION_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("AUTH_SESSION_SECRET_KEY must be set")
	}
	goth.ConfigureSessionStore(secretKey)

	return http.NewAuthService(secretKey, users, provs)
}

func (s *rezServer) registerJobs(
	sync rez.ProviderSyncService,
	oncall rez.OncallService,
	debriefs rez.DebriefService,
) error {
	river.RegisterPeriodicJob(sync.MakeSyncProviderDataPeriodicJob(), sync.SyncProviderData)

	river.RegisterPeriodicJob(oncall.MakeScanShiftsPeriodicJob(), oncall.HandlePeriodicScanShifts)
	river.RegisterWorkerFunc(oncall.HandleEnsureShiftHandoverReminderSent)
	river.RegisterWorkerFunc(oncall.HandleEnsureShiftHandoverSent)
	river.RegisterWorkerFunc(oncall.HandleGenerateShiftMetrics)

	river.RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	river.RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	river.RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)

	return nil
}

func (s *rezServer) start() error {
	ctx := context.Background()

	if jobsErr := s.jobs.Start(access.SystemContext(ctx)); jobsErr != nil {
		return fmt.Errorf("failed to start background jobs client: %w", jobsErr)
	}

	if serverErr := s.httpServer.Start(access.AnonymousContext(ctx)); serverErr != nil {
		return fmt.Errorf("http Server error: %w", serverErr)
	}

	return nil
}

func (s *rezServer) Stop() {
	timeout := time.Duration(s.opts.StopTimeoutSeconds) * time.Second
	ctx, cancelStopCtx := context.WithTimeout(context.Background(), timeout)
	defer cancelStopCtx()

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
		s.db.Close()
	}
}

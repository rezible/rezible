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
	"github.com/rezible/rezible/internal/hocuspocus"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/oidc"
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

	slackSocketListener *slack.SocketModeListener
}

func newRezibleServer(opts *Options) *rezServer {
	rez.DebugMode = opts.DebugMode
	if opts.DebugMode {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return &rezServer{opts: opts}
}

func (s *rezServer) Start(ctx context.Context) error {
	if setupErr := s.setup(); setupErr != nil {
		return fmt.Errorf("failed to setup server: %w", setupErr)
	}

	if jobsErr := s.jobs.Start(access.SystemContext(ctx)); jobsErr != nil {
		return fmt.Errorf("failed to start background jobs client: %w", jobsErr)
	}

	if s.slackSocketListener != nil {
		if slErr := s.slackSocketListener.Start(ctx); slErr != nil {
			return fmt.Errorf("failed to start slack socket listener: %w", slErr)
		}
	}

	return s.httpServer.Start(ctx)
}

func (s *rezServer) Stop(ctx context.Context) {
	timeout := time.Duration(s.opts.StopTimeoutSeconds) * time.Second
	timeoutCtx, cancelStopCtx := context.WithTimeout(ctx, timeout)
	defer cancelStopCtx()

	if s.jobs != nil {
		if jobsErr := s.jobs.Stop(timeoutCtx); jobsErr != nil {
			log.Error().Err(jobsErr).Msg("failed to stop jobs client")
		}
	}

	if s.slackSocketListener != nil {
		if slErr := s.slackSocketListener.Stop(timeoutCtx); slErr != nil {
			log.Error().Err(slErr).Msg("failed to stop slack socket listener")
		}
	}

	if s.httpServer != nil {
		if srvErr := s.httpServer.Stop(timeoutCtx); srvErr != nil {
			log.Error().Err(srvErr).Msg("failed to stop http server")
		}
	}

	if s.db != nil {
		s.db.Close()
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

	configs, pcErr := postgres.NewProviderConfigService(dbc)
	if pcErr != nil {
		return fmt.Errorf("failed to create provider configs: %w", pcErr)
	}
	syncSvc := datasync.NewProviderSyncService(dbc, providers.NewProviderLoader(configs))

	orgs, orgsErr := postgres.NewOrganizationsService(dbc)
	if orgsErr != nil {
		return fmt.Errorf("postgres.NewOrganizationsService: %w", orgsErr)
	}

	users, usersErr := postgres.NewUserService(dbc, orgs)
	if usersErr != nil {
		return fmt.Errorf("postgres.NewUserService: %w", usersErr)
	}

	auth, authErr := s.makeAuthService(ctx, orgs, users)
	if authErr != nil {
		return fmt.Errorf("http.NewAuthService: %w", authErr)
	}

	events, eventsErr := postgres.NewEventsService(dbc, users)
	if eventsErr != nil {
		return fmt.Errorf("postgres.NewEventsService: %w", eventsErr)
	}

	annos, annosErr := postgres.NewEventAnnotationsService(dbc, events)
	if annosErr != nil {
		return fmt.Errorf("postgres.NewEventAnnotationsService: %w", annosErr)
	}

	chat, chatErr := slack.NewChatService(users, annos)
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

	incidents, incidentsErr := postgres.NewIncidentService(dbc, jobSvc, chat, users)
	if incidentsErr != nil {
		return fmt.Errorf("postgres.NewIncidentService: %w", incidentsErr)
	}

	rosters, rostersErr := postgres.NewOncallRostersService(dbc, jobSvc)
	if rostersErr != nil {
		return fmt.Errorf("postgres.NewOncallRostersService: %w", rostersErr)
	}

	shifts, shiftsErr := postgres.NewOncallShiftsService(dbc, jobSvc, chat)
	if shiftsErr != nil {
		return fmt.Errorf("postgres.NewOncallShiftsService: %w", shiftsErr)
	}

	oncallMetrics, oncallMetricsErr := postgres.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return fmt.Errorf("postgres.NewOncallMetricsService: %w", oncallMetricsErr)
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

	playbooks, playbooksErr := postgres.NewPlaybookService(dbc)
	if playbooksErr != nil {
		return fmt.Errorf("postgres.NewPlaybookService: %w", playbooksErr)
	}

	jobsErr := s.registerJobs(syncSvc, shifts, oncallMetrics, debriefs)
	if jobsErr != nil {
		return fmt.Errorf("registering jobs: %w", jobsErr)
	}

	frontendFS, feFSErr := http.GetEmbeddedFrontendFiles()
	if feFSErr != nil {
		return fmt.Errorf("failed to get embedded frontend files: %w", feFSErr)
	}

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	srv := http.NewServer(listenAddr, auth)

	apiHandler := api.NewHandler(dbc, auth, orgs, configs, users, incidents, debriefs, rosters, shifts, oncallMetrics, events, annos, docs, retros, components, alerts, playbooks)
	srv.MountOpenApi(apiHandler)
	srv.MountDocuments(docs)
	srv.MountMCP(eino.NewMCPHandler(auth))
	srv.MountStaticFrontend(frontendFS)

	if slack.UseSocketMode() {
		sml, listenerErr := slack.NewSocketModeEventListener(chat)
		if listenerErr != nil {
			return fmt.Errorf("slack.NewSocketModeEventListener: %w", listenerErr)
		}
		s.slackSocketListener = sml
	} else {
		webhookListener, whErr := slack.NewWebhookEventListener(chat)
		if whErr != nil {
			return fmt.Errorf("slack.NewWebhookEventListener: %w", whErr)
		}
		srv.AddWebhookPathHandler("/slack", webhookListener.Handler())
	}

	s.httpServer = srv

	return nil
}

func authProviderEnabled(name string) bool {
	return strings.ToLower(os.Getenv("AUTH_ENABLE_"+strings.ToUpper(name))) == "true"
}

func (s *rezServer) makeAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (rez.AuthService, error) {
	var provs []rez.AuthSessionProvider

	secretKey := os.Getenv("AUTH_SESSION_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("AUTH_SESSION_SECRET_KEY must be set")
	}

	if authProviderEnabled("saml") {
		samlProv, spErr := saml.NewAuthSessionProvider(ctx)
		if spErr != nil {
			return nil, fmt.Errorf("saml.NewAuthSessionProvider: %w", spErr)
		}
		provs = append(provs, samlProv)
	}

	if authProviderEnabled("google_oidc") {
		clientID := os.Getenv("GOOGLE_OIDC_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_OIDC_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			return nil, fmt.Errorf("client id/secret env vars not set")
		}

		cfg := oidc.ProviderConfig{
			SessionSecret: secretKey,
			ClientID:      clientID,
			ClientSecret:  clientSecret,
		}
		googleProv, googleErr := oidc.NewGoogleAuthSessionProvider(ctx, cfg)
		if googleErr != nil {
			return nil, fmt.Errorf("oidc.NewGoogleProvider: %w", googleErr)
		}

		provs = append(provs, googleProv)
	}

	return http.NewAuthService(secretKey, orgs, users, provs)
}

func (s *rezServer) registerJobs(
	sync rez.ProviderSyncService,
	shifts rez.OncallShiftsService,
	oncallMetrics rez.OncallMetricsService,
	debriefs rez.DebriefService,
) error {
	river.RegisterPeriodicJob(sync.MakeSyncProviderDataPeriodicJob(), sync.SyncProviderData)

	river.RegisterPeriodicJob(shifts.MakeScanShiftsPeriodicJob(), shifts.HandlePeriodicScanShifts)
	river.RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverReminderSent)
	river.RegisterWorkerFunc(shifts.HandleEnsureShiftHandoverSent)
	river.RegisterWorkerFunc(oncallMetrics.HandleGenerateShiftMetrics)

	river.RegisterWorkerFunc(debriefs.HandleGenerateDebriefResponse)
	river.RegisterWorkerFunc(debriefs.HandleGenerateSuggestions)
	river.RegisterWorkerFunc(debriefs.HandleSendDebriefRequests)

	return nil
}

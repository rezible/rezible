package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
	"github.com/spf13/cobra"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/eino"
	"github.com/rezible/rezible/internal/hocuspocus"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/oidc"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/internal/saml"
	"github.com/rezible/rezible/internal/slack"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "runs the rezible server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRezibleServer(cmd.Context()); err != nil {
			log.Fatal().Err(err).Msg("failed to serve")
		}
	},
}

// TODO: remove this
type Options struct {
	DebugMode                   bool   `doc:"App Debug Mode" name:"debug" default:"false"`
	Host                        string `doc:"Hostname to listen on." default:"localhost"`
	Port                        string `doc:"Port to listen on." short:"p" default:"8888"`
	StopTimeoutSeconds          int    `doc:"Timeout in seconds to wait before cancelling" default:"10"`
	DocumentServerAddress       string `doc:"Document server address" name:"document_server_address" default:"localhost:8889"`
	DocumentServerWebhookSecret string `doc:"Document server webhook secret" name:"document_server_webhook_secret"`
	DatabaseUrl                 string `doc:"Database connection url" name:"db_url"`
}

func loadOptions(ctx context.Context) Options {
	opts := Options{
		DebugMode:                   os.Getenv("REZ_DEBUG") == "true",
		Host:                        "localhost",
		Port:                        "8888",
		DatabaseUrl:                 os.Getenv("DB_URL"),
		StopTimeoutSeconds:          10,
		DocumentServerAddress:       "localhost:8889",
		DocumentServerWebhookSecret: os.Getenv("DOCUMENT_SERVER_WEBHOOK_SECRET"),
	}

	rez.DebugMode = opts.DebugMode
	if opts.DebugMode {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return opts
}

type rezServer struct {
	opts Options

	errChan   chan error
	listeners map[string]listener
	closers   map[string]io.Closer
}

type listener interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func runRezibleServer(ctx context.Context) error {
	ctx = access.AnonymousContext(ctx)
	opts := loadOptions(ctx)

	s := &rezServer{
		opts:      opts,
		errChan:   make(chan error),
		listeners: make(map[string]listener),
		closers:   make(map[string]io.Closer),
	}

	if setupErr := s.setup(ctx); setupErr != nil {
		return fmt.Errorf("setup failed: %w", setupErr)
	}

	go s.Start(ctx)

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case waitErr := <-s.errChan:
		err = fmt.Errorf("server error: %w", waitErr)
	}

	stopErr := s.Stop()
	if stopErr != nil {
		log.Error().Err(stopErr).Msg("Failed to stop rezible server")
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (s *rezServer) Start(ctx context.Context) {
	p := pool.New().
		WithErrors().
		WithContext(ctx)

	for _, l := range s.listeners {
		p.Go(l.Start)
	}

	s.errChan <- p.Wait()
}

func (s *rezServer) Stop() error {
	timeout := time.Duration(s.opts.StopTimeoutSeconds) * time.Second
	timeoutCtx, cancelStopCtx := context.WithTimeout(context.Background(), timeout)
	defer cancelStopCtx()

	var err error
	for name, l := range s.listeners {
		if listenerErr := l.Stop(timeoutCtx); listenerErr != nil && !errors.Is(listenerErr, context.Canceled) {
			err = errors.Join(err, fmt.Errorf("stopping %s: %w", name, listenerErr))
		}
	}

	for name, c := range s.closers {
		if closeErr := c.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing %s: %w", name, closeErr))
		}
	}

	return err
}

func (s *rezServer) setup(ctx context.Context) error {
	pgClient, poolErr := postgres.Open(ctx, s.opts.DatabaseUrl)
	if poolErr != nil {
		return fmt.Errorf("failed to open db: %w", poolErr)
	}
	s.closers["database_client"] = pgClient

	jobSvc, jobSvcErr := river.NewJobService(pgClient.Pool)
	if jobSvcErr != nil {
		return fmt.Errorf("failed to create job service: %w", jobSvcErr)
	}
	s.listeners["jobs"] = jobSvc

	dbc := pgClient.Client()

	configs, pcErr := db.NewProviderConfigService(dbc)
	if pcErr != nil {
		return fmt.Errorf("failed to create provider configs: %w", pcErr)
	}
	syncSvc := datasync.NewProviderSyncService(dbc, providers.NewProviderLoader(configs))

	orgs, orgsErr := db.NewOrganizationsService(dbc)
	if orgsErr != nil {
		return fmt.Errorf("postgres.NewOrganizationsService: %w", orgsErr)
	}

	users, usersErr := db.NewUserService(dbc, orgs)
	if usersErr != nil {
		return fmt.Errorf("postgres.NewUserService: %w", usersErr)
	}

	auth, authErr := makeAuthService(ctx, orgs, users)
	if authErr != nil {
		return fmt.Errorf("http.NewAuthService: %w", authErr)
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

	lms, lmsErr := eino.NewLanguageModelService(ctx)
	if lmsErr != nil {
		return fmt.Errorf("eino.NewLanguageModelService: %w", lmsErr)
	}

	docs, docsErr := hocuspocus.NewDocumentsService(s.opts.DocumentServerAddress, dbc, auth, users)
	if docsErr != nil {
		return fmt.Errorf("prosemirror.NewDocumentsService: %w", docsErr)
	}

	incidents, incidentsErr := db.NewIncidentService(dbc, jobSvc, users)
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

	chat, chatErr := slack.NewChatService(jobSvc, users, incidents, annos, components)
	if chatErr != nil {
		return fmt.Errorf("postgres.NewChatService: %w", chatErr)
	}

	shifts, shiftsErr := db.NewOncallShiftsService(dbc, jobSvc, chat)
	if shiftsErr != nil {
		return fmt.Errorf("postgres.NewOncallShiftsService: %w", shiftsErr)
	}

	oncallMetrics, oncallMetricsErr := db.NewOncallMetricsService(dbc, jobSvc, shifts)
	if oncallMetricsErr != nil {
		return fmt.Errorf("postgres.NewOncallMetricsService: %w", oncallMetricsErr)
	}

	debriefs, debriefsErr := db.NewDebriefService(dbc, jobSvc, lms)
	if debriefsErr != nil {
		return fmt.Errorf("postgres.NewDebriefService: %w", debriefsErr)
	}

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

	frontendFS, feFSErr := http.GetEmbeddedFrontendFiles()
	if feFSErr != nil {
		return fmt.Errorf("failed to get embedded frontend files: %w", feFSErr)
	}

	river.RegisterJobWorkers(chat, syncSvc, shifts, oncallMetrics, debriefs)

	listenAddr := net.JoinHostPort(s.opts.Host, s.opts.Port)
	srv := http.NewServer(listenAddr, auth)

	apiHandler := api.NewHandler(dbc, auth, orgs, configs, users, incidents, debriefs, rosters, shifts, oncallMetrics, events, annos, docs, retros, components, alerts, playbooks)
	srv.MountOpenApi(apiHandler)
	srv.MountDocuments(docs)
	srv.MountMCP(eino.NewMCPHandler(auth))
	srv.MountStaticFrontend(frontendFS)

	if chat.EnableEventListener() {
		sml, listenerErr := chat.MakeEventListener()
		if listenerErr != nil {
			return fmt.Errorf("slack.NewSocketModeEventListener: %w", listenerErr)
		}
		s.listeners["chat events"] = sml
	} else {
		whHandler, whErr := slack.NewWebhookEventHandler(chat)
		if whErr != nil {
			return fmt.Errorf("slack.NewWebhookEventListener: %w", whErr)
		}
		srv.AddWebhookPathHandler("/slack", whHandler.Handler())
	}

	s.listeners["httpServer"] = srv

	return nil
}

func authProviderEnabled(name string) bool {
	return strings.ToLower(os.Getenv("AUTH_ENABLE_"+strings.ToUpper(name))) == "true"
}

func makeAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (rez.AuthService, error) {
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

package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	sessionCookieName = "saml_session"

	IdPMetadataUrlEnv = "SAML_IDP_METADATA_URL"
	CertFileEnv       = "SAML_CERT_FILE"
	CertKeyFileEnv    = "SAML_CERT_KEY_FILE"
)

var (
	userMapping = &ent.User{
		Name:  "y",
		Email: "y",
	}
)

type Config struct {
	IdPMetadataUrl string
	CertFile       string
	CertKeyFile    string
}

type AuthSessionProvider struct {
	pathBase string
	mw       *samlsp.Middleware
}

func NewAuthSessionProvider(ctx context.Context) (*AuthSessionProvider, error) {
	p := &AuthSessionProvider{
		pathBase: "/auth/saml/",
	}

	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	if mwErr := p.createSamlMiddleware(ctx, cfg); mwErr != nil {
		return nil, fmt.Errorf("failed to create saml middleware: %w", mwErr)
	}

	return p, nil
}

func loadConfig() (Config, error) {
	var cfg Config

	for _, v := range []string{IdPMetadataUrlEnv, CertFileEnv, CertKeyFileEnv} {
		if os.Getenv(v) == "" {
			return cfg, fmt.Errorf("missing environment variable: %s", v)
		}
	}

	cfg.IdPMetadataUrl = os.Getenv(IdPMetadataUrlEnv)
	cfg.CertFile = os.Getenv(CertFileEnv)
	cfg.CertKeyFile = os.Getenv(CertKeyFileEnv)

	return cfg, nil
}

func (p *AuthSessionProvider) Name() string {
	return "saml"
}

func (p *AuthSessionProvider) createSamlMiddleware(ctx context.Context, cfg Config) error {
	appUrl, appUrlErr := url.Parse(rez.BackendUrl)
	if appUrlErr != nil {
		return fmt.Errorf("bad app url: %w", appUrlErr)
	}

	keyPair, pairErr := tls.LoadX509KeyPair(cfg.CertFile, cfg.CertKeyFile)
	if pairErr != nil {
		return fmt.Errorf("failed to load keypair: %w", pairErr)
	}
	cert, certErr := x509.ParseCertificate(keyPair.Certificate[0])
	if certErr != nil {
		return fmt.Errorf("failed to parse keypair cert: %w", certErr)
	}
	keyPair.Leaf = cert

	privateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("failed to cast *rsa.PrivateKey")
	}

	idpMetadataURL, idpUrlErr := url.Parse(cfg.IdPMetadataUrl)
	if idpUrlErr != nil {
		return fmt.Errorf("bad idp url: %w", idpUrlErr)
	}

	idpMetadata, metadataErr := samlsp.FetchMetadata(ctx, http.DefaultClient, *idpMetadataURL)
	if metadataErr != nil {
		return fmt.Errorf("fetching idp metadata: %w", metadataErr)
	}

	opts := samlsp.Options{
		URL:                *appUrl,
		Key:                privateKey,
		Certificate:        cert,
		CookieName:         sessionCookieName,
		IDPMetadata:        idpMetadata,
		SignRequest:        true,
		DefaultRedirectURI: rez.FrontendUrl,
	}

	mw, mwErr := samlsp.New(opts)
	if mwErr != nil {
		return fmt.Errorf("samlsp.New: %w", mwErr)
	}

	mw.OnError = p.handleSamlError

	// fix redirect urls
	resolveMountedPath := func(path string) url.URL {
		return *appUrl.ResolveReference(&url.URL{Path: p.pathBase + path})
	}
	mw.ServiceProvider.SloURL = resolveMountedPath("slo")
	mw.ServiceProvider.AcsURL = resolveMountedPath("acs")
	mw.ServiceProvider.AcsURL = resolveMountedPath("metadata")

	p.mw = mw

	return nil
}

func (p *AuthSessionProvider) handleSamlError(w http.ResponseWriter, _ *http.Request, err error) {
	var parseErr *saml.InvalidResponseError
	if errors.As(err, &parseErr) {
		log.Printf("WARNING: received invalid saml response: %s (now: %s) %s",
			parseErr.Response, parseErr.Now, parseErr.PrivateErr)
	}
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func (p *AuthSessionProvider) StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	p.mw.HandleStartAuthFlow(w, r)
}

func (p *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated rez.AuthSessionCreatedFn) bool {
	if r.URL.Path == p.mw.ServiceProvider.MetadataURL.Path {
		p.mw.ServeMetadata(w, r)
		return true
	}

	if r.URL.Path == p.mw.ServiceProvider.AcsURL.Path {
		if acsErr := p.handleServeACS(w, r, onCreated); acsErr != nil {
			log.Error().Err(acsErr).Msgf("failed to handle serve acs")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return true
	}

	return false
}

func (p *AuthSessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) {
	if delErr := p.mw.Session.DeleteSession(w, r); delErr != nil {
		log.Error().Err(delErr).Msgf("failed to delete saml session")
	}
}

// mostly taken from samlsp
func (p *AuthSessionProvider) handleServeACS(w http.ResponseWriter, r *http.Request, onCreated rez.AuthSessionCreatedFn) error {
	if parseErr := r.ParseForm(); parseErr != nil {
		return fmt.Errorf("parse form: %w", parseErr)
	}

	rt := p.mw.RequestTracker
	sp := p.mw.ServiceProvider

	possibleRequestIDs := []string{}
	if sp.AllowIDPInitiated {
		possibleRequestIDs = append(possibleRequestIDs, "")
	}

	trackedRequests := rt.GetTrackedRequests(r)
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}

	assertion, parseAssnErr := sp.ParseResponse(r, possibleRequestIDs)
	if parseAssnErr != nil {
		return fmt.Errorf("failed to parse assertion response: %w", parseAssnErr)
	}

	redirectUri := sp.DefaultRedirectURI

	if trackedRequestIndex := r.Form.Get("RelayState"); trackedRequestIndex != "" {
		trackedRequest, trackErr := rt.GetTrackedRequest(r, trackedRequestIndex)
		if trackErr != nil {
			if !(errors.Is(trackErr, http.ErrNoCookie) && sp.AllowIDPInitiated) {
				return fmt.Errorf("failed to get tracked request: %w", trackErr)
			}
			if uri := r.Form.Get("RelayState"); uri != "" {
				redirectUri = uri
			}
		} else {
			if stopErr := rt.StopTrackingRequest(w, r, trackedRequestIndex); stopErr != nil {
				return fmt.Errorf("failed to stop tracking request: %w", stopErr)
			}
			redirectUri = trackedRequest.URI
		}
	}

	if createErr := p.mw.Session.CreateSession(w, r, assertion); createErr != nil {
		return fmt.Errorf("failed to create session: %w", createErr)
	}

	cs, sessErr := p.createSession(assertion)
	if sessErr != nil {
		return fmt.Errorf("failed to convert assertion to auth session: %w", sessErr)
	}

	onCreated(cs.user, cs.expiresAt, redirectUri)

	return nil
}

type createdSession struct {
	user      *ent.User
	expiresAt time.Time
}

func (p *AuthSessionProvider) createSession(a *saml.Assertion) (*createdSession, error) {
	sp, spOk := p.mw.Session.(samlsp.CookieSessionProvider)
	if !spOk {
		return nil, fmt.Errorf("failed to get cookie session provider")
	}

	sess, sessErr := sp.Codec.New(a)
	if sessErr != nil {
		return nil, fmt.Errorf("failed to create session: %w", sessErr)
	}

	sa, saOk := sess.(samlsp.SessionWithAttributes)
	if !saOk {
		return nil, fmt.Errorf("saml: session does not implement samlsp.SessionWithAttributes")
	}

	claims, claimsOk := sess.(samlsp.JWTSessionClaims)
	if !claimsOk {
		return nil, fmt.Errorf("session does not implement samlsp.JWTSessionClaims")
	}

	attr := sa.GetAttributes()
	cs := &createdSession{
		user: &ent.User{
			Name:  attr.Get("firstName"),
			Email: attr.Get("email"),
		},
		expiresAt: time.Unix(claims.ExpiresAt, 0),
	}

	return cs, nil
}

func (p *AuthSessionProvider) GetUserMapping() *ent.User {
	return userMapping
}

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
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	sessionCookieName = "saml_session"
)

var (
	userMapping = &ent.User{
		Name:  "y",
		Email: "y",
	}
)

type Config struct {
	IdPMetadataUrl string `json:"idp_metadata_url"`
	CertFile       string `json:"cert_file"`
	CertKeyFile    string `json:"cert_key_file"`
}

type SessionProvider struct {
	appUrl *url.URL
	mw     *samlsp.Middleware
}

func NewAuthSessionProvider(ctx context.Context, cfg Config) (*SessionProvider, error) {
	appUrl, appUrlErr := url.Parse(rez.BackendUrl)
	if appUrlErr != nil {
		return nil, fmt.Errorf("bad app url: %w", appUrlErr)
	}

	p := &SessionProvider{
		appUrl: appUrl,
	}

	mw, mwErr := p.createSamlMiddleware(ctx, cfg)
	if mwErr != nil {
		return nil, fmt.Errorf("failed to create saml middleware: %w", mwErr)
	}

	return &SessionProvider{
		mw: mw,
	}, nil
}

func (p *SessionProvider) createSamlMiddleware(ctx context.Context, cfg Config) (*samlsp.Middleware, error) {
	cert, kpErr := loadCert(cfg.CertFile, cfg.CertKeyFile)
	if kpErr != nil {
		return nil, fmt.Errorf("cert error: %w", kpErr)
	}

	privateKey, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast *rsa.PrivateKey")
	}

	idpMetadataURL, idpUrlErr := url.Parse(cfg.IdPMetadataUrl)
	if idpUrlErr != nil {
		return nil, fmt.Errorf("bad idp url: %w", idpUrlErr)
	}

	idpMetadata, metadataErr := samlsp.FetchMetadata(ctx, http.DefaultClient, *idpMetadataURL)
	if metadataErr != nil {
		return nil, fmt.Errorf("fetching idp metadata: %w", metadataErr)
	}

	opts := samlsp.Options{
		URL:                *p.appUrl,
		Key:                privateKey,
		Certificate:        cert.Leaf,
		CookieName:         sessionCookieName,
		IDPMetadata:        idpMetadata,
		SignRequest:        true,
		DefaultRedirectURI: rez.FrontendUrl,
	}

	mw, mwErr := samlsp.New(opts)
	if mwErr != nil {
		return nil, fmt.Errorf("samlsp.New: %w", mwErr)
	}

	// fix redirect urls
	resolveMountedPath := func(path string) url.URL {
		return *(p.appUrl.ResolveReference(&url.URL{Path: "/auth" + path}))
	}
	mw.ServiceProvider.SloURL = resolveMountedPath("/saml/slo")
	mw.ServiceProvider.AcsURL = resolveMountedPath("/saml/acs")
	mw.ServiceProvider.AcsURL = resolveMountedPath("/saml/metadata")

	mw.OnError = func(w http.ResponseWriter, _ *http.Request, err error) {
		var parseErr *saml.InvalidResponseError
		if errors.As(err, &parseErr) {
			log.Printf("WARNING: received invalid saml response: %s (now: %s) %s",
				parseErr.Response, parseErr.Now, parseErr.PrivateErr)
		}
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}

	return mw, nil
}

func loadCert(certFile, keyFile string) (*tls.Certificate, error) {
	keyPair, pairErr := tls.LoadX509KeyPair(certFile, keyFile)
	if pairErr != nil {
		return nil, fmt.Errorf("failed to load keypair: %w", pairErr)
	}
	parsedCert, parseErr := x509.ParseCertificate(keyPair.Certificate[0])
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse keypair cert: %w", parseErr)
	}
	keyPair.Leaf = parsedCert
	return &keyPair, nil
}

func (p *SessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated rez.AuthSessionCreatedFn) bool {
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

func (p *SessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) {
	// TODO: logout
}

// mostly taken from samlsp
func (p *SessionProvider) handleServeACS(w http.ResponseWriter, r *http.Request, onCreated rez.AuthSessionCreatedFn) error {
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

	user, expiresAt, sessErr := p.createSessionFromAssertion(assertion)
	if sessErr != nil {
		return fmt.Errorf("failed to convert assertion to auth session: %w", sessErr)
	}

	onCreated(user, expiresAt, redirectUri)

	return nil
}

func (p *SessionProvider) StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	p.mw.HandleStartAuthFlow(w, r)
}

func (p *SessionProvider) createSessionFromAssertion(a *saml.Assertion) (*ent.User, time.Time, error) {
	var expiresAt time.Time
	sp, ok := p.mw.Session.(samlsp.CookieSessionProvider)
	if !ok {
		return nil, expiresAt, fmt.Errorf("failed to get cookie session provider")
	}

	sess, sessErr := sp.Codec.New(a)
	if sessErr != nil {
		return nil, expiresAt, fmt.Errorf("failed to create session: %w", sessErr)
	}

	sa, ok := sess.(samlsp.SessionWithAttributes)
	if !ok {
		return nil, expiresAt, fmt.Errorf("saml: session does not implement samlsp.SessionWithAttributes")
	}

	claims, claimsOk := sess.(samlsp.JWTSessionClaims)
	if !claimsOk {
		return nil, expiresAt, fmt.Errorf("session does not implement samlsp.JWTSessionClaims")
	}

	attr := sa.GetAttributes()
	user := &ent.User{
		Name:  attr.Get("firstName"),
		Email: attr.Get("email"),
	}
	expiresAt = time.Unix(claims.ExpiresAt, 0)

	return user, expiresAt, nil
}

func (p *SessionProvider) GetUserMapping() *ent.User {
	return userMapping
}

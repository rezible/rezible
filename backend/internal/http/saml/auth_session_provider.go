package saml

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

type AuthSessionProviderConfig struct {
	idpMetadata *saml.EntityDescriptor
	appUrl      url.URL
	pathBase    string
	keyPair     *tls.Certificate
	cert        *x509.Certificate
	key         crypto.Signer
}

func (c *AuthSessionProviderConfig) resolveMountedPath(path string) (*url.URL, error) {
	parsed, parseErr := url.Parse(c.pathBase + path)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse: %w", parseErr)
	}
	return c.appUrl.ResolveReference(parsed), nil
}

func loadEnvConfig(ctx context.Context) (*AuthSessionProviderConfig, error) {
	idPMetadataUrl := rez.Config.GetString("auth.saml.idp_metadata_url")
	idpMetadata, mdErr := fetchIdpMetadata(ctx, idPMetadataUrl)
	if mdErr != nil {
		return nil, fmt.Errorf("fetching idp metadata: %w", mdErr)
	}

	certFile := rez.Config.GetString("auth.saml.cert_file")
	certKeyFile := rez.Config.GetString("auth.saml.cert_key_file")

	keyPair, pairErr := tls.LoadX509KeyPair(certFile, certKeyFile)
	if pairErr != nil {
		return nil, fmt.Errorf("failed to load keypair: %w", pairErr)
	}

	cert, certErr := x509.ParseCertificate(keyPair.Certificate[0])
	if certErr != nil {
		return nil, fmt.Errorf("failed to parse keypair cert: %w", certErr)
	}
	keyPair.Leaf = cert

	privateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast *rsa.PrivateKey")
	}

	appUrl, appUrlErr := url.Parse(rez.Config.AppUrl())
	if appUrlErr != nil {
		return nil, fmt.Errorf("failed to parse backend url: %w", appUrlErr)
	}

	pathBase, baseErr := url.JoinPath(rez.Config.ApiRouteBase(), rez.Config.AuthRouteBase(), "/saml/")
	if baseErr != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", baseErr)
	}

	cfg := &AuthSessionProviderConfig{
		appUrl:      *appUrl,
		pathBase:    pathBase,
		idpMetadata: idpMetadata,
		keyPair:     &keyPair,
		cert:        cert,
		key:         privateKey,
	}

	return cfg, nil
}

func fetchIdpMetadata(ctx context.Context, metadataUrl string) (*saml.EntityDescriptor, error) {
	idpMetadataURL, idpUrlErr := url.Parse(metadataUrl)
	if idpUrlErr != nil {
		return nil, fmt.Errorf("bad idp url: %w", idpUrlErr)
	}

	// TODO: use samlsp.ParseMetadata with a cache
	idpMetadata, metadataErr := samlsp.FetchMetadata(ctx, http.DefaultClient, *idpMetadataURL)
	if metadataErr != nil {
		return nil, fmt.Errorf("fetching idp metadata: %w", metadataErr)
	}

	return idpMetadata, nil
}

type AuthSessionProvider struct {
	displayName string
	providerId  string
	mw          *samlsp.Middleware
}

func NewAuthSessionProvider(ctx context.Context) (*AuthSessionProvider, error) {
	cfg, cfgErr := loadEnvConfig(ctx)
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	p := &AuthSessionProvider{
		// TODO: get these from config
		displayName: "SAML",
		providerId:  "saml",
	}

	mw, mwErr := p.createSamlMiddleware(cfg)
	if mwErr != nil {
		return nil, fmt.Errorf("failed to create saml middleware: %w", mwErr)
	}
	p.mw = mw

	return p, nil
}

func (p *AuthSessionProvider) DisplayName() string {
	return p.displayName
}

func (p *AuthSessionProvider) Id() string {
	return p.providerId
}

func (p *AuthSessionProvider) UserMapping() *ent.User {
	return userMapping
}

func (p *AuthSessionProvider) createSamlMiddleware(cfg *AuthSessionProviderConfig) (*samlsp.Middleware, error) {
	opts := samlsp.Options{
		URL:                cfg.appUrl,
		Key:                cfg.key,
		Certificate:        cfg.cert,
		CookieName:         sessionCookieName,
		IDPMetadata:        cfg.idpMetadata,
		SignRequest:        true,
		DefaultRedirectURI: rez.Config.AppUrl(),
	}
	if !rez.Config.DebugMode() {
		opts.CookieSameSite = http.SameSiteLaxMode
	}

	mw, mwErr := samlsp.New(opts)
	if mwErr != nil {
		return nil, fmt.Errorf("samlsp.New: %w", mwErr)
	}

	mw.OnError = p.onError

	sloUrl, sloUrlErr := cfg.resolveMountedPath("slo")
	if sloUrlErr != nil {
		return nil, fmt.Errorf("slo path: %w", sloUrlErr)
	}
	mw.ServiceProvider.SloURL = *sloUrl

	acsUrl, acsUrlErr := cfg.resolveMountedPath("acs")
	if acsUrlErr != nil {
		return nil, fmt.Errorf("acs path: %w", acsUrlErr)
	}
	mw.ServiceProvider.AcsURL = *acsUrl

	metadataUrl, metadataUrlErr := cfg.resolveMountedPath("metadata")
	if metadataUrlErr != nil {
		return nil, fmt.Errorf("metadata path: %w", metadataUrlErr)
	}
	mw.ServiceProvider.MetadataURL = *metadataUrl

	return mw, nil
}

func (p *AuthSessionProvider) onError(w http.ResponseWriter, r *http.Request, err error) {
	if sessErr := p.ClearSession(w, r); sessErr != nil {
		log.Error().Err(sessErr).Msg("failed to clear session")
	}
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

func (p *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated func(session rez.AuthProviderSession)) bool {
	if r.URL.Path == p.mw.ServiceProvider.MetadataURL.Path {
		p.mw.ServeMetadata(w, r)
		return true
	}

	if r.URL.Path == p.mw.ServiceProvider.AcsURL.Path {
		if sessErr := p.handleServeACS(w, r, onCreated); sessErr != nil {
			p.onError(w, r, sessErr)
		}
		return true
	}

	return false
}

func (p *AuthSessionProvider) SessionExists(r *http.Request) bool {
	sess, sessErr := p.mw.Session.GetSession(r)
	return sess != nil && sessErr == nil
}

func (p *AuthSessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) error {
	return p.mw.Session.DeleteSession(w, r)
}

// mostly taken from samlsp
func (p *AuthSessionProvider) handleServeACS(w http.ResponseWriter, r *http.Request, onCreated func(session rez.AuthProviderSession)) error {
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

	cs, sessErr := p.createSession(assertion, redirectUri)
	if sessErr != nil {
		return fmt.Errorf("failed to convert assertion to auth session: %w", sessErr)
	}

	onCreated(*cs)

	return nil
}

func (p *AuthSessionProvider) createSession(a *saml.Assertion, redirectUrl string) (*rez.AuthProviderSession, error) {
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

	if verifyErr := p.verifyClaims(claims); verifyErr != nil {
		return nil, fmt.Errorf("failed to verify claims: %w", verifyErr)
	}

	attr := sa.GetAttributes()

	email := attr.Get("email")
	domain := strings.Split(email, "@")[1]
	id := attr.Get("id")

	po := ent.Organization{
		ProviderID: id,
		Name:       domain,
	}

	pu := ent.User{
		ProviderID: claims.Subject,
		Name:       attr.Get("firstName"),
		Email:      email,
	}

	ps := &rez.AuthProviderSession{
		Organization: po,
		User:         pu,
		ExpiresAt:    time.Unix(claims.ExpiresAt, 0),
		RedirectUrl:  redirectUrl,
	}

	return ps, nil
}

func (p *AuthSessionProvider) verifyClaims(claims samlsp.JWTSessionClaims) error {
	appUrl := rez.Config.AppUrl()
	if !claims.VerifyAudience(appUrl, true) {
		return fmt.Errorf("audience '%s'", claims.Audience)
	}

	if !claims.VerifyIssuer(appUrl, true) {
		return fmt.Errorf("issuer '%s'", claims.Issuer)
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return fmt.Errorf("expiry '%v'", claims.ExpiresAt)
	}

	if !claims.VerifyNotBefore(time.Now().Unix(), true) {
		return fmt.Errorf("expiry '%v'", claims.ExpiresAt)
	}

	return nil
}

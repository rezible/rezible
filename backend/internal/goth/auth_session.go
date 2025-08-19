package goth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func ConfigureSessionStore(secretKey string) {
	maxAge := 86400 * 30 // 30 days

	store := sessions.NewCookieStore([]byte(secretKey))
	store.MaxAge(maxAge)
	// store.Options.Domain = rez.BackendUrl
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = !rez.DebugMode

	gothic.Store = store
}

type AuthSessionProvider struct {
	callbackPath string
	provider     goth.Provider
}

type Config struct {
	PathBase string

	Provider     string `json:"provider"`
	ClientKey    string `json:"client_key"`
	ClientSecret string `json:"client_secret"`
	DiscoveryUrl string `json:"oidc_discovery_url,omitempty"`
}

func NewAuthSessionProvider(cfg Config) (*AuthSessionProvider, error) {
	callbackPath := fmt.Sprintf("%s/%s", cfg.PathBase, "callback")
	p, provErr := registerProvider(cfg, callbackPath)
	if provErr != nil {
		return nil, provErr
	}

	return &AuthSessionProvider{
		callbackPath: callbackPath,
		provider:     p,
	}, nil
}

func registerProvider(cfg Config, callbackPath string) (goth.Provider, error) {
	provName := strings.ToLower(cfg.Provider)
	callbackUrl, urlErr := url.JoinPath(rez.BackendUrl, callbackPath)
	if urlErr != nil {
		return nil, fmt.Errorf("creating callback url: %w", urlErr)
	}

	scopes := []string{"user:email"}
	switch provName {
	case "github":
		return github.New(cfg.ClientKey, cfg.ClientSecret, callbackUrl, scopes...), nil
	}
	return nil, fmt.Errorf("unknown provider: %s", provName)
}

func (s *AuthSessionProvider) Name() string {
	return s.provider.Name()
}

func (s *AuthSessionProvider) getProviderSession(r *http.Request) (goth.Session, error) {
	marshalledSess, getErr := gothic.GetFromSession(s.provider.Name(), r)
	if getErr != nil {
		return nil, fmt.Errorf("gothic.GetFromSession: %w", getErr)
	}

	return s.provider.UnmarshalSession(marshalledSess)
}

func (s *AuthSessionProvider) setProviderSession(w http.ResponseWriter, r *http.Request, sess goth.Session) error {
	return gothic.StoreInSession(s.provider.Name(), sess.Marshal(), r, w)
}

func (s *AuthSessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) {
	if logoutErr := gothic.Logout(w, r); logoutErr != nil {
		log.Error().Err(logoutErr).Msg("logout failed")
	}
}

var userMapping = &ent.User{
	Name:  "y",
	Email: "y",
}

func (s *AuthSessionProvider) GetUserMapping() *ent.User {
	return userMapping
}

func (s *AuthSessionProvider) StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	redirect, redirectErr := s.createProviderSessionRedirect(w, r)
	if redirectErr != nil {
		log.Error().Err(redirectErr).Msg("could not create provider session redirect")
		http.Error(w, redirectErr.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func (s *AuthSessionProvider) createProviderSessionRedirect(w http.ResponseWriter, r *http.Request) (string, error) {
	state := r.URL.Query().Get("state")
	if state == "" {
		nonceBytes := make([]byte, 64)
		_, err := io.ReadFull(rand.Reader, nonceBytes)
		if err != nil {
			return "", fmt.Errorf("could not generate session state nonce: %w", err)
		}
		state = base64.URLEncoding.EncodeToString(nonceBytes)
	}

	sess, sessErr := s.provider.BeginAuth(state)
	if sessErr != nil {
		return "", fmt.Errorf("provider begin auth: %w", sessErr)
	}
	authUrl, urlErr := sess.GetAuthURL()
	if urlErr != nil {
		return "", fmt.Errorf("getting auth url: %w", urlErr)
	}
	if storeErr := s.setProviderSession(w, r, sess); storeErr != nil {
		return "", fmt.Errorf("storing provider session: %w", storeErr)
	}
	return authUrl, nil
}

func (s *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, cs rez.AuthSessionCreatedFn) bool {
	if r.URL.Path == s.callbackPath {
		cbErr := s.handleFlowCallback(w, r, cs)
		if cbErr == nil {
			return true
		}
		s.ClearSession(w, r)
		log.Error().Err(cbErr).Msg("could not handle oauth2 callback")
		return false
	}
	return false
}

func (s *AuthSessionProvider) handleFlowCallback(w http.ResponseWriter, r *http.Request, onCreated rez.AuthSessionCreatedFn) error {
	sess, sessErr := s.getProviderSession(r)
	if sessErr != nil {
		return fmt.Errorf("getting provider session: %w", sessErr)
	}

	if validateErr := s.validateRequestSessionState(r, sess); validateErr != nil {
		return fmt.Errorf("validating request session: %w", validateErr)
	}

	sessUser, fetchErr := s.fetchSessionUser(w, r, sess)
	if fetchErr != nil {
		return fmt.Errorf("fetching user: %w", fetchErr)
	}

	if sessUser.Email == "" {
		return errors.New("missing email")
	}

	user := &ent.User{
		Name:  sessUser.Name,
		Email: sessUser.Email,
	}

	onCreated(user, sessUser.ExpiresAt, rez.FrontendUrl)

	return nil
}

func (s *AuthSessionProvider) validateRequestSessionState(r *http.Request, sess goth.Session) error {
	rawAuthURL, urlErr := sess.GetAuthURL()
	if urlErr != nil {
		return fmt.Errorf("get auth url: %w", urlErr)
	}

	authURL, authUrlErr := url.Parse(rawAuthURL)
	if authUrlErr != nil {
		return fmt.Errorf("parse auth url: %w", authUrlErr)
	}

	urlState := authURL.Query().Get("state")
	if urlState == "" {
		return fmt.Errorf("missing state parameter")
	}

	params := r.URL.Query()
	reqState := params.Get("state")
	if reqState == "" && r.Method == http.MethodPost {
		reqState = r.FormValue("state")
	}

	if urlState != reqState {
		return errors.New("state mismatch")
	}
	return nil
}

func (s *AuthSessionProvider) fetchSessionUser(w http.ResponseWriter, r *http.Request, sess goth.Session) (*goth.User, error) {
	user, fetchErr := s.provider.FetchUser(sess)
	if fetchErr == nil {
		// user found with existing session data
		return &user, nil
	}

	params, paramsErr := getRequestParams(r)
	if paramsErr != nil {
		return nil, fmt.Errorf("getting request params: %w", paramsErr)
	}

	_, authErr := sess.Authorize(s.provider, params)
	if authErr != nil {
		return nil, fmt.Errorf("authorizing session: %w", authErr)
	}

	if storeErr := s.setProviderSession(w, r, sess); storeErr != nil {
		return nil, fmt.Errorf("failed to store provider session: %w", storeErr)
	}

	user, fetchErr = s.provider.FetchUser(sess)
	if fetchErr != nil {
		return nil, fmt.Errorf("fetching user after auth: %w", fetchErr)
	}
	return &user, nil
}

func getRequestParams(r *http.Request) (url.Values, error) {
	params := r.URL.Query()
	if params.Encode() == "" && r.Method == http.MethodPost {
		if parseErr := r.ParseForm(); parseErr != nil {
			return nil, fmt.Errorf("parsing request form: %w", parseErr)
		}
		params = r.Form
	}
	return params, nil
}

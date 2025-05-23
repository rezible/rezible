package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"

	rez "github.com/rezible/rezible"
)

// TODO: implement this properly

var (
	userMapping = &ent.User{
		Name:  "y",
		Email: "y",
	}
)

type AuthSessionProvider struct {
	provider goth.Provider
	handler  *chi.Mux
}

type Config struct {
	Provider     string `json:"provider"`
	ClientKey    string `json:"client_key"`
	ClientSecret string `json:"client_secret"`
	DiscoveryUrl string `json:"oidc_discovery_url,omitempty"`
}

func NewAuthSessionProvider(cfg Config) (*AuthSessionProvider, error) {
	p, provErr := registerProvider(cfg)
	if provErr != nil {
		return nil, provErr
	}

	return &AuthSessionProvider{
		provider: p,
	}, nil
}

func registerProvider(cfg Config) (goth.Provider, error) {
	provName := strings.ToLower(cfg.Provider)
	callbackUrl := fmt.Sprintf("%s/auth/callback", rez.BackendUrl)
	var provider goth.Provider
	scopes := []string{"user:email"}
	switch provName {
	case "github":
		provider = github.New(cfg.ClientKey, cfg.ClientSecret, callbackUrl, scopes...)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provName)
	}
	return provider, nil
}

func (s *AuthSessionProvider) Name() string {
	return s.provider.Name()
}

func (s *AuthSessionProvider) GetUserMapping() *ent.User {
	return userMapping
}

func (s *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, cs rez.AuthSessionCreatedFn) bool {
	if r.URL.Path == "/auth/callback" {
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

func (s *AuthSessionProvider) StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	redirectUrl, redirectErr := s.createProviderSessionRedirect(w, r)
	if redirectErr != nil {
		log.Error().Err(redirectErr).Msg("could not create provider session redirect")
		http.Error(w, redirectErr.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}

func (s *AuthSessionProvider) handleFlowCallback(w http.ResponseWriter, r *http.Request, onCreated rez.AuthSessionCreatedFn) error {
	sess, sessErr := s.getProviderSession(r)
	if sessErr != nil {
		return fmt.Errorf("getting provider session: %w", sessErr)
	}

	if validateErr := validateRequestSessionState(r, sess); validateErr != nil {
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
		Email: sessUser.Email,
	}

	onCreated(user, sessUser.ExpiresAt, rez.FrontendUrl)
	return nil
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
	if storeErr := s.storeProviderSession(w, r, sess); storeErr != nil {
		return "", fmt.Errorf("storing provider session: %w", storeErr)
	}
	return authUrl, nil
}

func (s *AuthSessionProvider) storeProviderSession(w http.ResponseWriter, r *http.Request, sess goth.Session) error {
	// TODO: dont use gothic, manage session store internally
	return gothic.StoreInSession(s.provider.Name(), sess.Marshal(), r, w)
}

func (s *AuthSessionProvider) getProviderSession(r *http.Request) (goth.Session, error) {
	// TODO: dont use gothic, manage session store internally
	marshalledSess, getErr := gothic.GetFromSession(s.provider.Name(), r)
	if getErr != nil {
		return nil, fmt.Errorf("get from session: %w", getErr)
	}

	sess, unmarshalErr := s.provider.UnmarshalSession(marshalledSess)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshalling session: %w", unmarshalErr)
	}

	return sess, nil
}

func validateRequestSessionState(r *http.Request, sess goth.Session) error {
	rawAuthURL, urlErr := sess.GetAuthURL()
	if urlErr != nil {
		return fmt.Errorf("get auth url: %w", urlErr)
	}

	authURL, authUrlErr := url.Parse(rawAuthURL)
	if authUrlErr != nil {
		return fmt.Errorf("parse auth url: %w", authUrlErr)
	}
	originalState := authURL.Query().Get("state")
	if originalState == "" {
		return fmt.Errorf("missing state parameter")
	}

	params := r.URL.Query()
	reqState := params.Get("state")
	if reqState == "" && r.Method == http.MethodPost {
		reqState = r.FormValue("state")
	}

	if originalState != reqState {
		return errors.New("state mismatch")
	}
	return nil
}

func (s *AuthSessionProvider) fetchSessionUser(w http.ResponseWriter, r *http.Request, sess goth.Session) (*goth.User, error) {
	user, fetchErr := s.provider.FetchUser(sess)
	if fetchErr == nil {
		// user can be found with existing session data
		return &user, nil
	}

	if updateErr := s.updateSessionAuth(w, r, sess); updateErr != nil {
		return nil, fmt.Errorf("updating session authz: %w", updateErr)
	}

	user, fetchErr = s.provider.FetchUser(sess)
	if fetchErr != nil {
		return nil, fmt.Errorf("fetching user after auth: %w", fetchErr)
	}
	return &user, nil
}

func (s *AuthSessionProvider) updateSessionAuth(w http.ResponseWriter, r *http.Request, sess goth.Session) error {
	params := r.URL.Query()
	if params.Encode() == "" && r.Method == http.MethodPost {
		if parseErr := r.ParseForm(); parseErr != nil {
			return fmt.Errorf("could not parse request form: %w", parseErr)
		}
		params = r.Form
	}

	_, authErr := sess.Authorize(s.provider, params)
	if authErr != nil {
		return fmt.Errorf("failed to authorize: %w", authErr)
	}

	if storeErr := s.storeProviderSession(w, r, sess); storeErr != nil {
		return fmt.Errorf("failed to store provider session: %w", storeErr)
	}

	return nil
}

func (s *AuthSessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) {
	if logoutErr := gothic.Logout(w, r); logoutErr != nil {
		log.Error().Err(logoutErr).Msg("logout failed")
	}
}

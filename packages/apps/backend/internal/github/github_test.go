package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/testkit/mocks"
)

func TestValidateConfig_MissingCredentials(t *testing.T) {
	cases := []map[string]any{
		{},
		{"app": map[string]any{}},
		{"app": map[string]any{"app_id": float64(123)}},
		{"app": map[string]any{"app_id": float64(123), "client_id": "cid"}},
		{"app": map[string]any{"app_id": float64(123), "client_id": "cid", "client_secret": "cs"}},
	}
	intg := &integration{}
	for _, cfg := range cases {
		err := intg.ValidateConfig(cfg)
		assert.Error(t, err, "expected error for config: %v", cfg)
	}
}

func TestValidateConfig_ValidAppCredentials(t *testing.T) {
	intg := &integration{}
	cfg := map[string]any{
		"app": map[string]any{
			"app_id":          float64(123),
			"client_id":       "client-id",
			"client_secret":   "client-secret",
			"private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----",
		},
	}
	require.NoError(t, intg.ValidateConfig(cfg))
}

func TestOAuth2Config(t *testing.T) {
	intg := &integration{cfg: Config{
		App: struct {
			AppID         int64  `koanf:"app_id"`
			ClientID      string `koanf:"client_id"`
			ClientSecret  string `koanf:"client_secret"`
			PrivateKeyPEM string `koanf:"private_key_pem"`
		}{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
	}}
	intg.oauth2Config = intg.loadOAuthConfig()

	cfg := intg.OAuth2Config()
	require.NotNil(t, cfg)
	assert.Equal(t, "client-id", cfg.ClientID)
	assert.Equal(t, "client-secret", cfg.ClientSecret)
	assert.Equal(t, "https://github.com/login/oauth/authorize", cfg.Endpoint.AuthURL)
	assert.Equal(t, "https://github.com/login/oauth/access_token", cfg.Endpoint.TokenURL)

	authURL := cfg.AuthCodeURL("state-value")
	assert.Contains(t, authURL, "client_id=client-id")
	assert.Contains(t, authURL, "state=state-value")
}

func TestExtractIntegrationOptionsFromToken(t *testing.T) {
	intg := &integration{
		cfg: Config{
			App: struct {
				AppID         int64  `koanf:"app_id"`
				ClientID      string `koanf:"client_id"`
				ClientSecret  string `koanf:"client_secret"`
				PrivateKeyPEM string `koanf:"private_key_pem"`
			}{AppID: 123},
		},
		listUserInstallations: func(_ context.Context, token string) ([]*github.Installation, error) {
			assert.Equal(t, "access-token", token)
			return []*github.Installation{
				{
					ID:      github.Ptr[int64](456),
					AppID:   github.Ptr[int64](123),
					Account: &github.User{Login: github.Ptr("myorg")},
				},
			}, nil
		},
	}

	options, err := intg.ExtractIntegrationOptionsFromToken(&oauth2.Token{AccessToken: "access-token"})

	require.NoError(t, err)
	require.Len(t, options, 1)
	assert.Equal(t, "456", options[0].ExternalRef)
	assert.Equal(t, "myorg", options[0].DisplayName)
	assert.Equal(t, "myorg", options[0].Config[configOrg])
	assert.Equal(t, int64(456), options[0].Config[configInstallationID])
}

func TestExtractIntegrationOptionsFromToken_NoInstallations(t *testing.T) {
	intg := &integration{
		listUserInstallations: func(_ context.Context, _ string) ([]*github.Installation, error) {
			return nil, nil
		},
	}

	_, err := intg.ExtractIntegrationOptionsFromToken(&oauth2.Token{AccessToken: "access-token"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid github app installations")
}

func TestExtractIntegrationOptionsFromToken_MultipleInstallations(t *testing.T) {
	intg := &integration{
		listUserInstallations: func(_ context.Context, _ string) ([]*github.Installation, error) {
			return []*github.Installation{
				{ID: github.Ptr[int64](1), Account: &github.User{Login: github.Ptr("org-one")}},
				{ID: github.Ptr[int64](2), Account: &github.User{Login: github.Ptr("org-two")}},
			}, nil
		},
	}

	options, err := intg.ExtractIntegrationOptionsFromToken(&oauth2.Token{AccessToken: "access-token"})

	require.NoError(t, err)
	require.Len(t, options, 2)
	assert.Equal(t, "1", options[0].ExternalRef)
	assert.Equal(t, "2", options[1].ExternalRef)
}

func makePushPayload(t *testing.T, after, fullName, ref, org string, installationID int64) []byte {
	t.Helper()
	ts := github.Timestamp{Time: time.Now()}
	event := &github.PushEvent{
		After: github.Ptr(after),
		Ref:   github.Ptr(ref),
		Repo: &github.PushEventRepository{
			FullName: github.Ptr(fullName),
			Owner:    &github.User{Login: github.Ptr(org)},
		},
		HeadCommit: &github.HeadCommit{
			Timestamp: &ts,
		},
		Installation: &github.Installation{ID: github.Ptr(installationID)},
	}
	b, err := json.Marshal(event)
	require.NoError(t, err)
	return b
}

func makeIntegrationsServiceWithOrg(t *testing.T, org string, tenantID interface{ GetTenantID() interface{} }) *mockIntegrationsService {
	return &mockIntegrationsService{installationID: "123", org: org}
}

// TODO: just use a generated mock from testkit
type mockIntegrationsService struct {
	installationID string
	org            string
}

func (m *mockIntegrationsService) Configure(_ context.Context, _ rez.ConfigureIntegrationParams) (rez.ConfiguredIntegration, error) {
	return nil, nil
}
func (m *mockIntegrationsService) ListConfigured(_ context.Context, params rez.ListIntegrationsParams) ([]rez.ConfiguredIntegration, error) {
	if len(params.ExternalRefs) == 0 || params.ExternalRefs[0] == m.installationID {
		ci := &ConfiguredIntegration{
			intg: &ent.Integration{
				ID:              uuid.New(),
				TenantID:        1,
				Provider:        integrationName,
				DisplayName:     m.org,
				ExternalRef:     m.installationID,
				Config:          map[string]any{"org": m.org, "installation_id": float64(123)},
				UserPreferences: map[string]any{},
			},
		}
		return []rez.ConfiguredIntegration{ci}, nil
	}
	return nil, nil
}
func (m *mockIntegrationsService) GetConfigured(_ context.Context, _ uuid.UUID) (rez.ConfiguredIntegration, error) {
	return nil, nil
}
func (m *mockIntegrationsService) UpdateConfiguredPreferences(_ context.Context, _ uuid.UUID, _ map[string]any) (rez.ConfiguredIntegration, error) {
	return nil, nil
}
func (m *mockIntegrationsService) DeleteConfigured(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockIntegrationsService) StartOAuth2Flow(_ context.Context, _ string, _ *url.URL) (string, error) {
	return "", nil
}
func (m *mockIntegrationsService) CompleteOAuth2Flow(_ context.Context, _ string, _ rez.CompleteIntegrationOAuth2Params) (*rez.CompleteIntegrationOAuth2Result, error) {
	return nil, nil
}
func (m *mockIntegrationsService) SelectOAuth2Flow(_ context.Context, _ string, _ rez.SelectIntegrationOAuth2Params) (*rez.CompleteIntegrationOAuth2Result, error) {
	return nil, nil
}
func (m *mockIntegrationsService) GetChatService(_ context.Context) (rez.ChatService, error) {
	return nil, nil
}
func (m *mockIntegrationsService) GetVideoConferenceService(_ context.Context) (rez.VideoConferenceService, error) {
	return nil, nil
}

func TestPushProcessor_ValidPayload(t *testing.T) {
	const (
		org      = "myorg"
		fullName = "myorg/myrepo"
		after    = "abc123def456abc123def456abc123def456abc1"
		ref      = "refs/heads/main"
	)

	svcs := &rez.Services{
		Integrations: &mockIntegrationsService{org: org, installationID: "123"},
	}
	proc := &pushEventProcessor{services: svcs}

	payload := makePushPayload(t, after, fullName, ref, org, 123)
	events, err := proc.Process(context.Background(), rez.ProviderEvent{
		Provider: integrationName,
		Source:   "push",
		Payload:  payload,
	})

	require.NoError(t, err)
	require.Len(t, events, 1)
	ev := events[0]
	assert.Equal(t, ne.KindChangeEventObserved, ev.Kind)
	assert.NotEmpty(t, ev.SubjectRef)
	assert.NotEmpty(t, ev.ProviderEventRef)
	assert.NotEmpty(t, ev.Attributes)
	assert.Equal(t, after, ev.ProviderEventRef)
	assert.Equal(t, fmt.Sprintf("github:%s:%s", fullName, after), ev.SubjectRef)
}

func TestPushProcessor_BranchDeletion(t *testing.T) {
	svcs := &rez.Services{
		Integrations: &mockIntegrationsService{org: "org", installationID: "123"},
	}
	proc := &pushEventProcessor{services: svcs}

	payload := makePushPayload(t, zeroSHA, "org/repo", "refs/heads/main", "org", 123)
	events, err := proc.Process(context.Background(), rez.ProviderEvent{
		Provider: integrationName,
		Source:   "push",
		Payload:  payload,
	})

	require.NoError(t, err)
	assert.Nil(t, events)
}

func TestPushProcessor_InvalidJSON(t *testing.T) {
	svcs := &rez.Services{
		Integrations: &mockIntegrationsService{org: "org", installationID: "123"},
	}
	proc := &pushEventProcessor{services: svcs}

	_, err := proc.Process(context.Background(), rez.ProviderEvent{
		Provider: integrationName,
		Source:   "push",
		Payload:  []byte(`{invalid json`),
	})

	require.Error(t, err)
}

// --- PR event processor tests ---

func makePRPayload(t *testing.T, prNum int, fullName, title, org string, installationID int64) []byte {
	t.Helper()
	ts := github.Timestamp{Time: time.Now()}
	event := &github.PullRequestEvent{
		PullRequest: &github.PullRequest{
			Number:    github.Ptr(prNum),
			Title:     github.Ptr(title),
			CreatedAt: &ts,
		},
		Repo: &github.Repository{
			FullName: github.Ptr(fullName),
			Owner:    &github.User{Login: github.Ptr(org)},
		},
		Installation: &github.Installation{ID: github.Ptr(installationID)},
	}
	b, err := json.Marshal(event)
	require.NoError(t, err)
	return b
}

func TestPRProcessor_ValidPayload(t *testing.T) {
	const (
		org      = "myorg"
		fullName = "myorg/myrepo"
		prNum    = 42
		title    = "My PR"
	)

	svcs := &rez.Services{
		Integrations: &mockIntegrationsService{org: org, installationID: "123"},
	}
	proc := &pullRequestEventProcessor{services: svcs}

	payload := makePRPayload(t, prNum, fullName, title, org, 123)
	events, err := proc.Process(context.Background(), rez.ProviderEvent{
		Provider: integrationName,
		Source:   "pull_request",
		Payload:  payload,
	})

	require.NoError(t, err)
	require.Len(t, events, 1)
	ev := events[0]
	assert.Equal(t, ne.KindChangeEventObserved, ev.Kind)
	assert.NotEmpty(t, ev.SubjectRef)
	assert.NotEmpty(t, ev.ProviderEventRef)
	assert.NotEmpty(t, ev.Attributes)
	assert.Equal(t, fmt.Sprintf("pr:%d", prNum), ev.ProviderEventRef)
	assert.Equal(t, fmt.Sprintf("github:%s:pr:%d", fullName, prNum), ev.SubjectRef)
}

func TestPRProcessor_InvalidJSON(t *testing.T) {
	svcs := &rez.Services{
		Integrations: &mockIntegrationsService{org: "org", installationID: "123"},
	}
	proc := &pullRequestEventProcessor{services: svcs}

	_, err := proc.Process(context.Background(), rez.ProviderEvent{
		Provider: integrationName,
		Source:   "pull_request",
		Payload:  []byte(`{invalid json`),
	})

	require.Error(t, err)
}

// --- Webhook handler tests ---

func makeWebhookRequest(t *testing.T, body, secret, event, delivery string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", event)
	req.Header.Set("X-GitHub-Delivery", delivery)
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(body))
		req.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	return req
}

func TestWebhookHandler_ValidSignature(t *testing.T) {
	const secret = "test-secret"
	provEvs := mocks.NewMockProviderEventService(t)
	provEvs.On("Ingest", mock.Anything, mock.Anything).Return(nil).Once()

	h := newWebhookHandler(secret, &rez.Services{ProviderEvents: provEvs})
	req := makeWebhookRequest(t, `{"action":"push"}`, secret, "push", "delivery-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	const secret = "test-secret"
	provEvs := mocks.NewMockProviderEventService(t)

	h := newWebhookHandler(secret, &rez.Services{ProviderEvents: provEvs})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"action":"push"}`))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256="+strings.Repeat("0", 64))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWebhookHandler_EmptyBody(t *testing.T) {
	provEvs := mocks.NewMockProviderEventService(t)

	h := newWebhookHandler("", &rez.Services{ProviderEvents: provEvs})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWebhookHandler_CallsIngest(t *testing.T) {
	const (
		body     = `{"action":"push"}`
		event    = "push"
		delivery = "delivery-abc"
	)

	provEvs := mocks.NewMockProviderEventService(t)
	provEvs.On("Ingest", mock.Anything, mock.MatchedBy(func(ev rez.ProviderEvent) bool {
		return ev.Provider == integrationName &&
			ev.Source == event &&
			string(ev.Payload) == body &&
			ev.DedupeKey == delivery
	})).Return(nil).Once()

	h := newWebhookHandler("", &rez.Services{ProviderEvents: provEvs})
	req := makeWebhookRequest(t, body, "", event, delivery)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- IsAvailable tests ---

func TestIsAvailable_Disabled(t *testing.T) {
	i := &integration{cfg: Config{Enabled: false}}
	available, err := i.IsAvailable()
	require.NoError(t, err)
	assert.False(t, available)
}

func TestIsAvailable_Enabled(t *testing.T) {
	i := &integration{cfg: Config{
		Enabled: true,
		App: struct {
			AppID         int64  `koanf:"app_id"`
			ClientID      string `koanf:"client_id"`
			ClientSecret  string `koanf:"client_secret"`
			PrivateKeyPEM string `koanf:"private_key_pem"`
		}{
			AppID:         123,
			ClientID:      "cid",
			ClientSecret:  "cs",
			PrivateKeyPEM: "pem",
		},
	}}
	available, err := i.IsAvailable()
	require.NoError(t, err)
	assert.True(t, available)
}

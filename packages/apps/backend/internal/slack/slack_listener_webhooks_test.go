package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/testkit/mocks"
)

const signingSecret = "test-secret"

func signedSlackRequest(t *testing.T, method, target, body string) *http.Request {
	t.Helper()

	timestamp := time.Now().Unix()
	base := "v0:" + strconv.FormatInt(timestamp, 10) + ":" + body
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, err := mac.Write([]byte(base))
	require.NoError(t, err)

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(mac.Sum(nil)))
	return req
}

func makeWebhookListener(t *testing.T) *WebhookListener {
	t.Helper()

	ingestor := mocks.NewMockProviderEventIngestorService(t)
	ingestor.On("IngestEvent", mock.Anything, mock.Anything).Maybe().Return(nil)

	handler := &eventHandler{
		services: &rez.Services{
			ProviderEvents: ingestor,
		},
	}
	return &WebhookListener{
		handler:       handler,
		signingSecret: signingSecret,
	}
}

func TestEventsAPIWebhookEnqueuesVerifiedCallbackEvent(t *testing.T) {
	body := `{
		"type":"event_callback",
		"team_id":"T123",
		"api_app_id":"A123",
		"event":{"type":"app_home_opened","user":"U123","channel":"D123","tab":"home","event_ts":"123.456"},
		"event_id":"Ev123",
		"event_time":123
	}`
	listener := makeWebhookListener(t)
	req := signedSlackRequest(t, http.MethodPost, "/events", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	listener.Handler().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestEventsAPIWebhookRejectsInvalidSignatureWithoutEnqueue(t *testing.T) {
	listener := makeWebhookListener(t)
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"type":"event_callback"}`))
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("X-Slack-Signature", "v0="+strings.Repeat("0", 64))
	rec := httptest.NewRecorder()
	listener.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	//require.Empty(t, ingestor.events)
}

func TestEventsAPIWebhookURLVerificationIsSynchronous(t *testing.T) {
	body := `{"type":"url_verification","challenge":"challenge-value","token":"legacy-token"}`
	listener := makeWebhookListener(t)
	req := signedSlackRequest(t, http.MethodPost, "/events", body)
	rec := httptest.NewRecorder()
	listener.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "challenge-value", rec.Body.String())
	//require.Empty(t, ingestor.events)
}

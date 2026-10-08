package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type WebhookHandlerSuite struct {
	test.Suite
}

func TestWebhookHandlerSuite(t *testing.T) {
	suite.Run(t, &WebhookHandlerSuite{Suite: test.NewSuite()})
}

// installationLookup reads installations as the production lookup does.
type installationLookup struct {
	db rez.Database
}

func (l installationLookup) LookupInstallation(ctx context.Context, pred predicate.Integration) (*ent.Integration, error) {
	return l.db.Client(ctx).Integration.Query().Where(pred).Only(ctx)
}

func (l installationLookup) ListInstallations(ctx context.Context, preds ...predicate.Integration) ([]*ent.Integration, error) {
	return l.db.Client(ctx).Integration.Query().Where(preds...).All(ctx)
}

type webhookHarness struct {
	handler http.Handler
	events  *mocks.MockProviderEventPipelineService
	logs    *bytes.Buffer
}

func (s *WebhookHandlerSuite) newHarness(tdb rez.Database) *webhookHarness {
	logs := &bytes.Buffer{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	s.T().Cleanup(func() { slog.SetDefault(previous) })
	events := mocks.NewMockProviderEventPipelineService(s.T())
	handler := newWebhookHandler(test.NewClock(testReceivedAt), events, installationLookup{db: tdb})
	return &webhookHarness{handler: handler, events: events, logs: logs}
}

// install creates an installation of the named integration in the context's tenant with a webhook token.
func (s *WebhookHandlerSuite) install(ctx context.Context, tdb rez.Database, name string) (*ent.Integration, string) {
	return s.installWithConfig(ctx, tdb, name, `{"preset":"deployment"}`)
}

func (s *WebhookHandlerSuite) installWithConfig(ctx context.Context, tdb rez.Database, name string, config string) (*ent.Integration, string) {
	token, tokenHash, tokenErr := integrations.NewWebhookToken()
	s.Require().NoError(tokenErr)
	create := tdb.Client(ctx).Integration.Create().
		SetProvider(name).
		SetName(name).
		SetDisplayName(name).
		SetProviderInstallationRef(uuid.NewString()).
		SetInstallationConfig([]byte(config)).
		SetWebhookTokenHash(tokenHash)
	intg, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return intg, token
}

func (h *webhookHarness) post(token string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/"+token, bytes.NewReader(body))
	request = request.WithContext(execution.NewRootContext(request.Context(), execution.KindAnonymous, execution.SourceHTTP))
	recorder := httptest.NewRecorder()
	h.handler.ServeHTTP(recorder, request)
	return recorder
}

func (s *WebhookHandlerSuite) requireTokenNotLogged(h *webhookHarness, token string) {
	s.NotEmpty(h.logs.String(), "the delivery is logged")
	s.NotContains(h.logs.String(), token)
}

func (s *WebhookHandlerSuite) TestDeliveryQueuesTheReportInItsInstallationTenant() {
	_, tdb := s.SetupTestDatabase()
	ownerCtx, _ := s.SetupTestDatabase()
	intg, token := s.install(ownerCtx, tdb, integrationName)
	h := s.newHarness(tdb)

	var queued []rez.ProviderEvent
	h.events.EXPECT().IngestMany(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, events []rez.ProviderEvent) error {
			tenantID, hasTenant := execution.GetContext(ctx).TenantID()
			s.True(hasTenant)
			s.Equal(intg.TenantID, tenantID, "events are ingested in the installation's tenant")
			queued = events
			return nil
		}).
		Once()

	response := h.post(token, encodeReport(s.T(), testReport()))

	s.Equal(http.StatusNoContent, response.Code)
	s.Require().Len(queued, 1)
	s.Equal(intg.ID.String(), queued[0].ProviderNamespace)
	s.Contains(h.logs.String(), intg.ID.String())
	s.requireTokenNotLogged(h, token)
}

func (s *WebhookHandlerSuite) TestDeliveryWithUnrecognizedTokenIsNotFound() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newHarness(tdb)

	replaced, replacedToken := s.install(ctx, tdb, integrationName)
	_, currentHash, tokenErr := integrations.NewWebhookToken()
	s.Require().NoError(tokenErr)
	replaceToken := tdb.Client(ctx).Integration.UpdateOne(replaced).SetWebhookTokenHash(currentHash)
	s.Require().NoError(replaceToken.Exec(ctx))

	_, otherIntegrationToken := s.install(ctx, tdb, "alertmanager")

	unknownToken, _, unknownErr := integrations.NewWebhookToken()
	s.Require().NoError(unknownErr)

	tokens := map[string]string{
		"unknown":           unknownToken,
		"replaced":          replacedToken,
		"other integration": otherIntegrationToken,
	}
	for name, token := range tokens {
		response := h.post(token, encodeReport(s.T(), testReport()))
		s.Equal(http.StatusNotFound, response.Code, name)
		s.requireTokenNotLogged(h, token)
	}
}

func (s *WebhookHandlerSuite) TestOversizedDeliveryIsRejected() {
	ctx, tdb := s.SetupTestDatabase()
	_, token := s.install(ctx, tdb, integrationName)
	h := s.newHarness(tdb)

	report := testReport()
	report["notes"] = strings.Repeat("n", maxDeliveryBytes)

	response := h.post(token, encodeReport(s.T(), report))

	s.Equal(http.StatusRequestEntityTooLarge, response.Code)
	s.requireTokenNotLogged(h, token)
}

func (s *WebhookHandlerSuite) TestInvalidReportNamesTheFieldAndQueuesNothing() {
	ctx, tdb := s.SetupTestDatabase()
	_, token := s.install(ctx, tdb, integrationName)
	h := s.newHarness(tdb)

	report := testReport()
	report["repository"] = "checkout"
	response := h.post(token, encodeReport(s.T(), report))

	s.Equal(http.StatusBadRequest, response.Code)
	s.Equal("application/json", response.Header().Get("Content-Type"))
	var body map[string]string
	s.Require().NoError(json.Unmarshal(response.Body.Bytes(), &body))
	s.True(strings.HasPrefix(body["error"], "repository: "), body["error"])
	s.requireTokenNotLogged(h, token)

	notJSON := h.post(token, []byte("status=succeeded"))
	s.Equal(http.StatusBadRequest, notJSON.Code)
	s.Contains(notJSON.Body.String(), `"error":"body: `)
}

func (s *WebhookHandlerSuite) TestEnqueueFailureAsksForRetry() {
	ctx, tdb := s.SetupTestDatabase()
	_, token := s.install(ctx, tdb, integrationName)
	h := s.newHarness(tdb)

	h.events.EXPECT().IngestMany(mock.Anything, mock.Anything).
		Return(fmt.Errorf("insert failed")).
		Once()

	response := h.post(token, encodeReport(s.T(), testReport()))

	s.Equal(http.StatusServiceUnavailable, response.Code)
	s.requireTokenNotLogged(h, token)
}

func (s *WebhookHandlerSuite) TestStoredConfigWithoutAKnownPresetAsksForRetry() {
	configs := map[string]string{
		"missing preset":   `{}`,
		"malformed preset": `{"preset":["deployment"]}`,
		"unknown preset":   `{"preset":"incident"}`,
	}
	for name, config := range configs {
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			_, token := s.installWithConfig(ctx, tdb, integrationName, config)
			// The mock fails the test if anything is enqueued.
			h := s.newHarness(tdb)

			response := h.post(token, encodeReport(s.T(), testReport()))

			s.Equal(http.StatusServiceUnavailable, response.Code)
			s.requireTokenNotLogged(h, token)
		})
	}
}

package alertmanager

import (
	"bytes"
	"context"
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
	logger := slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	events := mocks.NewMockProviderEventPipelineService(s.T())
	handler := newWebhookHandler(logger, test.NewClock(testReceivedAt), events, installationLookup{db: tdb})
	return &webhookHarness{handler: handler, events: events, logs: logs}
}

// install creates an installation of the named integration in the context's tenant with a webhook token.
func (s *WebhookHandlerSuite) install(ctx context.Context, tdb rez.Database, name string) (*ent.Integration, string) {
	token, tokenHash, tokenErr := integrations.NewWebhookToken()
	s.Require().NoError(tokenErr)
	create := tdb.Client(ctx).Integration.Create().
		SetProvider(name).
		SetName(name).
		SetDisplayName(name).
		SetProviderInstallationRef(uuid.NewString()).
		SetInstallationConfig([]byte(`{}`)).
		SetWebhookTokenHash(tokenHash)
	intg, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return intg, token
}

func (h *webhookHarness) post(token string, body []byte) int {
	request := httptest.NewRequest(http.MethodPost, "/"+token, bytes.NewReader(body))
	request = request.WithContext(execution.NewRootContext(request.Context(), execution.KindAnonymous, execution.SourceHTTP))
	recorder := httptest.NewRecorder()
	h.handler.ServeHTTP(recorder, request)
	return recorder.Code
}

func (s *WebhookHandlerSuite) requireTokenNotLogged(h *webhookHarness, token string) {
	s.NotEmpty(h.logs.String(), "the delivery is logged")
	s.NotContains(h.logs.String(), token)
}

func (s *WebhookHandlerSuite) validDelivery(alerts ...map[string]any) []byte {
	if len(alerts) == 0 {
		alerts = append(alerts, testAlert("firing", map[string]string{"alertname": "HighErrorRate", "service": "checkout"}))
	}
	return encode(s.T(), testPayload(alerts...))
}

func (s *WebhookHandlerSuite) TestDeliveryQueuesValidAlertsInItsInstallationTenant() {
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

	invalid := testAlert("firing", map[string]string{"service": "checkout"})
	invalid["fingerprint"] = "bad-alert"
	valid := testAlert("firing", map[string]string{"alertname": "HighErrorRate", "service": "checkout"})
	status := h.post(token, s.validDelivery(invalid, valid))

	s.Equal(http.StatusNoContent, status)
	s.Require().Len(queued, 1)
	s.Equal(intg.ID.String(), queued[0].ProviderNamespace)
	s.Contains(h.logs.String(), "bad-alert", "the skipped alert is logged with its fingerprint")
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

	_, otherProviderToken := s.install(ctx, tdb, "github")

	unknownToken, _, unknownErr := integrations.NewWebhookToken()
	s.Require().NoError(unknownErr)

	for name, token := range map[string]string{
		"unknown":        unknownToken,
		"replaced":       replacedToken,
		"other provider": otherProviderToken,
	} {
		s.Equal(http.StatusNotFound, h.post(token, s.validDelivery()), name)
		s.requireTokenNotLogged(h, token)
	}
}

func (s *WebhookHandlerSuite) TestOversizedDeliveryIsRejected() {
	ctx, tdb := s.SetupTestDatabase()
	_, token := s.install(ctx, tdb, integrationName)
	h := s.newHarness(tdb)

	padding := strings.Repeat(" ", maxDeliveryBytes)
	oversized := fmt.Sprintf(`{"version":"4","alerts":[]%s}`, padding)

	s.Equal(http.StatusRequestEntityTooLarge, h.post(token, []byte(oversized)))
	s.requireTokenNotLogged(h, token)
}

func (s *WebhookHandlerSuite) TestInvalidPayloadQueuesNothing() {
	ctx, tdb := s.SetupTestDatabase()
	_, token := s.install(ctx, tdb, integrationName)
	h := s.newHarness(tdb)

	s.Equal(http.StatusBadRequest, h.post(token, []byte(`{"version":"3","alerts":[]}`)))
	s.requireTokenNotLogged(h, token)
}

func (s *WebhookHandlerSuite) TestEnqueueFailureAsksForRetry() {
	ctx, tdb := s.SetupTestDatabase()
	_, token := s.install(ctx, tdb, integrationName)
	h := s.newHarness(tdb)

	h.events.EXPECT().IngestMany(mock.Anything, mock.Anything).
		Return(fmt.Errorf("insert failed")).
		Once()

	s.Equal(http.StatusServiceUnavailable, h.post(token, s.validDelivery()))
	s.requireTokenNotLogged(h, token)
}

package alertmanager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

var (
	testInstallationID = uuid.MustParse("6f1c9a52-3a7e-4c8b-9a51-1d2e3f405162")
	testReceivedAt     = time.Date(2026, 6, 1, 10, 4, 30, 0, time.UTC)
	defaultSettings    = &installationSettings{serviceLabels: defaultServiceLabels}
)

func testAlert(status string, labels map[string]string) map[string]any {
	alert := map[string]any{
		"status":       status,
		"labels":       labels,
		"annotations":  map[string]string{"summary": "Checkout errors above 5%", "description": "5xx rate is 12%"},
		"startsAt":     "2026-06-01T10:00:00Z",
		"endsAt":       "0001-01-01T00:00:00Z",
		"generatorURL": "http://prometheus/graph?g0.expr=rate",
		"fingerprint":  "c0ffee",
	}
	if status == projections.AlertStateResolved {
		alert["endsAt"] = "2026-06-01T10:03:00Z"
	}
	return alert
}

func testPayload(alerts ...map[string]any) map[string]any {
	return map[string]any{
		"version":         "4",
		"status":          "firing",
		"truncatedAlerts": 0,
		"alerts":          alerts,
	}
}

func encode(t *testing.T, v any) []byte {
	t.Helper()
	encoded, encodeErr := json.Marshal(v)
	require.NoError(t, encodeErr)
	return encoded
}

type DeliverySuite struct {
	test.Suite
}

func TestDeliverySuite(t *testing.T) {
	suite.Run(t, &DeliverySuite{Suite: test.NewSuite()})
}

func (s *DeliverySuite) mapTestDelivery(settings *installationSettings, payload map[string]any, receivedAt time.Time) *mappedDelivery {
	s.T().Helper()
	delivery, mapErr := mapDelivery(testInstallationID, settings, encode(s.T(), payload), receivedAt)
	s.Require().NoError(mapErr)
	return delivery
}

// normalize maps a provider event to its normalized alert instance attributes.
func (s *DeliverySuite) normalize(event rez.ProviderEvent) (string, time.Time, projections.AlertInstanceEventAttributes) {
	s.T().Helper()
	normalized, processErr := EventProcessor{}.ProcessProviderEvent(s.T().Context(), event)
	s.Require().NoError(processErr)
	s.Require().Len(normalized, 1)
	var attrs projections.AlertInstanceEventAttributes
	s.Require().NoError(json.Unmarshal(normalized[0].Attributes, &attrs))
	s.Require().Equal(projections.KindAlertInstance, normalized[0].Kind)
	s.Require().Equal(event.ProviderEventRef, normalized[0].ProviderEventRef)
	return normalized[0].ProviderResourceRef, normalized[0].OccurredAt, attrs
}

func (s *DeliverySuite) TestMapFiringAndResolvedAlerts() {
	labels := map[string]string{"alertname": "HighErrorRate", "service": "checkout-api", "severity": "page"}
	payload := testPayload(testAlert("firing", labels), testAlert("resolved", labels))
	// The group status does not stand in for each alert's status.
	payload["status"] = "resolved"
	delivery := s.mapTestDelivery(defaultSettings, payload, testReceivedAt)
	s.Require().Empty(delivery.skipped)
	s.Require().Len(delivery.events, 2)

	firing := delivery.events[0]
	s.Require().Equal(ProviderName, firing.Provider)
	s.Require().Equal(testInstallationID.String(), firing.ProviderNamespace)
	s.Require().Equal(sourceAlerts, firing.ProviderEventSource)
	s.Require().Equal(testReceivedAt, firing.ReceivedAt)

	definitionRef, occurredAt, attrs := s.normalize(firing)
	s.Require().Equal("HighErrorRate", attrs.Title)
	s.Require().Equal("Checkout errors above 5%", attrs.Summary)
	s.Require().Equal("5xx rate is 12%", attrs.Description)
	s.Require().Equal("http://prometheus/graph?g0.expr=rate", attrs.Definition)
	s.Require().Equal(projections.AlertStateFiring, attrs.State)
	s.Require().Equal("c0ffee", attrs.InstanceID)
	s.Require().Equal(labels, attrs.Labels)
	s.Require().Equal("critical", attrs.Severity)
	s.Require().Equal("page", attrs.SeverityRef)
	s.Require().Equal(time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC), attrs.StartedAt)
	s.Require().Nil(attrs.EndedAt, "a firing alert's endsAt is not a resolution")
	s.Require().Nil(attrs.ResolutionTimeoutSeconds)
	s.Require().Equal(testReceivedAt, occurredAt)
	expectedService := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: testInstallationID.String(),
			ResourceRef:       "service:checkout-api",
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: "checkout-api",
		LinkingAttributes: projections.LinkingAttributes{
			"service.name": "checkout-api",
		},
	}
	s.Require().Equal([]projections.EntityObservation{expectedService}, attrs.ObservedEntities)

	resolvedRef, resolvedAt, resolved := s.normalize(delivery.events[1])
	endedAt := time.Date(2026, 6, 1, 10, 3, 0, 0, time.UTC)
	s.Require().Equal(definitionRef, resolvedRef)
	s.Require().Equal(projections.AlertStateResolved, resolved.State)
	s.Require().Equal(&endedAt, resolved.EndedAt)
	s.Require().Equal(endedAt, resolvedAt)
}

func (s *DeliverySuite) TestMapSeverity() {
	cases := map[string]string{
		"critical": "critical", "PAGE": "critical", "p1": "critical", "Sev1": "critical", "high": "critical",
		"warning": "warning", "warn": "warning", "p2": "warning", "P3": "warning", "medium": "warning",
		"info": "info", "none": "info", "low": "info", "p4": "info", "p5": "info",
		"": "unknown", "error": "unknown", "sev2": "unknown",
	}
	for raw, expected := range cases {
		s.Require().Equal(expected, string(mapSeverity(raw)), "severity %q", raw)
	}
}

func (s *DeliverySuite) TestMapSkipsInvalidAlerts() {
	valid := testAlert("firing", map[string]string{"alertname": "HighErrorRate"})
	withChange := func(change func(map[string]any)) map[string]any {
		alert := testAlert("firing", map[string]string{"alertname": "HighErrorRate"})
		change(alert)
		return alert
	}
	invalid := []map[string]any{
		withChange(func(a map[string]any) { a["labels"] = map[string]string{"service": "checkout"} }),
		withChange(func(a map[string]any) { a["labels"] = map[string]string{"alertname": " "} }),
		withChange(func(a map[string]any) { a["labels"] = map[string]any{"alertname": 5} }),
		withChange(func(a map[string]any) { delete(a, "fingerprint") }),
		withChange(func(a map[string]any) { a["status"] = "pending" }),
		withChange(func(a map[string]any) { a["startsAt"] = "yesterday" }),
		withChange(func(a map[string]any) { a["startsAt"] = "0001-01-01T00:00:00Z" }),
		withChange(func(a map[string]any) { a["status"] = "resolved"; a["endsAt"] = "0001-01-01T00:00:00Z" }),
		withChange(func(a map[string]any) { a["status"] = "resolved"; a["endsAt"] = "2026-06-01T09:59:59Z" }),
	}
	alerts := append([]map[string]any{valid}, invalid...)
	delivery := s.mapTestDelivery(defaultSettings, testPayload(alerts...), testReceivedAt)

	s.Require().Len(delivery.events, 1)
	s.Require().Len(delivery.skipped, len(invalid))
	for _, skipped := range delivery.skipped {
		s.Require().NotEmpty(skipped.reason)
	}
	s.Require().Equal("c0ffee", delivery.skipped[0].fingerprint, "a skipped alert reports its fingerprint")
	s.Require().Empty(delivery.skipped[3].fingerprint)
}

func (s *DeliverySuite) TestMapRejectsInvalidPayloads() {
	payloads := map[string]string{
		"not json":         `{"version":`,
		"version 3":        `{"version":"3","alerts":[]}`,
		"no version":       `{"alerts":[]}`,
		"no alerts":        `{"version":"4"}`,
		"null alerts":      `{"version":"4","alerts":null}`,
		"alerts not array": `{"version":"4","alerts":{}}`,
	}
	for name, payload := range payloads {
		_, mapErr := mapDelivery(testInstallationID, defaultSettings, []byte(payload), testReceivedAt)
		s.Require().Error(mapErr, name)
	}

	delivery, mapErr := mapDelivery(testInstallationID, defaultSettings, []byte(`{"version":"4","alerts":[],"extra":true}`), testReceivedAt)
	s.Require().NoError(mapErr, "unknown fields are ignored")
	s.Require().Empty(delivery.events)
}

func (s *DeliverySuite) TestMapTruncatedPayloadKeepsItsAlerts() {
	payload := testPayload(testAlert("firing", map[string]string{"alertname": "HighErrorRate"}))
	payload["truncatedAlerts"] = 3
	delivery := s.mapTestDelivery(defaultSettings, payload, testReceivedAt)
	s.Require().Equal(3, delivery.truncatedAlerts)
	s.Require().Len(delivery.events, 1)
}

func (s *DeliverySuite) TestMapEventIdentity() {
	labels := map[string]string{"alertname": "HighErrorRate"}
	firingRef := func(alert map[string]any, receivedAt time.Time) string {
		return s.mapTestDelivery(defaultSettings, testPayload(alert), receivedAt).events[0].ProviderEventRef
	}
	firing := testAlert("firing", labels)
	sameMinute := testReceivedAt.Add(20 * time.Second)
	nextMinute := testReceivedAt.Add(time.Minute)

	s.Require().Equal(firingRef(firing, testReceivedAt), firingRef(firing, sameMinute))
	s.Require().NotEqual(firingRef(firing, testReceivedAt), firingRef(firing, nextMinute))
	s.Require().Regexp(`^[0-9a-f]{32}-202606011004$`, firingRef(firing, testReceivedAt))

	// Equivalent timestamps have one identity.
	offset := testAlert("firing", labels)
	offset["startsAt"] = "2026-06-01T20:00:00.000+10:00"
	s.Require().Equal(firingRef(firing, testReceivedAt), firingRef(offset, testReceivedAt))

	resolved := testAlert("resolved", labels)
	resolvedOffset := testAlert("resolved", labels)
	resolvedOffset["startsAt"] = "2026-06-01T10:00:00.000000Z"
	resolvedOffset["endsAt"] = "2026-06-01T12:03:00+02:00"
	resolvedRef := firingRef(resolved, testReceivedAt)
	s.Require().Equal(resolvedRef, firingRef(resolved, nextMinute.Add(time.Hour)), "a resolution's identity does not depend on receipt")
	s.Require().Equal(resolvedRef, firingRef(resolvedOffset, testReceivedAt))
	s.Require().Regexp(`^[0-9a-f]{32}$`, resolvedRef)
}

func (s *DeliverySuite) TestMapResolvesService() {
	serviceOf := func(settings *installationSettings, labels map[string]string) *projections.EntityObservation {
		labels["alertname"] = "HighErrorRate"
		delivery := s.mapTestDelivery(settings, testPayload(testAlert("firing", labels)), testReceivedAt)
		_, _, attrs := s.normalize(delivery.events[0])
		if len(attrs.ObservedEntities) == 0 {
			return nil
		}
		s.Require().Len(attrs.ObservedEntities, 1)
		return &attrs.ObservedEntities[0]
	}

	checkout := serviceOf(defaultSettings, map[string]string{"service": "Checkout_API"})
	s.Require().Equal("service:checkout-api", checkout.Ref.ResourceRef)
	s.Require().Equal("Checkout_API", checkout.DisplayName)
	s.Require().Equal(projections.LinkingAttributes{"service.name": "checkout-api"}, checkout.LinkingAttributes)

	// Empty and invalid values fall back to the next configured label.
	fallback := serviceOf(defaultSettings, map[string]string{"service": "---", "service_name": "", "app": "  Search..API  "})
	s.Require().Equal("service:search-api", fallback.Ref.ResourceRef)

	s.Require().Equal("service:cart", serviceOf(defaultSettings, map[string]string{"app_kubernetes_io_name": "cart"}).Ref.ResourceRef)
	s.Require().Nil(serviceOf(defaultSettings, map[string]string{"job": "checkout"}), "job is not a default service label")

	custom := &installationSettings{serviceLabels: []string{"job"}}
	s.Require().Equal("service:checkout", serviceOf(custom, map[string]string{"job": "checkout", "service": "search"}).Ref.ResourceRef)

	disabled := &installationSettings{serviceLabels: []string{}}
	s.Require().Nil(serviceOf(disabled, map[string]string{"service": "checkout"}))
}

func (s *DeliverySuite) TestServiceAliasesShareOneDefinition() {
	definitionOf := func(labels map[string]string) (string, rez.ProviderResourceRef) {
		delivery := s.mapTestDelivery(defaultSettings, testPayload(testAlert("firing", labels)), testReceivedAt)
		ref, _, attrs := s.normalize(delivery.events[0])
		if len(attrs.ObservedEntities) == 0 {
			return ref, rez.ProviderResourceRef{}
		}
		return ref, attrs.ObservedEntities[0].Ref
	}

	checkoutRef, checkoutService := definitionOf(map[string]string{"alertname": "HighErrorRate", "service": "Checkout_API"})
	aliasRef, aliasService := definitionOf(map[string]string{"alertname": "HighErrorRate", "service_name": "checkout-api"})
	s.Require().Equal(checkoutRef, aliasRef)
	s.Require().Equal(checkoutService, aliasService)
	s.Require().Regexp(`^alert:[0-9a-f]{64}$`, checkoutRef)

	searchRef, _ := definitionOf(map[string]string{"alertname": "HighErrorRate", "service": "search"})
	s.Require().NotEqual(checkoutRef, searchRef)

	noServiceRef, _ := definitionOf(map[string]string{"alertname": "HighErrorRate"})
	s.Require().NotEqual(checkoutRef, noServiceRef)
	s.Require().Equal("alert:"+identityDigest("HighErrorRate", nil), noServiceRef)
}

func (s *DeliverySuite) TestParseInstallationSettings() {
	missing, missingErr := parseInstallationSettings(nil)
	s.Require().NoError(missingErr)
	s.Require().Equal(defaultServiceLabels, missing.serviceLabels)

	other, otherErr := parseInstallationSettings(map[string]any{"other": true})
	s.Require().NoError(otherErr)
	s.Require().Equal(defaultServiceLabels, other.serviceLabels)

	empty, emptyErr := parseInstallationSettings(map[string]any{"service_labels": []any{}})
	s.Require().NoError(emptyErr)
	s.Require().Empty(empty.serviceLabels)

	custom, customErr := parseInstallationSettings(map[string]any{"service_labels": []any{"job", "service"}})
	s.Require().NoError(customErr)
	s.Require().Equal([]string{"job", "service"}, custom.serviceLabels)

	invalid := []any{
		nil,
		"service",
		map[string]any{"service": true},
		[]any{"service", 5},
		[]any{""},
		[]any{"  "},
		[]any{"service", "service"},
	}
	for _, value := range invalid {
		_, parseErr := parseInstallationSettings(map[string]any{"service_labels": value})
		s.Require().ErrorIs(parseErr, rez.ErrInvalidInput, "%#v", value)
	}
}

func (s *DeliverySuite) TestInstallIgnoresClientConfig() {
	i := &Integration{}
	first, firstErr := i.ValidateInstallationConfig([]byte(`{"token":"client-supplied"}`))
	s.Require().NoError(firstErr)
	second, secondErr := i.ValidateInstallationConfig(nil)
	s.Require().NoError(secondErr)

	s.Require().NotEqual(first.InstallationTargetRef().ResourceRef, second.InstallationTargetRef().ResourceRef)
	s.Require().NoError(first.InstallationTargetRef().Validate())
	encoded, encodeErr := first.Encode()
	s.Require().NoError(encodeErr)
	s.Require().JSONEq(`{}`, string(encoded))
}

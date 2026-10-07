package alertmanager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
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

func mapTestDelivery(t *testing.T, settings *installationSettings, payload map[string]any, receivedAt time.Time) *mappedDelivery {
	t.Helper()
	delivery, mapErr := mapDelivery(testInstallationID, settings, encode(t, payload), receivedAt)
	require.NoError(t, mapErr)
	return delivery
}

// normalize maps a provider event to its normalized alert instance attributes.
func normalize(t *testing.T, event rez.ProviderEvent) (string, time.Time, projections.AlertInstanceEventAttributes) {
	t.Helper()
	normalized, processErr := EventProcessor{}.ProcessProviderEvent(t.Context(), event)
	require.NoError(t, processErr)
	require.Len(t, normalized, 1)
	var attrs projections.AlertInstanceEventAttributes
	require.NoError(t, json.Unmarshal(normalized[0].Attributes, &attrs))
	require.Equal(t, projections.KindAlertInstance, normalized[0].Kind)
	require.Equal(t, event.ProviderEventRef, normalized[0].ProviderEventRef)
	return normalized[0].ProviderResourceRef, normalized[0].OccurredAt, attrs
}

func TestMapFiringAndResolvedAlerts(t *testing.T) {
	labels := map[string]string{"alertname": "HighErrorRate", "service": "checkout-api", "severity": "page"}
	payload := testPayload(testAlert("firing", labels), testAlert("resolved", labels))
	// The group status does not stand in for each alert's status.
	payload["status"] = "resolved"
	delivery := mapTestDelivery(t, defaultSettings, payload, testReceivedAt)
	require.Empty(t, delivery.skipped)
	require.Len(t, delivery.events, 2)

	firing := delivery.events[0]
	require.Equal(t, ProviderName, firing.Provider)
	require.Equal(t, testInstallationID.String(), firing.ProviderNamespace)
	require.Equal(t, sourceAlerts, firing.ProviderEventSource)
	require.Equal(t, testReceivedAt, firing.ReceivedAt)

	definitionRef, occurredAt, attrs := normalize(t, firing)
	require.Equal(t, "HighErrorRate", attrs.Title)
	require.Equal(t, "Checkout errors above 5%", attrs.Summary)
	require.Equal(t, "5xx rate is 12%", attrs.Description)
	require.Equal(t, "http://prometheus/graph?g0.expr=rate", attrs.Definition)
	require.Equal(t, projections.AlertStateFiring, attrs.State)
	require.Equal(t, "c0ffee", attrs.InstanceID)
	require.Equal(t, labels, attrs.Labels)
	require.Equal(t, "critical", attrs.Severity)
	require.Equal(t, "page", attrs.SeverityRef)
	require.Equal(t, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC), attrs.StartedAt)
	require.Nil(t, attrs.EndedAt, "a firing alert's endsAt is not a resolution")
	require.Nil(t, attrs.ResolutionTimeoutSeconds)
	require.Equal(t, testReceivedAt, occurredAt)
	expectedService := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: testInstallationID.String(),
			ResourceRef:       "service:checkout-api",
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: "checkout-api",
	}
	require.Equal(t, []projections.EntityObservation{expectedService}, attrs.ObservedEntities)

	resolvedRef, resolvedAt, resolved := normalize(t, delivery.events[1])
	endedAt := time.Date(2026, 6, 1, 10, 3, 0, 0, time.UTC)
	require.Equal(t, definitionRef, resolvedRef)
	require.Equal(t, projections.AlertStateResolved, resolved.State)
	require.Equal(t, &endedAt, resolved.EndedAt)
	require.Equal(t, endedAt, resolvedAt)
}

func TestMapSeverity(t *testing.T) {
	cases := map[string]string{
		"critical": "critical", "PAGE": "critical", "p1": "critical", "Sev1": "critical", "high": "critical",
		"warning": "warning", "warn": "warning", "p2": "warning", "P3": "warning", "medium": "warning",
		"info": "info", "none": "info", "low": "info", "p4": "info", "p5": "info",
		"": "unknown", "error": "unknown", "sev2": "unknown",
	}
	for raw, expected := range cases {
		require.Equal(t, expected, string(mapSeverity(raw)), "severity %q", raw)
	}
}

func TestMapSkipsInvalidAlerts(t *testing.T) {
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
	delivery := mapTestDelivery(t, defaultSettings, testPayload(alerts...), testReceivedAt)

	require.Len(t, delivery.events, 1)
	require.Len(t, delivery.skipped, len(invalid))
	for _, skipped := range delivery.skipped {
		require.NotEmpty(t, skipped.reason)
	}
	require.Equal(t, "c0ffee", delivery.skipped[0].fingerprint, "a skipped alert reports its fingerprint")
	require.Empty(t, delivery.skipped[3].fingerprint)
}

func TestMapRejectsInvalidPayloads(t *testing.T) {
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
		require.Error(t, mapErr, name)
	}

	delivery, mapErr := mapDelivery(testInstallationID, defaultSettings, []byte(`{"version":"4","alerts":[],"extra":true}`), testReceivedAt)
	require.NoError(t, mapErr, "unknown fields are ignored")
	require.Empty(t, delivery.events)
}

func TestMapTruncatedPayloadKeepsItsAlerts(t *testing.T) {
	payload := testPayload(testAlert("firing", map[string]string{"alertname": "HighErrorRate"}))
	payload["truncatedAlerts"] = 3
	delivery := mapTestDelivery(t, defaultSettings, payload, testReceivedAt)
	require.Equal(t, 3, delivery.truncatedAlerts)
	require.Len(t, delivery.events, 1)
}

func TestMapEventIdentity(t *testing.T) {
	labels := map[string]string{"alertname": "HighErrorRate"}
	firingRef := func(alert map[string]any, receivedAt time.Time) string {
		return mapTestDelivery(t, defaultSettings, testPayload(alert), receivedAt).events[0].ProviderEventRef
	}
	firing := testAlert("firing", labels)
	sameMinute := testReceivedAt.Add(20 * time.Second)
	nextMinute := testReceivedAt.Add(time.Minute)

	require.Equal(t, firingRef(firing, testReceivedAt), firingRef(firing, sameMinute))
	require.NotEqual(t, firingRef(firing, testReceivedAt), firingRef(firing, nextMinute))
	require.Regexp(t, `^[0-9a-f]{32}-202606011004$`, firingRef(firing, testReceivedAt))

	// Equivalent timestamps have one identity.
	offset := testAlert("firing", labels)
	offset["startsAt"] = "2026-06-01T20:00:00.000+10:00"
	require.Equal(t, firingRef(firing, testReceivedAt), firingRef(offset, testReceivedAt))

	resolved := testAlert("resolved", labels)
	resolvedOffset := testAlert("resolved", labels)
	resolvedOffset["startsAt"] = "2026-06-01T10:00:00.000000Z"
	resolvedOffset["endsAt"] = "2026-06-01T12:03:00+02:00"
	resolvedRef := firingRef(resolved, testReceivedAt)
	require.Equal(t, resolvedRef, firingRef(resolved, nextMinute.Add(time.Hour)), "a resolution's identity does not depend on receipt")
	require.Equal(t, resolvedRef, firingRef(resolvedOffset, testReceivedAt))
	require.Regexp(t, `^[0-9a-f]{32}$`, resolvedRef)
}

func TestMapResolvesService(t *testing.T) {
	serviceOf := func(settings *installationSettings, labels map[string]string) *projections.EntityObservation {
		labels["alertname"] = "HighErrorRate"
		delivery := mapTestDelivery(t, settings, testPayload(testAlert("firing", labels)), testReceivedAt)
		_, _, attrs := normalize(t, delivery.events[0])
		if len(attrs.ObservedEntities) == 0 {
			return nil
		}
		require.Len(t, attrs.ObservedEntities, 1)
		return &attrs.ObservedEntities[0]
	}

	checkout := serviceOf(defaultSettings, map[string]string{"service": "Checkout_API"})
	require.Equal(t, "service:checkout-api", checkout.Ref.ResourceRef)
	require.Equal(t, "Checkout_API", checkout.DisplayName)

	// Empty and invalid values fall back to the next configured label.
	fallback := serviceOf(defaultSettings, map[string]string{"service": "---", "service_name": "", "app": "  Search..API  "})
	require.Equal(t, "service:search-api", fallback.Ref.ResourceRef)

	require.Equal(t, "service:cart", serviceOf(defaultSettings, map[string]string{"app_kubernetes_io_name": "cart"}).Ref.ResourceRef)
	require.Nil(t, serviceOf(defaultSettings, map[string]string{"job": "checkout"}), "job is not a default service label")

	custom := &installationSettings{serviceLabels: []string{"job"}}
	require.Equal(t, "service:checkout", serviceOf(custom, map[string]string{"job": "checkout", "service": "search"}).Ref.ResourceRef)

	disabled := &installationSettings{serviceLabels: []string{}}
	require.Nil(t, serviceOf(disabled, map[string]string{"service": "checkout"}))
}

func TestServiceAliasesShareOneDefinition(t *testing.T) {
	definitionOf := func(labels map[string]string) (string, rez.ProviderResourceRef) {
		delivery := mapTestDelivery(t, defaultSettings, testPayload(testAlert("firing", labels)), testReceivedAt)
		ref, _, attrs := normalize(t, delivery.events[0])
		if len(attrs.ObservedEntities) == 0 {
			return ref, rez.ProviderResourceRef{}
		}
		return ref, attrs.ObservedEntities[0].Ref
	}

	checkoutRef, checkoutService := definitionOf(map[string]string{"alertname": "HighErrorRate", "service": "Checkout_API"})
	aliasRef, aliasService := definitionOf(map[string]string{"alertname": "HighErrorRate", "service_name": "checkout-api"})
	require.Equal(t, checkoutRef, aliasRef)
	require.Equal(t, checkoutService, aliasService)
	require.Regexp(t, `^alert:[0-9a-f]{64}$`, checkoutRef)

	searchRef, _ := definitionOf(map[string]string{"alertname": "HighErrorRate", "service": "search"})
	require.NotEqual(t, checkoutRef, searchRef)

	noServiceRef, _ := definitionOf(map[string]string{"alertname": "HighErrorRate"})
	require.NotEqual(t, checkoutRef, noServiceRef)
	require.Equal(t, "alert:"+identityDigest("HighErrorRate", nil), noServiceRef)
}

func TestParseInstallationSettings(t *testing.T) {
	missing, missingErr := parseInstallationSettings(nil)
	require.NoError(t, missingErr)
	require.Equal(t, defaultServiceLabels, missing.serviceLabels)

	other, otherErr := parseInstallationSettings(map[string]any{"other": true})
	require.NoError(t, otherErr)
	require.Equal(t, defaultServiceLabels, other.serviceLabels)

	empty, emptyErr := parseInstallationSettings(map[string]any{"service_labels": []any{}})
	require.NoError(t, emptyErr)
	require.Empty(t, empty.serviceLabels)

	custom, customErr := parseInstallationSettings(map[string]any{"service_labels": []any{"job", "service"}})
	require.NoError(t, customErr)
	require.Equal(t, []string{"job", "service"}, custom.serviceLabels)

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
		require.ErrorIs(t, parseErr, rez.ErrInvalidInput, "%#v", value)
	}
}

func TestInstallIgnoresClientConfig(t *testing.T) {
	i := &Integration{}
	first, firstErr := i.ValidateInstallationConfig([]byte(`{"token":"client-supplied"}`))
	require.NoError(t, firstErr)
	second, secondErr := i.ValidateInstallationConfig(nil)
	require.NoError(t, secondErr)

	require.NotEqual(t, first.InstallationTargetRef().ResourceRef, second.InstallationTargetRef().ResourceRef)
	require.NoError(t, first.InstallationTargetRef().Validate())
	encoded, encodeErr := first.Encode()
	require.NoError(t, encodeErr)
	require.JSONEq(t, `{}`, string(encoded))
}

package alertmanager

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

// webhookPayload is the Alertmanager webhook payload, version 4. Alerts are decoded one at a time so one
// malformed alert does not reject the rest.
type webhookPayload struct {
	Version         string            `json:"version"`
	TruncatedAlerts int               `json:"truncatedAlerts"`
	Alerts          []json.RawMessage `json:"alerts"`
}

type webhookAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// alertEventAttributes are the provider event attributes for one alert. Times are canonical UTC, and the
// service is resolved from the installation's settings at receive time, so processing needs no installation
// configuration.
type alertEventAttributes struct {
	Fingerprint  string            `json:"fingerprint"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	StartsAt     time.Time         `json:"starts_at"`
	EndsAt       *time.Time        `json:"ends_at,omitempty"`
	GeneratorURL string            `json:"generator_url,omitempty"`
	// ServiceName is the normalized service name, empty when no service was resolved.
	ServiceName string                         `json:"service_name,omitempty"`
	Service     *projections.EntityObservation `json:"service,omitempty"`
}

type mappedDelivery struct {
	events          []rez.ProviderEvent
	skipped         []skippedAlert
	truncatedAlerts int
}

type skippedAlert struct {
	fingerprint string
	reason      string
}

// mapDelivery maps one webhook delivery to provider events, one per valid alert. An error rejects the
// whole payload; invalid alerts are skipped and reported instead.
func mapDelivery(integrationID uuid.UUID, settings *installationSettings, body []byte, receivedAt time.Time) (*mappedDelivery, error) {
	var payload webhookPayload
	if decodeErr := json.Unmarshal(body, &payload); decodeErr != nil {
		return nil, fmt.Errorf("decode payload: %w", decodeErr)
	}
	if payload.Version != "4" {
		return nil, fmt.Errorf("unsupported webhook version %q", payload.Version)
	}
	if payload.Alerts == nil {
		return nil, fmt.Errorf("payload has no alerts array")
	}

	m := alertMapper{
		namespace:  integrationID.String(),
		settings:   settings,
		receivedAt: receivedAt.UTC(),
	}
	delivery := &mappedDelivery{truncatedAlerts: payload.TruncatedAlerts}
	for _, rawAlert := range payload.Alerts {
		event, mapErr := m.mapAlert(rawAlert)
		if mapErr != nil {
			skipped := skippedAlert{fingerprint: rawAlertFingerprint(rawAlert), reason: mapErr.Error()}
			delivery.skipped = append(delivery.skipped, skipped)
			continue
		}
		delivery.events = append(delivery.events, *event)
	}
	return delivery, nil
}

// rawAlertFingerprint reads the fingerprint of an alert that may not decode.
func rawAlertFingerprint(raw json.RawMessage) string {
	var alert struct {
		Fingerprint any `json:"fingerprint"`
	}
	if decodeErr := json.Unmarshal(raw, &alert); decodeErr != nil {
		return ""
	}
	fingerprint, _ := alert.Fingerprint.(string)
	return fingerprint
}

type alertMapper struct {
	namespace  string
	settings   *installationSettings
	receivedAt time.Time
}

func (m *alertMapper) mapAlert(raw json.RawMessage) (*rez.ProviderEvent, error) {
	var alert webhookAlert
	if decodeErr := json.Unmarshal(raw, &alert); decodeErr != nil {
		return nil, fmt.Errorf("decode alert: %w", decodeErr)
	}
	attrs, validateErr := m.validateAlert(alert)
	if validateErr != nil {
		return nil, validateErr
	}
	attrs.ServiceName, attrs.Service = m.resolveService(alert.Labels)

	encoded, encodeErr := json.Marshal(attrs)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode alert attributes: %w", encodeErr)
	}
	event := &rez.ProviderEvent{
		Provider:            ProviderName,
		ProviderNamespace:   m.namespace,
		ProviderEventSource: sourceAlerts,
		ProviderEventRef:    m.eventRef(attrs),
		Attributes:          encoded,
		ReceivedAt:          m.receivedAt,
	}
	return event, nil
}

func (m *alertMapper) validateAlert(alert webhookAlert) (*alertEventAttributes, error) {
	if strings.TrimSpace(alert.Labels["alertname"]) == "" {
		return nil, errors.New("missing labels.alertname")
	}
	if alert.Fingerprint == "" {
		return nil, errors.New("missing fingerprint")
	}
	if alert.Status != projections.AlertStateFiring && alert.Status != projections.AlertStateResolved {
		return nil, fmt.Errorf("invalid status %q", alert.Status)
	}
	startsAt, startErr := parseAlertTime(alert.StartsAt)
	if startErr != nil {
		return nil, fmt.Errorf("invalid startsAt: %w", startErr)
	}
	attrs := &alertEventAttributes{
		Fingerprint:  alert.Fingerprint,
		Status:       alert.Status,
		Labels:       alert.Labels,
		Annotations:  alert.Annotations,
		StartsAt:     startsAt,
		GeneratorURL: alert.GeneratorURL,
	}
	// A firing alert's endsAt is Alertmanager's expiry estimate, not a resolution.
	if alert.Status == projections.AlertStateResolved {
		endsAt, endErr := parseAlertTime(alert.EndsAt)
		if endErr != nil {
			return nil, fmt.Errorf("invalid endsAt: %w", endErr)
		}
		if endsAt.Before(startsAt) {
			return nil, errors.New("endsAt is before startsAt")
		}
		attrs.EndsAt = &endsAt
	}
	return attrs, nil
}

func parseAlertTime(value string) (time.Time, error) {
	parsed, parseErr := time.Parse(time.RFC3339Nano, value)
	if parseErr != nil {
		return time.Time{}, parseErr
	}
	if parsed.IsZero() {
		return time.Time{}, errors.New("zero time")
	}
	return parsed.UTC(), nil
}

// eventRef identifies an observation. Firing observations are deduplicated within a UTC minute of receipt;
// a resolution has one identity however often it is delivered.
func (m *alertMapper) eventRef(attrs *alertEventAttributes) string {
	startsAt := attrs.StartsAt.Format(time.RFC3339Nano)
	if attrs.EndsAt == nil {
		hash := identityDigest(attrs.Fingerprint, projections.AlertStateFiring, startsAt)
		return hash[:32] + "-" + m.receivedAt.Format("200601021504")
	}
	hash := identityDigest(attrs.Fingerprint, projections.AlertStateResolved, startsAt, attrs.EndsAt.Format(time.RFC3339Nano))
	return hash[:32]
}

// resolveService returns the first configured label value that normalizes to a nonempty service name, and
// its observation.
func (m *alertMapper) resolveService(labels map[string]string) (string, *projections.EntityObservation) {
	for _, label := range m.settings.serviceLabels {
		value, present := labels[label]
		if !present {
			continue
		}
		name := normalizeServiceName(value)
		if name == "" {
			continue
		}
		observation := &projections.EntityObservation{
			Ref: rez.ProviderResourceRef{
				Provider:          ProviderName,
				ProviderNamespace: m.namespace,
				ResourceRef:       "service:" + name,
			},
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: value,
		}
		return name, observation
	}
	return "", nil
}

// normalizeServiceName lower-cases the value, replaces every run of characters outside [a-z0-9] with `-`
// and trims `-`.
func normalizeServiceName(value string) string {
	var b strings.Builder
	separated := false
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			separated = false
		} else if !separated {
			b.WriteByte('-')
			separated = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// identityDigest is the SHA-256 hex digest of the JSON-encoded tuple.
func identityDigest(parts ...any) string {
	encoded, encodeErr := json.Marshal(parts)
	if encodeErr != nil {
		// Strings and nil always encode.
		panic(fmt.Sprintf("encode identity tuple: %v", encodeErr))
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

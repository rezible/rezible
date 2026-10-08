package webhook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	maxNameLength = 200
	maxURLLength  = 2000
	// maxFutureSkew is how far after receipt a reported time may be.
	maxFutureSkew = 5 * time.Minute
)

var (
	repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,100}/[A-Za-z0-9._-]{1,100}$`)
	shaPattern        = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
)

// deploymentReport is the provider event attributes of one report.
type deploymentReport struct {
	// DeploymentRef identifies the deployment: its reports share it.
	DeploymentRef string                                `json:"deployment_ref"`
	Deployment    projections.DeploymentEventAttributes `json:"deployment"`
}

// deploymentPreset is the `deployment` preset: each delivery is one report of a deployment's progress.
type deploymentPreset struct{}

func (deploymentPreset) MapDelivery(integrationID uuid.UUID, body []byte, receivedAt time.Time) (*rez.ProviderEvent, error) {
	var fields reportFields
	if decodeErr := json.Unmarshal(body, &fields); decodeErr != nil || fields == nil {
		return nil, &fieldError{field: "body", reason: "must be a JSON object"}
	}
	m := reportMapper{
		namespace:  integrationID.String(),
		fields:     fields,
		receivedAt: receivedAt.UTC(),
	}
	report, mapErr := m.mapReport()
	if mapErr != nil {
		return nil, mapErr
	}
	encoded, encodeErr := json.Marshal(report)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode deployment report: %w", encodeErr)
	}
	event := &rez.ProviderEvent{
		Provider:            ProviderName,
		ProviderNamespace:   m.namespace,
		ProviderEventSource: presetDeployment,
		ProviderEventRef:    report.DeploymentRef + ":" + report.Deployment.Status,
		Attributes:          encoded,
		ReceivedAt:          m.receivedAt,
	}
	return event, nil
}

func (deploymentPreset) ProcessEvent(prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var report deploymentReport
	if decodeErr := json.Unmarshal(prov.Attributes, &report); decodeErr != nil {
		return nil, fmt.Errorf("decode deployment report: %w", decodeErr)
	}
	encodedAttrs, encodeErr := projections.EncodeAttributes(report.Deployment)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode deployment attributes: %w", encodeErr)
	}
	// The deployment's time is when it finished, or when it started while in progress. Mapping sets the one
	// its status implies.
	occurredAt := prov.ReceivedAt
	if report.Deployment.Finished() && report.Deployment.FinishedAt != nil {
		occurredAt = *report.Deployment.FinishedAt
	} else if !report.Deployment.Finished() && report.Deployment.StartedAt != nil {
		occurredAt = *report.Deployment.StartedAt
	}
	result := &ent.NormalizedEvent{
		Provider:            ProviderName,
		ProviderNamespace:   prov.ProviderNamespace,
		ProviderResourceRef: report.DeploymentRef,
		ProviderEventSource: prov.ProviderEventSource,
		ProviderEventRef:    prov.ProviderEventRef,
		Kind:                projections.KindDeployment,
		OccurredAt:          occurredAt,
		ReceivedAt:          prov.ReceivedAt,
		Attributes:          encodedAttrs,
	}
	return ent.NormalizedEvents{result}, nil
}

// reportFields are a report's raw top-level fields. Unknown fields are ignored.
type reportFields map[string]json.RawMessage

// optional returns the trimmed string value of a field, empty when it is absent, null or empty.
func (f reportFields) optional(name string) (string, error) {
	raw, present := f[name]
	if !present {
		return "", nil
	}
	var value *string
	if decodeErr := json.Unmarshal(raw, &value); decodeErr != nil {
		return "", &fieldError{field: name, reason: "must be a string"}
	}
	if value == nil {
		return "", nil
	}
	return strings.TrimSpace(*value), nil
}

func (f reportFields) required(name string) (string, error) {
	value, valueErr := f.optional(name)
	if valueErr != nil {
		return "", valueErr
	}
	if value == "" {
		return "", &fieldError{field: name, reason: "is required"}
	}
	return value, nil
}

type reportMapper struct {
	namespace  string
	fields     reportFields
	receivedAt time.Time
}

func (m *reportMapper) mapReport() (*deploymentReport, error) {
	externalID, idErr := m.fields.required("id")
	if idErr != nil {
		return nil, idErr
	}
	if utf8.RuneCountInString(externalID) > maxNameLength {
		return nil, &fieldError{field: "id", reason: fmt.Sprintf("must be at most %d characters", maxNameLength)}
	}
	serviceValue, serviceName, serviceErr := m.name("service")
	if serviceErr != nil {
		return nil, serviceErr
	}
	environmentValue, environmentName, environmentErr := m.name("environment")
	if environmentErr != nil {
		return nil, environmentErr
	}
	status, statusErr := m.status()
	if statusErr != nil {
		return nil, statusErr
	}

	attrs := projections.DeploymentEventAttributes{
		ExternalID: externalID,
		Status:     status,
		Service:    m.serviceObservation(serviceValue, serviceName),
		Environment: projections.DeploymentEnvironment{
			Name:        environmentName,
			DisplayName: environmentValue,
		},
	}
	if repositoryErr := m.repository(&attrs); repositoryErr != nil {
		return nil, repositoryErr
	}
	if detailsErr := m.details(&attrs); detailsErr != nil {
		return nil, detailsErr
	}
	if timesErr := m.times(&attrs); timesErr != nil {
		return nil, timesErr
	}

	report := &deploymentReport{
		DeploymentRef: deploymentRef(serviceName, environmentName, externalID),
		Deployment:    attrs,
	}
	return report, nil
}

// name reads a required name field, returning its value and normalized name.
func (m *reportMapper) name(field string) (string, string, error) {
	value, valueErr := m.fields.required(field)
	if valueErr != nil {
		return "", "", valueErr
	}
	if utf8.RuneCountInString(value) > maxNameLength {
		return "", "", &fieldError{field: field, reason: fmt.Sprintf("must be at most %d characters", maxNameLength)}
	}
	normalized := projections.NormalizeServiceName(value)
	if normalized == "" {
		return "", "", &fieldError{field: field, reason: "must contain a letter or digit"}
	}
	return value, normalized, nil
}

func (m *reportMapper) status() (string, error) {
	status, statusErr := m.fields.required("status")
	if statusErr != nil {
		return "", statusErr
	}
	switch status {
	case projections.DeploymentStatusStarted, projections.DeploymentStatusSucceeded, projections.DeploymentStatusFailed:
		return status, nil
	default:
		return "", &fieldError{field: "status", reason: "must be started, succeeded or failed"}
	}
}

func (m *reportMapper) serviceObservation(value string, name string) projections.EntityObservation {
	return projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: m.namespace,
			ResourceRef:       "service:" + name,
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: value,
		LinkingAttributes: projections.LinkingAttributes{
			projections.LinkingAttributeServiceName: name,
		},
	}
}

func (m *reportMapper) repository(attrs *projections.DeploymentEventAttributes) error {
	value, valueErr := m.fields.optional("repository")
	if valueErr != nil || value == "" {
		return valueErr
	}
	if !repositoryPattern.MatchString(value) {
		return &fieldError{field: "repository", reason: "must be owner/name, each 1-100 characters of letters, digits, '.', '_' or '-'"}
	}
	fullName := projections.NormalizeRepositoryFullName(value)
	attrs.Repository = &projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: m.namespace,
			ResourceRef:       "repository:" + fullName,
		},
		Category:          kne.CategoryCode,
		Kind:              "repository",
		DisplayName:       value,
		LinkingAttributes: projections.RepositoryLinkingAttributes(value),
	}
	return nil
}

func (m *reportMapper) details(attrs *projections.DeploymentEventAttributes) error {
	sha, shaErr := m.fields.optional("sha")
	if shaErr != nil {
		return shaErr
	}
	if sha != "" && !shaPattern.MatchString(sha) {
		return &fieldError{field: "sha", reason: "must be 40 hexadecimal characters"}
	}
	attrs.Sha = strings.ToLower(sha)

	version, versionErr := m.fields.optional("version")
	if versionErr != nil {
		return versionErr
	}
	if utf8.RuneCountInString(version) > maxNameLength {
		return &fieldError{field: "version", reason: fmt.Sprintf("must be at most %d characters", maxNameLength)}
	}
	attrs.Version = version

	link, linkErr := m.fields.optional("url")
	if linkErr != nil {
		return linkErr
	}
	if link != "" {
		if utf8.RuneCountInString(link) > maxURLLength {
			return &fieldError{field: "url", reason: fmt.Sprintf("must be at most %d characters", maxURLLength)}
		}
		parsed, parseErr := url.Parse(link)
		if parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return &fieldError{field: "url", reason: "must be an http or https URL"}
		}
	}
	attrs.URL = link
	return nil
}

// times reads started_at and finished_at, defaulting the one the status implies to the receive time.
func (m *reportMapper) times(attrs *projections.DeploymentEventAttributes) error {
	startedAt, startedErr := m.time("started_at")
	if startedErr != nil {
		return startedErr
	}
	finishedAt, finishedErr := m.time("finished_at")
	if finishedErr != nil {
		return finishedErr
	}
	if attrs.Status == projections.DeploymentStatusStarted && finishedAt != nil {
		return &fieldError{field: "finished_at", reason: "must be absent from a started report"}
	}
	receivedAt := m.receivedAt
	finishedSent := finishedAt != nil
	if attrs.Finished() && finishedAt == nil {
		finishedAt = &receivedAt
	}
	if !attrs.Finished() && startedAt == nil {
		startedAt = &receivedAt
	}
	// Checked after defaults, so a start later than the receive time that stands in for finished_at is
	// rejected rather than dropped.
	if startedAt != nil && finishedAt != nil && finishedAt.Before(*startedAt) {
		if finishedSent {
			return &fieldError{field: "finished_at", reason: "must not be before started_at"}
		}
		return &fieldError{field: "started_at", reason: "must not be after the time the report is received, which stands in for the missing finished_at"}
	}
	attrs.StartedAt = startedAt
	attrs.FinishedAt = finishedAt
	return nil
}

func (m *reportMapper) time(field string) (*time.Time, error) {
	value, valueErr := m.fields.optional(field)
	if valueErr != nil || value == "" {
		return nil, valueErr
	}
	parsed, parseErr := time.Parse(time.RFC3339, value)
	if parseErr != nil {
		return nil, &fieldError{field: field, reason: "must be an RFC3339 time"}
	}
	if parsed.After(m.receivedAt.Add(maxFutureSkew)) {
		return nil, &fieldError{field: field, reason: "must not be more than 5 minutes after now"}
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

// deploymentRef identifies a deployment by its normalized service and environment and the caller's ID:
// `deployment:` and the SHA-256 hex digest of the JSON-encoded tuple.
func deploymentRef(serviceName, environmentName, externalID string) string {
	encoded, encodeErr := json.Marshal([]string{serviceName, environmentName, externalID})
	if encodeErr != nil {
		// Strings always encode.
		panic(fmt.Sprintf("encode deployment identity: %v", encodeErr))
	}
	sum := sha256.Sum256(encoded)
	return "deployment:" + hex.EncodeToString(sum[:])
}

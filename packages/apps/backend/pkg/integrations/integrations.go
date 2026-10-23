package integrations

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
)

var (
	ErrUnknownIntegration     = fmt.Errorf("%w: unknown integration", errs.ErrNotFound)
	ErrCapabilityNotSupported = fmt.Errorf("%w: integration capability not supported", errs.ErrUnprocessableInput)
	ErrCapabilityNotInstalled = fmt.Errorf("%w: no installed integration has the capability", errs.ErrNotFound)
)

type (
	IntegrationInstallState struct {
		Installed   []rez.InstalledIntegration
		Preferences *ent.OrganizationPreferences
	}

	// IntegrationWithInstallRequirements is implemented by integrations that can only be newly installed in some tenant states.
	// Existing installations can always be updated or removed.
	IntegrationWithInstallRequirements interface {
		CheckInstallRequirements(*IntegrationInstallState) error
	}

	IntegrationWithWebhookHandler interface {
		WebhookHandler() http.Handler
	}

	IntegrationWithProviderEventQuerier interface {
		MakeProviderEventQuerier(*ent.Integration) (rez.ProviderEventQuerier, error)
	}

	IntegrationWithAgentToolProvider interface {
		GetAvailableAgentTools(context.Context, []rez.InstalledIntegration, rez.GetAvailableAiAgentToolsParams) ([]ai.Tool, error)
	}

	// IntegrationWithScheduledSync is implemented by integrations whose sources are synced hourly, as well as
	// when an installation is saved.
	IntegrationWithScheduledSync interface {
		ScheduledSyncSources() []string
	}

	// InstalledIntegrationWithMetadata exposes read-only facts about an installation. They are returned to anyone
	// who can see the installation, so they must never be secret.
	InstalledIntegrationWithMetadata interface {
		Metadata() map[string]string
	}
)

// As returns the installed integration as capability T, or ErrCapabilityNotSupported.
func As[T any](ii rez.InstalledIntegration) (T, error) {
	capability, supported := ii.(T)
	if !supported {
		return capability, fmt.Errorf("%w: %s does not implement %s", ErrCapabilityNotSupported, ii.Integration().Name, reflect.TypeFor[T]())
	}
	return capability, nil
}

// LookupInstallationAs returns the installation as capability T.
// It returns errs.ErrNotFound or ErrCapabilityNotSupported.
func LookupInstallationAs[T any](ctx context.Context, getter rez.InstalledIntegrationGetter, id uuid.UUID) (T, error) {
	ii, getErr := getter.GetInstalledIntegration(ctx, id)
	if getErr != nil {
		var zero T
		return zero, getErr
	}
	return As[T](ii)
}

// FindInstallationAs returns the organization's one installation with capability T, and the installation as T.
// It returns ErrCapabilityNotInstalled when there is none, and an error when there is more than one.
func FindInstallationAs[T any](ctx context.Context, getter rez.InstalledIntegrationGetter) (rez.InstalledIntegration, T, error) {
	var zero T
	installed, listErr := getter.ListAllInstalled(ctx)
	if listErr != nil {
		return nil, zero, fmt.Errorf("list installed integrations: %w", listErr)
	}
	var found []rez.InstalledIntegration
	for _, ii := range installed {
		if _, supported := ii.(T); supported {
			found = append(found, ii)
		}
	}
	switch len(found) {
	case 0:
		return nil, zero, fmt.Errorf("%w: %s", ErrCapabilityNotInstalled, reflect.TypeFor[T]())
	case 1:
		return found[0], found[0].(T), nil
	default:
		return nil, zero, fmt.Errorf("%d installations implement %s", len(found), reflect.TypeFor[T]())
	}
}

// LogSource is the organization's log reader.
type LogSource struct {
	Installation *ent.Integration
	Querier      LogQuerier
}

// LookupLogSource returns the organization's log reader. When logs cannot be read, it returns an
// *UnavailableError saying why, worked out without calling a provider.
func LookupLogSource(ctx context.Context, getter rez.InstalledIntegrationGetter) (*LogSource, error) {
	ii, querier, findErr := FindInstallationAs[LogQuerier](ctx, getter)
	if findErr != nil {
		if errors.Is(findErr, ErrCapabilityNotInstalled) {
			return nil, &UnavailableError{Reason: "no logs integration installed"}
		}
		return nil, findErr
	}
	if availableErr := querier.CheckLogsAvailable(); availableErr != nil {
		return nil, &UnavailableError{Reason: availableErr.Error()}
	}
	return &LogSource{Installation: ii.Integration(), Querier: querier}, nil
}

// UnavailableError says why a source cannot be read.
type UnavailableError struct {
	Reason string
}

func (e *UnavailableError) Error() string {
	return e.Reason
}

const (
	// CapabilityDeployments is what an installation that reports deployments, such as a webhook with the
	// deployment preset, says it provides.
	CapabilityDeployments = "deployments"
	// CapabilityCodeChanges is what an installation that reports merged code changes, such as a code forge,
	// says it provides.
	CapabilityCodeChanges = "code_changes"
)

// Sources says which sources the organization's installations can read, and why not.
type Sources struct {
	Logs    rezai.SourceAvailability
	Changes rezai.SourceAvailability
	// CodeChanges is whether merged code changes are reported, so deployments can be linked to them.
	CodeChanges rezai.SourceAvailability
}

// CheckSources says which sources can be read, and why not, without calling a provider. An installation that
// cannot be loaded makes the sources unavailable rather than failing.
func CheckSources(ctx context.Context, getter rez.InstalledIntegrationGetter) Sources {
	installed, listErr := getter.ListAllInstalled(ctx)
	if listErr != nil {
		slog.ErrorContext(ctx, "list installed integrations for sources", "error", listErr)
		unavailable := rezai.SourceAvailability{Reason: "installed integrations could not be loaded"}
		return Sources{Logs: unavailable, Changes: unavailable, CodeChanges: unavailable}
	}
	sources := Sources{
		Logs:        rezai.SourceAvailability{Available: true},
		Changes:     rezai.SourceAvailability{Reason: "no deployments integration installed"},
		CodeChanges: rezai.SourceAvailability{Reason: "no code changes integration installed"},
	}
	if _, logsErr := LookupLogSource(ctx, getter); logsErr != nil {
		var unavailable *UnavailableError
		if !errors.As(logsErr, &unavailable) {
			slog.ErrorContext(ctx, "look up log source for sources", "error", logsErr)
			unavailable = &UnavailableError{Reason: "the logs integration could not be loaded"}
		}
		sources.Logs = rezai.SourceAvailability{Reason: unavailable.Reason}
	}
	for _, ii := range installed {
		if slices.Contains(ii.Capabilities(), CapabilityDeployments) {
			sources.Changes = rezai.SourceAvailability{Available: true}
		}
		if slices.Contains(ii.Capabilities(), CapabilityCodeChanges) {
			sources.CodeChanges = rezai.SourceAvailability{Available: true}
		}
	}
	return sources
}

const webhookTokenBytes = 32

// NewWebhookToken returns a random URL-safe token for an installation's webhook URL and the hash stored for it.
// The token itself is never stored.
func NewWebhookToken() (string, []byte, error) {
	raw := make([]byte, webhookTokenBytes)
	if _, readErr := rand.Read(raw); readErr != nil {
		return "", nil, fmt.Errorf("read random token: %w", readErr)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, WebhookTokenHash(token), nil
}

// WebhookTokenHash is the stored form of a webhook URL token.
func WebhookTokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// DecodeUserSettings decodes an installation's user settings by the json tags of T. Unknown keys and mistyped
// values are invalid input; missing settings decode to the zero value.
func DecodeUserSettings[T any](userSettings []byte) (*T, error) {
	var settings T
	if len(userSettings) == 0 {
		return &settings, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(userSettings))
	decoder.DisallowUnknownFields()
	if decodeErr := decoder.Decode(&settings); decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid settings: %w", errs.ErrInvalidInput, decodeErr)
	}
	return &settings, nil
}

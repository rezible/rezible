package integrations

import (
	"bytes"
	"cmp"
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
	"strings"

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
)

type (
	IntegrationInstallState struct {
		Installed   []rez.InstalledIntegration
		Preferences *ent.OrganizationPreferences
		// Target is the new installation's target, or nil when it isn't known yet.
		Target *rez.ProviderResourceRef
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

// DirectMessageSenders are the installations that can send members direct messages.
type DirectMessageSenders struct {
	Senders []DirectMessageSender
	// Unresolved holds the errors of installations that could not be resolved, any of which might send.
	Unresolved []error
}

// ListDirectMessageSenders returns the installations that can send direct messages, in installation order.
// An installation of an integration that is no longer registered is skipped.
func ListDirectMessageSenders(ctx context.Context, installations rez.IntegrationService) (*DirectMessageSenders, error) {
	installed, listErr := installations.ListInstallations(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("list installations: %w", listErr)
	}
	senders := &DirectMessageSenders{}
	for _, installation := range installed {
		ii, resolveErr := installations.AsInstalledIntegration(installation)
		if errors.Is(resolveErr, ErrUnknownIntegration) {
			continue
		}
		if resolveErr != nil {
			senders.Unresolved = append(senders.Unresolved, resolveErr)
			continue
		}
		if sender, supported := ii.(DirectMessageSender); supported {
			senders.Senders = append(senders.Senders, sender)
		}
	}
	return senders, nil
}

// Description says how members are sent direct messages, as the first sender describes it. It is empty when
// there is no sender. known is false when there is no sender but an unresolved installation might be one.
func (s *DirectMessageSenders) Description() (description string, known bool) {
	if len(s.Senders) > 0 {
		return s.Senders[0].DirectMessageDescription(), true
	}
	return "", len(s.Unresolved) == 0
}

// LogSource is an installation that reads logs.
type LogSource struct {
	Installation *ent.Integration
	Querier      LogQuerier
}

// LookupLogSource returns the reader of a service's logs: of the installations that can read logs, the oldest
// that states the service, or else the oldest. statesService says whether an installation states the service.
// When logs cannot be read, it returns an *UnavailableError saying why, worked out without calling a provider.
func LookupLogSource(ctx context.Context, getter rez.InstalledIntegrationGetter, statesService func(*ent.Integration) (bool, error)) (*LogSource, error) {
	installed, listErr := getter.ListAllInstalled(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("list installed integrations: %w", listErr)
	}
	sources, unavailable := logSources(installed)
	if unavailable != nil {
		return nil, unavailable
	}
	for _, source := range sources {
		states, statesErr := statesService(source.Installation)
		if statesErr != nil {
			return nil, fmt.Errorf("check whether installation %s states the service: %w", source.Installation.ID, statesErr)
		}
		if states {
			return source, nil
		}
	}
	return sources[0], nil
}

// logSources returns the installations that can read logs, oldest first by created_at, then ID. When none
// can, it says why: none is installed, or the oldest installation's own reason.
func logSources(installed []rez.InstalledIntegration) ([]*LogSource, *UnavailableError) {
	var candidates []*LogSource
	for _, ii := range installed {
		if querier, supported := ii.(LogQuerier); supported {
			candidates = append(candidates, &LogSource{Installation: ii.Integration(), Querier: querier})
		}
	}
	if len(candidates) == 0 {
		return nil, &UnavailableError{Reason: "no logs integration installed"}
	}
	slices.SortFunc(candidates, func(a, b *LogSource) int {
		return cmp.Or(a.Installation.CreatedAt.Compare(b.Installation.CreatedAt), strings.Compare(a.Installation.ID.String(), b.Installation.ID.String()))
	})

	var sources []*LogSource
	var unavailable *UnavailableError
	for _, candidate := range candidates {
		if availableErr := candidate.Querier.CheckLogsAvailable(); availableErr != nil {
			if unavailable == nil {
				unavailable = &UnavailableError{Reason: availableErr.Error()}
			}
			continue
		}
		sources = append(sources, candidate)
	}
	if len(sources) == 0 {
		return nil, unavailable
	}
	return sources, nil
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
	if _, unavailable := logSources(installed); unavailable != nil {
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

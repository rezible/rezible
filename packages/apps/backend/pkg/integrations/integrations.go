package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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

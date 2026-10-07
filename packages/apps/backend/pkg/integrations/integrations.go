package integrations

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

var (
	ErrUnknownIntegration     = fmt.Errorf("%w: unknown integration", rez.ErrNotFound)
	ErrCapabilityNotSupported = fmt.Errorf("%w: integration capability not supported", rez.ErrUnprocessableInput)
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
// It returns rez.ErrNotFound or ErrCapabilityNotSupported.
func LookupInstallationAs[T any](ctx context.Context, getter rez.InstalledIntegrationGetter, id uuid.UUID) (T, error) {
	ii, getErr := getter.GetInstalledIntegration(ctx, id)
	if getErr != nil {
		var zero T
		return zero, getErr
	}
	return As[T](ii)
}

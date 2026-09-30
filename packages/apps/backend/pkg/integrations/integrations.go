package integrations

import (
	"context"
	"fmt"
	"net/http"

	"github.com/firebase/genkit/go/ai"

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

package http

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/http/oidc"
	oapiv1 "github.com/rezible/rezible/openapi/v1"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (UserAuthSessionService, error) {
		orgs := do.MustInvoke[rez.OrganizationService](i)
		users := do.MustInvoke[rez.UserService](i)
		return oidc.NewAuthSessionService(orgs, users)
	}),
	do.Lazy(func(i do.Injector) (*Server, error) {
		telemetry := do.MustInvoke[rez.TelemetryService](i)
		auth := do.MustInvoke[UserAuthSessionService](i)
		v1Handler := do.MustInvoke[oapiv1.Handler](i)
		ipr := do.MustInvoke[*integrations.PackageRegistry](i)
		return NewServer(telemetry, auth, v1Handler, ipr.GetWebhookHandlers())
	}),
)

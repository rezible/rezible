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
		return oidc.NewAuthSessionService(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (*Server, error) {
		return NewServer(
			do.MustInvoke[rez.Config](i),
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[UserAuthSessionService](i),
			do.MustInvoke[oapiv1.Handler](i),
			do.MustInvoke[*integrations.PackageRegistry](i).GetWebhookHandlers(),
		)
	}),
)

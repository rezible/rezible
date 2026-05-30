package apiv1

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (oapiv1.Handler, error) {
		handler := NewHandler(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.AlertService](i),
			do.MustInvoke[rez.OrganizationService](i),
			do.MustInvoke[rez.UserService](i),
			do.MustInvoke[rez.DocumentsService](i),
			do.MustInvoke[rez.DebriefService](i),
			do.MustInvoke[rez.IncidentService](i),
			do.MustInvoke[rez.IntegrationsService](i),
			do.MustInvoke[rez.EventsService](i),
			do.MustInvoke[rez.EventAnnotationsService](i),
			do.MustInvoke[rez.OncallRostersService](i),
			do.MustInvoke[rez.OncallShiftsService](i),
			do.MustInvoke[rez.OncallMetricsService](i),
			do.MustInvoke[rez.PlaybookService](i),
			do.MustInvoke[rez.RetrospectiveService](i),
			do.MustInvoke[rez.SystemTopologyService](i),
		)
		return handler, nil
	}))

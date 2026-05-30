package db

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (rez.IntegrationsService, error) {
		return NewIntegrationsService(
			do.MustInvoke[rez.ConfigLoader](i),
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[*integrations.PackageRegistry](i))
	}),
	do.Lazy(func(i do.Injector) (rez.ProviderEventService, error) {
		return NewProviderEventService(
			do.MustInvoke[rez.TelemetryService](i),
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationsService](i),
			do.MustInvoke[*projections.EventProjectionHandlerRegistry](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.OrganizationService, error) {
		return NewOrganizationsService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.UserService, error) {
		return NewUserService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.OrganizationService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.TeamService, error) {
		return NewTeamService(do.MustInvoke[*ent.Client](i))
	}),
	do.Lazy(func(i do.Injector) (rez.EventsService, error) {
		return NewEventsService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.UserService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.EventAnnotationsService, error) {
		return NewEventAnnotationsService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.EventsService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.IncidentService, error) {
		return NewIncidentService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.UserService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.OncallRostersService, error) {
		return NewOncallRostersService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.SystemTopologyService, error) {
		return NewSystemTopologyService(
			do.MustInvoke[*ent.Client](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.OncallShiftsService, error) {
		return NewOncallShiftsService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
			do.MustInvoke[rez.IntegrationsService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.OncallMetricsService, error) {
		return NewOncallMetricsService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.OncallShiftsService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.DebriefService, error) {
		return NewDebriefService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.RetrospectiveService, error) {
		return NewRetrospectiveService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.MessageService](i),
			do.MustInvoke[rez.IncidentService](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.AlertService, error) {
		return NewAlertService(do.MustInvoke[*ent.Client](i))
	}),
	do.Lazy(func(i do.Injector) (rez.PlaybookService, error) {
		return NewPlaybookService(do.MustInvoke[*ent.Client](i))
	}),
	do.Lazy(func(i do.Injector) (rez.DocumentsService, error) {
		return NewDocumentsService(
			do.MustInvoke[rez.ConfigLoader](i),
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.TeamService](i),
		)
	}),
)

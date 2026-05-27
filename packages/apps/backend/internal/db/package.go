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
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobsService](i),
			do.MustInvoke[*integrations.PackageRegistry](i))
	}),
	do.Lazy(func(i do.Injector) (rez.ProviderEventService, error) {
		return NewProviderEventService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobsService](i),
			do.MustInvoke[rez.IntegrationsService](i),
			do.MustInvoke[*projections.EventProjectionHandlerRegistry](i),
		)
	}),
	do.Lazy(func(i do.Injector) (rez.OrganizationService, error) {
		return NewOrganizationsService(
			do.MustInvoke[*ent.Client](i),
			do.MustInvoke[rez.JobsService](i),
		)
	}),
)

/*
	svcs.Organizations, svcErr = db.NewOrganizationsService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewOrganizationsService: %w", svcErr)
	}

	svcs.Users, svcErr = db.NewUserService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewUserService: %w", svcErr)
	}

	svcs.Teams, svcErr = db.NewTeamService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewTeamService: %w", svcErr)
	}

	svcs.Events, svcErr = db.NewEventsService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewEventsService: %w", svcErr)
	}

	svcs.EventAnnotations, svcErr = db.NewEventAnnotationsService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewEventAnnotationsService: %w", svcErr)
	}

	svcs.Incidents, svcErr = db.NewIncidentService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewIncidentService: %w", svcErr)
	}

	svcs.OncallRosters, svcErr = db.NewOncallRostersService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewOncallRostersService: %w", svcErr)
	}

	svcs.Topology, svcErr = db.NewSystemTopologyService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewTopologyService: %w", svcErr)
	}

	svcs.OncallShifts, svcErr = db.NewOncallShiftsService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewOncallShiftsService: %w", svcErr)
	}

	svcs.OncallMetrics, svcErr = db.NewOncallMetricsService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewOncallMetricsService: %w", svcErr)
	}

	svcs.Debriefs, svcErr = db.NewDebriefService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewDebriefService: %w", svcErr)
	}

	svcs.Retros, svcErr = db.NewRetrospectiveService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewRetrospectiveService: %w", svcErr)
	}

	svcs.Alerts, svcErr = db.NewAlertService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewAlertService: %w", svcErr)
	}

	svcs.Playbooks, svcErr = db.NewPlaybookService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewPlaybookService: %w", svcErr)
	}

	svcs.Documents, svcErr = db.NewDocumentsService(svcs)
	if svcErr != nil {
		return fmt.Errorf("db.NewDocumentsService: %w", svcErr)
	}
*/

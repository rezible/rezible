package datasync

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/uuid"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/jobs"
)

type Syncer struct {
	db *ent.Client
}

func NewSyncerService(db *ent.Client) *Syncer {
	return &Syncer{db: db}
}

func (s *Syncer) SyncIntegrationsData(ctx context.Context, args jobs.SyncIntegrationsData) error {
	ctx = makeSyncContext(ctx, args.IgnoreHistory, args.CreateDefaults)

	// Sync a single integration
	if args.IntegrationId != uuid.Nil {
		intg, queryErr := s.db.Integration.Get(ctx, args.IntegrationId)
		if queryErr != nil {
			return fmt.Errorf("query args integrations: %w", queryErr)
		}
		return s.syncData(access.TenantContext(ctx, intg.TenantID), ent.Integrations{intg})
	}

	// Sync all integrations for one organization
	if args.OrganizationId != uuid.Nil {
		org, orgErr := s.db.Organization.Get(ctx, args.OrganizationId)
		if orgErr != nil {
			return fmt.Errorf("query args organization: %w", orgErr)
		}
		return s.syncTenantIntegrations(ctx, org.TenantID)
	}

	// Sync all tenants
	tenantIds, tenantsErr := s.db.Tenant.Query().IDs(ctx)
	if tenantsErr != nil {
		return fmt.Errorf("querying tenants: %w", tenantsErr)
	}
	for _, tenantId := range tenantIds {
		if syncErr := s.syncTenantIntegrations(ctx, tenantId); syncErr != nil {
			slog.Error("failed to sync tenant integrations",
				"error", syncErr,
				"tenantId", tenantId,
			)
		}
	}

	return nil
}

func (s *Syncer) syncTenantIntegrations(sysCtx context.Context, tenantId int) error {
	ctx := access.TenantContext(sysCtx, tenantId)
	intgs, intgsErr := s.db.Integration.Query().All(ctx)
	if intgsErr != nil {
		return fmt.Errorf("querying integrations: %w", intgsErr)
	}
	return s.syncData(ctx, intgs)
}

func (s *Syncer) syncData(ctx context.Context, intgs ent.Integrations) error {
	names := make([]string, len(intgs))
	for i, intg := range intgs {
		names[i] = intg.Name
	}

	usersProvs, usersErr := integrations.GetUserDataProviders(ctx, intgs)
	if usersErr != nil {
		slog.Error("failed to load user data providers", "error", usersErr)
	}
	slog.Debug("sync user data", "providers", len(usersProvs))
	for _, prov := range usersProvs {
		if syncErr := syncUsers(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("user provider (%s): %w", reflect.TypeOf(prov).String(), syncErr)
		}
	}

	teamsProviders, teamsErr := integrations.GetTeamDataProviders(ctx, intgs)
	if teamsErr != nil {
		slog.Error("failed to load teams data providers", "error", teamsErr)
	} else if len(teamsProviders) > 0 {
		for _, teams := range teamsProviders {
			if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
				return fmt.Errorf("teams: %w", syncErr)
			}
		}
	}

	oncallProviders, oncallErr := integrations.GetOncallDataProviders(ctx, intgs)
	if oncallErr != nil {
		slog.Error("failed to load oncall data providers", "error", oncallErr)
	} else if len(oncallProviders) > 0 {
		for _, oncall := range oncallProviders {
			if syncErr := syncOncallRosters(ctx, s.db, oncall); syncErr != nil {
				return fmt.Errorf("oncall rosters: %w", syncErr)
			}
			if syncErr := syncOncallShifts(ctx, s.db, oncall); syncErr != nil {
				return fmt.Errorf("oncall shifts: %w", syncErr)
			}
		}
	}

	componentsProviders, componentsErr := integrations.GetSystemComponentsDataProviders(ctx, intgs)
	if componentsErr != nil {
		slog.Error("failed to load components data providers", "error", componentsErr)
	} else if len(componentsProviders) > 0 {
		for _, components := range componentsProviders {
			if syncErr := syncSystemComponents(ctx, s.db, components); syncErr != nil {
				return fmt.Errorf("system components: %w", syncErr)
			}
		}
	}

	alertsProviders, alertsErr := integrations.GetAlertDataProviders(ctx, intgs)
	if alertsErr != nil {
		slog.Error("failed to load alerts data providers", "error", alertsErr)
	} else if len(alertsProviders) > 0 {
		for _, alerts := range alertsProviders {
			if syncErr := syncAlerts(ctx, s.db, alerts); syncErr != nil {
				return fmt.Errorf("alerts: %w", syncErr)
			}
			if syncErr := syncAlertInstances(ctx, s.db, alerts); syncErr != nil {
				return fmt.Errorf("alert instances: %w", syncErr)
			}
		}
	}

	playbooksProviders, playbooksErr := integrations.GetPlaybookDataProviders(ctx, intgs)
	if playbooksErr != nil {
		slog.Error("failed to load playbooks data providers", "error", playbooksErr)
	} else if len(playbooksProviders) > 0 {
		for _, playbooks := range playbooksProviders {
			if syncErr := syncPlaybooks(ctx, s.db, playbooks); syncErr != nil {
				return fmt.Errorf("playbooks: %w", syncErr)
			}
		}
	}

	incidentsProviders, incidentsErr := integrations.GetIncidentDataProviders(ctx, intgs)
	if incidentsErr != nil {
		slog.Error("failed to load incidents data providers", "error", incidentsErr)
	} else if len(incidentsProviders) > 0 {
		for _, incidents := range incidentsProviders {
			if syncErr := syncIncidentRoles(ctx, s.db, incidents); syncErr != nil {
				return fmt.Errorf("incident roles: %w", syncErr)
			}
			if syncErr := syncIncidents(ctx, s.db, incidents); syncErr != nil {
				return fmt.Errorf("incidents: %w", syncErr)
			}
		}
	}

	if shouldCreateDefaults(ctx) {
		if defaultsErr := syncRequiredDefaultData(ctx, s.db); defaultsErr != nil {
			return fmt.Errorf("create required default data: %w", defaultsErr)
		}
	}

	return nil
}

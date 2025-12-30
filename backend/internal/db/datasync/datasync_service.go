package datasync

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/jobs"
)

type Syncer struct {
	db *ent.Client
}

func NewSyncer(db *ent.Client) *Syncer {
	return &Syncer{db: db}
}

func (s *Syncer) SyncIntegrationsData(ctx context.Context, args jobs.SyncIntegrationsData) error {
	if args.Hard {
		ctx = context.WithValue(ctx, ignoreHistoryKey{}, true)
	}
	if args.CreateDefaults {
		ctx = context.WithValue(ctx, createDefaultsKey{}, true)
	}

	if args.IntegrationId != uuid.Nil {
		intg, queryErr := s.db.Integration.Get(access.SystemContext(ctx), args.IntegrationId)
		if queryErr != nil {
			return fmt.Errorf("query args integrations: %w", queryErr)
		}
		ctx = access.TenantContext(ctx, intg.TenantID)
		return s.syncData(ctx, ent.Integrations{intg})
	}

	if args.OrganizationId != uuid.Nil {
		org, orgErr := s.db.Organization.Get(access.SystemContext(ctx), args.OrganizationId)
		if orgErr != nil {
			return fmt.Errorf("query args organization: %w", orgErr)
		}
		ctx = access.TenantContext(ctx, org.TenantID)
		intgs, intgsErr := s.db.Integration.Query().All(ctx)
		if intgsErr != nil {
			return fmt.Errorf("querying integrations: %w", intgsErr)
		}
		return s.syncData(ctx, intgs)
	}

	// Sync all tenants
	tenantIds, tenantsErr := s.db.Tenant.Query().IDs(access.SystemContext(ctx))
	if tenantsErr != nil {
		return fmt.Errorf("querying tenants: %w", tenantsErr)
	}
	for _, tenantId := range tenantIds {
		tenantCtx := access.TenantContext(ctx, tenantId)
		intgs, intgsErr := s.db.Integration.Query().All(tenantCtx)
		if intgsErr != nil {
			return fmt.Errorf("querying integrations: %w", intgsErr)
		}
		if syncErr := s.syncData(tenantCtx, intgs); syncErr != nil {
			log.Error().
				Err(syncErr).
				Msg("failed to sync integrations data")
		}
	}

	return nil
}

func (s *Syncer) syncData(ctx context.Context, intgs ent.Integrations) error {
	names := make([]string, len(intgs))
	for i, intg := range intgs {
		names[i] = intg.Name
	}

	usersProviders, usersErr := integrations.GetUserDataProviders(ctx, intgs)
	if usersErr != nil {
		log.Error().Err(usersErr).Msg("failed to load user data providers")
	} else if len(usersProviders) > 0 {
		for _, prov := range usersProviders {
			if syncErr := syncUsers(ctx, s.db, prov); syncErr != nil {
				return fmt.Errorf("user provider (name): %w", syncErr)
			}
		}
	}

	teamsProviders, teamsErr := integrations.GetTeamDataProviders(ctx, intgs)
	if teamsErr != nil {
		log.Error().Err(teamsErr).Msg("failed to load teams data providers")
	} else if len(teamsProviders) > 0 {
		for _, teams := range teamsProviders {
			if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
				return fmt.Errorf("teams: %w", syncErr)
			}
		}
	}

	oncallProviders, oncallErr := integrations.GetOncallDataProviders(ctx, intgs)
	if oncallErr != nil {
		log.Error().Err(oncallErr).Msg("failed to load oncall data providers")
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
		log.Error().Err(componentsErr).Msg("failed to load components data providers")
	} else if len(componentsProviders) > 0 {
		for _, components := range componentsProviders {
			if syncErr := syncSystemComponents(ctx, s.db, components); syncErr != nil {
				return fmt.Errorf("system components: %w", syncErr)
			}
		}
	}

	alertsProviders, alertsErr := integrations.GetAlertDataProviders(ctx, intgs)
	if alertsErr != nil {
		log.Error().Err(alertsErr).Msg("failed to load alerts data providers")
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
		log.Error().Err(playbooksErr).Msg("failed to load playbooks data providers")
	} else if len(playbooksProviders) > 0 {
		for _, playbooks := range playbooksProviders {
			if syncErr := syncPlaybooks(ctx, s.db, playbooks); syncErr != nil {
				return fmt.Errorf("playbooks: %w", syncErr)
			}
		}
	}

	incidentsProviders, incidentsErr := integrations.GetIncidentDataProviders(ctx, intgs)
	if incidentsErr != nil {
		log.Error().Err(incidentsErr).Msg("failed to load incidents data providers")
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

	if ShouldCreateDefaults(ctx) {
		if defaultsErr := syncRequiredDefaultData(ctx, s.db); defaultsErr != nil {
			return fmt.Errorf("create required default data: %w", defaultsErr)
		}
	}

	return nil
}

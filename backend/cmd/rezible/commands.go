package main

import (
	"context"
	"fmt"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/internal/documents"
	"github.com/rezible/rezible/internal/river"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/openapi"
)

func printSpecCmd(ctx context.Context, opts *Options) error {
	spec, yamlErr := yaml.Marshal(openapi.MakeDefaultApi(&api.Handler{}).OpenAPI())
	if yamlErr != nil {
		return yamlErr
	}
	fmt.Println(string(spec))
	return nil
}

func withDatabase(ctx context.Context, opts *Options, fn func(db *postgres.Database) error) error {
	db, dbErr := postgres.Open(ctx, opts.DatabaseUrl)
	if dbErr != nil {
		return fmt.Errorf("failed to open database: %w", dbErr)
	}

	defer func(cdb *postgres.Database) {
		if closeErr := cdb.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close database connection")
		}
	}(db)

	return fn(db)
}

func migrateCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		if dbErr := db.RunEntMigrations(ctx); dbErr != nil {
			return fmt.Errorf("failed to run ent migrations: %w", dbErr)
		}

		if riverErr := river.RunMigrations(ctx, db.Pool); riverErr != nil {
			return fmt.Errorf("failed to run river migrations: %w", riverErr)
		}

		return nil
	})
}

func syncCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		const (
			syncUsers     = true
			syncOncall    = true
			syncIncidents = true
		)

		c := db.Client()
		pl := providers.NewProviderLoader(c.ProviderConfig)

		users, usersErr := postgres.NewUserService(c, pl)
		if usersErr != nil {
			return fmt.Errorf("user service: %w", usersErr)
		}

		if syncUsers {
			_, chatErr := documents.NewChatService(ctx, pl, users)
			if chatErr != nil {
				return fmt.Errorf("to create chat: %w", chatErr)
			}
			if syncErr := users.SyncData(ctx); syncErr != nil {
				return fmt.Errorf("users sync failed: %w", syncErr)
			}
		}

		if syncOncall {
			oncall, oncallErr := postgres.NewOncallService(ctx, c, nil, pl, nil, nil, users, nil)
			if oncallErr != nil {
				return fmt.Errorf("postgres.NewOncallService: %w", oncallErr)
			}
			if syncErr := oncall.SyncData(ctx); syncErr != nil {
				return fmt.Errorf("oncall sync failed: %w", syncErr)
			}
		}

		if syncIncidents {
			inc, incErr := postgres.NewIncidentService(ctx, c, nil, pl, nil, nil, users)
			if incErr != nil {
				return fmt.Errorf("postgres.NewIncidentService: %w", incErr)
			}
			if syncErr := inc.SyncData(ctx); syncErr != nil {
				return fmt.Errorf("incidents sync failed: %w", syncErr)
			}
		}

		return nil
	})
}

func loadConfigCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		return providers.LoadFromFile(ctx, db.Client(), ".dev_provider_configs.json")
	})
}

func seedCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		return seedDatabase(ctx, db)
	})
}

func seedDatabase(ctx context.Context, db *postgres.Database) error {
	usr, usrErr := db.Client().User.Query().First(ctx)
	tTeam, teamErr := db.Client().Team.Query().Where(team.Slug("test-team")).Only(ctx)
	if teamErr != nil {
		db.Client().Team.Create().
			SetTimezone("Australia/Perth").
			SetName("Test Team").
			SetSlug("test-team").
			SetChatChannelID("test").
			ExecX(ctx)
	}
	
	if usrErr == nil {
		if !usr.QueryTeams().Where(team.Slug("test-team")).ExistX(ctx) {
			usr.Update().AddTeams(tTeam).ExecX(ctx)
			log.Info().Msg("added user to test team")
		}
	}
	/*
		fakerOpts := []fakeropts.OptionFunc{
			fakeropts.WithStringLanguage(fakerintf.LangENG),
			fakeropts.WithGenerateUniqueValues(true),
		}

		numBetween := func(min, max int) int {
			return min + int(math.Round(float64(max-min)*rand.Float64()))
		}

		return ent.WithTx(ctx, db.Client(), func(tx *ent.Tx) error {
			numTeams := 20
			createTeams := make([]*ent.TeamCreate, numTeams)
			for i := 0; i < numTeams; i++ {
				name := faker.Word(fakerOpts...)
				createTeams[i] = tx.Team.Create().
					SetName(name).
					SetSlug(strings.ToLower(name))
			}
			teams := tx.Team.CreateBulk(createTeams...).SaveX(ctx)

			minUsersPerTeam := 4
			maxUsersPerTeam := 12
			teamUsers := make(map[uuid.UUID][]uuid.UUID)
			var createUsers []*ent.UserCreate
			for _, team := range teams {
				numUsers := numBetween(minUsersPerTeam, maxUsersPerTeam)
				for j := 0; j < numUsers; j++ {
					id := uuid.New()
					name := faker.Name(fakerOpts...)
					createUsers = append(createUsers, tx.User.Create().
						SetID(id).
						SetName(name).
						SetEmail(faker.Email(fakerOpts...)).
						AddTeams(team))
					if _, ok := teamUsers[team.ID]; !ok {
						teamUsers[team.ID] = make([]uuid.UUID, numUsers)
					}
					teamUsers[team.ID][j] = id
				}
			}
			tx.User.CreateBulk(createUsers...).SaveX(ctx)

			var createRosters []*ent.OncallRosterCreate
			createRosters = append(createRosters, tx.OncallRoster.Create().
				SetName("demo-roster").
				SetProviderID("bleh-demo").
				SetTimezone("Australia/Sydney").
				SetSlug("demo-roster"))
			for _, team := range teams {
				createRosters = append(createRosters, tx.OncallRoster.Create().
					SetName(team.Name+"-roster").
					SetSlug(team.Slug+"-roster").
					SetTimezone("Australia/Sydney").
					SetProviderID(team.ID.String()))
			}
			rosters := tx.OncallRoster.CreateBulk(createRosters...).SaveX(ctx)

			var createSchedules []*ent.OncallScheduleCreate
			for _, roster := range rosters {
				createSchedules = append(createSchedules, tx.OncallSchedule.Create().
					SetRoster(roster).
					SetName(roster.Name+"-schedule").
					SetProviderID(roster.ID.String()).
					SetTimezone("Australia/Sydney"))
			}
			schedules := tx.OncallSchedule.CreateBulk(createSchedules...).SaveX(ctx)

			var createScheduleParticipants []*ent.OncallScheduleParticipantCreate
			for i, team := range teams {
				schedule := schedules[i]
				for idx, userId := range teamUsers[team.ID] {
					createScheduleParticipants = append(createScheduleParticipants, tx.OncallScheduleParticipant.Create().
						SetSchedule(schedule).
						SetUserID(userId).
						SetIndex(idx))
				}
			}
			tx.OncallScheduleParticipant.CreateBulk(createScheduleParticipants...).SaveX(ctx)

			maxServicesPerTeam := 3
			var createServices []*ent.ServiceCreate
			for _, team := range teams {
				for j := 0; j < numBetween(1, maxServicesPerTeam); j++ {
					name := faker.Word(fakerOpts...)
					createServices = append(createServices, tx.Service.Create().
						SetName(name).
						SetOwnerTeam(team).
						SetSlug(strings.ToLower(name)))
				}
			}
			svcs := tx.Service.CreateBulk(createServices...).SaveX(ctx)

			sev1 := tx.IncidentSeverity.Create().SetName("Severity 1").SaveX(ctx)
			tx.IncidentSeverity.Create().SetName("Severity 2").ExecX(ctx)
			prodEnv := tx.Environment.Create().SetName("prod").SaveX(ctx)
			tx.Environment.Create().SetName("staging").SaveX(ctx)
			ownerRole := tx.IncidentRole.Create().SetName("Owner").SaveX(ctx)

			numIncidents := 20
			events := []struct {
				Title         string
				OffsetMinutes int
			}{
				{"Incident Declared", 120},
				{"Something Found", 110},
				{"Severity Set", 100},
				{"Impact Mitigated", 70},
				{"Impact Resolved", 40},
			}
			createIncidents := make([]*ent.IncidentCreate, numIncidents)
			createRetrospectives := make([]*ent.RetrospectiveCreate, numIncidents)
			createTeamAssignments := make([]*ent.IncidentTeamAssignmentCreate, numIncidents)
			createRoleAssignments := make([]*ent.IncidentRoleAssignmentCreate, numIncidents)
			createResourceImpact := make([]*ent.IncidentResourceImpactCreate, numIncidents)
			createIncidentEvents := make([]*ent.IncidentEventCreate, numIncidents*len(events))

			now := time.Now()
			for i := 0; i < numIncidents; i++ {
				svc := svcs[numBetween(0, len(svcs)-1)]
				word := faker.Word(fakerOpts...)
				incidentId := uuid.New()
				title := svc.Name + " " + word
				r := rand.Int()
				start := (time.Minute * time.Duration(r)) * time.Minute
				createIncidents[i] = tx.Incident.Create().
					SetID(incidentId).
					SetTitle(title).
					SetOpenedAt(now.Add(-start)).
					SetClosedAt(now.Add(-start).Add(time.Hour * 2)).
					SetSlug(svc.Slug + "-" + strings.ToLower(word)).
					SetProviderID(incidentId.String()).
					SetSummary("summary").
					AddEnvironments(prodEnv).
					SetSeverity(sev1)

				createRetrospectives[i] = tx.Retrospective.Create().
					SetDocumentName(strings.ToLower(word) + "-retrospective").
					SetIncidentID(incidentId)

				teamId := svc.QueryOwnerTeam().OnlyIDX(ctx)
				createTeamAssignments[i] = tx.IncidentTeamAssignment.Create().
					SetIncidentID(incidentId).
					SetTeamID(teamId)

				usr := teamUsers[teamId][numBetween(0, len(teamUsers[teamId])-1)]
				createRoleAssignments[i] = tx.IncidentRoleAssignment.Create().
					SetIncidentID(incidentId).
					SetRole(ownerRole).
					SetUserID(usr)

				createResourceImpact[i] = tx.IncidentResourceImpact.Create().
					SetIncidentID(incidentId).
					SetService(svc)

				//for ei, e := range events {
				//	t := time.Minute * time.Duration(e.OffsetMinutes)
				//	createIncidentEvents[(i*len(events))+ei] = client.IncidentEvent.Create().
				//		SetIncidentID(incidentId).
				//		SetTitle(e.Title).
				//		SetType(incidentevent.TypeIncident).
				//		SetTime(time.Now().Add(-t))
				//}
			}
			tx.Incident.CreateBulk(createIncidents...).SaveX(ctx)
			tx.Retrospective.CreateBulk(createRetrospectives...).SaveX(ctx)
			tx.IncidentTeamAssignment.CreateBulk(createTeamAssignments...).SaveX(ctx)
			tx.IncidentRoleAssignment.CreateBulk(createRoleAssignments...).SaveX(ctx)
			tx.IncidentResourceImpact.CreateBulk(createResourceImpact...).SaveX(ctx)
			tx.IncidentEvent.CreateBulk(createIncidentEvents...).SaveX(ctx)

			return nil
		})
	*/
	return nil
}

func makeExampleIncident(client *ent.Client, e *ent.Environment, sev1 *ent.IncidentSeverity) {
	ctx := context.Background()
	id := uuid.New()
	client.Incident.Create().
		SetID(id).
		SetTitle("Demo Incident").
		SetSlug("demo-incident").
		SetSummary("summary").
		AddEnvironments(e).
		SetSeverity(sev1).
		SaveX(ctx)

	client.Retrospective.Create().
		SetDocumentName("demo-incident-retrospective").
		SetIncidentID(id).
		SaveX(ctx)
}

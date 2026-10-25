package main

import (
	"context"
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/execution"
)

func createDevelopmentSession(ctx context.Context, a *Application) (*ent.UserAuthSession, error) {
	svc := a.mustInvoke[rez.AuthSessionService]()
	sess, sessionErr := svc.CreateFromUserAuthResponse(execution.NewSystemContext(ctx), makeDevelopmentAuthSession())
	if sessionErr != nil {
		return nil, fmt.Errorf("http development session: %w", sessionErr)
	}
	return sess, nil
}

func makeDevelopmentAuthSession() *rez.UserAuthProviderSession {
	return &rez.UserAuthProviderSession{
		User: ent.User{
			Email:          "test@dev.rezible.com",
			Name:           "Dev User",
			AuthProviderID: "dev-user",
		},
		Org: ent.Organization{
			Name:           "Dev Org",
			AuthProviderID: "dev-org",
		},
		Role:      user.RoleAdmin,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

// seedDevelopmentIdentity makes the development user an admin of an organization that has finished setup, when
// the auth skip mode is on.
func seedDevelopmentIdentity(ctx context.Context, a *Application) error {
	if !a.mustInvoke[rez.Config]().HttpServer.Auth.EnableDevSkipMode {
		return nil
	}

	sess, sessErr := createDevelopmentSession(ctx, a)
	if sessErr != nil {
		return fmt.Errorf("seed development identity: %w", sessErr)
	}
	ctx = execution.NewTenantContext(ctx, sess.TenantID)

	orgs := a.mustInvoke[rez.OrganizationService]()
	org, orgErr := orgs.Get(ctx, organization.ID(sess.OrganizationID))
	if orgErr != nil {
		return fmt.Errorf("load development organization: %w", orgErr)
	}

	client := a.mustInvoke[rez.Database]().Client(ctx)

	var prefsErr error
	if org.Edges.Preferences == nil {
		prefsErr = client.OrganizationPreferences.Create().
			SetOrganizationID(org.ID).
			SetInitialSetupAt(time.Now().UTC()).
			Exec(ctx)
	} else if org.Edges.Preferences.InitialSetupAt.IsZero() {
		prefsErr = org.Edges.Preferences.Update().
			SetInitialSetupAt(time.Now().UTC()).
			Exec(ctx)
	}
	if prefsErr != nil {
		return fmt.Errorf("complete development organization setup: %w", prefsErr)
	}

	return nil
}

package oidc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent/user"
)

func TestMakeSessionReadsOrganizationClaims(t *testing.T) {
	claims := func(t *testing.T, accessClaims string) *VerifiedProviderAuthClaims {
		vc := &VerifiedProviderAuthClaims{IdentityClaims: ProviderIdentityTokenClaims{Sub: "user-1"}}
		require.NoError(t, json.Unmarshal([]byte(accessClaims), &vc.AccessClaims))
		return vc
	}

	sess, sessErr := claims(t, `{"org_id":"org-1","org_name":"Acme","org_role":"admin","org_slug":"acme"}`).makeSession(true)
	require.NoError(t, sessErr)
	require.Equal(t, "org-1", sess.Org.AuthProviderID)
	require.Equal(t, "Acme", sess.Org.Name)
	require.Equal(t, user.RoleAdmin, sess.Role)
	require.Equal(t, "user-1", sess.User.AuthProviderID)

	for name, accessClaims := range map[string]string{
		"missing org_id":   `{"org_name":"Acme","org_role":"member"}`,
		"missing org_role": `{"org_id":"org-1","org_name":"Acme"}`,
		"unknown org_role": `{"org_id":"org-1","org_name":"Acme","org_role":"owner"}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, refusedErr := claims(t, accessClaims).makeSession(true)
			require.Error(t, refusedErr)
		})
	}

	single, singleErr := claims(t, `{}`).makeSession(false)
	require.NoError(t, singleErr)
	require.Empty(t, single.Role)
}

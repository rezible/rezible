package privacyrules

import (
	"context"
	"entgo.io/ent/entql"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/privacy"
)

func DenyIfNoAccessContext() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ac := access.FromContext(ctx)
		if ac == nil {
			return privacy.Denyf("access context is missing")
		}
		return privacy.Skip
	})
}

func AllowIfSystem() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ac := access.FromContext(ctx)
		if ac.IsSystem() {
			return privacy.Allow
		}
		return privacy.Skip
	})
}

func FilterTenantRule() privacy.QueryMutationRule {
	type TenantsFilter interface {
		WhereTenantID(entql.IntP)
	}
	return privacy.FilterFunc(func(ctx context.Context, f privacy.Filter) error {
		ac := access.FromContext(ctx)
		tenantId, hasTenant := ac.TenantId()
		if !hasTenant {
			return privacy.Denyf("missing tenant information in access context")
		}
		tf, isFilterable := f.(TenantsFilter)
		if !isFilterable {
			return privacy.Denyf("unexpected filter type %T", f)
		}

		// Make sure that a tenant reads only entities that have an edge to it.
		tf.WhereTenantID(entql.IntEQ(tenantId))

		return privacy.Skip
	})
}

/*
// DenyMismatchedTenants is a rule that runs only on create operations and returns a deny
// decision if the operation tries to add users to groups that are not in the same tenant.
func DenyMismatchedTenants() privacy.MutationRule {
	return privacy.GroupMutationRuleFunc(func(ctx context.Context, m *ent.GroupMutation) error {
		tid, exists := m.TenantID()
		if !exists {
			return privacy.Denyf("missing tenant information in mutation")
		}
		users := m.UsersIDs()
		// If there are no users in the mutation, skip this rule-check.
		if len(users) == 0 {
			return privacy.Skip
		}
		// Query the tenant-ids of all attached users. Expect all users to be connected to the same tenant
		// as the group. Note, we use privacy.DecisionContext to skip the FilterTenantRule defined above.
		ids, err := m.Client().User.Query().Where(user.IDIn(users...)).Select(user.FieldTenantID).Ints(privacy.DecisionContext(ctx, privacy.Allow))
		if err != nil {
			return privacy.Denyf("querying the tenant-ids %v", err)
		}
		if len(ids) != len(users) {
			return privacy.Denyf("one the attached users is not connected to a tenant %v", err)
		}
		for _, id := range ids {
			if id != tid {
				return privacy.Denyf("mismatch tenant-ids for group/users %d != %d", tid, id)
			}
		}
		// Skip to the next privacy rule (equivalent to return nil).
		return privacy.Skip
	})
}
*/

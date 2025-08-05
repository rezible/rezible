package rules

import (
	"context"
	"entgo.io/ent/entql"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/privacy"
)

func DenyIfNoAccessContext() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ac := access.GetAuthContext(ctx)
		if ac == nil {
			return privacy.Denyf("access context is missing")
		}
		return privacy.Skip
	})
}

func AllowIfSystemRole() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ac := access.GetAuthContext(ctx)
		if ac.HasRole(access.RoleSystem) {
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
		ac := access.GetAuthContext(ctx)
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

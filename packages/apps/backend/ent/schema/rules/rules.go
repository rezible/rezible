package rules

import (
	"context"

	"entgo.io/ent/entql"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/privacy"
)

func DenyIfNoAccessScope() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if !access.IsScoped(ctx) {
			return privacy.Denyf("access context is missing")
		}
		return privacy.Skip
	})
}

func DenyIfAnonymous() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if access.IsAnonymous(ctx) {
			return privacy.Deny
		}
		return privacy.Skip
	})
}

func AllowIfSystemRole() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if !access.IsSystem(ctx) {
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
		tf, isFilterable := f.(TenantsFilter)
		if !isFilterable {
			return privacy.Denyf("unexpected filter type %T", f)
		}

		tenantId, tenantSet := access.GetTenantId(ctx)
		if !tenantSet {
			return privacy.Denyf("missing tenant in access context")
		}
		// Make sure that a tenant reads only entities that have an edge to it.
		tf.WhereTenantID(entql.IntEQ(tenantId))

		return privacy.Skip
	})
}

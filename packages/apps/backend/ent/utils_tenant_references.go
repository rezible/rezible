package ent

import (
	"context"

	"entgo.io/ent/privacy"
)

// ReportMissingReferencesKey is the context key set by schema.ReportMissingReferences.
type ReportMissingReferencesKey struct{}

// tenantReferenceQuery is a typed query over the rows an edge's IDs name.
type tenantReferenceQuery[Q any] interface {
	Clone() Q
	Count(context.Context) (int, error)
}

// checkTenantReference returns a not found error when query, run under the caller's tenant, finds fewer than n
// rows. With schema.ReportMissingReferences, a shortfall is recounted across tenants and only a row in another
// tenant is not found; a missing row is left to its foreign key, whose violation names the constraint.
func checkTenantReference[Q tenantReferenceQuery[Q]](ctx context.Context, label string, n int, query Q) error {
	inTenant, countErr := query.Clone().Count(ctx)
	if countErr != nil {
		return countErr
	}
	if inTenant == n {
		return nil
	}
	if reportMissing, _ := ctx.Value(ReportMissingReferencesKey{}).(bool); reportMissing {
		all, allErr := query.Clone().Count(privacy.DecisionContext(ctx, privacy.Allow))
		if allErr != nil {
			return allErr
		}
		if all == inTenant {
			return nil
		}
	}
	return &NotFoundError{label}
}

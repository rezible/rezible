package ent

import (
	"context"
)

// tenantReferenceQuery is a typed query over the rows an edge's IDs name.
type tenantReferenceQuery[Q any] interface {
	Count(context.Context) (int, error)
}

// checkTenantReference returns a not found error when query, run under the caller's tenant, finds fewer than n
// rows.
func checkTenantReference[Q tenantReferenceQuery[Q]](ctx context.Context, label string, n int, query Q) error {
	inTenant, countErr := query.Count(ctx)
	if countErr != nil {
		return countErr
	}
	if inTenant == n {
		return nil
	}
	return &NotFoundError{label}
}

package jobs_test

import (
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/stretchr/testify/require"
)

type tenantJobArgs struct {
	jobs.TenantArgs
	Payload string `json:"payload" river:"unique"`
}

func (*tenantJobArgs) Kind() string { return "test-tenant-job" }

func TestSetExecutionContextArgs(t *testing.T) {
	ctx := execution.NewTenantContext(t.Context(), 101)
	args := &tenantJobArgs{Payload: "event"}
	require.NoError(t, jobs.SetContextualArgs(ctx, args))
	require.Equal(t, 101, args.TenantID)
	require.Equal(t, "event", args.Payload)

	otherCtx := execution.NewTenantContext(t.Context(), 202)
	require.NoError(t, jobs.SetContextualArgs(otherCtx, args))
	require.Equal(t, 202, args.TenantID)

	require.ErrorIs(t, jobs.SetContextualArgs(t.Context(), args), rez.ErrTenantContextMissing)
	require.Equal(t, 202, args.TenantID)

	ordinary := jobs.SyncIntegrationSourceEvents{Sources: []string{"events"}, SyncReason: "manual"}
	want := ordinary
	require.NoError(t, jobs.SetContextualArgs(t.Context(), ordinary))
	require.NoError(t, jobs.SetContextualArgs(ctx, &ordinary))
	require.Equal(t, want, ordinary)
}

func TestProcessProviderEventArgsRequirePointer(t *testing.T) {
	var args jobs.JobArgs = &jobs.ProcessProviderEventArgs{}
	require.Equal(t, "process-provider-event", args.Kind())
	_, valueIsJob := any(jobs.ProcessProviderEventArgs{}).(jobs.JobArgs)
	require.False(t, valueIsJob)
}

package river

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

type jobMetadata struct {
	EncodedExecutionContext []byte `json:"ec"`
}

type accessContextMiddleware struct {
	river.MiddlewareDefaults
}

func (m *accessContextMiddleware) InsertMany(ctx context.Context, params []*rivertype.JobInsertParams, doInner func(context.Context) ([]*rivertype.JobInsertResult, error)) ([]*rivertype.JobInsertResult, error) {
	for _, p := range params {
		encodedExec, encodeErr := m.encodeWorkerExecutionContext(ctx, p.Args)
		if encodeErr != nil {
			return nil, fmt.Errorf("job %q: %w", p.Kind, encodeErr)
		}

		var meta jobMetadata
		var jsonErr error
		if jsonErr = json.Unmarshal(p.Metadata, &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
		}

		meta.EncodedExecutionContext = encodedExec

		p.Metadata, jsonErr = json.Marshal(meta)
		if jsonErr != nil {
			return nil, fmt.Errorf("failed to marshal job metadata: %w", jsonErr)
		}
	}

	return doInner(ctx)
}

// encodeWorkerExecutionContext returns the identity the job's worker runs with, taken from the
// inserting context. System jobs encode nothing and run under the job client's system root context.
func (m *accessContextMiddleware) encodeWorkerExecutionContext(ctx context.Context, args rivertype.JobArgs) ([]byte, error) {
	kind := execution.KindTenant
	if identityArgs, hasIdentity := args.(jobs.ArgsWithWorkerExecutionContext); hasIdentity {
		kind = identityArgs.WorkerExecutionContext()
	}
	caller := execution.GetContext(ctx)
	switch kind {
	case execution.KindTenant:
		tenantID, tenantOK := caller.TenantID()
		if !tenantOK {
			return nil, rez.ErrTenantContextMissing
		}
		return execution.GetContext(execution.NewTenantContext(ctx, tenantID)).Encode()
	case execution.KindUser, execution.KindAiAgent:
		if caller.ActorKind != kind {
			return nil, fmt.Errorf("%w: %s job inserted by %s actor", rez.ErrForbidden, kind, caller.ActorKind)
		}
		return caller.Encode()
	case execution.KindSystem:
		if !caller.IsSystem() {
			return nil, fmt.Errorf("%w: system job inserted by %s actor", rez.ErrForbidden, caller.ActorKind)
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("%w: invalid worker execution context %q", rez.ErrForbidden, kind)
	}
}

func (m *accessContextMiddleware) Work(ctx context.Context, job *rivertype.JobRow, doInner func(context.Context) error) error {
	var meta jobMetadata
	if jsonErr := json.Unmarshal(job.Metadata, &meta); jsonErr != nil {
		return fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
	}
	if len(meta.EncodedExecutionContext) > 0 {
		exec, decodeErr := execution.DecodeContext(meta.EncodedExecutionContext)
		if decodeErr != nil {
			return fmt.Errorf("invalid execution context for job: %w", decodeErr)
		}
		//exec.Provenance.ParentKind = "job"
		//exec.Provenance.ParentID = fmt.Sprintf("%d", job.ID)
		ctx = execution.SetContext(ctx, exec)
	}
	return doInner(ctx)
}

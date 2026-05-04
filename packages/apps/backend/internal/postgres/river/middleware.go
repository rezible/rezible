package river

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rezible/rezible/execution"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type jobMetadata struct {
	EncodedExecutionContext []byte `json:"ec"`
}

type accessContextMiddleware struct {
	river.MiddlewareDefaults
}

func (m *accessContextMiddleware) InsertMany(ctx context.Context, params []*rivertype.JobInsertParams, doInner func(context.Context) ([]*rivertype.JobInsertResult, error)) ([]*rivertype.JobInsertResult, error) {
	encodedExec, encodeErr := execution.FromContext(ctx).Encode()
	if encodeErr != nil {
		return nil, fmt.Errorf("failed to encode execution context: %w", encodeErr)
	}
	for _, p := range params {
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

func (m *accessContextMiddleware) Work(ctx context.Context, job *rivertype.JobRow, doInner func(context.Context) error) error {
	var meta jobMetadata
	if jsonErr := json.Unmarshal(job.Metadata, &meta); jsonErr != nil {
		return fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
	}
	if len(meta.EncodedExecutionContext) > 0 {
		exec, decodeErr := execution.Decode(meta.EncodedExecutionContext)
		if decodeErr != nil {
			return fmt.Errorf("invalid execution context for job: %w", decodeErr)
		}
		exec.Provenance.ParentKind = "job"
		exec.Provenance.ParentID = fmt.Sprintf("%d", job.ID)
		ctx = execution.StoreInContext(ctx, exec)
	}
	return doInner(ctx)
}

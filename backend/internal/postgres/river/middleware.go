package river

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rezible/rezible/access"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type jobMetadata struct {
	EncodedAccessScope []byte `json:"ac"`
}

type accessContextMiddleware struct {
	river.MiddlewareDefaults
}

func (m *accessContextMiddleware) InsertMany(ctx context.Context, params []*rivertype.JobInsertParams, doInner func(context.Context) ([]*rivertype.JobInsertResult, error)) ([]*rivertype.JobInsertResult, error) {
	encodedScope, scopeErr := access.EncodeScope(ctx)
	if scopeErr != nil {
		return nil, fmt.Errorf("failed to encode scope: %w", scopeErr)
	}
	for _, p := range params {
		var meta jobMetadata
		var jsonErr error
		if jsonErr = json.Unmarshal(p.Metadata, &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
		}

		meta.EncodedAccessScope = encodedScope

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
	restoredCtx, restoreErr := access.RestoreScope(ctx, meta.EncodedAccessScope)
	if restoreErr != nil {
		return fmt.Errorf("invalid (anonymous) access context for job")
	}
	return doInner(restoredCtx)
}

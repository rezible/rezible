package river

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rezible/rezible/access"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/rs/zerolog/log"
)

type jobMetadata struct {
	AccessContext access.Context `json:"ac"`
}

type accessContextMiddleware struct {
	river.MiddlewareDefaults
}

func (m *accessContextMiddleware) InsertMany(ctx context.Context, params []*rivertype.JobInsertParams, doInner func(context.Context) ([]*rivertype.JobInsertResult, error)) ([]*rivertype.JobInsertResult, error) {
	ac := access.GetContext(ctx)
	for _, p := range params {
		var meta jobMetadata
		var jsonErr error
		if jsonErr = json.Unmarshal(p.Metadata, &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal job metadata: %w", jsonErr)
		}

		meta.AccessContext = ac

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
	if meta.AccessContext.IsAnonymous() {
		log.Debug().Msg("job access context is anonymous")
	}
	return doInner(access.SetContext(ctx, meta.AccessContext))
}

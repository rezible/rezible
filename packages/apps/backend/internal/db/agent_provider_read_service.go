package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	apr "github.com/rezible/rezible/ent/agentproviderread"
	at "github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
)

type AgentProviderReadService struct {
	db    rez.Database
	clock rez.Clock
}

func NewAgentProviderReadService(db rez.Database, clock rez.Clock) *AgentProviderReadService {
	return &AgentProviderReadService{db: db, clock: clock}
}

func (s *AgentProviderReadService) StartProviderRead(ctx context.Context, params rez.StartAgentProviderReadParams) (*ent.AgentProviderRead, int, error) {
	type started struct {
		read *ent.AgentProviderRead
		used int
	}
	result, txErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*started, error) {
		unlocked, unlockedErr := tx.AgentTurn.Get(ctx, params.AgentTurnID)
		if unlockedErr != nil {
			return nil, fmt.Errorf("load agent turn: %w", unlockedErr)
		}
		// Reads started together in one turn would each see the others' count; the session lock, also taken
		// by reconciliation and the turn worker, makes counting and inserting one step. The turn is read again
		// under the lock, since the worker may have finished it meanwhile.
		if lockErr := acquireAgentSessionTurnLock(ctx, s.db, unlocked.AgentSessionID); lockErr != nil {
			return nil, fmt.Errorf("lock agent session: %w", lockErr)
		}
		turn, turnErr := tx.AgentTurn.Get(ctx, params.AgentTurnID)
		if turnErr != nil {
			return nil, fmt.Errorf("reread agent turn: %w", turnErr)
		}
		if turn.Status != at.StatusRunning {
			return nil, fmt.Errorf("%w: provider reads can only be made by a running turn", errs.ErrConflict)
		}

		used, countErr := s.CountTurnProviderReads(ctx, turn.ID)
		if countErr != nil {
			return nil, countErr
		}
		if used >= rezai.MaxProviderReadsPerTurn {
			return nil, fmt.Errorf("%w: this turn has used its budget of %d provider reads", rezai.ErrToolRefused, rezai.MaxProviderReadsPerTurn)
		}

		createRead := tx.AgentProviderRead.Create().
			SetAgentSessionID(turn.AgentSessionID).
			SetAgentTurnID(turn.ID).
			SetToolName(params.ToolName).
			SetServiceEntityID(params.ServiceEntityID).
			SetInput(params.Input)
		read, createErr := createRead.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("record provider read: %w", createErr)
		}
		return &started{read: read, used: used + 1}, nil
	})
	if txErr != nil {
		return nil, 0, txErr
	}
	return result.read, result.used, nil
}

func (s *AgentProviderReadService) FinishProviderRead(ctx context.Context, params rez.FinishAgentProviderReadParams) error {
	status := apr.StatusSucceeded
	if params.Error != "" {
		status = apr.StatusFailed
	}
	updateRead := s.db.Client(ctx).AgentProviderRead.UpdateOneID(params.ID).
		SetStatus(status).
		SetFinishedAt(s.clock.Now())
	if len(params.Read) > 0 {
		updateRead.SetRead(params.Read)
	}
	if len(params.Result) > 0 {
		updateRead.SetResult(params.Result)
	}
	if params.Error != "" {
		updateRead.SetError(params.Error)
	}
	return updateRead.Exec(ctx)
}

func (s *AgentProviderReadService) CountTurnProviderReads(ctx context.Context, turnID uuid.UUID) (int, error) {
	queryReads := s.db.Client(ctx).AgentProviderRead.Query().
		Where(apr.AgentTurnID(turnID))
	count, countErr := queryReads.Count(ctx)
	if countErr != nil {
		return 0, fmt.Errorf("count turn provider reads: %w", countErr)
	}
	return count, nil
}

package db

import (
	"context"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	inc "github.com/rezible/rezible/ent/incident"
	sit "github.com/rezible/rezible/ent/situation"
	sitact "github.com/rezible/rezible/ent/situationaction"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/situations"
)

const situationIncidentRaiseReason = "linked incident"

// situationVerb changes a locked, unclosed situation at the captured processing time.
type situationVerb func(ctx context.Context, situation *ent.Situation, now time.Time) error

// changeSituation runs a lifecycle verb in a transaction on the situation's locked row, then requests its
// evaluation; a closed situation is a conflict. It returns the situation as GetSituation presents it.
func (s *SituationService) changeSituation(ctx context.Context, id uuid.UUID, verb situationVerb) (*ent.Situation, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Situation, error) {
		locked, lockErr := s.lockSituations(ctx, id)
		if lockErr != nil {
			return nil, lockErr
		}
		if locked[id].ClosedAt != nil {
			return nil, fmt.Errorf("%w: situation is closed", errs.ErrConflict)
		}
		if verbErr := verb(ctx, locked[id], s.clock.Now()); verbErr != nil {
			return nil, verbErr
		}
		if evaluateErr := s.requestEvaluation(ctx, id); evaluateErr != nil {
			return nil, evaluateErr
		}
		return s.GetSituation(ctx, id)
	})
}

func (s *SituationService) RaiseSituation(ctx context.Context, id uuid.UUID, params rez.RaiseSituationParams) (*ent.Situation, error) {
	return s.changeSituation(ctx, id, func(ctx context.Context, situation *ent.Situation, now time.Time) error {
		return s.raiseLocked(ctx, situation, params, now)
	})
}

// raiseLocked raises a candidate, clears a mute and starts a missing investigation when asked. Repeating it
// changes nothing.
func (s *SituationService) raiseLocked(ctx context.Context, situation *ent.Situation, params rez.RaiseSituationParams, now time.Time) error {
	if situation.MutedAt != nil {
		if unmuteErr := s.unmuteLocked(ctx, situation, now); unmuteErr != nil {
			return unmuteErr
		}
	}
	if situation.RaisedAt == nil {
		raiseSituation := s.db.Client(ctx).Situation.UpdateOneID(situation.ID).
			SetRaisedAt(now)
		if raiseErr := raiseSituation.Exec(ctx); raiseErr != nil {
			return fmt.Errorf("raise situation: %w", raiseErr)
		}
		if actionErr := s.recordAction(ctx, situation.ID, sitact.ActionRaised, params.Reason, now); actionErr != nil {
			return actionErr
		}
		situation.RaisedAt = &now
	}
	if !params.StartInvestigation {
		return nil
	}
	queryInvestigation := s.db.Client(ctx).SituationInvestigation.Query().
		Where(siti.SituationID(situation.ID))
	hasInvestigation, queryErr := queryInvestigation.Exist(ctx)
	if queryErr != nil {
		return fmt.Errorf("check situation investigation: %w", queryErr)
	}
	if hasInvestigation {
		return nil
	}
	return s.createInvestigation(ctx, situation.ID)
}

func (s *SituationService) SetSituationMute(ctx context.Context, id uuid.UUID, mute *rez.SituationMute) (*ent.Situation, error) {
	if mute != nil {
		if reasonErr := sit.MuteReasonValidator(mute.Reason); reasonErr != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidInput, reasonErr)
		}
	}
	return s.changeSituation(ctx, id, func(ctx context.Context, situation *ent.Situation, now time.Time) error {
		if mute == nil {
			if situation.MutedAt == nil {
				return nil
			}
			return s.unmuteLocked(ctx, situation, now)
		}
		if situation.MutedAt != nil && *situation.MuteReason == mute.Reason {
			return nil
		}
		muteSituation := s.db.Client(ctx).Situation.UpdateOneID(situation.ID).
			SetMuteReason(mute.Reason)
		if situation.MutedAt == nil {
			muteSituation.SetMutedAt(now)
		}
		if muteErr := muteSituation.Exec(ctx); muteErr != nil {
			return fmt.Errorf("mute situation: %w", muteErr)
		}
		return s.recordAction(ctx, situation.ID, sitact.ActionMuted, string(mute.Reason), now)
	})
}

func (s *SituationService) unmuteLocked(ctx context.Context, situation *ent.Situation, now time.Time) error {
	unmuteSituation := s.db.Client(ctx).Situation.UpdateOneID(situation.ID).
		ClearMutedAt().
		ClearMuteReason()
	if unmuteErr := unmuteSituation.Exec(ctx); unmuteErr != nil {
		return fmt.Errorf("unmute situation: %w", unmuteErr)
	}
	situation.MutedAt = nil
	situation.MuteReason = nil
	return s.recordAction(ctx, situation.ID, sitact.ActionUnmuted, "", now)
}

func (s *SituationService) SetSituationHold(ctx context.Context, id uuid.UUID, hold *rez.SituationHold) (*ent.Situation, error) {
	return s.changeSituation(ctx, id, func(ctx context.Context, situation *ent.Situation, now time.Time) error {
		client := s.db.Client(ctx)
		if hold == nil {
			if situation.HoldUntil == nil {
				return nil
			}
			clearHold := client.Situation.UpdateOneID(situation.ID).
				ClearHoldUntil()
			if clearErr := clearHold.Exec(ctx); clearErr != nil {
				return fmt.Errorf("clear situation hold: %w", clearErr)
			}
			return s.recordAction(ctx, situation.ID, sitact.ActionHoldCleared, "", now)
		}
		until := now.Add(situations.HoldDefault)
		if hold.Until != nil {
			if !hold.Until.After(now) {
				return fmt.Errorf("%w: a hold must end in the future", errs.ErrInvalidInput)
			}
			until = *hold.Until
		}
		if situation.HoldUntil != nil && situation.HoldUntil.Equal(until) {
			return nil
		}
		setHold := client.Situation.UpdateOneID(situation.ID).
			SetHoldUntil(until)
		if setErr := setHold.Exec(ctx); setErr != nil {
			return fmt.Errorf("hold situation: %w", setErr)
		}
		return s.recordAction(ctx, situation.ID, sitact.ActionHeld, "", now)
	})
}

// CloseSituation closes the situation as dismissed when muted, otherwise as stabilized. Closing again is a
// no-op when that decided reason matches the stored one, and a conflict otherwise.
func (s *SituationService) CloseSituation(ctx context.Context, id uuid.UUID, params rez.CloseSituationParams) (*ent.Situation, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Situation, error) {
		locked, lockErr := s.lockSituations(ctx, id)
		if lockErr != nil {
			return nil, lockErr
		}
		situation := locked[id]
		reason := sit.CloseReasonStabilized
		if situation.MutedAt != nil {
			reason = sit.CloseReasonDismissed
		}
		if situation.ClosedAt != nil {
			if *situation.CloseReason != reason {
				return nil, fmt.Errorf("%w: situation is already closed as %s", errs.ErrConflict, *situation.CloseReason)
			}
			return s.GetSituation(ctx, id)
		}
		now := s.clock.Now()
		if closeErr := s.closeLocked(ctx, situation, reason, params.Note, now, now); closeErr != nil {
			return nil, closeErr
		}
		return s.GetSituation(ctx, id)
	})
}

// closeLocked closes the situation at closedAt with the reason and records the decision at the processing
// time: merged for a merge, otherwise closed.
func (s *SituationService) closeLocked(ctx context.Context, situation *ent.Situation, reason sit.CloseReason, note string, closedAt, now time.Time) error {
	closeSituation := s.db.Client(ctx).Situation.UpdateOneID(situation.ID).
		SetClosedAt(closedAt).
		SetCloseReason(reason)
	if closeErr := closeSituation.Exec(ctx); closeErr != nil {
		return fmt.Errorf("close situation: %w", closeErr)
	}
	action := sitact.ActionClosed
	if reason == sit.CloseReasonMerged {
		action = sitact.ActionMerged
	}
	situation.ClosedAt = &closedAt
	situation.CloseReason = &reason
	return s.recordAction(ctx, situation.ID, action, note, now)
}

// SyncIncidentLinks applies an incident write's link changes, in the incident's transaction, and requests
// evaluation of the incident's unclosed situations and those it unlinked: any incident write, such as its
// resolution, can change whether they may close. Linking raises the situation and starts its
// investigation; linking a closed situation is a conflict. Unlinking changes no stage.
func (s *SituationService) SyncIncidentLinks(ctx context.Context, incidentID uuid.UUID, changes rez.IncidentSituationLinkChanges) error {
	added := mapset.NewSet(changes.Added...)
	removed := mapset.NewSet(changes.Removed...).Difference(added)
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if linkErr := s.applyIncidentLinks(ctx, incidentID, added, removed); linkErr != nil {
			return linkErr
		}
		queryAffected := tx.Situation.Query().
			Where(
				sit.ClosedAtIsNil(),
				sit.Or(sit.HasIncidentsWith(inc.ID(incidentID)), sit.IDIn(removed.ToSlice()...)),
			)
		affected, queryErr := queryAffected.IDs(ctx)
		if queryErr != nil {
			return fmt.Errorf("query incident situations: %w", queryErr)
		}
		for _, situationID := range affected {
			if evaluateErr := s.requestEvaluation(ctx, situationID); evaluateErr != nil {
				return evaluateErr
			}
		}
		return nil
	})
}

// applyIncidentLinks links and unlinks the situations for the incident.
func (s *SituationService) applyIncidentLinks(ctx context.Context, incidentID uuid.UUID, added, removed mapset.Set[uuid.UUID]) error {
	if added.IsEmpty() && removed.IsEmpty() {
		return nil
	}
	client := s.db.Client(ctx)
	// The incident row lock keeps concurrent writes to one incident's links in order. It is taken before the
	// situation rows, as everywhere else.
	lockIncident := client.Incident.Query().
		Where(inc.ID(incidentID)).
		ForUpdate()
	if _, lockErr := lockIncident.Only(ctx); lockErr != nil {
		return fmt.Errorf("lock incident: %w", lockErr)
	}
	locked, lockErr := s.lockSituations(ctx, added.Union(removed).ToSlice()...)
	if lockErr != nil {
		return lockErr
	}
	now := s.clock.Now()
	for situationID := range removed.Iter() {
		unlinkIncident := client.Situation.UpdateOneID(situationID).
			RemoveIncidentIDs(incidentID)
		if unlinkErr := unlinkIncident.Exec(ctx); unlinkErr != nil {
			return fmt.Errorf("unlink situation incident: %w", unlinkErr)
		}
	}
	for situationID := range added.Iter() {
		if linkErr := s.linkIncidentLocked(ctx, locked[situationID], incidentID, now); linkErr != nil {
			return linkErr
		}
	}
	return nil
}

func (s *SituationService) linkIncidentLocked(ctx context.Context, situation *ent.Situation, incidentID uuid.UUID, now time.Time) error {
	if situation.ClosedAt != nil {
		return fmt.Errorf("%w: a closed situation cannot be linked to an incident", errs.ErrConflict)
	}
	client := s.db.Client(ctx)
	queryLinked := client.Situation.Query().
		Where(sit.ID(situation.ID), sit.HasIncidentsWith(inc.ID(incidentID)))
	linked, linkedErr := queryLinked.Exist(ctx)
	if linkedErr != nil {
		return fmt.Errorf("check situation incident link: %w", linkedErr)
	}
	if !linked {
		linkIncident := client.Situation.UpdateOneID(situation.ID).
			AddIncidentIDs(incidentID)
		if linkErr := linkIncident.Exec(ctx); linkErr != nil {
			return fmt.Errorf("link situation incident: %w", linkErr)
		}
	}
	raiseParams := rez.RaiseSituationParams{StartInvestigation: true, Reason: situationIncidentRaiseReason}
	return s.raiseLocked(ctx, situation, raiseParams, now)
}

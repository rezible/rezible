package db

import (
	"context"
	"errors"

	rez "github.com/rezible/rezible"
)

func (s *IncidentService) registerMessageHandlers() error {
	eventsErr := s.msgs.AddEventHandlers(
		rez.NewEventHandler("db.IncidentService.OnIncidentUpdate", s.onIncidentUpdate))
	return errors.Join(eventsErr)
}

func (s *IncidentService) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	//msQuery := s.db.IncidentMilestone.Query().
	//	Where(incidentmilestone.IncidentID(ev.IncidentId))
	//milestones, msErr := msQuery.All(ctx)
	//if msErr != nil {
	//	return fmt.Errorf("incident milestone query: %w", msErr)
	//}
	//for _, m := range milestones {
	//	log.Debug().Str("milestone", m.String()).Msg("Incident milestone")
	//}
	return nil
}

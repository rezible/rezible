package db

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ale "github.com/rezible/rezible/ent/alertepisode"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	"github.com/rezible/rezible/pkg/errs"
)

// The methods in this file, with ent.SituationObservation.Ref, are where an observation reference meets the
// schema. A new observation kind adds its field to rez.SituationObservationRef, its edge to
// SituationObservation and to the observation's single-reference check, and a case to each of them. A
// reference only reaches them after its kind loaded from a registered source, so an unhandled kind is a
// programming error.

// setObservationRef sets the observed reference on a new observation.
func (s *SituationService) setObservationRef(create *ent.SituationObservationCreate, ref rez.SituationObservationRef) *ent.SituationObservationCreate {
	switch ref.Kind() {
	case schematypes.SituationObservationKindAlertEpisode:
		return create.SetAlertEpisodeID(ref.AlertEpisodeID)
	}
	panic(fmt.Sprintf("unhandled situation observation kind %q", ref.Kind()))
}

// observesAny matches the observations of any of the references.
func (s *SituationService) observesAny(refs ...rez.SituationObservationRef) predicate.SituationObservation {
	var alertEpisodeIDs []uuid.UUID
	for _, ref := range refs {
		switch ref.Kind() {
		case schematypes.SituationObservationKindAlertEpisode:
			alertEpisodeIDs = append(alertEpisodeIDs, ref.AlertEpisodeID)
		default:
			panic(fmt.Sprintf("unhandled situation observation kind %q", ref.Kind()))
		}
	}
	return sitobs.Or(sitobs.AlertEpisodeIDIn(alertEpisodeIDs...))
}

// observesSource matches the observations of the kind whose source is sourceID.
func (s *SituationService) observesSource(kind schematypes.SituationObservationKind, sourceID uuid.UUID) predicate.SituationObservation {
	switch kind {
	case schematypes.SituationObservationKindAlertEpisode:
		return sitobs.HasAlertEpisodeWith(ale.AlertDefinitionID(sourceID))
	}
	panic(fmt.Sprintf("unhandled situation observation kind %q", kind))
}

// observedRefs lists the references the situation observes.
func (s *SituationService) observedRefs(ctx context.Context, situationID uuid.UUID) ([]rez.SituationObservationRef, error) {
	observations, queryErr := s.db.Client(ctx).SituationObservation.Query().
		Where(sitobs.SituationID(situationID)).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query situation observations: %w", queryErr)
	}
	refs := make([]rez.SituationObservationRef, len(observations))
	for i, observation := range observations {
		refs[i] = observation.Ref()
	}
	return refs, nil
}

// LoadObservations describes the observations by reference, each from the source of its kind. An unknown
// kind or observation is invalid input.
func (s *SituationService) LoadObservations(ctx context.Context, refs []rez.SituationObservationRef, opts rez.LoadSituationObservationsOptions) (map[rez.SituationObservationRef]rez.SituationObservation, error) {
	unique := mapset.NewSet(refs...)
	byKind := make(map[schematypes.SituationObservationKind][]rez.SituationObservationRef)
	for _, ref := range unique.ToSlice() {
		byKind[ref.Kind()] = append(byKind[ref.Kind()], ref)
	}
	observations := make(map[rez.SituationObservationRef]rez.SituationObservation, len(refs))
	for kind, kindRefs := range byKind {
		source, ok := s.sources[kind]
		if !ok {
			return nil, fmt.Errorf("%w: unknown observation kind %q", errs.ErrInvalidInput, kind)
		}
		loaded, loadErr := source.LoadObservations(ctx, kindRefs, opts)
		if loadErr != nil {
			return nil, fmt.Errorf("load %s observations: %w", kind, loadErr)
		}
		for _, observation := range loaded {
			observations[observation.Ref] = observation
		}
	}
	if len(observations) != unique.Cardinality() {
		return nil, fmt.Errorf("%w: observation not found", errs.ErrInvalidInput)
	}
	return observations, nil
}

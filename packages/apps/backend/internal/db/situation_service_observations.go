package db

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alertepisode"
	osc "github.com/rezible/rezible/ent/observationscope"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sitobs "github.com/rezible/rezible/ent/situationobservation"
	"github.com/rezible/rezible/pkg/errs"
)

// The methods in this file, with ent.SituationObservation.Ref and Scope and WithScopes, are where an
// observation reference meets the schema. A new observation kind adds its field to rez.SituationObservationRef,
// its edges and single-reference checks, and a case to each of them; an unhandled kind is a programming error.

// setObservationRef sets the observed reference on a new observation.
func (s *SituationService) setObservationRef(create *ent.SituationObservationCreate, ref rez.SituationObservationRef) *ent.SituationObservationCreate {
	switch ref.Kind() {
	case schematypes.SituationObservationKindAlertEpisode:
		return create.SetAlertEpisodeID(ref.AlertEpisodeID)
	}
	panic(fmt.Sprintf("unhandled situation observation kind %q", ref.Kind()))
}

// setScopeRef sets the observation's reference on a new scope, returning the column that is unique by
// observation.
func (s *SituationService) setScopeRef(create *ent.ObservationScopeCreate, ref rez.SituationObservationRef) (*ent.ObservationScopeCreate, string) {
	switch ref.Kind() {
	case schematypes.SituationObservationKindAlertEpisode:
		return create.SetAlertEpisodeID(ref.AlertEpisodeID), osc.FieldAlertEpisodeID
	}
	panic(fmt.Sprintf("unhandled situation observation kind %q", ref.Kind()))
}

// scopeOf matches the observation's scope.
func (s *SituationService) scopeOf(ref rez.SituationObservationRef) predicate.ObservationScope {
	switch ref.Kind() {
	case schematypes.SituationObservationKindAlertEpisode:
		return osc.AlertEpisodeID(ref.AlertEpisodeID)
	}
	panic(fmt.Sprintf("unhandled situation observation kind %q", ref.Kind()))
}

// hasScopeWith matches the observations whose scope matches.
func (s *SituationService) hasScopeWith(preds ...predicate.ObservationScope) predicate.SituationObservation {
	return sitobs.Or(sitobs.HasAlertEpisodeWith(alertepisode.HasObservationScopeWith(preds...)))
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

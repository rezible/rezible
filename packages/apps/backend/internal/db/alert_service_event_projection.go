package db

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionAlertInstanceObserved = "alert_instance_observed"
	knowledgeEntityKindAlert       = "alert"
)

func (s *AlertService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	if projections.SubjectKindAlertInstance.Matches(event) {
		observed, validationErr := projections.DecodeAlertInstanceEvent(event)
		if validationErr != nil || observed == nil {
			return nil, fmt.Errorf("invalid event: %w", validationErr)
		}
		return s.handleAlertEventProjection(ctx, observed)
	}
	return nil, nil
}

func (s *AlertService) handleAlertEventProjection(ctx context.Context, ae *projections.AlertInstanceEvent) ([]rez.ProjectedEntityRef, error) {
	attrs := ae.Attributes

	alertSubject := ae.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "Alert")
	alertSubject.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindAlert,
		Reference:   attrs.ExternalRef,
		DisplayName: attrs.Title,
		Description: attrs.Description,
	}
	alertObservedEvidence := ent.KnowledgeEvidenceRef{
		Kind:            ke.EvidenceKindObserved,
		Assertion:       assertionAlertInstanceObserved,
		EffectiveAt:     ae.Event.OccurredAt,
		SubjectAliasRef: alertSubject,
	}

	var projEnts []rez.ProjectedEntityRef
	return projEnts, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		kneId, kneIdErr := s.knowledge.IngestDomainEntityEvidence(ctx, ae.Event, alertObservedEvidence)
		if kneIdErr != nil {
			return fmt.Errorf("ingest knowledge entity: %w", kneIdErr)
		}

		// TODO: use regular alert service update flow here instead

		upsert := s.db.Client(ctx).Alert.Create().
			SetKnowledgeEntityID(kneId).
			SetTitle(attrs.Title).
			SetDescription(attrs.Description).
			SetDefinition(attrs.Definition).
			OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
			UpdateNewValues()
		alertId, saveErr := upsert.ID(ctx)
		if saveErr != nil {
			return fmt.Errorf("upsert alert: %w", saveErr)
		}

		projEnts = append(projEnts, rez.ProjectedEntityRef{Kind: "alert", Id: alertId})
		return nil
	})
}

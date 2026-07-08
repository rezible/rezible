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

func (s *AlertService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) ([]rez.ProjectedDomainEntityRef, error) {
	if projections.SubjectKindAlert.Matches(event) {
		observed, validationErr := projections.DecodeAlertEvent(event)
		if validationErr != nil || observed == nil {
			return nil, fmt.Errorf("invalid event: %w", validationErr)
		}
		return s.handleAlertEventProjection(ctx, observed)
	}
	return nil, nil
}

func (s *AlertService) handleAlertEventProjection(ctx context.Context, ae *projections.AlertEvent) ([]rez.ProjectedDomainEntityRef, error) {
	attrs := ae.Attributes

	alertSubjectAliasRef := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           ae.Event.Provider,
		ProviderSubjectRef: ae.Event.ProviderSubjectRef,
	}
	alertObservedEvidence := rez.ProjectedKnowledgeEvidence{
		Kind:        ke.EvidenceKindObserved,
		Assertion:   assertionAlertInstanceObserved,
		EffectiveAt: ae.Event.OccurredAt,
		SubjectAlias: rez.ProjectedKnowledgeEvidenceSubjectAlias{
			Description: "Alert Opened",
			AliasRef:    alertSubjectAliasRef,
			SubjectEntityRef: &ent.KnowledgeEntityRef{
				Kind:        knowledgeEntityKindAlert,
				Reference:   attrs.ExternalRef,
				DisplayName: attrs.Title,
				Description: attrs.Description,
			},
		},
	}
	knowledgeEvidence := []rez.ProjectedKnowledgeEvidence{alertObservedEvidence}

	var projEnts []rez.ProjectedDomainEntityRef
	return projEnts, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		saveKnowledgeErr := s.knowledge.IngestProjectedEventEvidence(ctx, ae.Event, knowledgeEvidence)
		if saveKnowledgeErr != nil {
			return fmt.Errorf("save projected entity: %w", saveKnowledgeErr)
		}

		keId, lookupErr := s.knowledge.LookupEntityIdByAliasRefs(ctx, alertSubjectAliasRef)
		if lookupErr != nil {
			if ent.IsNotFound(lookupErr) {
				return projections.Retryable(fmt.Errorf("alert entity not found: %s", alertSubjectAliasRef.ProviderSubjectRef))
			}
			return fmt.Errorf("lookup alert entity alias: %w", lookupErr)
		}

		// TODO: use regular alert service update flow here instead

		upsert := s.db.Client(ctx).Alert.Create().
			SetKnowledgeEntityID(keId).
			SetTitle(attrs.Title).
			SetDescription(attrs.Description).
			SetDefinition(attrs.Definition).
			OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
			UpdateNewValues()
		alertId, saveErr := upsert.ID(ctx)
		if saveErr != nil {
			return fmt.Errorf("upsert alert: %w", saveErr)
		}
		projEnts = append(projEnts, rez.ProjectedDomainEntityRef{Kind: "alert", Id: alertId})
		return nil
	})
}

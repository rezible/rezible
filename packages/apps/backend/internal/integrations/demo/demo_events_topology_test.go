package demoprovider

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	rez "github.com/rezible/rezible"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
)

func TestDemoTopologyFollowsTheModel(t *testing.T) {
	components := makeDemoTopologyComponents()
	byRef := make(map[string]topologyComponentObservedPayload, len(components))
	hierarchies := mapset.NewSet[knowledgegraph.Hierarchy]()
	for _, component := range components {
		if _, exists := byRef[component.ResourceRef]; exists {
			t.Fatalf("duplicate demo component resource ref %q", component.ResourceRef)
		}
		byRef[component.ResourceRef] = component
		hierarchies.Add(knowledgegraph.CategorySpecOf(component.Category).Hierarchy)
	}
	for _, hierarchy := range knowledgegraph.Hierarchies {
		if !hierarchies.Contains(hierarchy) {
			t.Errorf("demo topology has no entity in the %s hierarchy", hierarchy)
		}
	}

	namesAnotherEntity := func(key string, value any) bool {
		text, isText := value.(string)
		_, isRef := byRef[text]
		return strings.HasSuffix(key, "_ref") || strings.HasSuffix(key, "_team") || (isText && isRef)
	}
	for _, component := range components {
		for key, value := range component.Properties {
			if namesAnotherEntity(key, value) {
				t.Errorf("component %q property %q names another entity", component.ResourceRef, key)
			}
		}
	}

	relationships := makeDemoTopologyRelationships(components)
	seenRelationshipRefs := mapset.NewSet[string]()
	usedPredicates := mapset.NewSet[knr.Predicate]()
	parents := make(map[string]int)
	for _, relationship := range relationships {
		if !seenRelationshipRefs.Add(relationship.ResourceRef) {
			t.Fatalf("duplicate demo relationship resource ref %q", relationship.ResourceRef)
		}
		usedPredicates.Add(relationship.Predicate)
		validErr := knowledgegraph.ValidateRelationship(relationship.Predicate, relationship.Source.Category, relationship.Target.Category)
		if validErr != nil {
			t.Errorf("relationship %q: %v", relationship.ResourceRef, validErr)
		}
		if relationship.Predicate == knr.PredicateContains {
			parents[relationship.Target.ResourceRef]++
		}
		for key, value := range relationship.Properties {
			if namesAnotherEntity(key, value) {
				t.Errorf("relationship %q property %q names another entity", relationship.ResourceRef, key)
			}
		}
	}
	for ref, count := range parents {
		if count > 1 {
			t.Errorf("entity %q has %d contains parents", ref, count)
		}
	}
	notFromTopology := []knr.Predicate{knr.PredicateMemberOf, knr.PredicateImpacts, knr.PredicateTouches}
	for _, spec := range knowledgegraph.PredicateSpecs() {
		if !usedPredicates.Contains(spec.Predicate) && !slices.Contains(notFromTopology, spec.Predicate) {
			t.Errorf("demo topology does not use %s", spec.Predicate)
		}
	}
}

func TestIncidentProcessorLeavesUnknownOptionalDatesNull(t *testing.T) {
	receivedAt := time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC)
	openedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	payload, marshalErr := json.Marshal(incidentObservedPayload{
		ResourceID:  "unknown-dates",
		Title:       "Unknown incident dates",
		SeverityRef: "SEV-3",
		TypeRef:     "Internal Tooling",
		OccurredAt:  openedAt,
	})
	if marshalErr != nil {
		t.Fatalf("marshal demo incident: %v", marshalErr)
	}

	providerEvent := rez.ProviderEvent{
		ProviderNamespace:   integrationName,
		ProviderEventSource: sourceIncidents,
		ReceivedAt:          receivedAt,
		Attributes:          payload,
	}
	events, processErr := (&eventProcessor{event: &providerEvent}).processIncident()
	if processErr != nil {
		t.Fatalf("process demo incident: %v", processErr)
	}
	if len(events) != 1 {
		t.Fatalf("processed %d normalized events, want 1", len(events))
	}
	decoded, decodeErr := projections.DecodeIncidentEvent(events[0])
	if decodeErr != nil {
		t.Fatalf("decode projected incident: %v", decodeErr)
	}
	if decoded.Attributes.SourceUpdatedAt != nil {
		t.Fatalf("source_updated_at = %v, want null", decoded.Attributes.SourceUpdatedAt)
	}
	if !decoded.Attributes.OpenedAt.Equal(openedAt) {
		t.Fatalf("opened_at = %s, want source time %s", decoded.Attributes.OpenedAt, openedAt)
	}
	if !events[0].ReceivedAt.Equal(receivedAt) {
		t.Fatalf("received_at = %s, want receipt time %s", events[0].ReceivedAt, receivedAt)
	}
}

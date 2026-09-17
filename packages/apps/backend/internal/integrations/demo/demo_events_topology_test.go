package demoprovider

import (
	"testing"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

func TestDemoTopologyCoversMapLevelsAndPreservesReferences(t *testing.T) {
	components := makeDemoTopologyComponents()
	byRef := make(map[string]topologyComponentObservedPayload, len(components))
	seenCategories := make(map[kne.Category]bool)
	for _, component := range components {
		if _, exists := byRef[component.ResourceRef]; exists {
			t.Fatalf("duplicate demo component resource ref %q", component.ResourceRef)
		}
		byRef[component.ResourceRef] = component
		seenCategories[component.Category] = true
		if component.Category == kne.Category("domain_object") {
			t.Fatalf("demo topology still emits removed domain_object category for %q", component.ResourceRef)
		}
	}

	for _, category := range []kne.Category{
		kne.CategorySystemFunction,
		kne.CategorySystem,
		kne.CategoryContainer,
		kne.CategoryInfrastructure,
		kne.CategoryComponent,
		kne.CategoryCode,
	} {
		if !seenCategories[category] {
			t.Errorf("demo topology does not cover category %q", category)
		}
	}

	for _, id := range []string{"search_indexer", "order_fulfillment_worker", "email_dispatch_worker"} {
		component, exists := byRef[componentRef(id)]
		if !exists {
			t.Errorf("missing independently running demo worker %q", id)
			continue
		}
		if component.Category != kne.CategoryContainer {
			t.Errorf("demo worker %q has category %q, want container", id, component.Category)
		}
	}

	searchRepository, exists := byRef["demo:code_repositories:search-api"]
	if !exists || searchRepository.Category != kne.CategoryCode {
		t.Fatalf("demo search repository is not a code entity: %#v", searchRepository)
	}
	if demoCodeChangeEvents[0].RepositoryRef != demoCodeRepositoryEvents[0].resourceRef() {
		t.Fatalf("demo code change does not reference the observed repository resource")
	}

	relationships := makeDemoTopologyRelationships(components)
	seenRelationshipRefs := make(map[string]bool, len(relationships))
	containsMemberships := make(map[string]int)
	for _, relationship := range relationships {
		if seenRelationshipRefs[relationship.ResourceRef] {
			t.Fatalf("duplicate demo relationship resource ref %q", relationship.ResourceRef)
		}
		seenRelationshipRefs[relationship.ResourceRef] = true
		if _, exists := byRef[relationship.Source.ResourceRef]; !exists {
			t.Errorf("relationship %q has unknown source %q", relationship.ResourceRef, relationship.Source.ResourceRef)
		}
		if _, exists := byRef[relationship.Target.ResourceRef]; !exists {
			t.Errorf("relationship %q has unknown target %q", relationship.ResourceRef, relationship.Target.ResourceRef)
		}
		if relationship.Predicate == knr.PredicateContains {
			containsMemberships[relationship.Target.ResourceRef]++
		}
	}

	for _, id := range []string{"search_api", "order_events_queue", "production_cluster"} {
		if containsMemberships[componentRef(id)] < 2 {
			t.Errorf("demo shared member %q has fewer than two explicit contains memberships", id)
		}
	}
}

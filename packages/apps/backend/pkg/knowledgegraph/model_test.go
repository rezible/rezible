package knowledgegraph

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/errs"
)

func TestEveryEnumValueHasOneSpec(t *testing.T) {
	categoryCounts := make(map[kne.Category]int)
	for _, spec := range CategorySpecs() {
		assert.NoError(t, kne.CategoryValidator(spec.Category))
		categoryCounts[spec.Category]++
	}
	for _, value := range kne.CategoryValues {
		assert.Equal(t, 1, categoryCounts[kne.Category(value)], value)
	}

	predicateCounts := make(map[knr.Predicate]int)
	for _, spec := range PredicateSpecs() {
		assert.NoError(t, knr.PredicateValidator(spec.Predicate))
		assert.Equal(t, spec.SameHierarchy, len(spec.Sources) == 0 && len(spec.Targets) == 0, spec.Predicate)
		predicateCounts[spec.Predicate]++
	}
	for _, value := range knr.PredicateValues {
		assert.Equal(t, 1, predicateCounts[knr.Predicate(value)], value)
	}
}

func TestValidateRelationship(t *testing.T) {
	const (
		function  = kne.CategorySystemFunction
		system    = kne.CategorySystem
		container = kne.CategoryContainer
		component = kne.CategoryComponent
		infra     = kne.CategoryInfrastructure
		code      = kne.CategoryCode
		actor     = kne.CategoryActor
		process   = kne.CategoryProcess
		concern   = kne.CategoryConcern
		decision  = kne.CategoryDecision
		signal    = kne.CategorySignal
	)
	tests := []struct {
		predicate knr.Predicate
		source    kne.Category
		target    kne.Category
		allowed   bool
	}{
		{knr.PredicateContains, function, system, true},
		{knr.PredicateContains, system, system, true},
		{knr.PredicateContains, system, component, true},
		{knr.PredicateContains, container, container, true},
		{knr.PredicateContains, component, component, true},
		{knr.PredicateContains, container, system, false},
		{knr.PredicateContains, component, container, false},
		{knr.PredicateContains, infra, infra, true},
		{knr.PredicateContains, system, infra, false},
		{knr.PredicateContains, container, code, false},
		{knr.PredicateContains, actor, actor, true},
		{knr.PredicateRunsOn, container, infra, true},
		{knr.PredicateRunsOn, component, infra, true},
		{knr.PredicateRunsOn, actor, infra, false},
		{knr.PredicateRunsOn, container, container, false},
		{knr.PredicateBuiltFrom, container, code, true},
		{knr.PredicateBuiltFrom, system, code, false},
		{knr.PredicateOwns, actor, container, true},
		{knr.PredicateOwns, actor, actor, false},
		{knr.PredicateMemberOf, actor, actor, true},
		{knr.PredicateMemberOf, container, actor, false},
		{knr.PredicateParticipatesIn, actor, process, true},
		{knr.PredicateParticipatesIn, container, process, false},
		{knr.PredicateUses, actor, container, true},
		{knr.PredicateUses, actor, function, true},
		{knr.PredicateUses, container, container, false},
		{knr.PredicateSupports, container, process, true},
		{knr.PredicateSupports, infra, process, true},
		{knr.PredicateSupports, actor, process, false},
		{knr.PredicateCalls, container, infra, true},
		{knr.PredicateCalls, function, container, false},
		{knr.PredicateReadsFrom, component, container, true},
		{knr.PredicateWritesTo, container, code, false},
		{knr.PredicatePublishesTo, container, container, true},
		{knr.PredicateConsumesFrom, actor, container, false},
		{knr.PredicateDependsOn, system, system, true},
		{knr.PredicateDependsOn, process, system, false},
		{knr.PredicateObserves, signal, container, true},
		{knr.PredicateObserves, container, container, false},
		{knr.PredicateObserves, signal, signal, false},
		{knr.PredicateConstrains, concern, container, true},
		{knr.PredicateConstrains, decision, container, true},
		{knr.PredicateConstrains, decision, concern, false},
		{knr.PredicateAddresses, decision, concern, true},
		{knr.PredicateAddresses, concern, concern, false},
		{knr.PredicateMitigates, container, concern, true},
		{knr.PredicateMitigates, concern, concern, false},
	}
	for _, tc := range tests {
		validErr := ValidateRelationship(tc.predicate, tc.source, tc.target)
		if tc.allowed {
			assert.NoError(t, validErr, "%s %s %s", tc.source, tc.predicate, tc.target)
		} else {
			assert.ErrorIs(t, validErr, errs.ErrInvalidInput, "%s %s %s", tc.source, tc.predicate, tc.target)
		}
	}
}

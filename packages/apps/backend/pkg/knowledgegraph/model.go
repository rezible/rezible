package knowledgegraph

import (
	"fmt"
	"slices"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/errs"
)

// Hierarchy is a tree of entities joined by contains. Every category belongs to at most one.
type Hierarchy string

const (
	HierarchyNone         Hierarchy = ""
	HierarchyLogical      Hierarchy = "logical"
	HierarchyDeployment   Hierarchy = "deployment"
	HierarchyCode         Hierarchy = "code"
	HierarchyOrganization Hierarchy = "organization"
	HierarchyProcess      Hierarchy = "process"
	HierarchyConcern      Hierarchy = "concern"
	HierarchyDecision     Hierarchy = "decision"
	HierarchyMonitoring   Hierarchy = "monitoring"
)

// Hierarchies lists every hierarchy, in display order.
var Hierarchies = []Hierarchy{
	HierarchyLogical,
	HierarchyDeployment,
	HierarchyCode,
	HierarchyOrganization,
	HierarchyProcess,
	HierarchyConcern,
	HierarchyDecision,
	HierarchyMonitoring,
}

// Level is a logical entity's position, fixed by its category. Only the logical hierarchy has levels.
type Level int

const (
	NoLevel Level = iota - 1
	LevelLandscape
	LevelSystem
	LevelRuntime
	LevelImplementation
)

// Family groups predicates for display.
type Family string

const (
	FamilyStructure      Family = "structure"
	FamilyPlacement      Family = "placement"
	FamilyImplementation Family = "implementation"
	FamilyOwnership      Family = "ownership"
	FamilyMembership     Family = "membership"
	FamilySupport        Family = "support"
	FamilyFlow           Family = "flow"
	FamilyObservation    Family = "observation"
	FamilyGovernance     Family = "governance"
)

// Dependency says which endpoint of a relationship is affected when the other fails.
type Dependency string

const (
	// DependencySourceOnTarget: if the target fails, the source is affected.
	DependencySourceOnTarget Dependency = "source_on_target"
	// DependencyTargetOnSource: if the source fails, the target is affected.
	DependencyTargetOnSource Dependency = "target_on_source"
	// DependencyMutual: either end may affect the other, weakly or asynchronously.
	DependencyMutual Dependency = "mutual"
	DependencyNone   Dependency = "none"
)

type CategorySpec struct {
	Category  kne.Category
	Hierarchy Hierarchy
	Level     Level
}

type PredicateSpec struct {
	Predicate  knr.Predicate
	Family     Family
	Dependency Dependency
	// RollsUp means the map draws it as a connection between the endpoints' visible ancestors.
	RollsUp bool
	// SameHierarchy means the endpoints must share a hierarchy (contains only); Sources and Targets are empty.
	SameHierarchy bool
	Sources       []kne.Category
	Targets       []kne.Category
}

var categorySpecs = []CategorySpec{
	{kne.CategorySystemFunction, HierarchyLogical, LevelLandscape},
	{kne.CategorySystem, HierarchyLogical, LevelSystem},
	{kne.CategoryContainer, HierarchyLogical, LevelRuntime},
	{kne.CategoryComponent, HierarchyLogical, LevelImplementation},
	{kne.CategoryInfrastructure, HierarchyDeployment, NoLevel},
	{kne.CategoryCode, HierarchyCode, NoLevel},
	{kne.CategoryActor, HierarchyOrganization, NoLevel},
	{kne.CategoryProcess, HierarchyProcess, NoLevel},
	{kne.CategoryConcern, HierarchyConcern, NoLevel},
	{kne.CategoryDecision, HierarchyDecision, NoLevel},
	{kne.CategorySignal, HierarchyMonitoring, NoLevel},
}

func categoriesExcept(excluded ...kne.Category) []kne.Category {
	var categories []kne.Category
	for _, value := range kne.CategoryValues {
		if category := kne.Category(value); !slices.Contains(excluded, category) {
			categories = append(categories, category)
		}
	}
	return categories
}

var (
	logicalAndInfra = []kne.Category{kne.CategorySystem, kne.CategoryContainer, kne.CategoryComponent, kne.CategoryInfrastructure}
	logical         = []kne.Category{kne.CategorySystemFunction, kne.CategorySystem, kne.CategoryContainer, kne.CategoryComponent}
	runnable        = []kne.Category{kne.CategoryContainer, kne.CategoryComponent}
	actor           = []kne.Category{kne.CategoryActor}
)

var predicateSpecs = []PredicateSpec{
	{Predicate: knr.PredicateContains, Family: FamilyStructure, Dependency: DependencyNone, SameHierarchy: true},
	{Predicate: knr.PredicateRunsOn, Family: FamilyPlacement, Dependency: DependencySourceOnTarget,
		Sources: runnable, Targets: []kne.Category{kne.CategoryInfrastructure}},
	{Predicate: knr.PredicateBuiltFrom, Family: FamilyImplementation, Dependency: DependencyNone,
		Sources: runnable, Targets: []kne.Category{kne.CategoryCode}},
	{Predicate: knr.PredicateOwns, Family: FamilyOwnership, Dependency: DependencyNone,
		Sources: actor, Targets: categoriesExcept(kne.CategoryActor)},
	{Predicate: knr.PredicateMemberOf, Family: FamilyMembership, Dependency: DependencyNone,
		Sources: actor, Targets: actor},
	{Predicate: knr.PredicateParticipatesIn, Family: FamilyMembership, Dependency: DependencyNone,
		Sources: actor, Targets: []kne.Category{kne.CategoryProcess}},
	{Predicate: knr.PredicateUses, Family: FamilyFlow, Dependency: DependencySourceOnTarget, RollsUp: true,
		Sources: actor, Targets: logical},
	{Predicate: knr.PredicateSupports, Family: FamilySupport, Dependency: DependencyTargetOnSource,
		Sources: logicalAndInfra, Targets: []kne.Category{kne.CategoryProcess}},
	{Predicate: knr.PredicateCalls, Family: FamilyFlow, Dependency: DependencySourceOnTarget, RollsUp: true,
		Sources: logicalAndInfra, Targets: logicalAndInfra},
	{Predicate: knr.PredicateReadsFrom, Family: FamilyFlow, Dependency: DependencySourceOnTarget, RollsUp: true,
		Sources: logicalAndInfra, Targets: logicalAndInfra},
	{Predicate: knr.PredicateWritesTo, Family: FamilyFlow, Dependency: DependencySourceOnTarget, RollsUp: true,
		Sources: logicalAndInfra, Targets: logicalAndInfra},
	{Predicate: knr.PredicatePublishesTo, Family: FamilyFlow, Dependency: DependencyMutual, RollsUp: true,
		Sources: logicalAndInfra, Targets: logicalAndInfra},
	{Predicate: knr.PredicateConsumesFrom, Family: FamilyFlow, Dependency: DependencyMutual, RollsUp: true,
		Sources: logicalAndInfra, Targets: logicalAndInfra},
	{Predicate: knr.PredicateDependsOn, Family: FamilyFlow, Dependency: DependencySourceOnTarget, RollsUp: true,
		Sources: logicalAndInfra, Targets: logicalAndInfra},
	{Predicate: knr.PredicateObserves, Family: FamilyObservation, Dependency: DependencyNone,
		Sources: []kne.Category{kne.CategorySignal}, Targets: categoriesExcept(kne.CategorySignal)},
	{Predicate: knr.PredicateConstrains, Family: FamilyGovernance, Dependency: DependencyNone,
		Sources: []kne.Category{kne.CategoryConcern, kne.CategoryDecision},
		Targets: categoriesExcept(kne.CategoryConcern, kne.CategoryDecision)},
	{Predicate: knr.PredicateAddresses, Family: FamilyGovernance, Dependency: DependencyNone,
		Sources: []kne.Category{kne.CategoryDecision}, Targets: []kne.Category{kne.CategoryConcern}},
	{Predicate: knr.PredicateMitigates, Family: FamilyGovernance, Dependency: DependencyNone,
		Sources: categoriesExcept(kne.CategoryConcern), Targets: []kne.Category{kne.CategoryConcern}},
}

// CategorySpecOf returns the category's spec. An unknown category has no hierarchy and no level.
func CategorySpecOf(category kne.Category) CategorySpec {
	for _, spec := range categorySpecs {
		if spec.Category == category {
			return spec
		}
	}
	return CategorySpec{Category: category, Hierarchy: HierarchyNone, Level: NoLevel}
}

// PredicateSpecOf returns the predicate's spec. An unknown predicate allows no endpoints.
func PredicateSpecOf(predicate knr.Predicate) PredicateSpec {
	for _, spec := range predicateSpecs {
		if spec.Predicate == predicate {
			return spec
		}
	}
	return PredicateSpec{Predicate: predicate, Dependency: DependencyNone}
}

func CategorySpecs() []CategorySpec {
	return slices.Clone(categorySpecs)
}

func PredicateSpecs() []PredicateSpec {
	return slices.Clone(predicateSpecs)
}

// HierarchyCategories returns the hierarchy's categories, in spec order.
func HierarchyCategories(h Hierarchy) []kne.Category {
	var categories []kne.Category
	for _, spec := range categorySpecs {
		if spec.Hierarchy == h {
			categories = append(categories, spec.Category)
		}
	}
	return categories
}

// Allows reports whether the predicate may connect the categories. It does not check that a contains joins
// two different entities; callers know the IDs.
func (p PredicateSpec) Allows(source, target kne.Category) bool {
	if p.SameHierarchy {
		sourceSpec, targetSpec := CategorySpecOf(source), CategorySpecOf(target)
		if sourceSpec.Hierarchy == HierarchyNone || sourceSpec.Hierarchy != targetSpec.Hierarchy {
			return false
		}
		return sourceSpec.Level <= targetSpec.Level
	}
	return slices.Contains(p.Sources, source) && slices.Contains(p.Targets, target)
}

// ValidateRelationship returns errs.ErrInvalidInput when the predicate may not connect the categories.
func ValidateRelationship(predicate knr.Predicate, source, target kne.Category) error {
	if !PredicateSpecOf(predicate).Allows(source, target) {
		return fmt.Errorf("%w: %s may not connect %s to %s", errs.ErrInvalidInput, predicate, source, target)
	}
	return nil
}

// PredicatesWithDependency returns the predicates whose dependency is any of the given ones.
func PredicatesWithDependency(dependencies ...Dependency) []knr.Predicate {
	var predicates []knr.Predicate
	for _, spec := range predicateSpecs {
		if slices.Contains(dependencies, spec.Dependency) {
			predicates = append(predicates, spec.Predicate)
		}
	}
	return predicates
}

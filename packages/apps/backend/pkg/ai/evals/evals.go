package evals

import (
	"cmp"
	"fmt"
	"slices"

	rezai "github.com/rezible/rezible/pkg/ai"
)

var scenarios = []rezai.EvalScenario{
	&AlertsInsufficientContext{},
}

func List() []rezai.EvalScenarioDefinition {
	definitions := make([]rezai.EvalScenarioDefinition, 0, len(scenarios))
	for _, scenario := range scenarios {
		definitions = append(definitions, scenario.Definition())
	}
	slices.SortFunc(definitions, func(a, b rezai.EvalScenarioDefinition) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return definitions
}

func Lookup(name string) (rezai.EvalScenario, error) {
	for _, scenario := range scenarios {
		if scenario.Definition().Name == name {
			return scenario, nil
		}
	}
	return nil, fmt.Errorf("scenario %q not found", name)
}

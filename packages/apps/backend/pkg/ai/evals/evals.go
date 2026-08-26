package evals

import (
	"cmp"
	"fmt"
	"slices"

	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	scenarioFn         func() rezai.EvalScenario
	scenarioPtr[T any] interface {
		*T
		rezai.EvalScenario
	}
)

func defineScenario[T any, PT scenarioPtr[T]]() scenarioFn {
	return func() rezai.EvalScenario {
		return PT(new(T))
	}
}

var scenarioFuncs = []scenarioFn{
	defineScenario[AlertsInsufficientContext](),
}

func List() []rezai.EvalScenarioDefinition {
	definitions := make([]rezai.EvalScenarioDefinition, 0, len(scenarioFuncs))
	for _, fn := range scenarioFuncs {
		definitions = append(definitions, fn().Definition())
	}
	slices.SortFunc(definitions, func(a, b rezai.EvalScenarioDefinition) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return definitions
}

func Lookup(name string) (rezai.EvalScenario, error) {
	for _, fn := range scenarioFuncs {
		scenario := fn()
		if scenario.Definition().Name == name {
			return scenario, nil
		}
	}
	return nil, fmt.Errorf("scenario %q not found", name)
}

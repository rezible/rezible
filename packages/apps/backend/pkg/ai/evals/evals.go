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
	defineScenario[InvestigationInsufficientContext](),
}

func List() []rezai.EvalScenario {
	definitions := make([]rezai.EvalScenario, 0, len(scenarioFuncs))
	for _, fn := range scenarioFuncs {
		definitions = append(definitions, fn())
	}
	slices.SortFunc(definitions, func(a, b rezai.EvalScenario) int {
		return cmp.Compare(a.Definition().Name, b.Definition().Name)
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

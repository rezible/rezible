package evals

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

// Scenarios are named after their file.
//
//go:embed scenarios/*.json
var scenarioFiles embed.FS

// Scenario is an agent evaluation: the environment it sets up, and phases that act, wait for the agent and
// check.
type Scenario struct {
	Name        string `json:"-"`
	Description string `json:"description"`
	Agent       string `json:"agent"`
	// Events are ingested before the setup.
	Events []ScenarioEvent `json:"events"`
	// Logs are the services' log lines, read through the evaluation integration for the whole scenario.
	Logs []ScenarioLog `json:"logs,omitempty"`
	// Setup is the agent's own, decoded by its harness.
	Setup  json.RawMessage `json:"setup"`
	Phases []ScenarioPhase `json:"phases"`
}

// ScenarioEvent is a normalized event in Rezible's provider-neutral shape: a pkg/projections kind and that
// kind's attributes. The evaluation integration completes its provider coordinates, and those of the entity
// references in its attributes.
type ScenarioEvent struct {
	// Ref names the event within the scenario, and is its provider event ref.
	Ref         string `json:"ref"`
	Kind        string `json:"kind"`
	ResourceRef string `json:"resource_ref"`
	// OccurredAt defaults to the repetition's reference time.
	OccurredAt *time.Time      `json:"occurred_at,omitempty"`
	Attributes json.RawMessage `json:"attributes"`
}

// ScenarioPhase optionally ingests events and asks a question, then waits for the agent and runs its checks.
type ScenarioPhase struct {
	Events []ScenarioEvent `json:"events,omitempty"`
	Ask    string          `json:"ask,omitempty"`
	Check  []CheckSpec     `json:"check"`
}

// acts reports whether the phase gives the agent anything to respond to.
func (p ScenarioPhase) acts() bool {
	return len(p.Events) > 0 || p.Ask != ""
}

// ScenarioTargets returns every embedded scenario as a target, in name order.
func ScenarioTargets() ([]Target, error) {
	names, namesErr := scenarioNames()
	if namesErr != nil {
		return nil, namesErr
	}
	harnesses := newAgentHarnesses(&agentHarnessDeps{})
	targets := make([]Target, 0, len(names))
	for _, name := range names {
		scenario, loadErr := loadScenario(name, time.Now(), harnesses)
		if loadErr != nil {
			return nil, loadErr
		}
		targets = append(targets, Target{Name: name, Kind: TargetKindAgent, Description: scenario.Description})
	}
	return targets, nil
}

// HasScenario reports whether a scenario is embedded under the name.
func HasScenario(name string) bool {
	names, namesErr := scenarioNames()
	return namesErr == nil && slices.Contains(names, name)
}

func scenarioNames() ([]string, error) {
	files, globErr := fs.Glob(scenarioFiles, "scenarios/*.json")
	if globErr != nil {
		return nil, fmt.Errorf("list scenarios: %w", globErr)
	}
	names := make([]string, len(files))
	for index, file := range files {
		names[index] = strings.TrimSuffix(strings.TrimPrefix(file, "scenarios/"), ".json")
	}
	slices.Sort(names)
	return names, nil
}

// loadScenario reads the embedded scenario, renders its times from the reference time and validates it
// against its agent's harness.
func loadScenario(name string, referenceTime time.Time, harnesses map[string]AgentHarness) (*Scenario, error) {
	raw, readErr := scenarioFiles.ReadFile("scenarios/" + name + ".json")
	if readErr != nil {
		return nil, fmt.Errorf("scenario %s: %w", name, readErr)
	}
	scenario, parseErr := parseScenario(raw, referenceTime, harnesses)
	if parseErr != nil {
		return nil, fmt.Errorf("scenario %s: %w", name, parseErr)
	}
	scenario.Name = name
	return scenario, nil
}

// timePlaceholder is {{now}} or {{<duration>}}, such as {{-30m}}: a time relative to the reference time.
var timePlaceholder = regexp.MustCompile(`\{\{([^{}]*)\}\}`)

func parseScenario(raw []byte, referenceTime time.Time, harnesses map[string]AgentHarness) (*Scenario, error) {
	var renderErr error
	rendered := timePlaceholder.ReplaceAllFunc(raw, func(match []byte) []byte {
		offset := strings.TrimSpace(string(match[2 : len(match)-2]))
		at := referenceTime
		if offset != "now" {
			parsed, parseErr := time.ParseDuration(offset)
			if parseErr != nil {
				if renderErr == nil {
					renderErr = fmt.Errorf("time placeholder %s: %w", match, parseErr)
				}
				return match
			}
			at = referenceTime.Add(parsed)
		}
		return []byte(at.UTC().Format(time.RFC3339Nano))
	})
	if renderErr != nil {
		return nil, renderErr
	}
	var scenario Scenario
	if decodeErr := decodeStrict(rendered, &scenario); decodeErr != nil {
		return nil, fmt.Errorf("decode: %w", decodeErr)
	}
	harness, known := harnesses[scenario.Agent]
	if !known {
		return nil, fmt.Errorf("agent %q cannot be evaluated", scenario.Agent)
	}
	if validErr := scenario.validate(harness); validErr != nil {
		return nil, validErr
	}
	for index := range scenario.Events {
		scenario.Events[index].defaultOccurredAt(referenceTime)
	}
	for phaseIndex := range scenario.Phases {
		for index := range scenario.Phases[phaseIndex].Events {
			scenario.Phases[phaseIndex].Events[index].defaultOccurredAt(referenceTime)
		}
	}
	return &scenario, nil
}

func (e *ScenarioEvent) defaultOccurredAt(referenceTime time.Time) {
	if e.OccurredAt == nil {
		at := referenceTime.UTC()
		e.OccurredAt = &at
	}
}

func (s *Scenario) validate(harness AgentHarness) error {
	if strings.TrimSpace(s.Description) == "" {
		return fmt.Errorf("no description")
	}
	refs := mapset.NewSet[string]()
	if eventsErr := validateEvents(s.Events, refs); eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	if logsErr := validateLogs(s.Logs); logsErr != nil {
		return fmt.Errorf("logs: %w", logsErr)
	}
	if setupErr := harness.ValidateSetup(s.Setup, refs.Clone()); setupErr != nil {
		return fmt.Errorf("setup: %w", setupErr)
	}
	if len(s.Phases) == 0 {
		return fmt.Errorf("no phases")
	}
	kinds := mapset.NewSet(harness.CheckKinds()...)
	kinds.Append(genericCheckKinds...)
	for index, phase := range s.Phases {
		if phaseErr := phase.validate(refs, kinds); phaseErr != nil {
			return fmt.Errorf("phase %d: %w", index+1, phaseErr)
		}
	}
	return nil
}

func (p ScenarioPhase) validate(refs mapset.Set[string], kinds mapset.Set[string]) error {
	if eventsErr := validateEvents(p.Events, refs); eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	if len(p.Check) == 0 {
		return fmt.Errorf("no checks")
	}
	ids := mapset.NewSet[string]()
	for index, check := range p.Check {
		if !kinds.Contains(check.Kind) {
			return fmt.Errorf("check %d: the agent has no check kind %q", index+1, check.Kind)
		}
		if checkErr := check.validate(); checkErr != nil {
			return fmt.Errorf("check %d: %w", index+1, checkErr)
		}
		if !ids.Add(check.id()) {
			return fmt.Errorf("check %d: %s appears twice", index+1, check.id())
		}
	}
	return nil
}

func validateEvents(events []ScenarioEvent, refs mapset.Set[string]) error {
	for index, event := range events {
		switch {
		case strings.TrimSpace(event.Ref) == "":
			return fmt.Errorf("event %d has no ref", index+1)
		case strings.TrimSpace(event.Kind) == "":
			return fmt.Errorf("event %s has no kind", event.Ref)
		case strings.TrimSpace(event.ResourceRef) == "":
			return fmt.Errorf("event %s has no resource_ref", event.Ref)
		case len(event.Attributes) == 0:
			return fmt.Errorf("event %s has no attributes", event.Ref)
		case !refs.Add(event.Ref):
			return fmt.Errorf("event ref %s appears twice", event.Ref)
		}
	}
	return nil
}

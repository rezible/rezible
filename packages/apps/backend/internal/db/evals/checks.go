package evals

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

// Check kinds a scenario can name. provider_reads applies to any agent; the rest are the investigation's.
const (
	CheckReportPublished   = "report_published"
	CheckAnswerPublished   = "answer_published"
	CheckOutputCount       = "output_count"
	CheckOutputsCite       = "outputs_cite"
	CheckCitesEntity       = "cites_entity"
	CheckProviderReads     = "provider_reads"
	CheckCitesProviderRead = "cites_provider_read"
	CheckNextChecks        = "next_checks"
	// CheckTurnCompleted is recorded for each turn a phase waits for; a scenario does not name it.
	CheckTurnCompleted = "turn_completed"
)

// CheckSpec is one check named in a scenario, with the parameters its kind takes.
type CheckSpec struct {
	Kind string `json:"kind"`
	// Output is the kind of output output_count counts: "finding" or "hypothesis".
	Output string `json:"output,omitempty"`
	// Min and Max bound a count. Either may be absent, but not both.
	Min        *int           `json:"min,omitempty"`
	Max        *int           `json:"max,omitempty"`
	EntityKind string         `json:"entity_kind,omitempty"`
	Name       string         `json:"name,omitempty"`
	Property   map[string]any `json:"property,omitempty"`
	// Service, Tool and Outcome select provider reads: the read service's display name, the tool, and
	// "succeeded" or "failed".
	Service string `json:"service,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Outcome string `json:"outcome,omitempty"`
}

// genericCheckKinds read any agent's session.
var genericCheckKinds = []string{CheckProviderReads}

// investigationCheckKinds read an investigation's outputs.
var investigationCheckKinds = []string{
	CheckReportPublished, CheckAnswerPublished, CheckOutputCount, CheckOutputsCite, CheckCitesEntity, CheckCitesProviderRead,
	CheckNextChecks,
}

// checkParameters lists the parameters each kind takes.
var checkParameters = map[string][]string{
	CheckReportPublished:   {},
	CheckAnswerPublished:   {},
	CheckOutputCount:       {"output", "min", "max"},
	CheckOutputsCite:       {},
	CheckCitesEntity:       {"entity_kind", "name", "property"},
	CheckProviderReads:     {"service", "tool", "outcome", "min", "max"},
	CheckCitesProviderRead: {"service", "tool"},
	CheckNextChecks:        {"min"},
}

func (c CheckSpec) given() []string {
	var given []string
	add := func(name string, present bool) {
		if present {
			given = append(given, name)
		}
	}
	add("output", c.Output != "")
	add("min", c.Min != nil)
	add("max", c.Max != nil)
	add("entity_kind", c.EntityKind != "")
	add("name", c.Name != "")
	add("property", len(c.Property) > 0)
	add("service", c.Service != "")
	add("tool", c.Tool != "")
	add("outcome", c.Outcome != "")
	return given
}

// UnmarshalJSON reads a check's parameters, filling those its kind assumes when absent: next_checks counts at
// least one. The service already enforces the maximum.
func (c *CheckSpec) UnmarshalJSON(data []byte) error {
	type plain CheckSpec
	var decoded plain
	if decodeErr := decodeStrict(data, &decoded); decodeErr != nil {
		return decodeErr
	}
	*c = CheckSpec(decoded)
	if c.Kind == CheckNextChecks && c.Min == nil {
		one := 1
		c.Min = &one
	}
	return nil
}

func (c CheckSpec) validate() error {
	allowed, known := checkParameters[c.Kind]
	if !known {
		return fmt.Errorf("unknown check kind %q", c.Kind)
	}
	for _, parameter := range c.given() {
		if !slices.Contains(allowed, parameter) {
			return fmt.Errorf("%s does not take %s", c.Kind, parameter)
		}
	}
	switch c.Kind {
	case CheckOutputCount:
		if c.Output != "finding" && c.Output != "hypothesis" {
			return fmt.Errorf("%s: output must be finding or hypothesis", c.Kind)
		}
		return c.validateBounds()
	case CheckProviderReads:
		if c.Outcome != "" && c.Outcome != "succeeded" && c.Outcome != "failed" {
			return fmt.Errorf("%s: outcome must be succeeded or failed", c.Kind)
		}
		return c.validateBounds()
	case CheckNextChecks:
		return c.validateBounds()
	case CheckCitesEntity:
		if c.EntityKind == "" {
			return fmt.Errorf("%s: no entity_kind", c.Kind)
		}
		if (c.Name == "") == (len(c.Property) == 0) {
			return fmt.Errorf("%s: give a name or properties, not both", c.Kind)
		}
	}
	return nil
}

func (c CheckSpec) validateBounds() error {
	if c.Min == nil && c.Max == nil {
		return fmt.Errorf("%s: no min or max", c.Kind)
	}
	if c.Min != nil && c.Max != nil && *c.Min > *c.Max {
		return fmt.Errorf("%s: min is above max", c.Kind)
	}
	return nil
}

// id is the kind and its distinguishing parameters.
func (c CheckSpec) id() string {
	parts := []string{c.Kind}
	switch c.Kind {
	case CheckOutputCount:
		parts = append(parts, c.Output)
	case CheckCitesEntity:
		parts = append(parts, c.EntityKind)
		if c.Name != "" {
			parts = append(parts, c.Name)
		}
		for _, key := range slices.Sorted(maps.Keys(c.Property)) {
			parts = append(parts, fmt.Sprintf("%s=%v", key, c.Property[key]))
		}
	case CheckProviderReads, CheckCitesProviderRead:
		for _, value := range []string{c.Service, c.Tool, c.Outcome} {
			if value != "" {
				parts = append(parts, value)
			}
		}
	}
	return strings.Join(parts, ":")
}

// Records is what checks read about a session: what any agent's session has, and what its agent produced.
type Records struct {
	// Reads are the session's provider reads.
	Reads         []RecordedRead
	Investigation InvestigationRecords
}

// InvestigationRecords is an investigation's published outputs.
type InvestigationRecords struct {
	// Reports holds every report published, oldest first.
	Reports []RecordedOutput
	// Findings, Hypotheses and Answers hold the latest version of each. An answer is a finding written for a
	// person's question.
	Findings   []RecordedOutput
	Hypotheses []RecordedOutput
	Answers    []RecordedOutput
	// LatestCompletedTurnID is the newest turn that completed, or nil.
	LatestCompletedTurnID uuid.UUID
}

// RecordedOutput is one published output and what it cites.
type RecordedOutput struct {
	TurnID   uuid.UUID
	Title    string
	Text     string
	Evidence []CitedEvidence
	Reads    []RecordedRead
	// NextChecks are a report's next checks, each its check and reason.
	NextChecks []string
}

// CitedEvidence is the entities cited evidence is about: its subject, or a relationship's two ends.
type CitedEvidence struct {
	Entities []RecordedEntity
}

type RecordedEntity struct {
	Kind       string
	Name       string
	Properties map[string]any
}

type RecordedRead struct {
	Tool    string
	Service string
	Outcome string
}

func (r InvestigationRecords) cited() []RecordedOutput {
	return slices.Concat(r.Reports, r.Findings, r.Hypotheses, r.Answers)
}

// Evaluate computes the check over the records. checkID is the ID the scenario run records it under.
func (c CheckSpec) Evaluate(checkID string, records Records) Check {
	check := Check{ID: checkID}
	investigation := records.Investigation
	switch c.Kind {
	case CheckReportPublished:
		published := slices.ContainsFunc(investigation.Reports, func(report RecordedOutput) bool {
			return strings.TrimSpace(report.Text) != ""
		})
		check.Passed = published
		check.Expected = "a report with text"
		check.Observed = fmt.Sprintf("%d reports", len(investigation.Reports))
	case CheckAnswerPublished:
		var answers []string
		for _, answer := range investigation.Answers {
			if investigation.LatestCompletedTurnID != uuid.Nil && answer.TurnID == investigation.LatestCompletedTurnID {
				answers = append(answers, answer.Title)
			}
		}
		check.Passed = len(answers) > 0
		check.Expected = "an answer from the latest completed turn"
		check.Observed = answers
	case CheckOutputCount:
		outputs := investigation.Findings
		if c.Output == "hypothesis" {
			outputs = investigation.Hypotheses
		}
		check.Passed = c.withinBounds(len(outputs))
		check.Expected = c.boundsText()
		check.Observed = len(outputs)
	case CheckOutputsCite:
		outputs := slices.Concat(investigation.Findings, investigation.Hypotheses)
		uncited := make([]string, 0)
		for _, output := range outputs {
			if len(output.Evidence) == 0 && len(output.Reads) == 0 {
				uncited = append(uncited, output.Title)
			}
		}
		check.Passed = len(outputs) > 0 && len(uncited) == 0
		check.Expected = "findings or hypotheses, each citing evidence or a provider read"
		check.Observed = map[string]any{"published": len(outputs), "uncited": uncited}
	case CheckCitesEntity:
		check.Passed = slices.ContainsFunc(investigation.cited(), func(output RecordedOutput) bool {
			return slices.ContainsFunc(output.Evidence, func(evidence CitedEvidence) bool {
				return slices.ContainsFunc(evidence.Entities, c.matchesEntity)
			})
		})
		check.Expected = c.entityText()
		check.Observed = citedEntityNames(investigation.cited())
	case CheckProviderReads:
		matching := 0
		for _, read := range records.Reads {
			if c.matchesRead(read) {
				matching++
			}
		}
		check.Passed = c.withinBounds(matching)
		check.Expected = c.boundsText()
		check.Observed = matching
	case CheckCitesProviderRead:
		check.Passed = slices.ContainsFunc(investigation.cited(), func(output RecordedOutput) bool {
			return slices.ContainsFunc(output.Reads, c.matchesRead)
		})
		check.Expected = "an output citing a matching provider read"
		check.Observed = len(records.Reads)
	case CheckNextChecks:
		nextChecks := 0
		if len(investigation.Reports) > 0 {
			nextChecks = len(investigation.Reports[len(investigation.Reports)-1].NextChecks)
		}
		check.Passed = c.withinBounds(nextChecks)
		check.Expected = c.boundsText() + " next checks in the latest report"
		check.Observed = nextChecks
	}
	check.Summary = fmt.Sprintf("Expected %v.", check.Expected)
	if !check.Passed {
		check.Summary = fmt.Sprintf("Expected %v; observed %v.", check.Expected, check.Observed)
	}
	return check
}

func (c CheckSpec) withinBounds(count int) bool {
	return (c.Min == nil || count >= *c.Min) && (c.Max == nil || count <= *c.Max)
}

func (c CheckSpec) boundsText() string {
	switch {
	case c.Min != nil && c.Max != nil:
		return fmt.Sprintf("between %d and %d", *c.Min, *c.Max)
	case c.Min != nil:
		return fmt.Sprintf("at least %d", *c.Min)
	default:
		return fmt.Sprintf("at most %d", *c.Max)
	}
}

func (c CheckSpec) matchesEntity(entity RecordedEntity) bool {
	if entity.Kind != c.EntityKind {
		return false
	}
	if c.Name != "" {
		return entity.Name == c.Name
	}
	for key, want := range c.Property {
		if !reflect.DeepEqual(entity.Properties[key], want) {
			return false
		}
	}
	return true
}

func (c CheckSpec) entityText() string {
	if c.Name != "" {
		return fmt.Sprintf("an output citing the %s %q", c.EntityKind, c.Name)
	}
	return fmt.Sprintf("an output citing a %s with %v", c.EntityKind, c.Property)
}

func (c CheckSpec) matchesRead(read RecordedRead) bool {
	return (c.Service == "" || read.Service == c.Service) &&
		(c.Tool == "" || read.Tool == c.Tool) &&
		(c.Outcome == "" || read.Outcome == c.Outcome)
}

func citedEntityNames(outputs []RecordedOutput) []string {
	names := make([]string, 0)
	for _, output := range outputs {
		for _, evidence := range output.Evidence {
			for _, entity := range evidence.Entities {
				name := entity.Kind + " " + entity.Name
				if !slices.Contains(names, name) {
					names = append(names, name)
				}
			}
		}
	}
	return names
}

// PhaseCheckIDs returns the ID each check of each phase is recorded under: its kind and parameters, with the
// phase appended when an earlier phase already recorded that ID.
func (s *Scenario) PhaseCheckIDs() [][]string {
	seen := mapset.NewSet[string]()
	ids := make([][]string, len(s.Phases))
	for phaseIndex, phase := range s.Phases {
		for _, check := range phase.Check {
			id := check.id()
			if !seen.Add(id) {
				id = fmt.Sprintf("%s@phase%d", id, phaseIndex+1)
			}
			ids[phaseIndex] = append(ids[phaseIndex], id)
		}
	}
	return ids
}

// InvestigationOutputs is what an investigation published, for people to read in a result.
type InvestigationOutputs struct {
	Report     string   `json:"report"`
	NextChecks []string `json:"next_checks"`
	Findings   []string `json:"findings"`
	Hypotheses []string `json:"hypotheses"`
	Answers    []string `json:"answers"`
}

// Outputs returns the latest report and the latest version of each other output.
func (r InvestigationRecords) Outputs() InvestigationOutputs {
	outputs := InvestigationOutputs{NextChecks: []string{}, Findings: []string{}, Hypotheses: []string{}, Answers: []string{}}
	if len(r.Reports) > 0 {
		latest := r.Reports[len(r.Reports)-1]
		outputs.Report = latest.Text
		outputs.NextChecks = append(outputs.NextChecks, latest.NextChecks...)
	}
	titled := func(output RecordedOutput) string {
		return output.Title + "\n" + output.Text
	}
	for _, finding := range r.Findings {
		outputs.Findings = append(outputs.Findings, titled(finding))
	}
	for _, hypothesis := range r.Hypotheses {
		outputs.Hypotheses = append(outputs.Hypotheses, titled(hypothesis))
	}
	for _, answer := range r.Answers {
		outputs.Answers = append(outputs.Answers, titled(answer))
	}
	return outputs
}

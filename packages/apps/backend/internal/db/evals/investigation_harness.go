package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	at "github.com/rezible/rezible/ent/agentturn"
	inv "github.com/rezible/rezible/ent/investigation"
	ifind "github.com/rezible/rezible/ent/investigationfinding"
	ifv "github.com/rezible/rezible/ent/investigationfindingversion"
	ihyp "github.com/rezible/rezible/ent/investigationhypothesis"
	ihv "github.com/rezible/rezible/ent/investigationhypothesisversion"
	ior "github.com/rezible/rezible/ent/investigationoutputreference"
	irep "github.com/rezible/rezible/ent/investigationreport"
	ke "github.com/rezible/rezible/ent/knowledgeentity"
	kev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
)

// investigationHarness evaluates the investigation agent. Its session starts when a situation is raised with
// an investigation.
type investigationHarness struct {
	agentHarnessBase
	situations     rez.SituationService
	investigations rez.InvestigationService
}

func newInvestigationHarness(deps *agentHarnessDeps) *investigationHarness {
	return &investigationHarness{
		db:             deps.db,
		situations:     deps.situations,
		investigations: deps.investigations,
	}
}

func (h *investigationHarness) CheckKinds() []string {
	return investigationCheckKinds
}

// investigationSetup is the situation the investigation starts on: a title and the setup events whose signals it holds.
type investigationSetup struct {
	Situation struct {
		Title   string   `json:"title"`
		Signals []string `json:"signals"`
	} `json:"situation"`
}

func (h *investigationHarness) decodeSetup(raw json.RawMessage, setupEventRefs mapset.Set[string]) (*investigationSetup, error) {
	var setup investigationSetup
	if decodeErr := decodeStrict(raw, &setup); decodeErr != nil {
		return nil, fmt.Errorf("decode: %w", decodeErr)
	}
	if strings.TrimSpace(setup.Situation.Title) == "" {
		return nil, fmt.Errorf("the situation has no title")
	}
	if len(setup.Situation.Signals) == 0 {
		return nil, fmt.Errorf("the situation has no signals")
	}
	for _, ref := range setup.Situation.Signals {
		if !setupEventRefs.Contains(ref) {
			return nil, fmt.Errorf("the situation names %q, which is not a setup event", ref)
		}
	}
	return &setup, nil
}

func (h *investigationHarness) ValidateSetup(raw json.RawMessage, setupEventRefs mapset.Set[string]) error {
	_, decodeErr := h.decodeSetup(raw, setupEventRefs)
	return decodeErr
}

// Setup creates the situation raised with an investigation, and returns the investigation's session.
func (h *investigationHarness) Setup(ctx context.Context, raw json.RawMessage, signals map[string][]uuid.UUID) (uuid.UUID, error) {
	refs := mapset.NewSet[string]()
	for ref := range signals {
		refs.Add(ref)
	}
	setup, decodeErr := h.decodeSetup(raw, refs)
	if decodeErr != nil {
		return uuid.Nil, decodeErr
	}
	var signalIDs []uuid.UUID
	for _, ref := range setup.Situation.Signals {
		if len(signals[ref]) == 0 {
			return uuid.Nil, fmt.Errorf("event %s produced no signal", ref)
		}
		signalIDs = append(signalIDs, signals[ref]...)
	}
	params := rez.CreateSituationParams{
		Title:        setup.Situation.Title,
		SeedEntityID: signalIDs[0],
		ObservationGroups: []rez.SituationObservationGroupParams{
			{Title: setup.Situation.Title, SignalEntityIDs: signalIDs},
		},
		Raise: &rez.RaiseSituationParams{StartInvestigation: true, Reason: "Raised by an evaluation scenario"},
	}
	situation, createErr := h.situations.CreateSituation(ctx, params)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create the situation: %w", createErr)
	}
	queryInvestigation := h.db.Client(ctx).SituationInvestigation.Query().
		Where(siti.SituationID(situation.ID)).
		WithInvestigation()
	link, linkErr := queryInvestigation.Only(ctx)
	if linkErr != nil {
		return uuid.Nil, fmt.Errorf("find the situation's investigation: %w", linkErr)
	}
	return link.Edges.Investigation.AgentSessionID, nil
}

func (h *investigationHarness) investigation(ctx context.Context, sessionID uuid.UUID) (*ent.Investigation, error) {
	queryInvestigation := h.db.Client(ctx).Investigation.Query().
		Where(inv.AgentSessionID(sessionID))
	investigation, queryErr := queryInvestigation.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("find the session's investigation: %w", queryErr)
	}
	return investigation, nil
}

func (h *investigationHarness) Ask(ctx context.Context, sessionID uuid.UUID, text, key string) error {
	investigation, findErr := h.investigation(ctx, sessionID)
	if findErr != nil {
		return findErr
	}
	params := rez.SubmitInvestigationUserInputParams{InvestigationID: investigation.ID, Text: text, SubmissionKey: key}
	if _, submitErr := h.investigations.SubmitInvestigationUserInput(ctx, params); submitErr != nil {
		return fmt.Errorf("submit the question: %w", submitErr)
	}
	return nil
}

func (h *investigationHarness) Outputs(records Records) any {
	return records.Investigation.Outputs()
}

// Records reads the investigation's published outputs, what each cites, and the session's provider reads.
func (h *investigationHarness) Records(ctx context.Context, sessionID uuid.UUID) (Records, error) {
	var records Records
	investigation, findErr := h.investigation(ctx, sessionID)
	if findErr != nil {
		return records, findErr
	}
	client := h.db.Client(ctx)

	queryReports := client.InvestigationReport.Query().
		Where(irep.InvestigationID(investigation.ID)).
		Order(ent.Asc(irep.FieldCreatedAt))
	reports, reportsErr := queryReports.All(ctx)
	if reportsErr != nil {
		return records, fmt.Errorf("query reports: %w", reportsErr)
	}
	// Versions newest first, so each output's first version is its latest.
	queryFindings := client.InvestigationFinding.Query().
		Where(ifind.InvestigationID(investigation.ID)).
		WithVersions(func(query *ent.InvestigationFindingVersionQuery) {
			query.Order(ent.Desc(ifv.FieldCreatedAt), ent.Desc(ifv.FieldID))
		})
	findings, findingsErr := queryFindings.All(ctx)
	if findingsErr != nil {
		return records, fmt.Errorf("query findings: %w", findingsErr)
	}
	queryHypotheses := client.InvestigationHypothesis.Query().
		Where(ihyp.InvestigationID(investigation.ID)).
		WithVersions(func(query *ent.InvestigationHypothesisVersionQuery) {
			query.Order(ent.Desc(ihv.FieldCreatedAt), ent.Desc(ihv.FieldID))
		})
	hypotheses, hypothesesErr := queryHypotheses.All(ctx)
	if hypothesesErr != nil {
		return records, fmt.Errorf("query hypotheses: %w", hypothesesErr)
	}
	reads, readsErr := h.loadReads(ctx, sessionID)
	if readsErr != nil {
		return records, readsErr
	}
	for _, read := range reads {
		records.Reads = append(records.Reads, read)
	}

	reportIDs := make([]uuid.UUID, 0, len(reports))
	for _, report := range reports {
		reportIDs = append(reportIDs, report.ID)
	}
	var findingVersionIDs, hypothesisVersionIDs []uuid.UUID
	for _, finding := range findings {
		if len(finding.Edges.Versions) > 0 {
			findingVersionIDs = append(findingVersionIDs, finding.Edges.Versions[0].ID)
		}
	}
	for _, hypothesis := range hypotheses {
		if len(hypothesis.Edges.Versions) > 0 {
			hypothesisVersionIDs = append(hypothesisVersionIDs, hypothesis.Edges.Versions[0].ID)
		}
	}
	queryReferences := client.InvestigationOutputReference.Query().
		Where(ior.Or(
			ior.ReportIDIn(reportIDs...),
			ior.FindingVersionIDIn(findingVersionIDs...),
			ior.HypothesisVersionIDIn(hypothesisVersionIDs...),
		))
	references, referencesErr := queryReferences.All(ctx)
	if referencesErr != nil {
		return records, fmt.Errorf("query output references: %w", referencesErr)
	}
	var evidenceIDs []uuid.UUID
	for _, reference := range references {
		if reference.KnowledgeEvidenceID != nil {
			evidenceIDs = append(evidenceIDs, *reference.KnowledgeEvidenceID)
		}
	}
	evidence, evidenceErr := h.loadCitedEvidence(ctx, evidenceIDs)
	if evidenceErr != nil {
		return records, evidenceErr
	}
	// cited returns what the output with the given ID cites.
	cited := func(outputID uuid.UUID) ([]CitedEvidence, []RecordedRead) {
		var citedEvidence []CitedEvidence
		var citedReads []RecordedRead
		for _, reference := range references {
			isOutput := (reference.ReportID != nil && *reference.ReportID == outputID) ||
				(reference.FindingVersionID != nil && *reference.FindingVersionID == outputID) ||
				(reference.HypothesisVersionID != nil && *reference.HypothesisVersionID == outputID)
			if !isOutput {
				continue
			}
			if reference.KnowledgeEvidenceID != nil {
				citedEvidence = append(citedEvidence, evidence[*reference.KnowledgeEvidenceID])
			}
			if reference.AgentProviderReadID != nil {
				if read, known := reads[*reference.AgentProviderReadID]; known {
					citedReads = append(citedReads, read)
				}
			}
		}
		return citedEvidence, citedReads
	}

	outputs := &records.Investigation
	for _, report := range reports {
		citedEvidence, citedReads := cited(report.ID)
		output := RecordedOutput{TurnID: report.AgentTurnID, Text: report.Text, Evidence: citedEvidence, Reads: citedReads}
		for _, next := range report.NextChecks {
			output.NextChecks = append(output.NextChecks, next.Check+" — "+next.Reason)
		}
		outputs.Reports = append(outputs.Reports, output)
	}
	for _, finding := range findings {
		if len(finding.Edges.Versions) == 0 {
			continue
		}
		latest := finding.Edges.Versions[0]
		citedEvidence, citedReads := cited(latest.ID)
		output := RecordedOutput{TurnID: latest.AgentTurnID, Title: latest.Title, Text: latest.Body, Evidence: citedEvidence, Reads: citedReads}
		// An answer is a finding written for a person's question.
		if finding.UserInputID != nil {
			outputs.Answers = append(outputs.Answers, output)
		} else {
			outputs.Findings = append(outputs.Findings, output)
		}
	}
	for _, hypothesis := range hypotheses {
		if len(hypothesis.Edges.Versions) == 0 {
			continue
		}
		latest := hypothesis.Edges.Versions[0]
		citedEvidence, citedReads := cited(latest.ID)
		output := RecordedOutput{TurnID: latest.AgentTurnID, Title: latest.Title, Text: latest.Justification, Evidence: citedEvidence, Reads: citedReads}
		outputs.Hypotheses = append(outputs.Hypotheses, output)
	}

	queryLatestTurn := client.AgentTurn.Query().
		Where(at.AgentSessionID(sessionID), at.StatusEQ(at.StatusCompleted)).
		Order(ent.Desc(at.FieldSequence))
	latestTurn, latestErr := queryLatestTurn.First(ctx)
	if latestErr != nil && !ent.IsNotFound(latestErr) {
		return records, fmt.Errorf("query latest completed turn: %w", latestErr)
	}
	if latestTurn != nil {
		outputs.LatestCompletedTurnID = latestTurn.ID
	}
	return records, nil
}

// loadCitedEvidence resolves each piece of evidence to the entities it is about: its subject entity, or both
// ends of its subject relationship.
func (h *investigationHarness) loadCitedEvidence(ctx context.Context, evidenceIDs []uuid.UUID) (map[uuid.UUID]CitedEvidence, error) {
	cited := make(map[uuid.UUID]CitedEvidence, len(evidenceIDs))
	if len(evidenceIDs) == 0 {
		return cited, nil
	}
	client := h.db.Client(ctx)
	queryEvidence := client.KnowledgeEvidence.Query().
		Where(kev.IDIn(evidenceIDs...))
	evidence, evidenceErr := queryEvidence.All(ctx)
	if evidenceErr != nil {
		return nil, fmt.Errorf("query cited evidence: %w", evidenceErr)
	}
	aliasIDs := make([]uuid.UUID, 0, len(evidence))
	for _, item := range evidence {
		aliasIDs = append(aliasIDs, item.SubjectAliasID)
	}
	queryAliases := client.KnowledgeSubjectAlias.Query().
		Where(ksa.IDIn(aliasIDs...))
	aliases, aliasesErr := queryAliases.All(ctx)
	if aliasesErr != nil {
		return nil, fmt.Errorf("query evidence subjects: %w", aliasesErr)
	}
	var relationshipIDs []uuid.UUID
	entityIDs := mapset.NewSet[uuid.UUID]()
	for _, alias := range aliases {
		if alias.EntityID != nil {
			entityIDs.Add(*alias.EntityID)
		}
		if alias.RelationshipID != nil {
			relationshipIDs = append(relationshipIDs, *alias.RelationshipID)
		}
	}
	queryRelationships := client.KnowledgeRelationship.Query().
		Where(knr.IDIn(relationshipIDs...))
	relationships, relationshipsErr := queryRelationships.All(ctx)
	if relationshipsErr != nil {
		return nil, fmt.Errorf("query evidence relationships: %w", relationshipsErr)
	}
	relationshipEnds := make(map[uuid.UUID][]uuid.UUID, len(relationships))
	for _, relationship := range relationships {
		ends := []uuid.UUID{relationship.SourceEntityID, relationship.TargetEntityID}
		relationshipEnds[relationship.ID] = ends
		entityIDs.Append(ends...)
	}
	queryEntities := client.KnowledgeEntity.Query().
		Where(ke.IDIn(entityIDs.ToSlice()...))
	entities, entitiesErr := queryEntities.All(ctx)
	if entitiesErr != nil {
		return nil, fmt.Errorf("query evidence entities: %w", entitiesErr)
	}
	recordedEntities := make(map[uuid.UUID]RecordedEntity, len(entities))
	for _, entity := range entities {
		recordedEntities[entity.ID] = RecordedEntity{Kind: entity.Kind, Name: entity.State.DisplayName, Properties: entity.State.Properties}
	}
	subjects := make(map[uuid.UUID][]uuid.UUID, len(aliases))
	for _, alias := range aliases {
		switch {
		case alias.EntityID != nil:
			subjects[alias.ID] = []uuid.UUID{*alias.EntityID}
		case alias.RelationshipID != nil:
			subjects[alias.ID] = relationshipEnds[*alias.RelationshipID]
		}
	}
	for _, item := range evidence {
		var about []RecordedEntity
		for _, entityID := range subjects[item.SubjectAliasID] {
			if entity, known := recordedEntities[entityID]; known {
				about = append(about, entity)
			}
		}
		cited[item.ID] = CitedEvidence{Entities: about}
	}
	return cited, nil
}

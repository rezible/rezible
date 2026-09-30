package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	inv "github.com/rezible/rezible/ent/investigation"
	"github.com/rezible/rezible/ent/predicate"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"

	at "github.com/rezible/rezible/ent/agentturn"
	invf "github.com/rezible/rezible/ent/investigationfinding"
	invfv "github.com/rezible/rezible/ent/investigationfindingversion"
	invfvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	invh "github.com/rezible/rezible/ent/investigationhypothesis"
	invhv "github.com/rezible/rezible/ent/investigationhypothesisversion"
	invor "github.com/rezible/rezible/ent/investigationoutputreference"
	invr "github.com/rezible/rezible/ent/investigationreport"
	invui "github.com/rezible/rezible/ent/investigationuserinput"
)

const investigationAnswerKeyPrefix = "answer:"

const maxInvestigationReportSummaryLength = 400

func (s *InvestigationService) PublishInvestigationReport(ctx context.Context, scope rez.InvestigationPublicationScope, params rez.PublishInvestigationReportParams) (*rez.InvestigationReportResult, error) {
	invId := scope.InvestigationID
	text := strings.TrimSpace(params.Text)
	if invId == uuid.Nil || scope.AgentTurnID == uuid.Nil || text == "" {
		return nil, fmt.Errorf("%w: investigation, turn and report text are required", rez.ErrInvalidInput)
	}
	summary := strings.TrimSpace(params.Summary)
	if utf8.RuneCountInString(summary) > maxInvestigationReportSummaryLength {
		return nil, fmt.Errorf("%w: report summary must be at most %d characters", rez.ErrInvalidInput, maxInvestigationReportSummaryLength)
	}
	evidenceIDs := s.normalizeInvestigationEvidenceIDs(params.EvidenceIDs)

	fpPayload := investigationReportFingerprintPayload{
		Text:        text,
		Summary:     summary,
		EvidenceIDs: evidenceIDs,
	}
	fingerprint, fingerprintErr := s.investigationOutputFingerprint("report", invId.String(), fpPayload)
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint investigation report: %w", fingerprintErr)
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.InvestigationReportResult, error) {
		current, turn, prepareErr := s.prepareInvestigationOutputWrite(ctx, invId, scope.AgentTurnID)
		if prepareErr != nil {
			return nil, prepareErr
		}

		reportPreds := []predicate.InvestigationReport{
			invr.InvestigationID(current.ID),
			invr.AgentTurnID(turn.ID),
			invr.Fingerprint(fingerprint),
		}
		queryReport := tx.InvestigationReport.Query().
			Where(reportPreds...).
			WithAgentTurn()
		existing, queryErr := queryReport.Only(ctx)
		if queryErr == nil {
			return s.investigationReportResult(ctx, existing)
		}
		if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("find repeated investigation report: %w", queryErr)
		}

		createReport := tx.InvestigationReport.Create().
			SetInvestigationID(current.ID).
			SetAgentTurnID(turn.ID).
			SetText(text).
			SetSummary(summary).
			SetFingerprint(fingerprint)
		created, saveErr := createReport.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save investigation report publication: %w", saveErr)
		}

		owner := investigationReferenceOwner{reportID: &created.ID}
		if referenceErr := s.saveInvestigationEvidenceIDs(ctx, owner, params.EvidenceIDs); referenceErr != nil {
			return nil, referenceErr
		}
		queryCreated := tx.InvestigationReport.Query().
			Where(invr.ID(created.ID)).
			WithAgentTurn()
		loaded, loadErr := queryCreated.Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load saved investigation report: %w", loadErr)
		}
		return s.investigationReportResult(ctx, loaded)
	})
}

func (s *InvestigationService) makeFindingVersionPreds(turnId uuid.UUID, fingerprint string, findingPred predicate.InvestigationFinding) []predicate.InvestigationFindingVersion {
	return []predicate.InvestigationFindingVersion{
		invfv.AgentTurnID(turnId),
		invfv.Fingerprint(fingerprint),
		invfv.HasFindingWith(findingPred),
	}
}

func (s *InvestigationService) PublishInvestigationFinding(ctx context.Context, scope rez.InvestigationPublicationScope, input rez.PublishInvestigationFindingParams) (*rez.InvestigationFindingVersion, error) {
	investigationID, turnID := scope.InvestigationID, scope.AgentTurnID
	params := rez.PublishInvestigationFindingParams{
		Key:               strings.TrimSpace(input.Key),
		Title:             strings.TrimSpace(input.Title),
		Body:              strings.TrimSpace(input.Body),
		EvidenceIDs:       s.normalizeInvestigationEvidenceIDs(input.EvidenceIDs),
		FindingReferences: s.normalizeFindingVersionReferences(input.FindingReferences),
	}
	if investigationID == uuid.Nil || turnID == uuid.Nil || params.Key == "" || params.Title == "" || params.Body == "" {
		return nil, fmt.Errorf("%w: investigation, turn, key, title and body are required", rez.ErrInvalidInput)
	}
	if strings.HasPrefix(params.Key, investigationAnswerKeyPrefix) {
		return nil, fmt.Errorf("%w: finding keys beginning with %q are reserved", rez.ErrInvalidInput, investigationAnswerKeyPrefix)
	}
	fingerprintPayload := investigationFindingFingerprintPayload{
		Key:               params.Key,
		Title:             params.Title,
		Body:              params.Body,
		EvidenceIDs:       params.EvidenceIDs,
		FindingReferences: s.fingerprintFindingReferences(params.FindingReferences),
	}
	fingerprint, fingerprintErr := s.investigationOutputFingerprint("finding", params.Key, fingerprintPayload)
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint investigation finding: %w", fingerprintErr)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.InvestigationFindingVersion, error) {
		current, turn, prepareErr := s.prepareInvestigationOutputWrite(ctx, investigationID, turnID)
		if prepareErr != nil {
			return nil, prepareErr
		}

		queryExisting := tx.InvestigationFindingVersion.Query().
			Where(s.makeFindingVersionPreds(turn.ID, fingerprint, invf.InvestigationID(current.ID))...).
			WithFinding().
			WithAgentTurn()
		existing, queryErr := queryExisting.Only(ctx)
		if queryErr == nil {
			return s.investigationFindingResult(ctx, existing)
		}
		if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("find repeated investigation finding: %w", queryErr)
		}
		stable, stableErr := s.findOrCreateInvestigationFinding(ctx, current.ID, params.Key, nil)
		if stableErr != nil {
			return nil, stableErr
		}

		createVersion := tx.InvestigationFindingVersion.Create().
			SetFindingID(stable.ID).
			SetAgentTurnID(turn.ID).
			SetTitle(params.Title).
			SetBody(params.Body).
			SetFingerprint(fingerprint)
		created, saveErr := createVersion.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save investigation finding version: %w", saveErr)
		}

		refOwner := investigationReferenceOwner{findingVersionID: &created.ID}
		if refErr := s.saveInvestigationEvidenceIDs(ctx, refOwner, params.EvidenceIDs); refErr != nil {
			return nil, refErr
		}
		if linkErr := s.saveFindingVersionReferences(ctx, created.ID, params.FindingReferences); linkErr != nil {
			return nil, linkErr
		}

		queryVersion := tx.InvestigationFindingVersion.Query().
			Where(invfv.ID(created.ID)).
			WithFinding().
			WithAgentTurn()
		loaded, loadErr := queryVersion.Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load saved investigation finding version: %w", loadErr)
		}
		return s.investigationFindingResult(ctx, loaded)
	})
}

func (s *InvestigationService) PublishInvestigationAnswer(ctx context.Context, scope rez.InvestigationPublicationScope, input rez.PublishInvestigationAnswerParams) (*rez.InvestigationFindingVersion, error) {
	investigationID, turnID := scope.InvestigationID, scope.AgentTurnID
	params := rez.PublishInvestigationAnswerParams{
		Title:             strings.TrimSpace(input.Title),
		Body:              strings.TrimSpace(input.Body),
		EvidenceIDs:       s.normalizeInvestigationEvidenceIDs(input.EvidenceIDs),
		FindingReferences: s.normalizeFindingVersionReferences(input.FindingReferences),
	}
	if investigationID == uuid.Nil || turnID == uuid.Nil || params.Title == "" || params.Body == "" {
		return nil, fmt.Errorf("%w: investigation, turn, title and body are required", rez.ErrInvalidInput)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.InvestigationFindingVersion, error) {
		current, turn, prepareErr := s.prepareInvestigationOutputWrite(ctx, investigationID, turnID)
		if prepareErr != nil {
			return nil, prepareErr
		}
		assignedQuery := tx.InvestigationUserInput.Query().
			Where(invui.InvestigationID(current.ID), invui.AgentTurnID(turn.ID))
		userInput, assignedErr := assignedQuery.Only(ctx)
		if ent.IsNotFound(assignedErr) {
			return nil, fmt.Errorf("%w: this turn has no assigned user question", rez.ErrInvalidInput)
		}
		if assignedErr != nil {
			return nil, fmt.Errorf("load assigned investigation question: %w", assignedErr)
		}
		key := investigationAnswerKeyPrefix + userInput.ID.String()
		fingerprintPayload := investigationAnswerFingerprintPayload{
			Title:             params.Title,
			Body:              params.Body,
			EvidenceIDs:       params.EvidenceIDs,
			FindingReferences: s.fingerprintFindingReferences(params.FindingReferences),
		}
		fingerprint, fingerprintErr := s.investigationOutputFingerprint("answer", userInput.ID.String(), fingerprintPayload)
		if fingerprintErr != nil {
			return nil, fmt.Errorf("fingerprint investigation answer: %w", fingerprintErr)
		}

		findingPred := invf.And(
			invf.InvestigationID(current.ID),
			invf.UserInputID(userInput.ID),
			invf.Key(key),
		)
		queryExisting := tx.InvestigationFindingVersion.Query().
			Where(s.makeFindingVersionPreds(turn.ID, fingerprint, findingPred)...).
			WithFinding().
			WithAgentTurn()
		existing, queryErr := queryExisting.Only(ctx)
		if queryErr == nil {
			return s.investigationFindingResult(ctx, existing)
		}
		if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("find repeated investigation answer: %w", queryErr)
		}
		inputID := userInput.ID
		stable, stableErr := s.findOrCreateInvestigationFinding(ctx, current.ID, key, &inputID)
		if stableErr != nil {
			return nil, stableErr
		}

		createVersion := tx.InvestigationFindingVersion.Create().
			SetFindingID(stable.ID).
			SetAgentTurnID(turn.ID).
			SetTitle(params.Title).
			SetBody(params.Body).
			SetFingerprint(fingerprint)
		created, saveErr := createVersion.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save investigation answer version: %w", saveErr)
		}
		if referenceErr := s.saveInvestigationEvidenceIDs(ctx, investigationReferenceOwner{findingVersionID: &created.ID}, params.EvidenceIDs); referenceErr != nil {
			return nil, referenceErr
		}
		if linkErr := s.saveFindingVersionReferences(ctx, created.ID, params.FindingReferences); linkErr != nil {
			return nil, linkErr
		}
		queryVersion := tx.InvestigationFindingVersion.Query().
			Where(invfv.ID(created.ID)).
			WithFinding().
			WithAgentTurn()
		loaded, loadErr := queryVersion.Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load saved investigation answer version: %w", loadErr)
		}
		return s.investigationFindingResult(ctx, loaded)
	})
}

func (s *InvestigationService) PublishInvestigationHypothesis(ctx context.Context, scope rez.InvestigationPublicationScope, params rez.PublishInvestigationHypothesisParams) (*rez.InvestigationHypothesisVersion, error) {
	investigationID, turnID := scope.InvestigationID, scope.AgentTurnID
	params.Key = strings.TrimSpace(params.Key)
	params.Title = strings.TrimSpace(params.Title)
	params.Justification = strings.TrimSpace(params.Justification)
	params.Status = invhv.Status(strings.TrimSpace(string(params.Status)))
	params.EvidenceIDs = s.normalizeInvestigationEvidenceIDs(params.EvidenceIDs)
	if investigationID == uuid.Nil || turnID == uuid.Nil || params.Key == "" || params.Title == "" || params.Justification == "" {
		return nil, fmt.Errorf("%w: investigation, turn, key, title and justification are required", rez.ErrInvalidInput)
	}
	switch params.Status {
	case invhv.StatusOpen, invhv.StatusSupported, invhv.StatusDisproven, invhv.StatusInconclusive:
	default:
		return nil, fmt.Errorf("%w: invalid hypothesis status %q", rez.ErrInvalidInput, params.Status)
	}
	fingerprintPayload := investigationHypothesisFingerprintPayload{
		Key:           params.Key,
		Title:         params.Title,
		Justification: params.Justification,
		Status:        params.Status,
		EvidenceIDs:   params.EvidenceIDs,
	}
	fingerprint, fingerprintErr := s.investigationOutputFingerprint("hypothesis", params.Key, fingerprintPayload)
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint investigation hypothesis: %w", fingerprintErr)
	}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*rez.InvestigationHypothesisVersion, error) {
		current, turn, prepareErr := s.prepareInvestigationOutputWrite(ctx, investigationID, turnID)
		if prepareErr != nil {
			return nil, prepareErr
		}

		lookupResult := func(preds ...predicate.InvestigationHypothesisVersion) (*rez.InvestigationHypothesisVersion, error) {
			queryVersion := tx.InvestigationHypothesisVersion.Query().
				Where(preds...).
				WithHypothesis().
				WithAgentTurn()
			ver, queryErr := queryVersion.Only(ctx)
			if queryErr != nil {
				return nil, queryErr
			}
			return s.investigationHypothesisResult(ctx, ver)
		}

		versionPreds := []predicate.InvestigationHypothesisVersion{
			invhv.AgentTurnID(turn.ID),
			invhv.Fingerprint(fingerprint),
			invhv.HasHypothesisWith(
				invh.InvestigationID(current.ID),
				invh.Key(params.Key),
			),
		}

		_, existingResultErr := lookupResult(versionPreds...)
		if existingResultErr != nil && !ent.IsNotFound(existingResultErr) {
			return nil, fmt.Errorf("find repeated investigation hypothesis: %w", existingResultErr)
		}

		stable, stableErr := s.findOrCreateInvestigationHypothesis(ctx, current.ID, params.Key)
		if stableErr != nil {
			return nil, stableErr
		}

		createVersion := tx.InvestigationHypothesisVersion.Create().
			SetHypothesisID(stable.ID).
			SetAgentTurnID(turn.ID).
			SetTitle(params.Title).
			SetJustification(params.Justification).
			SetStatus(params.Status).
			SetFingerprint(fingerprint)
		created, saveErr := createVersion.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save investigation hypothesis version: %w", saveErr)
		}

		refOwner := investigationReferenceOwner{hypothesisVersionID: &created.ID}
		if refErr := s.saveInvestigationEvidenceIDs(ctx, refOwner, params.EvidenceIDs); refErr != nil {
			return nil, refErr
		}
		return lookupResult(invhv.ID(created.ID))
	})
}

func (s *InvestigationService) ReadInvestigationReport(ctx context.Context, investigationID uuid.UUID, params rez.ReadInvestigationReportParams) (*rez.InvestigationReportResult, error) {
	selection := params.Selection
	if selection == "" {
		selection = rez.InvestigationReportSelectionLatest
	}
	if selection != rez.InvestigationReportSelectionLatest && selection != rez.InvestigationReportSelectionCompleted {
		return nil, fmt.Errorf("%w: report selection must be latest or completed", rez.ErrInvalidInput)
	}

	queryReports := s.db.Client(ctx).InvestigationReport.Query().
		Where(invr.InvestigationID(investigationID)).
		WithAgentTurn()
	reports, queryErr := queryReports.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list investigation report publications: %w", queryErr)
	}
	eligible := make([]*ent.InvestigationReport, 0, len(reports))
	for _, report := range reports {
		if report.Edges.AgentTurn == nil || !s.eligibleTurnStatus(report.Edges.AgentTurn.Status, selection == rez.InvestigationReportSelectionCompleted) {
			continue
		}
		eligible = append(eligible, report)
	}
	sort.Slice(eligible, func(i, j int) bool {
		return s.outputPublicationNewer(eligible[i].Edges.AgentTurn, eligible[i].CreatedAt, eligible[i].ID, eligible[j].Edges.AgentTurn, eligible[j].CreatedAt, eligible[j].ID)
	})
	if len(eligible) == 0 {
		return nil, nil
	}
	return s.investigationReportResult(ctx, eligible[0])
}

func (s *InvestigationService) GetInvestigationFindingVersion(ctx context.Context, investigationID, versionID uuid.UUID) (*rez.InvestigationFindingVersion, error) {
	if investigationID == uuid.Nil || versionID == uuid.Nil {
		return nil, fmt.Errorf("%w: investigation and version IDs are required", rez.ErrInvalidInput)
	}
	queryInvestigation := s.db.Client(ctx).Investigation.Query().
		Where(inv.ID(investigationID))
	invExists, lookupInvErr := queryInvestigation.Exist(ctx)
	if !invExists {
		return nil, fmt.Errorf("%w: investigation not found", rez.ErrNotFound)
	} else if lookupInvErr != nil {
		return nil, fmt.Errorf("load investigation for finding read: %w", lookupInvErr)
	}

	queryVersion := s.db.Client(ctx).InvestigationFindingVersion.Query().
		Where(invfv.ID(versionID), invfv.HasFindingWith(invf.InvestigationID(investigationID))).
		WithFinding().
		WithAgentTurn()
	version, queryErr := queryVersion.Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("%w: investigation finding version not found", rez.ErrNotFound)
		}
		return nil, fmt.Errorf("load investigation finding version: %w", queryErr)
	}
	current, currentErr := s.latestInvestigationFindingVersions(ctx, investigationID)
	if currentErr != nil {
		return nil, currentErr
	}
	invalidated, invalidatedErr := s.invalidatedFindingVersions(ctx, current)
	if invalidatedErr != nil {
		return nil, invalidatedErr
	}
	return s.investigationFindingResultWithInvalidations(ctx, version, invalidated)
}

func (s *InvestigationService) GetInvestigationHypothesisVersion(ctx context.Context, investigationID, versionID uuid.UUID) (*rez.InvestigationHypothesisVersion, error) {
	if investigationID == uuid.Nil || versionID == uuid.Nil {
		return nil, fmt.Errorf("%w: investigation and version IDs are required", rez.ErrInvalidInput)
	}
	queryVersion := s.db.Client(ctx).InvestigationHypothesisVersion.Query().
		Where(invhv.ID(versionID), invhv.HasHypothesisWith(invh.InvestigationID(investigationID))).
		WithHypothesis().
		WithAgentTurn()
	version, queryErr := queryVersion.Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("%w: investigation hypothesis version not found", rez.ErrNotFound)
		}
		return nil, fmt.Errorf("load investigation hypothesis version: %w", queryErr)
	}
	return s.investigationHypothesisResult(ctx, version)
}

func (s *InvestigationService) ListInvestigationFindings(ctx context.Context, investigationID uuid.UUID, params ent.ListParams) (*ent.ListResult[rez.InvestigationFindingVersion], error) {
	current, currentErr := s.latestInvestigationFindingVersions(ctx, investigationID)
	if currentErr != nil {
		return nil, currentErr
	}
	invalidated, invalidatedErr := s.invalidatedFindingVersions(ctx, current)
	if invalidatedErr != nil {
		return nil, invalidatedErr
	}
	page, pageSize := params.GetPage(), params.GetPageSize()
	start, end := s.outputPageBounds(len(current), page, pageSize)
	result := &ent.ListResult[rez.InvestigationFindingVersion]{
		Data:     make([]*rez.InvestigationFindingVersion, 0, end-start),
		Page:     page,
		PageSize: pageSize,
		Total:    len(current),
	}
	for _, version := range current[start:end] {
		item, itemErr := s.investigationFindingResultWithInvalidations(ctx, version, invalidated)
		if itemErr != nil {
			return nil, itemErr
		}
		result.Data = append(result.Data, item)
	}
	return result, nil
}

func (s *InvestigationService) ListInvestigationHypotheses(ctx context.Context, investigationID uuid.UUID, params ent.ListParams) (*ent.ListResult[rez.InvestigationHypothesisVersion], error) {
	current, currentErr := s.latestInvestigationHypothesisVersions(ctx, investigationID)
	if currentErr != nil {
		return nil, currentErr
	}
	page, pageSize := params.GetPage(), params.GetPageSize()
	start, end := s.outputPageBounds(len(current), page, pageSize)
	result := &ent.ListResult[rez.InvestigationHypothesisVersion]{
		Data:     make([]*rez.InvestigationHypothesisVersion, 0, end-start),
		Page:     page,
		PageSize: pageSize,
		Total:    len(current),
	}
	for _, version := range current[start:end] {
		item, itemErr := s.investigationHypothesisResult(ctx, version)
		if itemErr != nil {
			return nil, itemErr
		}
		result.Data = append(result.Data, item)
	}
	return result, nil
}

func (s *InvestigationService) prepareInvestigationOutputWrite(ctx context.Context, invId, turnID uuid.UUID) (*ent.Investigation, *ent.AgentTurn, error) {
	client := s.db.Client(ctx)
	current, investigationErr := client.Investigation.Get(ctx, invId)
	if investigationErr != nil {
		if ent.IsNotFound(investigationErr) {
			return nil, nil, fmt.Errorf("%w: investigation not found", rez.ErrNotFound)
		}
		return nil, nil, fmt.Errorf("load investigation for output publication: %w", investigationErr)
	}

	queryTurn := client.AgentTurn.Query().
		Where(at.ID(turnID)).
		Modify(func(s *sql.Selector) {
			s.ForUpdate()
		})
	turn, turnErr := queryTurn.Only(ctx)
	if turnErr != nil {
		if ent.IsNotFound(turnErr) {
			return nil, nil, fmt.Errorf("%w: agent turn not found", rez.ErrNotFound)
		}
		return nil, nil, fmt.Errorf("lock producing agent turn: %w", turnErr)
	}
	if turn.AgentSessionID != current.AgentSessionID {
		return nil, nil, fmt.Errorf("%w: agent turn does not belong to this investigation", rez.ErrConflict)
	}
	if turn.Status != at.StatusRunning {
		return nil, nil, fmt.Errorf("%w: output can only be published by a running turn", rez.ErrConflict)
	}
	return current, turn, nil
}

func (s *InvestigationService) findOrCreateInvestigationFinding(ctx context.Context, invId uuid.UUID, key string, inputID *uuid.UUID) (*ent.InvestigationFinding, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.InvestigationFinding, error) {
		lockKey := invId.String() + "\x1f" + key
		if lockErr := s.db.AcquireTxLocks(ctx, "investigation_output_finding", lockKey); lockErr != nil {
			return nil, fmt.Errorf("lock investigation finding identity: %w", lockErr)
		}

		queryExisting := s.db.Client(ctx).InvestigationFinding.Query().
			Where(invf.InvestigationID(invId), invf.Key(key))
		existing, queryExistingErr := queryExisting.Only(ctx)
		if queryExistingErr != nil && !ent.IsNotFound(queryExistingErr) {
			return nil, fmt.Errorf("load investigation finding identity: %w", queryExistingErr)
		} else if existing != nil {
			if inputID == nil && existing.UserInputID != nil || inputID != nil && (existing.UserInputID == nil || *existing.UserInputID != *inputID) {
				return nil, fmt.Errorf("%w: finding identity is already bound to another user input", rez.ErrConflict)
			}
			return existing, nil
		}

		createFinding := tx.InvestigationFinding.Create().
			SetInvestigationID(invId).
			SetKey(key)
		if inputID != nil {
			createFinding.SetUserInputID(*inputID)
		}
		created, saveErr := createFinding.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("create investigation finding identity: %w", saveErr)
		}
		return created, nil
	})
}

func (s *InvestigationService) findOrCreateInvestigationHypothesis(ctx context.Context, invId uuid.UUID, key string) (*ent.InvestigationHypothesis, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.InvestigationHypothesis, error) {
		lockKey := invId.String() + "\x1f" + key
		if lockErr := s.db.AcquireTxLocks(ctx, "investigation_output_hypothesis", lockKey); lockErr != nil {
			return nil, fmt.Errorf("lock investigation hypothesis identity: %w", lockErr)
		}
		queryHypothesis := tx.InvestigationHypothesis.Query().
			Where(invh.InvestigationID(invId), invh.Key(key))
		existing, queryErr := queryHypothesis.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("load investigation hypothesis identity: %w", queryErr)
		} else if existing != nil {
			return existing, nil
		}
		createHypothesis := tx.InvestigationHypothesis.Create().
			SetInvestigationID(invId).
			SetKey(key)
		created, saveErr := createHypothesis.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("create investigation hypothesis identity: %w", saveErr)
		}
		return created, nil
	})
}

func (s *InvestigationService) normalizeInvestigationEvidenceIDs(ids []uuid.UUID) []uuid.UUID {
	normalized := append([]uuid.UUID{}, ids...)
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].String() < normalized[j].String()
	})
	if len(normalized) < 2 {
		return normalized
	}
	unique := normalized[:1]
	for _, id := range normalized[1:] {
		if id != unique[len(unique)-1] {
			unique = append(unique, id)
		}
	}
	return unique
}

func (s *InvestigationService) normalizeFindingVersionReferences(refs []rez.FindingVersionReference) []rez.FindingVersionReference {
	normalized := make([]rez.FindingVersionReference, 0, len(refs))
	for _, ref := range refs {
		ref.Relation = invfvl.Relation(strings.TrimSpace(string(ref.Relation)))
		normalized = append(normalized, ref)
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Relation != normalized[j].Relation {
			return normalized[i].Relation < normalized[j].Relation
		}
		return normalized[i].VersionID.String() < normalized[j].VersionID.String()
	})
	if len(normalized) == 0 {
		return []rez.FindingVersionReference{}
	}
	unique := normalized[:0]
	for _, ref := range normalized {
		if len(unique) == 0 || unique[len(unique)-1] != ref {
			unique = append(unique, ref)
		}
	}
	return unique
}

type (
	findingVersionReferenceFingerprint struct {
		VersionID uuid.UUID       `json:"version_id"`
		Relation  invfvl.Relation `json:"relation"`
	}

	investigationReportFingerprintPayload struct {
		Text        string      `json:"text"`
		Summary     string      `json:"summary,omitempty"`
		EvidenceIDs []uuid.UUID `json:"evidence_ids"`
	}

	investigationFindingFingerprintPayload struct {
		Key               string                               `json:"key"`
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceIDs       []uuid.UUID                          `json:"evidence_ids"`
		FindingReferences []findingVersionReferenceFingerprint `json:"finding_references"`
	}

	investigationAnswerFingerprintPayload struct {
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceIDs       []uuid.UUID                          `json:"evidence_ids"`
		FindingReferences []findingVersionReferenceFingerprint `json:"finding_references"`
	}

	investigationHypothesisFingerprintPayload struct {
		Key           string       `json:"key"`
		Title         string       `json:"title"`
		Justification string       `json:"justification"`
		Status        invhv.Status `json:"status"`
		EvidenceIDs   []uuid.UUID  `json:"evidence_ids"`
	}

	investigationOutputFingerprintPayload struct {
		OutputType string `json:"output_type"`
		Identity   string `json:"identity"`
		Payload    any    `json:"payload"`
	}
)

func (s *InvestigationService) fingerprintFindingReferences(references []rez.FindingVersionReference) []findingVersionReferenceFingerprint {
	fingerprints := make([]findingVersionReferenceFingerprint, 0, len(references))
	for _, reference := range references {
		fingerprints = append(fingerprints, findingVersionReferenceFingerprint{
			VersionID: reference.VersionID,
			Relation:  reference.Relation,
		})
	}
	return fingerprints
}

func (s *InvestigationService) investigationOutputFingerprint(outputType, identity string, payload any) (string, error) {
	encoded, marshalErr := json.Marshal(investigationOutputFingerprintPayload{outputType, identity, payload})
	if marshalErr != nil {
		return "", fmt.Errorf("marshal normalized payload: %w", marshalErr)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

type investigationReferenceOwner struct {
	reportID            *uuid.UUID
	findingVersionID    *uuid.UUID
	hypothesisVersionID *uuid.UUID
}

func (s *InvestigationService) saveInvestigationEvidenceIDs(ctx context.Context, owner investigationReferenceOwner, ids []uuid.UUID) error {
	createFn := func(c *ent.InvestigationOutputReferenceCreate, i int) {
		if owner.reportID != nil {
			c.SetReportID(*owner.reportID)
		}
		if owner.findingVersionID != nil {
			c.SetFindingVersionID(*owner.findingVersionID)
		}
		if owner.hypothesisVersionID != nil {
			c.SetHypothesisVersionID(*owner.hypothesisVersionID)
		}
		c.SetKnowledgeEvidenceID(ids[i])
	}
	return s.db.Client(ctx).InvestigationOutputReference.MapCreateBulk(ids, createFn).Exec(ctx)
}

func (s *InvestigationService) saveFindingVersionReferences(ctx context.Context, sourceID uuid.UUID, refs []rez.FindingVersionReference) error {
	createFn := func(c *ent.InvestigationFindingVersionLinkCreate, i int) {
		c.SetSourceVersionID(sourceID)
		c.SetTargetVersionID(refs[i].VersionID)
		c.SetRelation(refs[i].Relation)
	}
	return s.db.Client(ctx).InvestigationFindingVersionLink.MapCreateBulk(refs, createFn).Exec(ctx)
}

func (s *InvestigationService) investigationReportResult(ctx context.Context, report *ent.InvestigationReport) (*rez.InvestigationReportResult, error) {
	turn := report.Edges.AgentTurn
	if turn == nil {
		loaded, loadErr := s.db.Client(ctx).InvestigationReport.QueryAgentTurn(report).Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load producing turn for report: %w", loadErr)
		}
		turn = loaded
	}
	evidenceIDs, evidenceErr := s.investigationEvidenceIDsForOwner(ctx, investigationReferenceOwner{reportID: &report.ID})
	if evidenceErr != nil {
		return nil, evidenceErr
	}
	return &rez.InvestigationReportResult{
		InvestigationPublicationMeta: s.investigationPublicationMeta(report.ID, turn, report.CreatedAt),
		Text:                         report.Text,
		Summary:                      report.Summary,
		EvidenceIDs:                  evidenceIDs,
	}, nil
}

func (s *InvestigationService) investigationFindingResult(ctx context.Context, version *ent.InvestigationFindingVersion) (*rez.InvestigationFindingVersion, error) {
	current, currentErr := s.latestInvestigationFindingVersions(ctx, version.Edges.Finding.InvestigationID)
	if currentErr != nil {
		return nil, currentErr
	}
	invalidated, invalidatedErr := s.invalidatedFindingVersions(ctx, current)
	if invalidatedErr != nil {
		return nil, invalidatedErr
	}
	return s.investigationFindingResultWithInvalidations(ctx, version, invalidated)
}

func (s *InvestigationService) investigationFindingResultWithInvalidations(ctx context.Context, version *ent.InvestigationFindingVersion, invalidated map[uuid.UUID][]uuid.UUID) (*rez.InvestigationFindingVersion, error) {
	finding := version.Edges.Finding
	if finding == nil {
		loaded, loadErr := s.db.Client(ctx).InvestigationFindingVersion.QueryFinding(version).Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load stable finding identity: %w", loadErr)
		}
		finding = loaded
	}
	turn := version.Edges.AgentTurn
	if turn == nil {
		loaded, loadErr := s.db.Client(ctx).InvestigationFindingVersion.QueryAgentTurn(version).Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load producing turn for finding: %w", loadErr)
		}
		turn = loaded
	}
	evidenceIDs, evidenceErr := s.investigationEvidenceIDsForOwner(ctx, investigationReferenceOwner{findingVersionID: &version.ID})
	if evidenceErr != nil {
		return nil, evidenceErr
	}
	findingReferences, findingReferenceErr := s.findingVersionReferencesForSource(ctx, version.ID)
	if findingReferenceErr != nil {
		return nil, findingReferenceErr
	}
	invalidatedBy := invalidated[version.ID]
	if invalidatedBy == nil {
		invalidatedBy = []uuid.UUID{}
	}
	return &rez.InvestigationFindingVersion{
		InvestigationPublicationMeta: s.investigationPublicationMeta(version.ID, turn, version.CreatedAt),
		FindingID:                    finding.ID, Key: finding.Key, UserInputID: finding.UserInputID,
		Title: version.Title, Body: version.Body,
		EvidenceIDs: evidenceIDs, FindingReferences: findingReferences,
		InvalidatedByVersionIDs: append([]uuid.UUID{}, invalidatedBy...),
	}, nil
}

func (s *InvestigationService) investigationHypothesisResult(ctx context.Context, version *ent.InvestigationHypothesisVersion) (*rez.InvestigationHypothesisVersion, error) {
	hypothesis := version.Edges.Hypothesis
	if hypothesis == nil {
		loaded, loadErr := s.db.Client(ctx).InvestigationHypothesisVersion.QueryHypothesis(version).Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load stable hypothesis identity: %w", loadErr)
		}
		hypothesis = loaded
	}
	turn := version.Edges.AgentTurn
	if turn == nil {
		loaded, loadErr := s.db.Client(ctx).InvestigationHypothesisVersion.QueryAgentTurn(version).Only(ctx)
		if loadErr != nil {
			return nil, fmt.Errorf("load producing turn for hypothesis: %w", loadErr)
		}
		turn = loaded
	}
	evidenceIDs, evidenceErr := s.investigationEvidenceIDsForOwner(ctx, investigationReferenceOwner{hypothesisVersionID: &version.ID})
	if evidenceErr != nil {
		return nil, evidenceErr
	}
	return &rez.InvestigationHypothesisVersion{
		InvestigationPublicationMeta: s.investigationPublicationMeta(version.ID, turn, version.CreatedAt),
		HypothesisID:                 hypothesis.ID, Key: hypothesis.Key, Title: version.Title,
		Justification: version.Justification, Status: version.Status, EvidenceIDs: evidenceIDs,
	}, nil
}

func (s *InvestigationService) investigationPublicationMeta(id uuid.UUID, turn *ent.AgentTurn, createdAt time.Time) rez.InvestigationPublicationMeta {
	return rez.InvestigationPublicationMeta{
		ID:          id,
		AgentTurnID: turn.ID,
		TurnStatus:  turn.Status,
		CreatedAt:   createdAt,
	}
}

func (s *InvestigationService) investigationEvidenceIDsForOwner(ctx context.Context, owner investigationReferenceOwner) ([]uuid.UUID, error) {
	var pred predicate.InvestigationOutputReference
	switch {
	case owner.reportID != nil:
		pred = invor.ReportID(*owner.reportID)
	case owner.findingVersionID != nil:
		pred = invor.FindingVersionID(*owner.findingVersionID)
	case owner.hypothesisVersionID != nil:
		pred = invor.HypothesisVersionID(*owner.hypothesisVersionID)
	default:
		return []uuid.UUID{}, nil
	}
	selectIds := s.db.Client(ctx).InvestigationOutputReference.Query().
		Where(pred).
		Select(invor.FieldKnowledgeEvidenceID)
	records, queryErr := selectIds.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("load investigation output evidence: %w", queryErr)
	}
	evidenceIDs := make([]uuid.UUID, 0, len(records))
	for _, record := range records {
		evidenceIDs = append(evidenceIDs, record.KnowledgeEvidenceID)
	}
	return s.normalizeInvestigationEvidenceIDs(evidenceIDs), nil
}

func (s *InvestigationService) findingVersionReferencesForSource(ctx context.Context, sourceID uuid.UUID) ([]rez.FindingVersionReference, error) {
	queryLinks := s.db.Client(ctx).InvestigationFindingVersionLink.Query().
		Where(invfvl.SourceVersionID(sourceID))
	links, queryErr := queryLinks.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("load finding version references: %w", queryErr)
	}
	refs := make([]rez.FindingVersionReference, len(links))
	for i, link := range links {
		refs[i] = rez.FindingVersionReference{
			VersionID: link.TargetVersionID,
			Relation:  link.Relation,
		}
	}
	return s.normalizeFindingVersionReferences(refs), nil
}

func (s *InvestigationService) latestInvestigationFindingVersions(ctx context.Context, invId uuid.UUID) ([]*ent.InvestigationFindingVersion, error) {
	queryVersions := s.db.Client(ctx).InvestigationFindingVersion.Query().
		Where(invfv.HasFindingWith(invf.InvestigationID(invId))).
		WithFinding().
		WithAgentTurn()
	versions, queryErr := queryVersions.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list investigation finding versions: %w", queryErr)
	}
	latest := make(map[uuid.UUID]*ent.InvestigationFindingVersion)
	for _, version := range versions {
		turn := version.Edges.AgentTurn
		finding := version.Edges.Finding
		if turn == nil || finding == nil || !s.eligibleTurnStatus(turn.Status, false) {
			continue
		}
		current, hasCurrent := latest[finding.ID]
		if !hasCurrent || s.outputPublicationNewer(turn, version.CreatedAt, version.ID, current.Edges.AgentTurn, current.CreatedAt, current.ID) {
			latest[finding.ID] = version
		}
	}
	result := make([]*ent.InvestigationFindingVersion, 0, len(latest))
	for _, version := range latest {
		result = append(result, version)
	}
	sort.Slice(result, func(i, j int) bool {
		return s.outputPublicationNewer(result[i].Edges.AgentTurn, result[i].CreatedAt, result[i].ID, result[j].Edges.AgentTurn, result[j].CreatedAt, result[j].ID)
	})
	return result, nil
}

func (s *InvestigationService) latestInvestigationHypothesisVersions(ctx context.Context, invId uuid.UUID) ([]*ent.InvestigationHypothesisVersion, error) {
	queryVersions := s.db.Client(ctx).InvestigationHypothesisVersion.Query().
		Where(invhv.HasHypothesisWith(invh.InvestigationID(invId))).
		WithHypothesis().
		WithAgentTurn()
	versions, queryErr := queryVersions.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list investigation hypothesis versions: %w", queryErr)
	}
	latest := make(map[uuid.UUID]*ent.InvestigationHypothesisVersion)
	for _, version := range versions {
		turn := version.Edges.AgentTurn
		hypothesis := version.Edges.Hypothesis
		if turn == nil || hypothesis == nil || !s.eligibleTurnStatus(turn.Status, false) {
			continue
		}
		current, hasCurrent := latest[hypothesis.ID]
		if !hasCurrent || s.outputPublicationNewer(turn, version.CreatedAt, version.ID, current.Edges.AgentTurn, current.CreatedAt, current.ID) {
			latest[hypothesis.ID] = version
		}
	}
	result := make([]*ent.InvestigationHypothesisVersion, 0, len(latest))
	for _, version := range latest {
		result = append(result, version)
	}
	sort.Slice(result, func(i, j int) bool {
		return s.outputPublicationNewer(result[i].Edges.AgentTurn, result[i].CreatedAt, result[i].ID, result[j].Edges.AgentTurn, result[j].CreatedAt, result[j].ID)
	})
	return result, nil
}

func (s *InvestigationService) invalidatedFindingVersions(ctx context.Context, current []*ent.InvestigationFindingVersion) (map[uuid.UUID][]uuid.UUID, error) {
	sourceIDs := make([]uuid.UUID, 0, len(current))
	for _, version := range current {
		sourceIDs = append(sourceIDs, version.ID)
	}
	if len(sourceIDs) == 0 {
		return map[uuid.UUID][]uuid.UUID{}, nil
	}
	queryLinks := s.db.Client(ctx).InvestigationFindingVersionLink.Query().
		Where(
			invfvl.SourceVersionIDIn(sourceIDs...),
			invfvl.RelationEQ(invfvl.RelationInvalidates),
		)
	links, queryErr := queryLinks.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list current finding invalidations: %w", queryErr)
	}
	byTarget := make(map[uuid.UUID][]uuid.UUID)
	for _, link := range links {
		byTarget[link.TargetVersionID] = append(byTarget[link.TargetVersionID], link.SourceVersionID)
	}
	for targetID := range byTarget {
		sort.Slice(byTarget[targetID], func(i, j int) bool { return byTarget[targetID][i].String() < byTarget[targetID][j].String() })
	}
	return byTarget, nil
}

func (s *InvestigationService) eligibleTurnStatus(status at.Status, completedOnly bool) bool {
	if completedOnly {
		return status == at.StatusCompleted
	}
	return status == at.StatusRunning || status == at.StatusCompleted
}

func (s *InvestigationService) outputPublicationNewer(leftTurn *ent.AgentTurn, leftCreatedAt time.Time, leftID uuid.UUID, rightTurn *ent.AgentTurn, rightCreatedAt time.Time, rightID uuid.UUID) bool {
	if leftTurn.Sequence != rightTurn.Sequence {
		return leftTurn.Sequence > rightTurn.Sequence
	}
	if !leftCreatedAt.Equal(rightCreatedAt) {
		return leftCreatedAt.After(rightCreatedAt)
	}
	return leftID.String() > rightID.String()
}

func (s *InvestigationService) outputPageBounds(total, page, pageSize int) (int, int) {
	if total == 0 || page <= 0 || pageSize <= 0 || page-1 > total/pageSize {
		return total, total
	}
	start := (page - 1) * pageSize
	if start >= total {
		return total, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}

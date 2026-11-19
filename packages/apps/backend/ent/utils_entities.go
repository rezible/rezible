package ent

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/alertepisode"
	"github.com/rezible/rezible/ent/incidentmilestone"
	invfvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/sequencecounter"
	vc "github.com/rezible/rezible/ent/videoconference"
)

// GetLatest returns the newest milestone of the given kinds, or of any kind when none are given. Equal
// timestamps are ordered by ID.
func (ims IncidentMilestones) GetLatest(kinds ...incidentmilestone.Kind) *IncidentMilestone {
	var latest *IncidentMilestone
	for _, im := range ims {
		if len(kinds) > 0 && !slices.Contains(kinds, im.Kind) {
			continue
		}
		if latest == nil || im.Timestamp.After(latest.Timestamp) ||
			(im.Timestamp.Equal(latest.Timestamp) && bytes.Compare(im.ID[:], latest.ID[:]) > 0) {
			latest = im
		}
	}
	return latest
}

// Ref is the reference of what the observation observes. Each observation kind sets its own edge field.
func (so *SituationObservation) Ref() schematypes.SituationObservationRef {
	var ref schematypes.SituationObservationRef
	if so.AlertEpisodeID != nil {
		ref.AlertEpisodeID = *so.AlertEpisodeID
	}
	return ref
}

// WithScopes loads each observation's scope with its entities and their knowledge entities.
func (soq *SituationObservationQuery) WithScopes() *SituationObservationQuery {
	return soq.WithAlertEpisode(func(q *AlertEpisodeQuery) {
		q.Select(alertepisode.FieldID).
			WithObservationScope(func(q *ObservationScopeQuery) {
				q.WithEntities(func(q *ObservationScopeEntityQuery) {
					q.WithKnowledgeEntity()
				})
			})
	})
}

// Scope is the observation's scope loaded by WithScopes, or nil when none is stored.
func (so *SituationObservation) Scope() *ObservationScope {
	if episode := so.Edges.AlertEpisode; episode != nil {
		return episode.Edges.ObservationScope
	}
	return nil
}

func (vcs VideoConferences) GetPrimary() *VideoConference {
	var active *VideoConference
	var latest *VideoConference
	for _, conference := range vcs {
		if latest == nil || conference.CreatedAt.After(latest.CreatedAt) {
			latest = conference
		}
		if conference.Status == vc.StatusActive {
			if active == nil || conference.CreatedAt.After(active.CreatedAt) {
				active = conference
			}
		}
	}
	if active != nil {
		return active
	}
	if latest != nil {
		return latest
	}
	return nil
}

func (ie IncidentEdges) GetPrimaryVideoConference() *VideoConference {
	//conferences, confErr := ie.VideoConferencesOrErr()
	//if confErr != nil || len(conferences) == 0 {
	//	return nil
	//}
	//return VideoConferences(conferences).GetPrimary()
	return nil
}

func (am *AgentMessage) MakeGenkitMessage() *ai.Message {
	return ai.NewMessage(ai.Role(am.Role), am.Metadata, am.Content...)
}

// CitedIDs returns the knowledge evidence or provider read each reference cites, in loaded order.
func (refs InvestigationOutputReferences) CitedIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		if ref.KnowledgeEvidenceID != nil {
			ids = append(ids, *ref.KnowledgeEvidenceID)
		} else if ref.AgentProviderReadID != nil {
			ids = append(ids, *ref.AgentProviderReadID)
		}
	}
	return ids
}

// KnowledgeEvidenceIDs returns the evidence IDs cited by the references, in loaded order.
func (refs InvestigationOutputReferences) KnowledgeEvidenceIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		if ref.KnowledgeEvidenceID != nil {
			ids = append(ids, *ref.KnowledgeEvidenceID)
		}
	}
	return ids
}

// InvalidatedByVersionIDs returns the source versions of the loaded incoming invalidation links.
func (v *InvestigationFindingVersion) InvalidatedByVersionIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(v.Edges.IncomingLinks))
	for _, link := range v.Edges.IncomingLinks {
		if link.Relation == invfvl.RelationInvalidates {
			ids = append(ids, link.SourceVersionID)
		}
	}
	return ids
}

// CurrentAnswerVersion returns the first loaded version of the input's answer finding.
// Investigation services load answer versions newest first.
func (ui *InvestigationUserInput) CurrentAnswerVersion() *InvestigationFindingVersion {
	for _, finding := range ui.Edges.Findings {
		if len(finding.Edges.Versions) > 0 {
			return finding.Edges.Versions[0]
		}
	}
	return nil
}

// Next returns the next number in the current tenant's sequence of the kind. Call it inside the transaction
// that creates the numbered entity: the upsert holds the counter row's lock until that transaction ends, so
// numbers are handed out one at a time per tenant and kind, and a rolled back create leaves no gap.
func (c *SequenceCounterClient) Next(ctx context.Context, kind sequencecounter.Kind) (int, error) {
	increment := c.Create().
		SetKind(kind).
		SetValue(1).
		OnConflictColumns(sequencecounter.FieldTenantID, sequencecounter.FieldKind).
		AddValue(1)
	id, upsertErr := increment.ID(ctx)
	if upsertErr != nil {
		return 0, fmt.Errorf("increment sequence counter: %w", upsertErr)
	}
	counter, getErr := c.Get(ctx, id)
	if getErr != nil {
		return 0, fmt.Errorf("read sequence counter: %w", getErr)
	}
	return counter.Value, nil
}

// formatReference returns the reference of the numbered entity, such as INC-12.
func formatReference(prefix string, n int) string {
	return prefix + "-" + strconv.Itoa(n)
}

// parseReference returns the number of a reference with the prefix in any letter case, or of a bare number.
func parseReference(prefix, s string) (int, bool) {
	digits := strings.TrimSpace(s)
	if len(digits) > len(prefix)+1 && strings.EqualFold(digits[:len(prefix)+1], prefix+"-") {
		digits = digits[len(prefix)+1:]
	}
	if digits == "" || strings.TrimLeft(digits, "0123456789") != "" {
		return 0, false
	}
	n, parseErr := strconv.Atoi(digits)
	if parseErr != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// Reference returns the incident's reference, such as INC-12.
func (i *Incident) Reference() string {
	return formatReference("INC", i.Number)
}

// ParseIncidentReference returns the number of an incident reference, such as INC-12, inc-12 or 12.
func ParseIncidentReference(s string) (int, bool) {
	return parseReference("INC", s)
}

// Reference returns the situation's reference, such as SIT-12.
func (s *Situation) Reference() string {
	return formatReference("SIT", s.Number)
}

// ParseSituationReference returns the number of a situation reference, such as SIT-12, sit-12 or 12.
func ParseSituationReference(s string) (int, bool) {
	return parseReference("SIT", s)
}

// ParentClaim returns the contains relationship chosen as the entity's parent, from its loaded target
// relationships; nil when it has none or they were not loaded.
func (ke *KnowledgeEntity) ParentClaim() *KnowledgeRelationship {
	for _, relationship := range ke.Edges.TargetRelationships {
		if relationship.IsParent {
			return relationship
		}
	}
	return nil
}

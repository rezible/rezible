package projections

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

func InternalEntityResourceRef(id uuid.UUID) rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:    "rezible",
		ResourceRef: id.String(),
	}
}

func InternalRelationshipResourceRef(sourceId uuid.UUID, targetId uuid.UUID) rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:    "rezible",
		ResourceRef: fmt.Sprintf("%s:%s", sourceId.String(), targetId.String()),
	}
}

type KnowledgeEntityLinkingAttributes struct {
	ID uuid.UUID
}

func (l KnowledgeEntityLinkingAttributes) Values() map[string]string {
	return map[string]string{"id": l.ID.String()}
}

type (
	CodeForgeEvent = Event[CodeForgeEventAttributes]

	// CodeForgeEventAttributes are the provider-neutral attributes persisted for repository observations.
	CodeForgeEventAttributes struct {
		DisplayName string `json:"display_name" validate:"required"`
		URL         string `json:"url"`
	}
)

const KindCodeForge = "code_forge"

func DecodeCodeForgeEvent(ev *ent.NormalizedEvent) (*CodeForgeEvent, error) {
	return DecodeEventAttributes[CodeForgeEventAttributes](ev)
}

type (
	// CodeChangeEvent is a normalized code change event from a code forge provider.
	CodeChangeEvent = Event[CodeChangeEventAttributes]

	// CodeChangeEventAttributes are the provider-neutral attributes persisted for code change events.
	CodeChangeEventAttributes struct {
		Repository       EntityObservation   `json:"repository" validate:"required"`
		DisplayName      string              `json:"display_name" validate:"required"`
		ImpactedEntities []EntityObservation `json:"impacted_entities"`
	}
)

const KindCodeChange = "code_change"

func DecodeCodeChangeEvent(ev *ent.NormalizedEvent) (*CodeChangeEvent, error) {
	return DecodeEventAttributes[CodeChangeEventAttributes](ev)
}

type (
	// UserEvent is a normalized user observation from an organization or chat provider.
	UserEvent = Event[UserEventAttributes]

	// UserEventAttributes are the provider-neutral attributes persisted for user observations.
	UserEventAttributes struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required"`
		ChatId   string `json:"chat_id"`
		Timezone string `json:"timezone"`
	}
)

// UserEntityLinkingAttributes contains the stable user identity values that
// can be shared by aliases from different providers.
type UserEntityLinkingAttributes struct {
	Email string
}

func (a UserEntityLinkingAttributes) Values() map[string]string {
	email := strings.ToLower(strings.TrimSpace(a.Email))
	if email == "" {
		return nil
	}
	return map[string]string{"user.email": email}
}

const KindUser = "user"

func DecodeUserEvent(ev *ent.NormalizedEvent) (*UserEvent, error) {
	return DecodeEventAttributes[UserEventAttributes](ev)
}

type (
	TeamEvent = Event[TeamEventAttributes]

	TeamEventAttributes struct {
		Name          string `json:"name" validate:"required"`
		Slug          string `json:"slug" validate:"required"`
		ChatChannelId string `json:"chat_channel_id"`
		Deleted       bool   `json:"deleted"`
	}
)

const KindTeam = "team"

func DecodeTeamEvent(ev *ent.NormalizedEvent) (*TeamEvent, error) {
	return DecodeEventAttributes[TeamEventAttributes](ev)
}

type (
	TeamMembershipEvent = Event[TeamMembershipEventAttributes]

	TeamMembershipEventAttributes struct {
		Team TeamMembershipTeamAttributes `json:"team" validate:"required"`
		User TeamMembershipUserAttributes `json:"user" validate:"required"`
		Role string                       `json:"role" validate:"oneof=admin member"`
	}

	TeamMembershipTeamAttributes struct {
		rez.ProviderResourceRef
		Name          string `json:"name" validate:"required"`
		Slug          string `json:"slug" validate:"required"`
		ChatChannelId string `json:"chat_channel_id"`
	}

	TeamMembershipUserAttributes struct {
		rez.ProviderResourceRef
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required"`
		ChatId   string `json:"chat_id"`
		Timezone string `json:"timezone"`
	}
)

const KindTeamMembership = "team_membership"

func DecodeTeamMembershipEvent(ev *ent.NormalizedEvent) (*TeamMembershipEvent, error) {
	return DecodeEventAttributes[TeamMembershipEventAttributes](ev)
}

type (
	// IncidentEvent is a normalized incident observation from an incident provider.
	IncidentEvent = Event[IncidentEventAttributes]

	// IncidentEventAttributes are the provider-neutral attributes persisted for incident observations.
	IncidentEventAttributes struct {
		Title           string     `json:"title" validate:"required"`
		Summary         string     `json:"summary"`
		SeverityRef     string     `json:"severity_ref" validate:"required"`
		TypeRef         string     `json:"type_ref" validate:"required"`
		OpenedAt        time.Time  `json:"opened_at"`
		ResponseState   string     `json:"response_state"`
		ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
		SourceUpdatedAt *time.Time `json:"source_updated_at,omitempty"`
		SourceURL       *string    `json:"source_url,omitempty"`
	}
)

const KindIncident = "incident"

func DecodeIncidentEvent(ev *ent.NormalizedEvent) (*IncidentEvent, error) {
	return DecodeEventAttributes[IncidentEventAttributes](ev)
}

type (
	// AlertInstanceEvent is a normalized alert observation from an alerting provider.
	AlertInstanceEvent = Event[AlertInstanceEventAttributes]

	// AlertInstanceEventAttributes are the provider-neutral attributes persisted for alert observations.
	AlertInstanceEventAttributes struct {
		// Title is the definition's stable name.
		Title       string `json:"title" validate:"required"`
		Description string `json:"description"`
		Definition  string `json:"definition"`
		State       string `json:"state" validate:"required,oneof=firing resolved"`
		// InstanceID is the provider's identity for the source instance (Alertmanager's fingerprint), when it has one.
		InstanceID string            `json:"instance_id,omitempty"`
		Labels     map[string]string `json:"labels,omitempty"`
		// Summary is the rendered text for this instance.
		Summary string `json:"summary,omitempty"`
		// IdentityGroupLabels is the source's proposal for which labels group instances. It applies only to a
		// definition with no identity group labels.
		IdentityGroupLabels []string `json:"identity_group_labels,omitempty"`
		// Severity is normalized; SeverityRef keeps the provider's raw value.
		Severity    string `json:"severity" validate:"required,oneof=unknown info warning critical"`
		SeverityRef string `json:"severity_ref,omitempty"`
		// StartedAt is the provider's window start. An integration whose source has none must synthesize a stable one.
		StartedAt time.Time `json:"started_at" validate:"required"`
		// EndedAt is the provider's end, read only from a resolved notification.
		EndedAt *time.Time `json:"ended_at,omitempty"`
		// ResolutionTimeoutSeconds is how long after the last firing notification a window is assumed ended: 0
		// never; omitted keeps the stored value, or the configured default for a new definition.
		ResolutionTimeoutSeconds *int                `json:"resolution_timeout_seconds,omitempty" validate:"omitempty,gte=0"`
		ObservedEntities         []EntityObservation `json:"observed_entities"`
	}
)

const (
	AlertStateFiring   = "firing"
	AlertStateResolved = "resolved"
)

const KindAlertInstance = "alert_instance"

func DecodeAlertInstanceEvent(ev *ent.NormalizedEvent) (*AlertInstanceEvent, error) {
	return DecodeEventAttributes[AlertInstanceEventAttributes](ev)
}

type (
	// SystemComponentEvent is a normalized system component observation from a topology provider.
	SystemComponentEvent = Event[SystemComponentEventAttributes]

	// SystemComponentEventAttributes are the provider-neutral attributes persisted for system component observations.
	SystemComponentEventAttributes struct {
		Category    kne.Category   `json:"category" validate:"required"`
		Kind        string         `json:"kind" validate:"required"`
		DisplayName string         `json:"display_name" validate:"required"`
		Description string         `json:"description"`
		Properties  map[string]any `json:"properties"`
	}
)

const KindSystemComponent = "system_component"

func DecodeSystemComponentEvent(ev *ent.NormalizedEvent) (*SystemComponentEvent, error) {
	return DecodeEventAttributes[SystemComponentEventAttributes](ev)
}

type (
	// SystemRelationshipEvent is a normalized system relationship observation from a topology provider.
	SystemRelationshipEvent = Event[SystemRelationshipEventAttributes]

	// SystemRelationshipEventAttributes are the provider-neutral attributes persisted for system relationship observations.
	SystemRelationshipEventAttributes struct {
		Predicate   knr.Predicate     `json:"predicate" validate:"required"`
		DisplayName string            `json:"display_name"`
		Description string            `json:"description"`
		Source      EntityObservation `json:"source" validate:"required"`
		Target      EntityObservation `json:"target" validate:"required"`
		Properties  map[string]any    `json:"properties"`
	}
)

const KindSystemRelationship = "system_relationship"

func DecodeSystemRelationshipEvent(ev *ent.NormalizedEvent) (*SystemRelationshipEvent, error) {
	return DecodeEventAttributes[SystemRelationshipEventAttributes](ev)
}

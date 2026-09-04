package projections

import (
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

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
		Title       string    `json:"title" validate:"required"`
		Summary     string    `json:"summary"`
		SeverityRef string    `json:"severity_ref" validate:"required"`
		TypeRef     string    `json:"type_ref" validate:"required"`
		OpenedAt    time.Time `json:"opened_at"`
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
		Title            string              `json:"title" validate:"required"`
		Description      string              `json:"description"`
		Definition       string              `json:"definition"`
		ObservedEntities []EntityObservation `json:"observed_entities"`
	}
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

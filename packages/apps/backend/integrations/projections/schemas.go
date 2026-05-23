package projections

import (
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
)

const attributeFieldNameTag = "attr"

type (
	// ChatMessage is a normalized chat message observed from a messaging provider.
	ChatMessage = Event[ChatMessageAttributes]

	// ChatMessageAttributes are the provider-neutral attributes persisted for chat message events.
	ChatMessageAttributes struct {
		ConversationExternalRef string `attr:"conversation_external_ref" validate:"required"`
		Body                    string `attr:"body" validate:"required"`
		SenderExternalRef       string `attr:"sender_external_ref"`
		ThreadExternalRef       string `attr:"thread_external_ref"`
	}
)

func DecodeChatMessage(ev *ent.NormalizedEvent) (*ChatMessage, error) {
	return decodeKind[ChatMessageAttributes](ev, ne.KindChatMessage)
}

type (
	// RepositoryObserved is a normalized repository observation from a code forge provider.
	RepositoryObserved = Event[RepositoryObservedAttributes]

	// RepositoryObservedAttributes are the provider-neutral attributes persisted for repository observations.
	RepositoryObservedAttributes struct {
		DisplayName string `attr:"display_name" validate:"required"`
		URL         string `attr:"url"`
	}
)

func DecodeRepositoryObserved(ev *ent.NormalizedEvent) (*RepositoryObserved, error) {
	return decodeKind[RepositoryObservedAttributes](ev, ne.KindRepositoryObserved)
}

type (
	// CodeChangeObserved is a normalized code change event from a code forge provider.
	CodeChangeObserved = Event[CodeChangeObservedAttributes]

	// CodeChangeObservedAttributes are the provider-neutral attributes persisted for code change events.
	CodeChangeObservedAttributes struct {
		RepositoryExternalRef string `attr:"repository_external_ref" validate:"required"`
		DisplayName           string `attr:"display_name" validate:"required"`
	}
)

func DecodeCodeChangeObserved(ev *ent.NormalizedEvent) (*CodeChangeObserved, error) {
	return decodeKind[CodeChangeObservedAttributes](ev, ne.KindChangeEventObserved)
}

type (
	// UserObserved is a normalized user observation from an organization or chat provider.
	UserObserved = Event[UserObservedAttributes]

	// UserObservedAttributes are the provider-neutral attributes persisted for user observations.
	UserObservedAttributes struct {
		Name     string `attr:"name" validate:"required"`
		Email    string `attr:"email" validate:"required"`
		ChatId   string `attr:"chat_id"`
		Timezone string `attr:"timezone"`
	}
)

func DecodeUserObserved(ev *ent.NormalizedEvent) (*UserObserved, error) {
	return decodeKind[UserObservedAttributes](ev, ne.KindUserObserved)
}

type (
	// IncidentObserved is a normalized incident observation from an incident provider.
	IncidentObserved = Event[IncidentObservedAttributes]

	// IncidentObservedAttributes are the provider-neutral attributes persisted for incident observations.
	IncidentObservedAttributes struct {
		Title       string `attr:"title" validate:"required"`
		ExternalRef string `attr:"external_ref" validate:"required"`
		Summary     string `attr:"summary"`
		SeverityRef string `attr:"severity_ref" validate:"required"`
		TypeRef     string `attr:"type_ref" validate:"required"`
	}
)

func DecodeIncidentObserved(ev *ent.NormalizedEvent) (*IncidentObserved, error) {
	return decodeKind[IncidentObservedAttributes](ev, ne.KindIncidentObserved)
}

type (
	// AlertObserved is a normalized alert observation from an alerting provider.
	AlertObserved = Event[AlertObservedAttributes]

	// AlertObservedAttributes are the provider-neutral attributes persisted for alert observations.
	AlertObservedAttributes struct {
		Title       string `attr:"title" validate:"required"`
		Description string `attr:"description"`
		Definition  string `attr:"definition"`
	}
)

func DecodeAlertObserved(ev *ent.NormalizedEvent) (*AlertObserved, error) {
	return decodeKind[AlertObservedAttributes](ev, ne.KindAlertObserved)
}

type (
	// SystemComponentObserved is a normalized system component observation from a topology provider.
	SystemComponentObserved = Event[SystemComponentObservedAttributes]

	// SystemComponentObservedAttributes are the provider-neutral attributes persisted for system component observations.
	SystemComponentObservedAttributes struct {
		ExternalRef string         `attr:"external_ref" validate:"required"`
		Kind        string         `attr:"kind" validate:"required"`
		DisplayName string         `attr:"display_name" validate:"required"`
		Description string         `attr:"description"`
		Properties  map[string]any `attr:"properties"`
	}
)

func DecodeSystemComponentObserved(ev *ent.NormalizedEvent) (*SystemComponentObserved, error) {
	return decodeKind[SystemComponentObservedAttributes](ev, ne.KindSystemComponentObserved)
}

type (
	// SystemRelationshipObserved is a normalized system relationship observation from a topology provider.
	SystemRelationshipObserved = Event[SystemRelationshipObservedAttributes]

	// SystemRelationshipObservedAttributes are the provider-neutral attributes persisted for system relationship observations.
	SystemRelationshipObservedAttributes struct {
		ExternalRef       string         `attr:"external_ref" validate:"required"`
		Kind              string         `attr:"kind" validate:"required"`
		DisplayName       string         `attr:"display_name"`
		Description       string         `attr:"description"`
		SourceExternalRef string         `attr:"source_external_ref" validate:"required"`
		SourceKind        string         `attr:"source_kind" validate:"required"`
		SourceDisplayName string         `attr:"source_display_name" validate:"required"`
		TargetExternalRef string         `attr:"target_external_ref" validate:"required"`
		TargetKind        string         `attr:"target_kind" validate:"required"`
		TargetDisplayName string         `attr:"target_display_name" validate:"required"`
		Properties        map[string]any `attr:"properties"`
	}
)

func DecodeSystemRelationshipObserved(ev *ent.NormalizedEvent) (*SystemRelationshipObserved, error) {
	return decodeKind[SystemRelationshipObservedAttributes](ev, ne.KindSystemRelationshipObserved)
}

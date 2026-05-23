package projections

import (
	"github.com/rezible/rezible/ent"
)

const attributeFieldNameTag = "attr"

type SubjectKind string

func (k SubjectKind) String() string {
	return string(k)
}

func (k SubjectKind) Matches(ev *ent.NormalizedEvent) bool {
	return SubjectKind(ev.SubjectKind) == k
}

const SubjectKindChatMessage SubjectKind = "chat_message"

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

func DecodeChatMessageEvent(ev *ent.NormalizedEvent) (*ChatMessage, error) {
	return DecodeSubjectAttributes[ChatMessageAttributes](ev)
}

type (
	CodeForgeEvent = Event[CodeForgeSubjectAttributes]

	// CodeForgeSubjectAttributes are the provider-neutral attributes persisted for repository observations.
	CodeForgeSubjectAttributes struct {
		DisplayName string `attr:"display_name" validate:"required"`
		URL         string `attr:"url"`
	}
)

const SubjectKindCodeForge SubjectKind = "code_forge"

func DecodeCodeForgeEvent(ev *ent.NormalizedEvent) (*CodeForgeEvent, error) {
	return DecodeSubjectAttributes[CodeForgeSubjectAttributes](ev)
}

type (
	// CodeChangeEvent is a normalized code change event from a code forge provider.
	CodeChangeEvent = Event[CodeChangeSubjectAttributes]

	// CodeChangeSubjectAttributes are the provider-neutral attributes persisted for code change events.
	CodeChangeSubjectAttributes struct {
		RepositoryExternalRef string `attr:"repository_external_ref" validate:"required"`
		DisplayName           string `attr:"display_name" validate:"required"`
	}
)

const SubjectKindCodeChange SubjectKind = "code_change"

func DecodeCodeChangeEvent(ev *ent.NormalizedEvent) (*CodeChangeEvent, error) {
	return DecodeSubjectAttributes[CodeChangeSubjectAttributes](ev)
}

type (
	// UserEvent is a normalized user observation from an organization or chat provider.
	UserEvent = Event[UserSubjectAttributes]

	// UserSubjectAttributes are the provider-neutral attributes persisted for user observations.
	UserSubjectAttributes struct {
		Name     string `attr:"name" validate:"required"`
		Email    string `attr:"email" validate:"required"`
		ChatId   string `attr:"chat_id"`
		Timezone string `attr:"timezone"`
	}
)

const SubjectKindUser SubjectKind = "User"

func DecodeUserEvent(ev *ent.NormalizedEvent) (*UserEvent, error) {
	return DecodeSubjectAttributes[UserSubjectAttributes](ev)
}

type (
	// IncidentEvent is a normalized incident observation from an incident provider.
	IncidentEvent = Event[IncidentSubjectAttributes]

	// IncidentSubjectAttributes are the provider-neutral attributes persisted for incident observations.
	IncidentSubjectAttributes struct {
		Title       string `attr:"title" validate:"required"`
		ExternalRef string `attr:"external_ref" validate:"required"`
		Summary     string `attr:"summary"`
		SeverityRef string `attr:"severity_ref" validate:"required"`
		TypeRef     string `attr:"type_ref" validate:"required"`
	}
)

const SubjectKindIncident SubjectKind = "Incident"

func DecodeIncidentEvent(ev *ent.NormalizedEvent) (*IncidentEvent, error) {
	return DecodeSubjectAttributes[IncidentSubjectAttributes](ev)
}

type (
	// AlertEvent is a normalized alert observation from an alerting provider.
	AlertEvent = Event[AlertSubjectAttributes]

	// AlertSubjectAttributes are the provider-neutral attributes persisted for alert observations.
	AlertSubjectAttributes struct {
		Title       string `attr:"title" validate:"required"`
		Description string `attr:"description"`
		Definition  string `attr:"definition"`
	}
)

const SubjectKindAlert SubjectKind = "Alert"

func DecodeAlertEvent(ev *ent.NormalizedEvent) (*AlertEvent, error) {
	return DecodeSubjectAttributes[AlertSubjectAttributes](ev)
}

type (
	// SystemComponentEvent is a normalized system component observation from a topology provider.
	SystemComponentEvent = Event[SystemComponentSubjectAttributes]

	// SystemComponentSubjectAttributes are the provider-neutral attributes persisted for system component observations.
	SystemComponentSubjectAttributes struct {
		ExternalRef string         `attr:"external_ref" validate:"required"`
		Kind        string         `attr:"kind" validate:"required"`
		DisplayName string         `attr:"display_name" validate:"required"`
		Description string         `attr:"description"`
		Properties  map[string]any `attr:"properties"`
	}
)

const SubjectKindSystemComponent SubjectKind = "SystemComponent"

func DecodeSystemComponentEvent(ev *ent.NormalizedEvent) (*SystemComponentEvent, error) {
	return DecodeSubjectAttributes[SystemComponentSubjectAttributes](ev)
}

type (
	// SystemRelationshipEvent is a normalized system relationship observation from a topology provider.
	SystemRelationshipEvent = Event[SystemRelationshipSubjectAttributes]

	// SystemRelationshipSubjectAttributes are the provider-neutral attributes persisted for system relationship observations.
	SystemRelationshipSubjectAttributes struct {
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

const SubjectKindSystemRelationship SubjectKind = "SystemRelationship"

func DecodeSystemRelationshipEvent(ev *ent.NormalizedEvent) (*SystemRelationshipEvent, error) {
	return DecodeSubjectAttributes[SystemRelationshipSubjectAttributes](ev)
}

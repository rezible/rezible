package projections

import (
	"time"

	"github.com/rezible/rezible/ent"
)

const attributeFieldNameTag = "json"

type (
	// ChatMessage is a normalized chat message observed from a messaging provider.
	ChatMessage = Event[ChatMessageAttributes]

	// ChatMessageAttributes are the provider-neutral attributes persisted for chat message events.
	ChatMessageAttributes struct {
		ConversationExternalRef string             `json:"conversation_external_ref" validate:"required"`
		Body                    string             `json:"body" validate:"required"`
		SenderExternalRef       string             `json:"sender_external_ref"`
		ThreadExternalRef       string             `json:"thread_external_ref"`
		RelatedEntities         []RelatedEntityRef `json:"related_entities"`
	}
)

func DecodeChatMessageEvent(ev *ent.NormalizedEvent) (*ChatMessage, error) {
	return DecodeSubjectAttributes[ChatMessageAttributes](ev)
}

type (
	CodeForgeEvent = Event[CodeForgeSubjectAttributes]

	// CodeForgeSubjectAttributes are the provider-neutral attributes persisted for repository observations.
	CodeForgeSubjectAttributes struct {
		DisplayName string `json:"display_name" validate:"required"`
		URL         string `json:"url"`
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
		RepositoryExternalRef string             `json:"repository_external_ref" validate:"required"`
		DisplayName           string             `json:"display_name" validate:"required"`
		RelatedEntities       []RelatedEntityRef `json:"related_entities"`
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
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required"`
		ChatId   string `json:"chat_id"`
		Timezone string `json:"timezone"`
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
		Title       string    `json:"title" validate:"required"`
		Summary     string    `json:"summary"`
		SeverityRef string    `json:"severity_ref" validate:"required"`
		TypeRef     string    `json:"type_ref" validate:"required"`
		OpenedAt    time.Time `json:"opened_at"`
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
		Title           string             `json:"title" validate:"required"`
		Description     string             `json:"description"`
		Definition      string             `json:"definition"`
		RelatedEntities []RelatedEntityRef `json:"related_entities"`
	}
)

const SubjectKindAlert SubjectKind = "Alert"

func DecodeAlertEvent(ev *ent.NormalizedEvent) (*AlertEvent, error) {
	return DecodeSubjectAttributes[AlertSubjectAttributes](ev)
}

type (
	PlaybookEvent = Event[PlaybookSubjectAttributes]

	PlaybookSubjectAttributes struct {
		Title         string   `json:"title" validate:"required"`
		Content       string   `json:"content" validate:"required"`
		RelatedAlerts []string `json:"related_alerts"`
	}
)

const SubjectKindPlaybook SubjectKind = "Playbook"

func DecodePlaybookEvent(ev *ent.NormalizedEvent) (*PlaybookEvent, error) {
	return DecodeSubjectAttributes[PlaybookSubjectAttributes](ev)
}

type (
	IncidentImpactEvent = Event[IncidentImpactSubjectAttributes]

	IncidentImpactSubjectAttributes struct {
		IncidentExternalRef string `json:"incident_external_ref" validate:"required"`
		EntityExternalRef   string `json:"entity_external_ref" validate:"required"`
		EntityKind          string `json:"entity_kind" validate:"required"`
		EntityDisplayName   string `json:"entity_display_name" validate:"required"`
		Source              string `json:"source"`
		Note                string `json:"note"`
	}
)

const SubjectKindIncidentImpact SubjectKind = "IncidentImpact"

func DecodeIncidentImpactEvent(ev *ent.NormalizedEvent) (*IncidentImpactEvent, error) {
	return DecodeSubjectAttributes[IncidentImpactSubjectAttributes](ev)
}

type (
	// SystemComponentEvent is a normalized system component observation from a topology provider.
	SystemComponentEvent = Event[SystemComponentSubjectAttributes]

	// SystemComponentSubjectAttributes are the provider-neutral attributes persisted for system component observations.
	SystemComponentSubjectAttributes struct {
		ExternalRef string         `json:"external_ref" validate:"required"`
		Kind        string         `json:"kind" validate:"required"`
		DisplayName string         `json:"display_name" validate:"required"`
		Description string         `json:"description"`
		Properties  map[string]any `json:"properties"`
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
		ExternalRef       string         `json:"external_ref" validate:"required"`
		Kind              string         `json:"kind" validate:"required"`
		DisplayName       string         `json:"display_name"`
		Description       string         `json:"description"`
		SourceExternalRef string         `json:"source_external_ref" validate:"required"`
		SourceKind        string         `json:"source_kind" validate:"required"`
		SourceDisplayName string         `json:"source_display_name" validate:"required"`
		TargetExternalRef string         `json:"target_external_ref" validate:"required"`
		TargetKind        string         `json:"target_kind" validate:"required"`
		TargetDisplayName string         `json:"target_display_name" validate:"required"`
		Properties        map[string]any `json:"properties"`
	}
)

const SubjectKindSystemRelationship SubjectKind = "SystemRelationship"

func DecodeSystemRelationshipEvent(ev *ent.NormalizedEvent) (*SystemRelationshipEvent, error) {
	return DecodeSubjectAttributes[SystemRelationshipSubjectAttributes](ev)
}

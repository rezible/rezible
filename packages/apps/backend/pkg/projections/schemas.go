package projections

import (
	"time"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

type SubjectKind string

func (k SubjectKind) String() string {
	return string(k)
}

func (k SubjectKind) Matches(ev *ent.NormalizedEvent) bool {
	return ev.SubjectKind == string(k)
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
		ExternalRef string `json:"external_ref" validate:"required"`
		Name        string `json:"name" validate:"required"`
		Email       string `json:"email" validate:"required"`
		ChatId      string `json:"chat_id"`
		Timezone    string `json:"timezone"`
	}
)

const SubjectKindUser SubjectKind = "user"

func DecodeUserEvent(ev *ent.NormalizedEvent) (*UserEvent, error) {
	return DecodeSubjectAttributes[UserSubjectAttributes](ev)
}

type (
	TeamEvent = Event[TeamSubjectAttributes]

	TeamSubjectAttributes struct {
		ExternalRef        string   `json:"external_ref" validate:"required"`
		Name               string   `json:"name" validate:"required"`
		Slug               string   `json:"slug" validate:"required"`
		ChatChannelId      string   `json:"chat_channel_id"`
		MemberExternalRefs []string `json:"member_external_refs"`
	}
)

const SubjectKindTeam SubjectKind = "team"

func DecodeTeamEvent(ev *ent.NormalizedEvent) (*TeamEvent, error) {
	return DecodeSubjectAttributes[TeamSubjectAttributes](ev)
}

type (
	TeamMembershipEvent = Event[TeamMembershipSubjectAttributes]

	TeamMembershipSubjectAttributes struct {
		Team TeamSubjectAttributes `json:"team" validate:"required"`
		User UserSubjectAttributes `json:"user" validate:"required"`
		Role string                `json:"role" validate:"oneof=admin member"`
	}
)

const SubjectKindTeamMembership SubjectKind = "team_membership"

func DecodeTeamMembershipEvent(ev *ent.NormalizedEvent) (*TeamMembershipEvent, error) {
	return DecodeSubjectAttributes[TeamMembershipSubjectAttributes](ev)
}

type (
	// IncidentEvent is a normalized incident observation from an incident provider.
	IncidentEvent = Event[IncidentSubjectAttributes]

	// IncidentSubjectAttributes are the provider-neutral attributes persisted for incident observations.
	IncidentSubjectAttributes struct {
		ExternalRef string    `json:"external_ref" validate:"required"`
		Title       string    `json:"title" validate:"required"`
		Summary     string    `json:"summary"`
		SeverityRef string    `json:"severity_ref" validate:"required"`
		TypeRef     string    `json:"type_ref" validate:"required"`
		OpenedAt    time.Time `json:"opened_at"`
	}
)

const SubjectKindIncident SubjectKind = "incident"

func DecodeIncidentEvent(ev *ent.NormalizedEvent) (*IncidentEvent, error) {
	return DecodeSubjectAttributes[IncidentSubjectAttributes](ev)
}

type (
	// AlertInstanceEvent is a normalized alert observation from an alerting provider.
	AlertInstanceEvent = Event[AlertInstanceSubjectAttributes]

	// AlertInstanceSubjectAttributes are the provider-neutral attributes persisted for alert observations.
	AlertInstanceSubjectAttributes struct {
		Title               string             `json:"title" validate:"required"`
		Description         string             `json:"description"`
		Definition          string             `json:"definition"`
		ExternalRef         string             `json:"external_ref" validate:"required"`
		InstanceExternalRef string             `json:"instance_external_ref"`
		RelatedEntities     []RelatedEntityRef `json:"related_entities"`
	}
)

const SubjectKindAlertInstance SubjectKind = "alert_instance"

func DecodeAlertInstanceEvent(ev *ent.NormalizedEvent) (*AlertInstanceEvent, error) {
	return DecodeSubjectAttributes[AlertInstanceSubjectAttributes](ev)
}

type (
	// SystemComponentEvent is a normalized system component observation from a topology provider.
	SystemComponentEvent = Event[SystemComponentSubjectAttributes]

	// SystemComponentSubjectAttributes are the provider-neutral attributes persisted for system component observations.
	SystemComponentSubjectAttributes struct {
		ExternalRef string         `json:"external_ref" validate:"required"`
		Category    kne.Category   `json:"category" validate:"required"`
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
		Predicate         knr.Predicate  `json:"predicate" validate:"required"`
		DisplayName       string         `json:"display_name"`
		Description       string         `json:"description"`
		SourceExternalRef string         `json:"source_external_ref" validate:"required"`
		SourceCategory    kne.Category   `json:"source_category" validate:"required"`
		SourceKind        string         `json:"source_kind" validate:"required"`
		SourceDisplayName string         `json:"source_display_name" validate:"required"`
		TargetExternalRef string         `json:"target_external_ref" validate:"required"`
		TargetCategory    kne.Category   `json:"target_category" validate:"required"`
		TargetKind        string         `json:"target_kind" validate:"required"`
		TargetDisplayName string         `json:"target_display_name" validate:"required"`
		Properties        map[string]any `json:"properties"`
	}
)

const SubjectKindSystemRelationship SubjectKind = "SystemRelationship"

func DecodeSystemRelationshipEvent(ev *ent.NormalizedEvent) (*SystemRelationshipEvent, error) {
	return DecodeSubjectAttributes[SystemRelationshipSubjectAttributes](ev)
}

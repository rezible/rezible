package projections

import (
	"fmt"
	"strconv"
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
		DisplayName       string            `json:"display_name" validate:"required"`
		URL               string            `json:"url"`
		LinkingAttributes LinkingAttributes `json:"linking_attributes,omitempty"`
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
		// Merge is set when the change was merged.
		Merge *CodeChangeMerge `json:"merge,omitempty"`
	}

	// CodeChangeMerge is what linking a merged change to deployments needs.
	CodeChangeMerge struct {
		MergedAt time.Time `json:"merged_at" validate:"required"`
		// MergeCommitSha is lower-case.
		MergeCommitSha string `json:"merge_commit_sha" validate:"required"`
		BaseRef        string `json:"base_ref" validate:"required"`
		// IntoDefaultBranch is whether the base ref is the repository's default branch.
		IntoDefaultBranch bool   `json:"into_default_branch"`
		Number            int    `json:"number" validate:"required"`
		URL               string `json:"url"`
		AuthorLogin       string `json:"author_login"`
	}
)

const KindCodeChange = "code_change"

// The merged code change entity's state properties. Every value is a string.
const (
	CodeChangePropertyMergedAt          = "merged_at"
	CodeChangePropertyMergeCommitSha    = "merge_commit_sha"
	CodeChangePropertyBaseRef           = "base_ref"
	CodeChangePropertyIntoDefaultBranch = "into_default_branch"
	CodeChangePropertyNumber            = "number"
	CodeChangePropertyURL               = "url"
	CodeChangePropertyAuthor            = "author"
)

// StateProperties are the merged code change entity's state properties, omitting an absent URL or author.
func (m CodeChangeMerge) StateProperties() map[string]any {
	properties := map[string]any{
		CodeChangePropertyMergedAt:          m.MergedAt.UTC().Format(time.RFC3339Nano),
		CodeChangePropertyMergeCommitSha:    m.MergeCommitSha,
		CodeChangePropertyBaseRef:           m.BaseRef,
		CodeChangePropertyIntoDefaultBranch: strconv.FormatBool(m.IntoDefaultBranch),
		CodeChangePropertyNumber:            strconv.Itoa(m.Number),
	}
	if m.URL != "" {
		properties[CodeChangePropertyURL] = m.URL
	}
	if m.AuthorLogin != "" {
		properties[CodeChangePropertyAuthor] = m.AuthorLogin
	}
	return properties
}

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

// Linking attributes shared by providers that name the same service or repository.
const (
	// LinkingAttributeServiceName is a service's name, normalized with NormalizeServiceName.
	LinkingAttributeServiceName = "service.name"
	// LinkingAttributeRepositoryFullName is a repository's `owner/name`, lower-cased.
	LinkingAttributeRepositoryFullName = "repository.full_name"
)

// LinkingAttributes are an observation's linking attribute values, already normalized by the provider.
type LinkingAttributes map[string]string

func (a LinkingAttributes) Values() map[string]string {
	return a
}

// NormalizeServiceName lower-cases the value, replaces every run of characters outside [a-z0-9] with `-`
// and trims `-`.
func NormalizeServiceName(value string) string {
	var b strings.Builder
	separated := false
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			separated = false
		} else if !separated {
			b.WriteByte('-')
			separated = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// NormalizeRepositoryFullName canonicalizes a repository's `owner/name` for matching across providers.
func NormalizeRepositoryFullName(fullName string) string {
	return strings.ToLower(strings.TrimSpace(fullName))
}

// RepositoryLinkingAttributes link a repository to other providers' observations of the same `owner/name`.
func RepositoryLinkingAttributes(fullName string) LinkingAttributes {
	normalized := NormalizeRepositoryFullName(fullName)
	if normalized == "" {
		return nil
	}
	return LinkingAttributes{
		LinkingAttributeRepositoryFullName: normalized,
	}
}

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

type (
	// DeploymentEvent is one report of a deployment's progress.
	DeploymentEvent = Event[DeploymentEventAttributes]

	// DeploymentEventAttributes are the provider-neutral attributes persisted for a deployment report. The
	// event's provider resource ref identifies the deployment.
	DeploymentEventAttributes struct {
		// ExternalID is the caller's ID for the deployment.
		ExternalID  string                `json:"external_id" validate:"required"`
		Status      string                `json:"status" validate:"required,oneof=started succeeded failed"`
		Service     EntityObservation     `json:"service" validate:"required"`
		Environment DeploymentEnvironment `json:"environment" validate:"required"`
		Repository  *EntityObservation    `json:"repository,omitempty"`
		Sha         string                `json:"sha,omitempty"`
		Version     string                `json:"version,omitempty"`
		URL         string                `json:"url,omitempty"`
		StartedAt   *time.Time            `json:"started_at,omitempty"`
		FinishedAt  *time.Time            `json:"finished_at,omitempty"`
	}

	// DeploymentEnvironment names where a deployment went. It is not an entity.
	DeploymentEnvironment struct {
		// Name is normalized with NormalizeServiceName.
		Name        string `json:"name" validate:"required"`
		DisplayName string `json:"display_name" validate:"required"`
	}
)

const KindDeployment = "deployment"

// The deployment entity's state properties. Every value is a string.
const (
	DeploymentPropertyEnvironment = "environment"
	DeploymentPropertyStatus      = "status"
	DeploymentPropertyExternalID  = "external_id"
	DeploymentPropertySha         = "sha"
	DeploymentPropertyVersion     = "version"
	DeploymentPropertyURL         = "url"
	DeploymentPropertyRepository  = "repository"
	DeploymentPropertyStartedAt   = "started_at"
	DeploymentPropertyFinishedAt  = "finished_at"
)

const (
	DeploymentStatusStarted   = "started"
	DeploymentStatusSucceeded = "succeeded"
	DeploymentStatusFailed    = "failed"
)

// Finished reports whether the report's status is terminal.
func (a DeploymentEventAttributes) Finished() bool {
	return a.Status == DeploymentStatusSucceeded || a.Status == DeploymentStatusFailed
}

// StateProperties are the deployment entity's state properties, omitting absent values. A report is complete,
// so they replace the previous report's properties whole.
func (a DeploymentEventAttributes) StateProperties() map[string]any {
	properties := map[string]any{
		DeploymentPropertyEnvironment: a.Environment.Name,
		DeploymentPropertyStatus:      a.Status,
		DeploymentPropertyExternalID:  a.ExternalID,
	}
	optional := map[string]string{
		DeploymentPropertySha:     a.Sha,
		DeploymentPropertyVersion: a.Version,
		DeploymentPropertyURL:     a.URL,
	}
	if a.Repository != nil {
		optional[DeploymentPropertyRepository] = a.Repository.LinkingAttributes[LinkingAttributeRepositoryFullName]
	}
	if a.StartedAt != nil {
		optional[DeploymentPropertyStartedAt] = a.StartedAt.UTC().Format(time.RFC3339Nano)
	}
	if a.FinishedAt != nil {
		optional[DeploymentPropertyFinishedAt] = a.FinishedAt.UTC().Format(time.RFC3339Nano)
	}
	for key, value := range optional {
		if value != "" {
			properties[key] = value
		}
	}
	return properties
}

func DecodeDeploymentEvent(ev *ent.NormalizedEvent) (*DeploymentEvent, error) {
	return DecodeEventAttributes[DeploymentEventAttributes](ev)
}

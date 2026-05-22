package projections

import "github.com/rezible/rezible/ent"

const (
	attrDisplayName = "display_name"
	attrURL         = "url"
)

type (
	RepositoryObserved           = Event[RepositoryObservedAttributes]
	RepositoryObservedAttributes struct {
		DisplayName string
		URL         string
	}
)

func (a RepositoryObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrDisplayName: a.DisplayName,
		attrURL:         a.URL,
	}
}

func DecodeRepositoryObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	displayName, nameErr := requiredString(ev, attrDisplayName)
	if nameErr != nil {
		return nil, nameErr
	}
	url, urlErr := optionalString(ev, attrURL)
	if urlErr != nil {
		return nil, urlErr
	}
	attrs := RepositoryObservedAttributes{
		DisplayName: displayName,
		URL:         url,
	}
	return attrs, nil
}

type (
	// TODO: update to CodeForgeEvent

	ChangeEventObserved           = Event[ChangeEventObservedAttributes]
	ChangeEventObservedAttributes struct {
		RepositoryExternalRef string
		DisplayName           string
	}
)

const (
	attrRepositoryExternalRef = "repository_external_ref"
)

func (a ChangeEventObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrDisplayName:           a.DisplayName,
		attrRepositoryExternalRef: a.RepositoryExternalRef,
	}
}

func DecodeChangeEventObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	repoRef, repoRefErr := requiredString(ev, attrRepositoryExternalRef)
	if repoRefErr != nil {
		return nil, repoRefErr
	}
	displayName, displayNameErr := requiredString(ev, attrDisplayName)
	if displayNameErr != nil {
		return nil, displayNameErr
	}
	attrs := ChangeEventObservedAttributes{
		RepositoryExternalRef: repoRef,
		DisplayName:           displayName,
	}
	return attrs, nil
}

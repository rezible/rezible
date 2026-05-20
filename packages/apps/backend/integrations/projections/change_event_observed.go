package projections

import "github.com/rezible/rezible/ent"

// TODO: update to CodeForgeEvent

const attrRepositoryExternalRef = "repository_external_ref"

type (
	ChangeEventObserved           = Event[ChangeEventObservedAttributes]
	ChangeEventObservedAttributes struct {
		RepositoryExternalRef string
		DisplayName           string
	}
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

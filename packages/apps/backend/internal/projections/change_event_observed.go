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

func decodeChangeEventObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	attrs, err := decodeChangeEventObserved(ev)
	if err != nil {
		return nil, err
	}
	return ChangeEventObserved{Event: ev, Attributes: attrs}, nil
}

func decodeChangeEventObserved(ev *ent.NormalizedEvent) (ChangeEventObservedAttributes, error) {
	if err := rejectUnsupportedAttributes(ev, attrRepositoryExternalRef, attrDisplayName); err != nil {
		return ChangeEventObservedAttributes{}, err
	}
	repoRef, err := requiredString(ev, attrRepositoryExternalRef)
	if err != nil {
		return ChangeEventObservedAttributes{}, err
	}
	displayName, err := requiredString(ev, attrDisplayName)
	if err != nil {
		return ChangeEventObservedAttributes{}, err
	}
	return ChangeEventObservedAttributes{
		RepositoryExternalRef: repoRef,
		DisplayName:           displayName,
	}, nil
}

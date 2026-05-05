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

func decodeRepositoryObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	attrs, err := decodeRepositoryObserved(ev)
	if err != nil {
		return nil, err
	}
	return RepositoryObserved{Event: ev, Attributes: attrs}, nil
}

func decodeRepositoryObserved(ev *ent.NormalizedEvent) (RepositoryObservedAttributes, error) {
	if err := rejectUnsupportedAttributes(ev, attrDisplayName, attrURL); err != nil {
		return RepositoryObservedAttributes{}, err
	}
	displayName, err := requiredString(ev, attrDisplayName)
	if err != nil {
		return RepositoryObservedAttributes{}, err
	}
	url, err := optionalString(ev, attrURL)
	if err != nil {
		return RepositoryObservedAttributes{}, err
	}
	return RepositoryObservedAttributes{
		DisplayName: displayName,
		URL:         url,
	}, nil
}

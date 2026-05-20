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

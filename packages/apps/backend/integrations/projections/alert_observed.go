package projections

import "github.com/rezible/rezible/ent"

const (
	attrDescription = "description"
	attrDefinition  = "definition"
)

type (
	AlertObserved           = Event[AlertObservedAttributes]
	AlertObservedAttributes struct {
		Title       string
		Description string
		Definition  string
	}
)

func (a AlertObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrTitle:       a.Title,
		attrDescription: a.Description,
		attrDefinition:  a.Definition,
	}
}

func DecodeAlertObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	title, titleErr := requiredString(ev, attrTitle)
	if titleErr != nil {
		return nil, titleErr
	}
	description, descriptionErr := optionalString(ev, attrDescription)
	if descriptionErr != nil {
		return nil, descriptionErr
	}
	definition, definitionErr := optionalString(ev, attrDefinition)
	if definitionErr != nil {
		return nil, definitionErr
	}
	attrs := AlertObservedAttributes{
		Title:       title,
		Description: description,
		Definition:  definition,
	}
	return attrs, nil
}

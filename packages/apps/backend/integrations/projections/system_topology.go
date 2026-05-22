package projections

import (
	"github.com/rezible/rezible/ent"
)

const (
	attrExternalRef = "external_ref"
	attrKind        = "kind"
	attrProperties  = "properties"

	attrSourceExternalRef = "source_external_ref"
	attrSourceKind        = "source_kind"
	attrSourceDisplayName = "source_display_name"
	attrTargetExternalRef = "target_external_ref"
	attrTargetKind        = "target_kind"
	attrTargetDisplayName = "target_display_name"
)

type (
	SystemComponentObserved           = Event[SystemComponentObservedAttributes]
	SystemComponentObservedAttributes struct {
		ExternalRef string
		Kind        string
		DisplayName string
		Description string
		Properties  map[string]any
	}
)

func (a SystemComponentObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrExternalRef: a.ExternalRef,
		attrKind:        a.Kind,
		attrDisplayName: a.DisplayName,
		attrDescription: a.Description,
		attrProperties:  a.Properties,
	}
}

func DecodeSystemComponentObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	externalRef, externalRefErr := requiredString(ev, attrExternalRef)
	if externalRefErr != nil {
		return nil, externalRefErr
	}
	kind, kindErr := requiredString(ev, attrKind)
	if kindErr != nil {
		return nil, kindErr
	}
	displayName, displayNameErr := requiredString(ev, attrDisplayName)
	if displayNameErr != nil {
		return nil, displayNameErr
	}
	description, descriptionErr := optionalString(ev, attrDescription)
	if descriptionErr != nil {
		return nil, descriptionErr
	}
	properties, propertiesErr := optionalMap(ev, attrProperties)
	if propertiesErr != nil {
		return nil, propertiesErr
	}
	return SystemComponentObservedAttributes{
		ExternalRef: externalRef,
		Kind:        kind,
		DisplayName: displayName,
		Description: description,
		Properties:  properties,
	}, nil
}

type (
	SystemRelationshipObserved           = Event[SystemRelationshipObservedAttributes]
	SystemRelationshipObservedAttributes struct {
		ExternalRef       string
		Kind              string
		DisplayName       string
		Description       string
		SourceExternalRef string
		SourceKind        string
		SourceDisplayName string
		TargetExternalRef string
		TargetKind        string
		TargetDisplayName string
		Properties        map[string]any
	}
)

func (a SystemRelationshipObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrExternalRef:       a.ExternalRef,
		attrKind:              a.Kind,
		attrDisplayName:       a.DisplayName,
		attrDescription:       a.Description,
		attrSourceExternalRef: a.SourceExternalRef,
		attrSourceKind:        a.SourceKind,
		attrSourceDisplayName: a.SourceDisplayName,
		attrTargetExternalRef: a.TargetExternalRef,
		attrTargetKind:        a.TargetKind,
		attrTargetDisplayName: a.TargetDisplayName,
		attrProperties:        a.Properties,
	}
}

func DecodeSystemRelationshipObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	externalRef, externalRefErr := requiredString(ev, attrExternalRef)
	if externalRefErr != nil {
		return nil, externalRefErr
	}
	kind, kindErr := requiredString(ev, attrKind)
	if kindErr != nil {
		return nil, kindErr
	}
	displayName, displayNameErr := optionalString(ev, attrDisplayName)
	if displayNameErr != nil {
		return nil, displayNameErr
	}
	sourceExternalRef, sourceExternalRefErr := requiredString(ev, attrSourceExternalRef)
	if sourceExternalRefErr != nil {
		return nil, sourceExternalRefErr
	}
	sourceKind, sourceKindErr := requiredString(ev, attrSourceKind)
	if sourceKindErr != nil {
		return nil, sourceKindErr
	}
	sourceDisplayName, sourceDisplayNameErr := requiredString(ev, attrSourceDisplayName)
	if sourceDisplayNameErr != nil {
		return nil, sourceDisplayNameErr
	}
	targetExternalRef, targetExternalRefErr := requiredString(ev, attrTargetExternalRef)
	if targetExternalRefErr != nil {
		return nil, targetExternalRefErr
	}
	targetKind, targetKindErr := requiredString(ev, attrTargetKind)
	if targetKindErr != nil {
		return nil, targetKindErr
	}
	targetDisplayName, targetDisplayNameErr := requiredString(ev, attrTargetDisplayName)
	if targetDisplayNameErr != nil {
		return nil, targetDisplayNameErr
	}
	description, descriptionErr := optionalString(ev, attrDescription)
	if descriptionErr != nil {
		return nil, descriptionErr
	}
	properties, propertiesErr := optionalMap(ev, attrProperties)
	if propertiesErr != nil {
		return nil, propertiesErr
	}
	return SystemRelationshipObservedAttributes{
		ExternalRef:       externalRef,
		Kind:              kind,
		DisplayName:       displayName,
		Description:       description,
		SourceExternalRef: sourceExternalRef,
		SourceKind:        sourceKind,
		SourceDisplayName: sourceDisplayName,
		TargetExternalRef: targetExternalRef,
		TargetKind:        targetKind,
		TargetDisplayName: targetDisplayName,
		Properties:        properties,
	}, nil
}

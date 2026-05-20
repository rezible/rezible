package projections

import "github.com/rezible/rezible/ent"

const (
	attrTitle        = "title"
	attrSummary      = "summary"
	attrSeverityName = "severity_name"
	attrSeverityRank = "severity_rank"
	attrTypeName     = "type_name"
)

type (
	IncidentObserved           = Event[IncidentObservedAttributes]
	IncidentObservedAttributes struct {
		Title        string
		Summary      string
		SeverityName string
		SeverityRank int
		TypeName     string
	}
)

func (a IncidentObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrTitle:        a.Title,
		attrSummary:      a.Summary,
		attrSeverityName: a.SeverityName,
		attrSeverityRank: a.SeverityRank,
		attrTypeName:     a.TypeName,
	}
}

func DecodeIncidentObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	title, titleErr := requiredString(ev, attrTitle)
	if titleErr != nil {
		return nil, titleErr
	}
	summary, summaryErr := optionalString(ev, attrSummary)
	if summaryErr != nil {
		return nil, summaryErr
	}
	severityName, severityNameErr := requiredString(ev, attrSeverityName)
	if severityNameErr != nil {
		return nil, severityNameErr
	}
	severityRank, severityRankErr := requiredInt(ev, attrSeverityRank)
	if severityRankErr != nil {
		return nil, severityRankErr
	}
	typeName, typeNameErr := requiredString(ev, attrTypeName)
	if typeNameErr != nil {
		return nil, typeNameErr
	}
	attrs := IncidentObservedAttributes{
		Title:        title,
		Summary:      summary,
		SeverityName: severityName,
		SeverityRank: severityRank,
		TypeName:     typeName,
	}
	return attrs, nil
}

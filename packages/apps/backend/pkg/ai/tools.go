package ai

import (
	"github.com/rezible/rezible/ent/schema/schematypes"
)

type ToolDefinition[I any, O any] struct {
	name        string
	description string
}

func (d ToolDefinition[I, O]) Name() string {
	return d.name
}

func (d ToolDefinition[I, O]) Description() string {
	return d.description
}

func defineTool[D ToolDefinition[I, O], I any, O any](name, description string) D {
	return D{name: name, description: description}
}

type (
	SaveAlertInvestigationReportInput struct {
		Report schematypes.AlertInvestigationReport `json:"report"`
	}

	SaveAlertInvestigationReportOutput struct {
		Saved bool `json:"saved"`
	}
)

var SaveAlertInvestigationReportTool = defineTool[ToolDefinition[SaveAlertInvestigationReportInput, SaveAlertInvestigationReportOutput]](
	"save_alert_investigation_report",
	"Save the alert investigation report for this agent session.",
)

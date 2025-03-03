package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentevent"
	oapi "github.com/rezible/rezible/openapi"
)

type incidentEventsHandler struct {
	db *ent.Client
}

func newIncidentEventsHandler(db *ent.Client) *incidentEventsHandler {
	return &incidentEventsHandler{db}
}

func (h *incidentEventsHandler) ListIncidentEvents(ctx context.Context, request *oapi.ListIncidentEventsRequest) (*oapi.ListIncidentEventsResponse, error) {
	var resp oapi.ListIncidentEventsResponse

	query := h.db.IncidentEvent.Query().
		Where(incidentevent.IncidentID(request.Id))
	events, eventsErr := query.All(ctx)
	if eventsErr != nil {
		return nil, detailError("failed to list incident events", eventsErr)
	}
	resp.Body.Data = make([]oapi.IncidentEvent, len(events))
	for i, e := range events {
		resp.Body.Data[i] = oapi.IncidentEventFromEnt(e)
	}

	return &resp, nil
}

func (h *incidentEventsHandler) CreateIncidentEvent(ctx context.Context, request *oapi.CreateIncidentEventRequest) (*oapi.CreateIncidentEventResponse, error) {
	var resp oapi.CreateIncidentEventResponse

	attr := request.Body.Attributes
	create := h.db.IncidentEvent.Create().
		SetIncidentID(request.Id).
		SetTitle(attr.Title).
		SetTimestamp(attr.Timestamp)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentEventFromEnt(created)

	return &resp, nil
}

func (h *incidentEventsHandler) UpdateIncidentEvent(ctx context.Context, request *oapi.UpdateIncidentEventRequest) (*oapi.UpdateIncidentEventResponse, error) {
	var resp oapi.UpdateIncidentEventResponse

	attr := request.Body.Attributes
	update := h.db.IncidentEvent.UpdateOneID(request.Id).
		SetNillableTitle(attr.Title).
		SetNillableTimestamp(attr.Timestamp)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update incident event", updateErr)
	}
	resp.Body.Data = oapi.IncidentEventFromEnt(updated)

	return &resp, nil
}

func (h *incidentEventsHandler) DeleteIncidentEvent(ctx context.Context, request *oapi.DeleteIncidentEventRequest) (*oapi.DeleteIncidentEventResponse, error) {
	var resp oapi.DeleteIncidentEventResponse

	if deleteErr := h.db.IncidentEvent.DeleteOneID(request.Id).Exec(ctx); deleteErr != nil {
		return nil, detailError("failed to delete incident event", deleteErr)
	}

	return &resp, nil
}

// TODO: allow customising these
var incidentEventFactorCategories = []oapi.IncidentEventContributingFactorCategory{
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "Organizational Pressures",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Time Pressure",
						Description: "Team was pushing to meet end-of-sprint commitments, leading to rushed testing",
						Examples: []string{
							"End of quarter release pressure",
							"Multiple concurrent project deadlines",
							"Customer commitment deadlines",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Resource Constraints",
						Description: "Only one engineer familiar with the system was available during the incident",
						Examples: []string{
							"Team understaffing",
							"Limited expert availability",
							"Budget constraints affecting tooling",
						},
					},
				},
			},
		},
	},
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "Knowledge & Visibility",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Missing Information",
						Description: "",
						Examples: []string{
							"Outdated runbooks",
							"Undocumented system interactions",
							"Missing architectural diagrams",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Monitoring Gap",
						Description: "No alerts existed for gradual database connection pool exhaustion",
						Examples: []string{
							"Insufficient metrics coverage",
							"Missing threshold alerts",
							"Incomplete logging",
						},
					},
				},
			},
		},
	},
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "Process & Coordination",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Communication Breakdown",
						Description: "Team wasn't sure when to involve senior engineers or wake up team leads",
						Examples: []string{
							"Unclear escalation path",
							"Unclear incident roles",
							"Communication tool issues",
							"Cross-team coordination challenges",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Process Uncertainty",
						Description: "No clear threshold for when to revert the deployment vs. trying to fix forward",
						Examples: []string{
							"Missing playbooks",
							"Undefined rollback criteria",
							"Unclear decision authority",
							"Undefined incident severity levels",
						},
					},
				},
			},
		},
	},
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "Technical Complexity",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "System Opacity",
						Description: "Initial Redis timeout led to unexpected cascading failures in multiple services",
						Examples: []string{
							"Complex failure cascade",
							"Hidden dependencies",
							"Unclear failure modes",
							"Complex state management",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Technical Debt",
						Description: "Old monitoring system couldn't be easily updated to catch new failure modes",
						Examples: []string{
							"Legacy system constraints",
							"Outdated infrastructure",
							"Hard-to-maintain code",
							"Technical workarounds",
						},
					},
				},
			},
		},
	},
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "Change Management",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Configuration Complexity",
						Description: "Multiple similar configuration parameters made it easy to modify the wrong one",
						Examples: []string{
							"Risky configuration surface",
							"Complex configuration options",
							"Manual configuration steps",
							"Configuration drift",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Testing Limitations",
						Description: "Test environment didn't accurately reflect production load patterns",
						Examples: []string{
							"Missing test coverage",
							"Environment differences",
							"Limited load testing",
						},
					},
				},
			},
		},
	},
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "Human Factors",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Cognitive Load",
						Description: "High volume of low-priority alerts led to missing critical warnings",
						Examples: []string{
							"Alert fatigue",
							"Information overload",
							"Fatigue during long incident",
							"Multiple concurrent issues",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Experience Mismatch",
						Description: "Engineer was handling an incident in a system they rarely work with",
						Examples: []string{
							"Unfamiliar territory",
							"New team members",
							"Cross-team coverage",
							"Rare system interactions",
						},
					},
				},
			},
		},
	},
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentEventContributingFactorCategoryAttributes{
			Label: "External Factors",
			FactorTypes: []oapi.IncidentEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Vendor Issues",
						Description: "Unable to quickly scale due to cloud provider quota restrictions",
						Examples: []string{
							"Cloud provider limitations",
							"Provider outages",
							"API limitations",
							"Third-party dependencies",
						},
					},
				},
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentEventContributingFactorTypeAttributes{
						Label:       "Customer Behaviour",
						Description: "New customer onboarding caused unexpected load spikes",
						Examples: []string{
							"Unexpected/Changed usage patterns",
							"Traffic spikes",
							"Customer configuration issues",
						},
					},
				},
			},
		},
	},
}

func (h *incidentEventsHandler) ListIncidentEventContributingFactors(ctx context.Context, request *oapi.ListIncidentEventContributingFactorsRequest) (*oapi.ListIncidentEventContributingFactorsResponse, error) {
	var resp oapi.ListIncidentEventContributingFactorsResponse

	resp.Body.Data = incidentEventFactorCategories

	return &resp, nil
}

package apiv1

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"

	itle "github.com/rezible/rezible/ent/incidenttimelineevent"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type incidentTimelineHandler struct {
	db rez.Database
}

func newIncidentTimelineHandler(db rez.Database) *incidentTimelineHandler {
	return &incidentTimelineHandler{db: db}
}

func (h *incidentTimelineHandler) ListIncidentTimelineEvents(ctx context.Context, request *oapi.ListIncidentTimelineEventsRequest) (*oapi.ListIncidentTimelineEventsResponse, error) {
	var resp oapi.ListIncidentTimelineEventsResponse

	query := h.db.Client(ctx).IncidentTimelineEvent.Query().
		Where(itle.IncidentID(request.Id)).
		WithTopologyContext()
	events, eventsErr := query.All(ctx)
	if eventsErr != nil {
		return nil, oapi.Error(ctx, "failed to list incident events", eventsErr)
	}
	resp.Body.Data = make([]oapi.IncidentTimelineEvent, len(events))
	for i, e := range events {
		resp.Body.Data[i] = oapi.IncidentTimelineEventFromEnt(e)
	}

	return &resp, nil
}

func (h *incidentTimelineHandler) getEventSequence(ctx context.Context, incidentId uuid.UUID, timestamp time.Time) (int, error) {
	query := h.db.Client(ctx).IncidentTimelineEvent.Query().
		Where(itle.And(itle.IncidentID(incidentId), itle.Timestamp(timestamp)))

	num, countErr := query.Count(ctx)
	if countErr != nil {
		return -1, countErr
	}
	return num + 1, nil
}

func (h *incidentTimelineHandler) CreateIncidentTimelineEvent(ctx context.Context, request *oapi.CreateIncidentTimelineEventRequest) (*oapi.CreateIncidentTimelineEventResponse, error) {
	var resp oapi.CreateIncidentTimelineEventResponse

	attr := request.Body.Attributes

	kind := itle.Kind(attr.Kind)
	if kindErr := itle.KindValidator(kind); kindErr != nil {
		return nil, oapi.Error(ctx, "invalid kind", kindErr)
	}

	// userId := requestUserId(ctx, h.auth)

	sequence, seqErr := h.getEventSequence(ctx, request.Id, attr.Timestamp)
	if seqErr != nil {
		return nil, oapi.Error(ctx, "failed to get sequence for incident event", seqErr)
	}

	create := h.db.Client(ctx).IncidentTimelineEvent.Create().
		SetIncidentID(request.Id).
		SetTitle(attr.Title).
		SetKind(kind).
		SetIsKey(attr.IsKey).
		SetTimestamp(attr.Timestamp).
		SetSequence(sequence)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		slog.Error("failed to create", "error", createErr)
		return nil, oapi.Error(ctx, "failed to create incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentTimelineEventFromEnt(created)

	return &resp, nil
}

func (h *incidentTimelineHandler) UpdateIncidentTimelineEvent(ctx context.Context, request *oapi.UpdateIncidentTimelineEventRequest) (*oapi.UpdateIncidentTimelineEventResponse, error) {
	var resp oapi.UpdateIncidentTimelineEventResponse

	attr := request.Body.Attributes

	update := h.db.Client(ctx).IncidentTimelineEvent.UpdateOneID(request.Id).
		SetNillableTitle(attr.Title).
		SetNillableTimestamp(attr.Timestamp)

	if attr.Kind != nil {
		kind := itle.Kind(*attr.Kind)
		if kindErr := itle.KindValidator(kind); kindErr != nil {
			return nil, oapi.Error(ctx, "invalid kind", kindErr)
		}
		update.SetKind(kind)
	}

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to update incident event", updateErr)
	}
	resp.Body.Data = oapi.IncidentTimelineEventFromEnt(updated)

	return &resp, nil
}

func (h *incidentTimelineHandler) DeleteIncidentTimelineEvent(ctx context.Context, request *oapi.DeleteIncidentTimelineEventRequest) (*oapi.DeleteIncidentTimelineEventResponse, error) {
	var resp oapi.DeleteIncidentTimelineEventResponse

	if deleteErr := h.db.Client(ctx).IncidentTimelineEvent.DeleteOneID(request.Id).Exec(ctx); deleteErr != nil {
		return nil, oapi.Error(ctx, "failed to delete incident event", deleteErr)
	}

	return &resp, nil
}

// TODO: allow customising these
var incidentEventFactorCategories = []oapi.IncidentTimelineEventContributingFactorCategory{
	{
		Id: uuid.New(),
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "Organizational Pressures",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "Knowledge & Visibility",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "Process & Coordination",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "Technical Complexity",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "Change Management",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "Human Factors",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
		Attributes: oapi.IncidentTimelineEventContributingFactorCategoryAttributes{
			Label: "External Factors",
			FactorTypes: []oapi.IncidentTimelineEventContributingFactorType{
				{
					Id: uuid.New(),
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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
					Attributes: oapi.IncidentTimelineEventContributingFactorTypeAttributes{
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

func (h *incidentTimelineHandler) GetIncidentTimelineEventMetadata(ctx context.Context, request *oapi.GetIncidentTimelineEventMetadataRequest) (*oapi.GetIncidentTimelineEventMetadataResponse, error) {
	var resp oapi.GetIncidentTimelineEventMetadataResponse

	resp.Body.Data = oapi.IncidentTimelineEventMetadata{
		ContributingFactorCategories: incidentEventFactorCategories,
	}

	return &resp, nil
}

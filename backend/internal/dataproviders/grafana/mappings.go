package grafana

import (
	"time"

	"github.com/rezible/rezible/ent"
)

// mappings respresent an object with non-zero fields for data that a provider can map

var (
	timeSupported = time.Now()
	strSupported  = "y"
)

var (
	incidentRoleDataMapping = ent.IncidentRole{
		ArchiveTime: timeSupported,
		Name:        strSupported,
		Required:    true,
	}
	incidentUserDataMapping = ent.User{
		Email:  strSupported,
		ChatID: strSupported,
		Name:   strSupported,
	}

	incidentSeverityMapping  = ent.IncidentSeverity{Name: strSupported}
	incidentTypeMapping      = ent.IncidentType{Name: strSupported}
	incidentTaskMapping      = ent.Task{Title: strSupported}
	incidentMilestoneMapping = ent.IncidentMilestone{Time: timeSupported}
	incidentTagMapping       = ent.IncidentTag{
		Key:   strSupported,
		Value: strSupported,
	}
	incidentRoleAssignmentMapping = ent.IncidentRoleAssignment{
		Edges: ent.IncidentRoleAssignmentEdges{
			Role: &incidentRoleDataMapping,
			User: &incidentUserDataMapping,
		},
	}

	incidentDataMapping = ent.Incident{
		Title:         strSupported,
		Summary:       strSupported,
		OpenedAt:      timeSupported,
		ModifiedAt:    timeSupported,
		ClosedAt:      timeSupported,
		ProviderID:    strSupported,
		ChatChannelID: strSupported,
		Edges: ent.IncidentEdges{
			Severity:        &incidentSeverityMapping,
			Type:            &incidentTypeMapping,
			RoleAssignments: []*ent.IncidentRoleAssignment{&incidentRoleAssignmentMapping},
			Milestones:      []*ent.IncidentMilestone{&incidentMilestoneMapping},
			Tasks:           []*ent.Task{&incidentTaskMapping},
			TagAssignments:  []*ent.IncidentTag{&incidentTagMapping},
		},
	}
)

var (
	rosterMapping = ent.OncallRoster{}

	shiftMapping = ent.OncallShift{}
)

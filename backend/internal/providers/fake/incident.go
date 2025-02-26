package fakeprovider

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type IncidentDataProvider struct {
	onIncidentUpdatedFn rez.DataProviderResourceUpdatedCallback
	fakeRoles           []*ent.IncidentRole
	fakeIncidents       []*ent.Incident
}

type IncidentDataProviderConfig struct {
}

func NewIncidentDataProvider(cfg IncidentDataProviderConfig) (*IncidentDataProvider, error) {
	p := &IncidentDataProvider{
		onIncidentUpdatedFn: func(id string, m time.Time) {
			log.Warn().Msg("no onIncidentUpdated function")
		},
	}

	p.makeFakeIncidentRoles()
	p.makeFakeIncidents()

	return p, nil
}

func (p *IncidentDataProvider) makeFakeIncidentRoles() {
	role := &ent.IncidentRole{Name: "Role 1"}

	p.fakeRoles = []*ent.IncidentRole{role}
}

func (p *IncidentDataProvider) makeFakeIncidents() {
	severity := &ent.IncidentSeverity{
		Name:        "Severity 1",
		Description: "a severity",
	}
	incType := &ent.IncidentType{Name: "Default"}
	tasks := []*ent.Task{{Title: "A Task"}}

	tags := []*ent.IncidentTag{{Key: "foo", Value: "bar"}}

	user := &ent.User{
		Name:   "User 1",
		Email:  "user@example.com",
		ChatID: "foo",
	}

	roles := []*ent.IncidentRoleAssignment{
		{Edges: ent.IncidentRoleAssignmentEdges{
			Role: p.fakeRoles[0],
			User: user,
		}},
	}

	milestones := []*ent.IncidentMilestone{
		{
			Type:  "",
			Time:  time.Now().Add(-8 * time.Hour),
			Edges: ent.IncidentMilestoneEdges{},
		},
	}

	inc1 := &ent.Incident{
		ProviderID:    "test-incident",
		Slug:          "test-incident",
		Title:         "Test Incident",
		Private:       false,
		Summary:       "a test incident",
		OpenedAt:      time.Now().Add(-8 * time.Hour),
		ModifiedAt:    time.Now().Add(-7 * time.Hour),
		ClosedAt:      time.Now().Add(-7 * time.Hour),
		ChatChannelID: "",
		Edges: ent.IncidentEdges{
			Severity:        severity,
			Type:            incType,
			Tasks:           tasks,
			RoleAssignments: roles,
			TagAssignments:  tags,
			Milestones:      milestones,
		},
	}

	p.fakeIncidents = []*ent.Incident{inc1}
}

func (p *IncidentDataProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{}
}

func (p *IncidentDataProvider) SetOnIncidentUpdatedCallback(cb rez.DataProviderResourceUpdatedCallback) {
	p.onIncidentUpdatedFn = cb
}

func (p *IncidentDataProvider) IncidentDataMapping() *ent.Incident {
	return &incidentDataMapping
}

func (p *IncidentDataProvider) IncidentRoleDataMapping() *ent.IncidentRole {
	return &incidentRoleDataMapping
}

func (p *IncidentDataProvider) PullIncidents(ctx context.Context) iter.Seq2[*ent.Incident, error] {
	return func(yield func(i *ent.Incident, err error) bool) {
		for _, inc := range p.fakeIncidents {
			yield(inc, nil)
		}
	}
}

func (p *IncidentDataProvider) GetIncidentByID(ctx context.Context, id string) (*ent.Incident, error) {
	for _, inc := range p.fakeIncidents {
		if inc.ProviderID == id {
			return inc, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (p *IncidentDataProvider) GetRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	return p.fakeRoles, nil
}

package fakeprovider

import (
	"context"
	"fmt"
	"iter"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gosimple/slug"

	"github.com/rezible/rezible/ent"
)

type IncidentDataProvider struct {
	roles      []*ent.IncidentRole
	severities []*ent.IncidentSeverity
	types      []*ent.IncidentType
	tags       []*ent.IncidentTag
	users      []*ent.User

	incidents []*ent.Incident
}

type IncidentDataProviderConfig struct {
}

func NewIncidentDataProvider(intg *ent.Integration) (*IncidentDataProvider, error) {
	p := &IncidentDataProvider{}

	p.makeFakeData()
	p.makeFakeIncidents()

	return p, nil
}

func (p *IncidentDataProvider) makeFakeData() {
	p.roles = []*ent.IncidentRole{
		{
			Name: "Incident Commander",
		},
	}
	p.severities = []*ent.IncidentSeverity{
		{
			Name:        "Severity 1",
			Description: "highest severity",
		},
	}
	p.types = []*ent.IncidentType{
		{
			Name: "Default",
		},
	}
	p.tags = []*ent.IncidentTag{
		{Key: faker.Word(), Value: faker.Word()},
	}
	p.users = []*ent.User{
		{
			Name:   faker.Name(),
			Email:  faker.Email(),
			ChatID: faker.Username(),
		},
	}
}

func (p *IncidentDataProvider) makeFakeRoleAssignments() []*ent.IncidentRoleAssignment {
	return []*ent.IncidentRoleAssignment{
		{
			Edges: ent.IncidentRoleAssignmentEdges{
				Role: p.roles[0],
				User: p.users[0],
			},
		},
	}
}

func (p *IncidentDataProvider) makeIncidentMilestones(start, end time.Time) []*ent.IncidentMilestone {
	return []*ent.IncidentMilestone{
		{
			Kind:      "",
			Timestamp: time.Now().Add(-8 * time.Hour),
			Edges:     ent.IncidentMilestoneEdges{},
		},
	}
}

func (p *IncidentDataProvider) makeFakeIncidents() {
	numIncidents := rand.Intn(10)
	p.incidents = make([]*ent.Incident, numIncidents)
	for i := 0; i < numIncidents; i++ {
		openedAt := time.Now().Add(-8 * time.Hour)
		closedAt := time.Now().Add(-7 * time.Hour)

		title := faker.Word() + "-rpc outage"
		incSlug := slug.MakeLang(title, "en")

		p.incidents[i] = &ent.Incident{
			Title:      title,
			ExternalID: fmt.Sprintf("fake-%d", i+1),
			Slug:       incSlug,
			Summary:    faker.Sentence(),
			OpenedAt:   openedAt,
			Edges: ent.IncidentEdges{
				Severity:        p.severities[rand.Intn(len(p.severities))],
				Type:            p.types[rand.Intn(len(p.types))],
				TagAssignments:  p.tags,
				RoleAssignments: p.makeFakeRoleAssignments(),
				Milestones:      p.makeIncidentMilestones(openedAt, closedAt),
			},
		}
	}
}

func (p *IncidentDataProvider) IncidentDataMapping() *ent.Incident {
	return &incidentDataMapping
}

func (p *IncidentDataProvider) IncidentRoleDataMapping() *ent.IncidentRole {
	return &incidentRoleDataMapping
}

func (p *IncidentDataProvider) PullIncidents(ctx context.Context) iter.Seq2[*ent.Incident, error] {
	return func(yield func(i *ent.Incident, err error) bool) {
		for _, inc := range p.incidents {
			yield(inc, nil)
		}
	}
}

func (p *IncidentDataProvider) GetIncidentByID(ctx context.Context, id string) (*ent.Incident, error) {
	for _, inc := range p.incidents {
		if inc.ExternalID == id {
			return inc, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (p *IncidentDataProvider) ListIncidentRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	return p.roles, nil
}

package fakeprovider

import (
	"context"
	"iter"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type TeamDataProvider struct{}

type TeamDataProviderConfig struct{}

var (
	userDebugEmailOnlyDataMapping = ent.User{
		Email: "y",
	}

	teamDataMapping = &ent.Team{
		Name:  "y",
		Edges: ent.TeamEdges{Users: ent.Users{&userDebugEmailOnlyDataMapping}},
	}
)

func NewTeamsDataProvider(intg *ent.Integration) (*TeamDataProvider, error) {
	return &TeamDataProvider{}, nil
}

func (p *TeamDataProvider) TeamDataMapping() *ent.Team {
	return teamDataMapping
}

func (p *TeamDataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	fakeTeam1 := &ent.Team{
		ExternalID: "test-team",
		Name:       "Test Team",
		Edges:      ent.TeamEdges{Users: ent.Users{}},
	}
	if debugEmail := rez.Config.GetString("DEBUG_USER_EMAIL"); debugEmail != "" {
		fakeTeam1.Edges.Users = append(fakeTeam1.Edges.Users, &ent.User{Email: debugEmail})
	}
	return func(yield func(*ent.Team, error) bool) {
		yield(fakeTeam1, nil)
	}
}

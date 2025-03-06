package fakeprovider

import (
	"context"
	"github.com/rezible/rezible/ent"
	"iter"
	"os"
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

func NewTeamsDataProvider(cfg TeamDataProviderConfig) (*TeamDataProvider, error) {
	return &TeamDataProvider{}, nil
}

func (p *TeamDataProvider) TeamDataMapping() *ent.Team {
	return teamDataMapping
}

func (p *TeamDataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	fakeTeam1 := &ent.Team{
		ProviderID: "test-team",
		Name:       "Test Team 1",
		Edges:      ent.TeamEdges{Users: ent.Users{}},
	}
	debugEmail := os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL")
	if debugEmail != "" {
		fakeTeam1.Edges.Users = append(fakeTeam1.Edges.Users, &ent.User{Email: debugEmail})
	}
	return func(yield func(*ent.Team, error) bool) {
		yield(fakeTeam1, nil)
	}
}

package fakeprovider

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"iter"
	"time"
)

type OncallDataProvider struct {
	providerUserMap map[string]*ent.User
}

type OncallDataProviderConfig struct{}

func NewOncallDataProvider(cfg OncallDataProviderConfig) (*OncallDataProvider, error) {
	p := &OncallDataProvider{
		providerUserMap: make(map[string]*ent.User),
	}

	return p, nil
}

func (p *OncallDataProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{}
}

func (p *OncallDataProvider) RosterDataMapping() *ent.OncallRoster {
	return &rosterMapping
}

func (p *OncallDataProvider) UserShiftDataMapping() *ent.OncallUserShift {
	return &shiftMapping
}

func (p *OncallDataProvider) PullRosters(ctx context.Context) iter.Seq2[*ent.OncallRoster, error] {
	return func(yield func(*ent.OncallRoster, error) bool) {

	}
}

func (p *OncallDataProvider) PullShiftsForRoster(ctx context.Context, rosterId string, from, to time.Time) iter.Seq2[*ent.OncallUserShift, error] {
	return func(yield func(*ent.OncallUserShift, error) bool) {

	}
}

func (p *OncallDataProvider) FetchOncallersForRoster(ctx context.Context, rosterId string) ([]*ent.User, error) {
	users := make([]*ent.User, 0)

	return users, nil
}

func (p *OncallDataProvider) ListRosters(ctx context.Context) ([]*ent.OncallRoster, error) {
	rosters := make([]*ent.OncallRoster, 0)

	return rosters, nil
}

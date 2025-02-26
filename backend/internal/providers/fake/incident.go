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

	incidentMappingSupport *ent.Incident

	userIdEmails map[string]string
}

type IncidentDataProviderConfig struct {
}

func NewIncidentDataProvider(cfg IncidentDataProviderConfig) (*IncidentDataProvider, error) {
	p := &IncidentDataProvider{
		userIdEmails: make(map[string]string),
		onIncidentUpdatedFn: func(id string, m time.Time) {
			log.Warn().Msg("no onIncidentUpdated function")
		},
	}

	return p, nil
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

	}
}

func (p *IncidentDataProvider) GetIncidentByID(ctx context.Context, id string) (*ent.Incident, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *IncidentDataProvider) GetRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	roles := make([]*ent.IncidentRole, 0)

	return roles, nil
}

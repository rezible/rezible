package jira

import (
	"context"
	"fmt"
	"iter"

	"github.com/andygrunwald/go-jira"
	"github.com/go-viper/mapstructure/v2"

	"github.com/rezible/rezible/ent"
)

type TicketDataProvider struct {
	client *jira.Client
}

func NewTicketDataProvider(ctx context.Context, intg *ent.Integration) (*TicketDataProvider, error) {
	var cfg IntegrationConfig
	if cfgErr := mapstructure.Decode(intg.Config, &cfg); cfgErr != nil {
		return nil, cfgErr
	}
	tp := jira.BasicAuthTransport{
		Username: cfg.ApiUsername,
		Password: cfg.ApiToken,
	}
	jc, jiraErr := jira.NewClient(tp.Client(), cfg.ProjectUrl)
	if jiraErr != nil {
		return nil, fmt.Errorf("create jira client: %w", jiraErr)
	}
	p := &TicketDataProvider{client: jc}

	return p, nil
}

var (
	ticketDataMapping = &ent.Ticket{}
)

func (p *TicketDataProvider) TeamDataMapping() *ent.Ticket {
	return ticketDataMapping
}

func (p *TicketDataProvider) PullTickets(ctx context.Context) iter.Seq2[*ent.Ticket, error] {
	return func(yield func(*ent.Ticket, error) bool) {

	}
}

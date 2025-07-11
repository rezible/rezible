package jira

import (
	"context"
	"fmt"
	"iter"

	"github.com/andygrunwald/go-jira"

	"github.com/rezible/rezible/ent"
)

type TicketDataProvider struct {
	client *jira.Client
}

type TicketDataProviderConfig struct {
	ProjectUrl  string `json:"project_url"`
	ApiUsername string `json:"api_username"`
	ApiToken    string `json:"api_token"`
}

func NewTicketDataProvider(ctx context.Context, cfg TicketDataProviderConfig) (*TicketDataProvider, error) {
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

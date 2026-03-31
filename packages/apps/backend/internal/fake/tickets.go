package fakeprovider

import (
	"context"
	"iter"

	"github.com/rezible/rezible/ent"
)

type TicketDataProvider struct {
}

type TicketDataProviderConfig struct {
}

func NewTicketDataProvider(intg *ent.Integration) (*TicketDataProvider, error) {
	return &TicketDataProvider{}, nil
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

package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	"github.com/rezible/rezible/ent/playbook"
	"github.com/rezible/rezible/pkg/projections"
)

type PlaybookService struct {
	db rez.Database
}

func NewPlaybookService(db rez.Database) (*PlaybookService, error) {
	s := &PlaybookService{
		db: db,
	}

	return s, nil
}

func (s *PlaybookService) ListPlaybooks(ctx context.Context, params rez.ListPlaybooksParams) ([]*ent.Playbook, int, error) {
	query := s.db.Client(ctx).Playbook.Query().
		Where()

	qCtx := params.GetQueryContext(ctx)
	count, queryErr := query.Count(qCtx)
	if queryErr != nil {
		return nil, 0, fmt.Errorf("count: %w", queryErr)
	}
	playbooks := make([]*ent.Playbook, 0)
	if count > 0 {
		playbooks, queryErr = query.All(qCtx)
	}
	if queryErr != nil {
		return nil, 0, fmt.Errorf("query: %w", queryErr)
	}
	return playbooks, count, nil
}

func (s *PlaybookService) GetPlaybook(ctx context.Context, id uuid.UUID) (*ent.Playbook, error) {
	return s.db.Client(ctx).Playbook.Get(ctx, id)
}

type saveablePlaybookQuery interface {
	Save(context.Context) (*ent.Playbook, error)
}

func (s *PlaybookService) SetPlaybook(ctx context.Context, playbook *ent.Playbook) (*ent.Playbook, error) {
	var q saveablePlaybookQuery
	if playbook.ID == uuid.Nil {
		q = s.db.Client(ctx).Playbook.Create().
			SetTitle(playbook.Title).
			SetContent(playbook.Content)
	} else {
		q = s.db.Client(ctx).Playbook.UpdateOneID(playbook.ID).
			SetTitle(playbook.Title).
			SetContent(playbook.Content)
	}
	return q.Save(ctx)
}

func (s *PlaybookService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) error {
	if !projections.SubjectKindPlaybook.Matches(event) {
		return nil
	}
	decoded, validationErr := projections.DecodePlaybookEvent(event)
	if validationErr != nil || decoded == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	attrs := decoded.Attributes

	dbc := s.db.Client(ctx)

	queryExisting := dbc.Playbook.Query().
		Where(playbook.Title(attrs.Title))
	existing, queryErr := queryExisting.Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return fmt.Errorf("query playbook: %w", queryErr)
	}

	alertIDs := make([]uuid.UUID, 0, len(attrs.RelatedAlerts))
	for _, alertRef := range attrs.RelatedAlerts {
		queryAlias := dbc.KnowledgeEntityAlias.Query().
			Where(knea.Provider(event.Provider), knea.ProviderSubjectRef(alertRef))

		alias, aliasErr := queryAlias.Only(ctx)
		if aliasErr != nil && !ent.IsNotFound(aliasErr) {
			return fmt.Errorf("query alert alias: %w", aliasErr)
		}
		if alias == nil {
			continue
		}
		queryAlert := dbc.Alert.Query().
			Where(alert.KnowledgeEntityID(alias.EntityID))
		a, alertErr := queryAlert.Only(ctx)
		if alertErr != nil && !ent.IsNotFound(alertErr) {
			return fmt.Errorf("query related alert: %w", alertErr)
		}
		if a != nil {
			alertIDs = append(alertIDs, a.ID)
		}
	}

	var mutator ent.EntityMutator[*ent.Playbook, *ent.PlaybookMutation]
	if existing == nil {
		mutator = dbc.Playbook.Create()
	} else {
		mutator = existing.Update()
	}

	m := mutator.Mutation()
	m.SetTitle(attrs.Title)
	m.SetContent([]byte(attrs.Content))
	if len(alertIDs) > 0 {
		m.AddAlertIDs(alertIDs...)
	}

	if _, saveErr := mutator.Save(ctx); saveErr != nil {
		return fmt.Errorf("save playbook: %w", saveErr)
	}
	return nil
}

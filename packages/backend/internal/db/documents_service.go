package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	da "github.com/rezible/rezible/ent/documentaccess"
	"github.com/rs/zerolog/log"
)

type DocumentsServiceConfig struct {
}

type DocumentsService struct {
	cfg DocumentsServiceConfig

	db    *ent.Client
	auth  rez.AuthService
	teams rez.TeamService
}

func NewDocumentsService(db *ent.Client, auth rez.AuthService, teams rez.TeamService) (*DocumentsService, error) {
	svc := &DocumentsService{
		db:    db,
		auth:  auth,
		teams: teams,
		cfg:   DocumentsServiceConfig{},
	}

	if cfgErr := rez.Config.Unmarshal("documents", &svc.cfg); cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	return svc, nil
}

func (s *DocumentsService) GetDocumentAccess(ctx context.Context, docId uuid.UUID) (*ent.DocumentAccess, error) {
	sess := s.auth.GetAuthSession(ctx)
	doc, docErr := s.GetDocument(ctx, docId)
	if docErr != nil {
		return nil, fmt.Errorf("get document: %w", docErr)
	}
	if !doc.AccessRestricted {
		defaultAccess := &ent.DocumentAccess{
			DocumentID: docId,
			UserID:     sess.UserId(),
			CanView:    true,
			CanEdit:    true,
			CanManage:  true,
		}
		return defaultAccess, nil
	}
	for _, scope := range sess.Scopes() {
		parts := strings.Split(scope, ":")
		if parts[0] != "document" || parts[1] != docId.String() {
			continue
		}
		id, idErr := uuid.Parse(parts[2])
		if idErr != nil {
			return nil, fmt.Errorf("invalid document access id: %w", idErr)
		}
		acc, getErr := s.db.DocumentAccess.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("document access: %w", getErr)
		}
		return acc, nil
	}
	bestAccess, accessesErr := s.getBestDocumentAccess(ctx, docId, sess.UserId())
	if accessesErr != nil {
		return nil, fmt.Errorf("failed to get document accesses: %w", accessesErr)
	}
	return bestAccess, nil
}

func (s *DocumentsService) getBestDocumentAccess(ctx context.Context, docId uuid.UUID, userId uuid.UUID) (*ent.DocumentAccess, error) {
	accessQuery := s.db.DocumentAccess.Query().
		Where(da.DocumentID(docId)).
		Where(da.Or(da.UserID(userId), da.TeamIDNotNil()))
	accesses, accessesErr := accessQuery.All(ctx)
	if accessesErr != nil {
		return nil, accessesErr
	}

	var availableAccesses ent.DocumentAccesses
	teamAccesses := map[uuid.UUID]*ent.DocumentAccess{}
	var accessTeamIds []uuid.UUID
	for _, acc := range accesses {
		if acc.UserID != uuid.Nil && acc.UserID == userId {
			availableAccesses = append(availableAccesses, acc)
		}
		if acc.TeamID != uuid.Nil {
			accessTeamIds = append(accessTeamIds, acc.TeamID)
			teamAccesses[acc.TeamID] = acc
		}
	}
	if len(accessTeamIds) > 0 {
		listParams := rez.ListTeamsParams{TeamIds: accessTeamIds, UserIds: []uuid.UUID{userId}}
		teams, teamsErr := s.teams.List(ctx, listParams)
		if teamsErr != nil {
			log.Error().Err(teamsErr).
				Str("docId", docId.String()).
				Str("userId", userId.String()).
				Msg("Error listing document access teams")
		} else {
			for _, team := range teams {
				if acc, ok := teamAccesses[team.ID]; ok {
					availableAccesses = append(availableAccesses, acc)
				}
			}
		}
	}
	if len(availableAccesses) == 0 {
		return nil, nil
	}
	highest := accesses[0]
	for _, a := range accesses {
		if a.CanManage {
			return a, nil
		}
		if a.CanEdit && !highest.CanEdit {
			highest = a
		}
	}
	return highest, nil
}

func (s *DocumentsService) GetDocument(ctx context.Context, id uuid.UUID) (*ent.Document, error) {
	return s.db.Document.Get(ctx, id)
}

type documentMutator interface {
	Save(ctx context.Context) (*ent.Document, error)
	Mutation() *ent.DocumentMutation
}

func (s *DocumentsService) SetDocument(ctx context.Context, id uuid.UUID, setFn func(*ent.DocumentMutation)) (*ent.Document, error) {
	var mutator documentMutator
	isNew := id == uuid.Nil
	if isNew {
		mutator = s.db.Document.Create().SetID(uuid.New())
	} else {
		curr, getErr := s.db.Document.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("fetch existing incident: %w", getErr)
		}
		mutator = s.db.Document.UpdateOne(curr)
	}

	mut := mutator.Mutation()
	setFn(mut)

	updated, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save incident: %w", saveErr)
	}

	return updated, nil
}

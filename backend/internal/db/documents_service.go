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

type DocumentsService struct {
	serverAddress string
	webhookSecret []byte

	db    *ent.Client
	auth  rez.AuthService
	teams rez.TeamService
}

func NewDocumentsService(db *ent.Client, auth rez.AuthService, teams rez.TeamService) (*DocumentsService, error) {
	webhookSecret := rez.Config.GetString("DOCUMENTS_API_SECRET")
	serverAddress := rez.Config.DocumentsServerAddress()

	svc := &DocumentsService{
		serverAddress: serverAddress,
		webhookSecret: []byte(webhookSecret),
		db:            db,
		auth:          auth,
		teams:         teams,
	}

	return svc, nil
}

func (s *DocumentsService) GetServerWebsocketAddress() string {
	return fmt.Sprintf("ws://%s", s.serverAddress)
}

func (s *DocumentsService) getDocumentAccesses(ctx context.Context, docId uuid.UUID, sess *rez.AuthSession) (ent.DocumentAccesses, error) {
	accessQuery := s.db.DocumentAccess.Query().
		Where(da.DocumentID(docId)).
		Where(da.Or(da.UserID(sess.UserId), da.TeamIDNotNil()))
	accesses, accessesErr := accessQuery.All(ctx)
	if accessesErr != nil {
		return nil, accessesErr
	}
	var sessAccesses ent.DocumentAccesses
	teamAccesses := map[uuid.UUID]*ent.DocumentAccess{}
	var accessTeamIds []uuid.UUID
	for _, acc := range accesses {
		if acc.UserID != uuid.Nil && acc.UserID == sess.UserId {
			sessAccesses = append(sessAccesses, acc)
		}
		if acc.TeamID != uuid.Nil {
			accessTeamIds = append(accessTeamIds, acc.TeamID)
			teamAccesses[acc.TeamID] = acc
		}
	}
	if len(accessTeamIds) > 0 {
		listParams := rez.ListTeamsParams{TeamIds: accessTeamIds, UserIds: []uuid.UUID{sess.UserId}}
		teams, teamsErr := s.teams.List(ctx, listParams)
		if teamsErr != nil {
			log.Error().Err(teamsErr).
				Str("docId", docId.String()).
				Str("userId", sess.UserId.String()).
				Msg("Error listing document access teams")
		} else {
			for _, team := range teams {
				if acc, ok := teamAccesses[team.ID]; ok {
					sessAccesses = append(sessAccesses, acc)
				}
			}
		}
	}
	return sessAccesses, nil
}

func (s *DocumentsService) getBestDocumentAccess(accesses ent.DocumentAccesses) *ent.DocumentAccess {
	if len(accesses) == 0 {
		return nil
	}
	highest := accesses[0]
	for _, a := range accesses {
		if a.CanManage {
			return a
		}
		if a.CanEdit && !highest.CanEdit {
			highest = a
		}
	}
	return highest
}

func (s *DocumentsService) getDocumentAccessScope(ctx context.Context, doc *ent.Document, sess *rez.AuthSession) (string, error) {
	if !doc.AccessRestricted {
		return "m", nil
	}
	sessAccesses, accessErr := s.getDocumentAccesses(ctx, doc.ID, sess)
	if accessErr != nil {
		return "", fmt.Errorf("failed to get document accesses: %w", accessErr)
	}
	highestAccess := s.getBestDocumentAccess(sessAccesses)
	if highestAccess == nil {
		return "", rez.ErrAuthSessionUnauthorized
	}
	return fmt.Sprintf("document:%s:%s", doc.ID, highestAccess.ID), nil
}

func (s *DocumentsService) GetDocumentAccess(ctx context.Context, docId uuid.UUID, sess *rez.AuthSession) (*ent.DocumentAccess, error) {
	for _, scope := range sess.Scopes {
		parts := strings.Split(scope, ":")
		if parts[0] != "document" || parts[1] != docId.String() {
			continue
		}
		id, idErr := uuid.Parse(parts[2])
		if idErr != nil {
			return nil, fmt.Errorf("invalid document access id: %w", idErr)
		}
		return s.getDocumentAccess(ctx, id)
	}
	sessAccesses, accessesErr := s.getDocumentAccesses(ctx, docId, sess)
	if accessesErr != nil {
		return nil, fmt.Errorf("failed to get document accesses: %w", accessesErr)
	}
	return s.getBestDocumentAccess(sessAccesses), nil
}

func (s *DocumentsService) GetDocument(ctx context.Context, id uuid.UUID) (*ent.Document, error) {
	return s.db.Document.Get(ctx, id)
}

func (s *DocumentsService) getDocumentAccess(ctx context.Context, id uuid.UUID) (*ent.DocumentAccess, error) {
	return s.db.DocumentAccess.Get(ctx, id)
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

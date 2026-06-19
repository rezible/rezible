package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"aidanwoods.dev/go-paseto"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	da "github.com/rezible/rezible/ent/documentaccess"
)

type DocumentsService struct {
	db    rez.Database
	sess  rez.AuthSessionService
	teams rez.TeamService

	documentsServerUrl     *url.URL
	editorSessionSecretKey paseto.V4AsymmetricSecretKey
}

func NewDocumentsService(cfg rez.Config, db rez.Database, sess rez.AuthSessionService, teams rez.TeamService) (*DocumentsService, error) {
	srvUrl, urlErr := url.Parse(cfg.Documents.ServerUrl)
	if urlErr != nil {
		return nil, fmt.Errorf("server url: %w", urlErr)
	}
	svc := &DocumentsService{
		db:                     db,
		sess:                   sess,
		teams:                  teams,
		documentsServerUrl:     srvUrl,
		editorSessionSecretKey: paseto.NewV4AsymmetricSecretKey(),
	}

	return svc, nil
}

func (s *DocumentsService) CreateDocumentEditorSession(ctx context.Context, docId uuid.UUID, userId uuid.UUID) (*rez.DocumentSession, error) {
	access, accessErr := s.GetUserDocumentAccess(ctx, docId, userId)
	if accessErr != nil {
		return nil, fmt.Errorf("get document access: %w", accessErr)
	}

	if !access.CanView && !access.CanEdit && !access.CanManage {
		return nil, fmt.Errorf("no access")
	}

	now := time.Now()
	token := paseto.NewToken()
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(now.Add(time.Hour))
	token.SetString("document_id", docId.String())
	token.SetString("user_id", userId.String())
	token.SetString("can_edit", fmt.Sprintf("%t", access.CanEdit))
	token.SetString("can_manage", fmt.Sprintf("%t", access.CanManage))

	tokenStr := token.V4Sign(s.editorSessionSecretKey, nil)

	sess := &rez.DocumentSession{
		DocumentName: docId.String(),
		ServerUrl:    s.documentsServerUrl,
		Token:        tokenStr,
	}
	return sess, nil
}

func (s *DocumentsService) GetUserDocumentAccess(ctx context.Context, docId uuid.UUID, userId uuid.UUID) (*ent.DocumentAccess, error) {
	doc, docErr := s.GetDocument(ctx, docId)
	if docErr != nil {
		return nil, fmt.Errorf("get document: %w", docErr)
	}
	if !doc.AccessRestricted {
		defaultAccess := &ent.DocumentAccess{
			DocumentID: docId,
			UserID:     userId,
			CanView:    true,
			CanEdit:    true,
			CanManage:  true,
		}
		return defaultAccess, nil
	}
	/*
		for _, scope := range execAuth.Scopes {
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
	*/
	bestAccess, accessesErr := s.getBestDocumentAccess(ctx, docId, userId)
	if accessesErr != nil {
		return nil, fmt.Errorf("failed to get document accesses: %w", accessesErr)
	}
	return bestAccess, nil
}

func (s *DocumentsService) getBestDocumentAccess(ctx context.Context, docId uuid.UUID, userId uuid.UUID) (*ent.DocumentAccess, error) {
	accessQuery := s.db.Client(ctx).DocumentAccess.Query().
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
			slog.Error("Error listing document access teams",
				"error", teamsErr,
				"docId", docId.String(),
				"userId", userId.String(),
			)
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
	return s.db.Client(ctx).Document.Get(ctx, id)
}

func (s *DocumentsService) SetDocument(ctx context.Context, id uuid.UUID, setFn func(*ent.DocumentMutation)) (*ent.Document, error) {
	var mutator ent.EntityMutator[*ent.Document, *ent.DocumentMutation]
	if id == uuid.Nil {
		mutator = s.db.Client(ctx).Document.Create().SetID(uuid.New())
	} else {
		mutator = s.db.Client(ctx).Document.UpdateOneID(id)
	}

	mut := mutator.Mutation()
	setFn(mut)

	updated, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save incident: %w", saveErr)
	}

	return updated, nil
}

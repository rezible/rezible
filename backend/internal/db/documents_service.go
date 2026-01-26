package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type DocumentsService struct {
	serverAddress string
	webhookSecret []byte

	db    *ent.Client
	users rez.UserService
}

func NewDocumentsService(db *ent.Client, users rez.UserService) (*DocumentsService, error) {
	webhookSecret := rez.Config.GetString("DOCUMENTS_API_SECRET")
	serverAddress := rez.Config.GetString("DOCUMENTS_SERVER_ADDRESS")

	svc := &DocumentsService{
		serverAddress: serverAddress,
		webhookSecret: []byte(webhookSecret),
		db:            db,
		users:         users,
	}

	return svc, nil
}

func (s *DocumentsService) GetServerWebsocketAddress() string {
	return fmt.Sprintf("ws://%s", s.serverAddress)
}

func (s *DocumentsService) GetUserDocumentAccess(ctx context.Context, userId uuid.UUID, docId uuid.UUID) (bool, error) {
	// TODO: do properly
	const readOnly = false
	return readOnly, nil
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

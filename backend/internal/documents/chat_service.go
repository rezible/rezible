package documents

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallusershift"
)

type ChatService struct {
	chat  rez.ChatProvider
	users rez.UserService
}

func NewChatService(ctx context.Context, chat rez.ChatProvider, users rez.UserService) (*ChatService, error) {
	s := &ChatService{
		chat:  chat,
		users: users,
	}

	chat.SetLookupUserFunc(users.GetByChatId)

	return s, nil
}

func (s *ChatService) createAnnotationLookup(ctx context.Context, id string) (uuid.UUID, []*ent.OncallUserShift, error) {
	usr, usrErr := s.users.GetByChatId(ctx, id)
	if usrErr != nil {
		return uuid.Nil, nil, usrErr
	}
	shiftIsActive := oncallusershift.And(oncallusershift.EndAtGT(time.Now()), oncallusershift.StartAtLT(time.Now()))
	shifts, shiftsErr := usr.QueryOncallShifts().WithRoster().Where(shiftIsActive).All(ctx)
	if shiftsErr != nil {
		return uuid.Nil, nil, shiftsErr
	}
	return usr.ID, shifts, nil
}

func (s *ChatService) SetCreateAnnotationFunc(fn rez.ChatInteractionFuncCreateAnnotation) {
	if s.chat != nil {
		s.chat.SetCreateAnnotationFunc(fn)
	}
}

func (s *ChatService) SendUserMessage(ctx context.Context, user *ent.User, msgText string) error {
	return s.chat.SendUserMessage(ctx, user.ChatID, msgText)
}

func (s *ChatService) SendMessage(ctx context.Context, id string, msg *rez.ContentNode) error {
	//TODO implement me
	return fmt.Errorf("not implemented")
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	return s.chat.SendOncallHandover(ctx, params)
}

func (s *ChatService) SendUserLinkMessage(ctx context.Context, user *ent.User, msgText string, linkUrl string, linkText string) error {
	return s.chat.SendUserLinkMessage(ctx, user.ChatID, msgText, linkUrl, linkText)
}

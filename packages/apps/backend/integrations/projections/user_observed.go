package projections

import "github.com/rezible/rezible/ent"

const (
	attrEmail    = "email"
	attrChatId   = "chat_id"
	attrTimezone = "timezone"
)

type (
	UserObserved           = Event[UserObservedAttributes]
	UserObservedAttributes struct {
		Email    string
		ChatId   string
		Timezone string
	}
)

func (a UserObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrEmail:    a.Email,
		attrChatId:   a.ChatId,
		attrTimezone: a.Timezone,
	}
}

func DecodeUserObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	email, emailErr := requiredString(ev, attrEmail)
	if emailErr != nil {
		return nil, emailErr
	}
	chatId, chatIdErr := optionalString(ev, attrChatId)
	if chatIdErr != nil {
		return nil, chatIdErr
	}
	tz, tzErr := optionalString(ev, attrTimezone)
	if tzErr != nil {
		return nil, tzErr
	}
	attrs := UserObservedAttributes{
		Email:    email,
		ChatId:   chatId,
		Timezone: tz,
	}
	return attrs, nil
}

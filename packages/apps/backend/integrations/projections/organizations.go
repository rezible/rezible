package projections

import "github.com/rezible/rezible/ent"

const (
	attrName     = "name"
	attrEmail    = "email"
	attrChatId   = "chat_id"
	attrTimezone = "timezone"
)

type (
	UserObserved           = Event[UserObservedAttributes]
	UserObservedAttributes struct {
		Name     string
		Email    string
		ChatId   string
		Timezone string
	}
)

func (a UserObservedAttributes) Encode() map[string]any {
	return map[string]any{
		attrName:     a.Name,
		attrEmail:    a.Email,
		attrChatId:   a.ChatId,
		attrTimezone: a.Timezone,
	}
}

func DecodeUserObservedEvent(ev *ent.NormalizedEvent) (any, error) {
	name, nameErr := requiredString(ev, attrName)
	if nameErr != nil {
		return nil, nameErr
	}
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
		Name:     name,
		Email:    email,
		ChatId:   chatId,
		Timezone: tz,
	}
	return attrs, nil
}

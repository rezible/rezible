package eventprojections

import "github.com/rezible/rezible/ent"

const (
	attrBody                    = "body"
	attrConversationExternalRef = "conversation_external_ref"
	attrSenderExternalRef       = "sender_external_ref"
	attrThreadExternalRef       = "thread_external_ref"
)

type (
	ChatMessage           = Event[ChatMessageAttributes]
	ChatMessageAttributes struct {
		ConversationExternalRef string
		Body                    string
		SenderExternalRef       string
		ThreadExternalRef       string
	}
)

func (a ChatMessageAttributes) Encode() map[string]any {
	return map[string]any{
		attrBody:                    a.Body,
		attrConversationExternalRef: a.ConversationExternalRef,
		attrSenderExternalRef:       a.SenderExternalRef,
		attrThreadExternalRef:       a.ThreadExternalRef,
	}
}

func DecodeChatMessageEvent(ev *ent.NormalizedEvent) (any, error) {
	if attrsErr := rejectUnsupportedAttributes(ev, attrConversationExternalRef, attrBody, attrSenderExternalRef, attrThreadExternalRef); attrsErr != nil {
		return nil, attrsErr
	}
	conversationRef, conversationRefErr := requiredString(ev, attrConversationExternalRef)
	if conversationRefErr != nil {
		return nil, conversationRefErr
	}
	body, bodyErr := requiredString(ev, attrBody)
	if bodyErr != nil {
		return nil, bodyErr
	}
	senderRef, senderRefErr := optionalString(ev, attrSenderExternalRef)
	if senderRefErr != nil {
		return nil, senderRefErr
	}
	threadRef, threadRefErr := optionalString(ev, attrThreadExternalRef)
	if threadRefErr != nil {
		return nil, threadRefErr
	}
	attrs := ChatMessageAttributes{
		ConversationExternalRef: conversationRef,
		Body:                    body,
		SenderExternalRef:       senderRef,
		ThreadExternalRef:       threadRef,
	}
	return ChatMessage{Event: ev, Attributes: attrs}, nil
}

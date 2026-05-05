package projections

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

func decodeChatMessageEvent(ev *ent.NormalizedEvent) (any, error) {
	attrs, err := decodeChatMessage(ev)
	if err != nil {
		return nil, err
	}
	return ChatMessage{Event: ev, Attributes: attrs}, nil
}

func decodeChatMessage(ev *ent.NormalizedEvent) (ChatMessageAttributes, error) {
	if err := rejectUnsupportedAttributes(ev, attrConversationExternalRef, attrBody, attrSenderExternalRef, attrThreadExternalRef); err != nil {
		return ChatMessageAttributes{}, err
	}
	conversationRef, err := requiredString(ev, attrConversationExternalRef)
	if err != nil {
		return ChatMessageAttributes{}, err
	}
	body, err := requiredString(ev, attrBody)
	if err != nil {
		return ChatMessageAttributes{}, err
	}
	senderRef, err := optionalString(ev, attrSenderExternalRef)
	if err != nil {
		return ChatMessageAttributes{}, err
	}
	threadRef, err := optionalString(ev, attrThreadExternalRef)
	if err != nil {
		return ChatMessageAttributes{}, err
	}
	return ChatMessageAttributes{
		ConversationExternalRef: conversationRef,
		Body:                    body,
		SenderExternalRef:       senderRef,
		ThreadExternalRef:       threadRef,
	}, nil
}

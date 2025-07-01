package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	rez "github.com/rezible/rezible"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallannotation"
)

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB

	createAnnotationCallbackID        = "create_annotation"
	createAnnotationConfirmCallbackID = "create_annotation_confirm"
)

type webhookHandler struct {
	signingSecret    string
	client           *slack.Client
	messageAnnotator rez.ChatMessageAnnotator
}

func newWebhookHandler(signingSecret string, client *slack.Client) *webhookHandler {
	return &webhookHandler{
		signingSecret: signingSecret,
		client:        client,
	}
}

func (h *webhookHandler) verifyWebhook(w http.ResponseWriter, r *http.Request) error {
	bodyReader := http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes)
	body, bodyErr := io.ReadAll(bodyReader)
	if bodyErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return bodyErr
	}

	sv, svErr := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if svErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return svErr
	}

	if _, writeErr := sv.Write(body); writeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return writeErr
	}

	if verificationErr := sv.Ensure(); verificationErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return verificationErr
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return nil
}

func (h *webhookHandler) handleOptions(w http.ResponseWriter, r *http.Request) {
	if verifyErr := h.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	log.Debug().Msg("get options")

	w.WriteHeader(http.StatusOK)
}

func (h *webhookHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if verifyErr := h.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ev, evErr := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if evErr != nil {
		log.Error().Err(evErr).Msg("failed to parse event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ev.Type == slackevents.URLVerification {
		var res *slackevents.ChallengeResponse
		if jsonErr := json.Unmarshal(body, &res); jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		if _, writeErr := w.Write([]byte(res.Challenge)); writeErr != nil {
			log.Error().Err(writeErr).Msg("failed to write challenge response")
		}
	} else if ev.Type == slackevents.AppRateLimited {
		log.Warn().Msg("slack app rate limited")
		w.WriteHeader(http.StatusOK)
	} else if ev.Type == slackevents.CallbackEvent {
		// TODO: queue processing of this
		go h.handleCallbackEvent(ev)
		w.WriteHeader(http.StatusOK)
	} else {
		log.Warn().Str("type", ev.Type).Msg("failed to handle event")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *webhookHandler) handleCallbackEvent(ev slackevents.EventsAPIEvent) {
	if ev.Type == slackevents.CallbackEvent {
		innerEvent := ev.InnerEvent
		switch data := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			h.handleAppMentionEvent(ev, data)
		}
	}
}

func (h *webhookHandler) handleAppMentionEvent(e slackevents.EventsAPIEvent, data *slackevents.AppMentionEvent) {
	fmt.Printf("mention event: %+v\n", data)
	_, _, msgErr := h.client.PostMessage(data.Channel, slack.MsgOptionText("hello", false))
	if msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to message")
	}
}

func (h *webhookHandler) handleInteractions(w http.ResponseWriter, r *http.Request) {
	if verifyErr := h.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	payload := r.PostFormValue("payload")
	if payload == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ic slack.InteractionCallback
	if jsonErr := json.Unmarshal([]byte(payload), &ic); jsonErr != nil {
		log.Debug().Err(jsonErr).Msg("failed to unmarshal interaction payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var handlerErr error
	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		handlerErr = h.handleMessageAction(ctx, &ic)
	case slack.InteractionTypeBlockActions:
		handlerErr = h.handleBlockActions(ctx, &ic)
	case slack.InteractionTypeViewSubmission:
		handlerErr = h.handleViewSubmission(ctx, &ic)
	default:
		log.Warn().Str("type", string(ic.Type)).Msg("unknown interaction type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if handlerErr != nil {
		log.Error().Err(handlerErr).Str("type", string(ic.Type)).Msg("failed to handle interaction")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *webhookHandler) handleMessageAction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationCallbackID:
		return h.handleCreateAnnotationAction(ctx, ic)
	}
	return fmt.Errorf("unknown message action: %s", ic.CallbackID)
}

func (h *webhookHandler) handleBlockActions(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationConfirmCallbackID:
		return h.handleCreateAnnotationModalBlockAction(ctx, ic)
	}
	return nil
}

func (h *webhookHandler) handleViewSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case createAnnotationConfirmCallbackID:
		return h.handleCreateAnnotationModalSubmission(ctx, ic)
	}
	return nil
}

func (h *webhookHandler) handleCreateAnnotationAction(ctx context.Context, ic *slack.InteractionCallback) error {
	view, viewErr := h.createAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	resp, respErr := h.client.OpenViewContext(ctx, ic.TriggerID, *view)
	if respErr != nil {
		if resp != nil && !resp.Ok && len(resp.ResponseMetadata.Messages) > 0 {
			log.Debug().
				Strs("messages", resp.ResponseMetadata.Messages).
				Msg("message action: open view failed")
		}
		return fmt.Errorf("open view: %w", respErr)
	}
	return nil
}

func convertSlackTs(ts string) time.Time {
	parts := strings.Split(ts, ".")
	if len(parts) < 2 {
		return time.Time{}
	}
	secs, parseErr := strconv.ParseInt(parts[0], 10, 32)
	if parseErr != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}

type createAnnotationMetadata struct {
	MsgId        string    `json:"mid"`
	MsgTimestamp time.Time `json:"mts"`
	RosterId     uuid.UUID `json:"rid"`
	AnnotationId uuid.UUID `json:"aid"`
}

func (h *webhookHandler) createAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback) (*slack.ModalViewRequest, error) {
	msgId := fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)

	rosters, event, infoErr := h.messageAnnotator.QueryUserChatMessageEventDetails(ctx, ic.User.ID, msgId)
	if infoErr != nil {
		return nil, fmt.Errorf("failed to get annotation information: %w", infoErr)
	}

	if len(rosters) == 0 {
		return &slack.ModalViewRequest{
			Type:  "modal",
			Title: plainTextBlock("No Oncall Rosters"),
			Blocks: slack.Blocks{BlockSet: []slack.Block{
				slack.NewSectionBlock(plainTextBlock("You are not a member of any oncall rosters"), nil, nil),
			}},
			Close:      plainTextBlock("Close"),
			CallbackID: createAnnotationConfirmCallbackID,
		}, nil
	}

	msgTime := convertSlackTs(ic.MessageTs)

	metadata := createAnnotationMetadata{
		MsgId:        msgId,
		MsgTimestamp: msgTime,
	}

	var rosterBlock slack.Block

	// TODO: allow selecting roster
	roster := rosters[0]

	rosterDetailsSection := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement("Editing annotation for ", nil),
		slack.NewRichTextSectionTextElement(roster.Name, &slack.RichTextSectionTextStyle{Bold: true}))
	rosterBlock = slack.NewRichTextBlock("anno_roster", rosterDetailsSection)
	/*
		if len(shifts) > 1 {
			shiftOptions := make([]*slack.OptionBlockObject, len(shifts))
			for i, sh := range shifts {
				rosterName := sh.Edges.Roster.Name
				shiftOptions[i] = slack.NewOptionBlockObject(sh.ID.String(), plainTextBlock(rosterName), nil)
			}

			shiftSelectElement := slack.NewOptionsSelectBlockElement(
				slack.OptTypeStatic,
				plainTextBlock("Select the roster shift to annotate"),
				"shift_roster_select",
				shiftOptions...)

			shiftBlock = slack.NewSectionBlock(plainTextBlock("Oncall Shift Rosters"), nil,
				slack.NewAccessory(shiftSelectElement),
				slack.SectionBlockOptionBlockID("shift_select"))
		}
	*/
	metadata.RosterId = roster.ID

	var curr *ent.OncallAnnotation

	if event != nil {
		rosterAnno, annoErr := event.QueryAnnotations().
			Where(oncallannotation.RosterID(roster.ID)).
			Only(ctx)
		if annoErr != nil && !ent.IsNotFound(annoErr) {
			return nil, fmt.Errorf("failed to query existing event annotation: %w", annoErr)
		}
		curr = rosterAnno
	}

	messageUserDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(ic.Message.User, nil),
		slack.NewRichTextSectionDateElement(msgTime.Unix(), " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(ic.Message.Text, &slack.RichTextSectionTextStyle{Italic: true}))

	inputBlock := slack.NewPlainTextInputBlockElement(nil, "notes_input_text")
	inputHint := plainTextBlock("You can edit this later")

	titleText := "Create Annotation"
	submitText := "Create"

	if curr != nil {
		inputBlock.WithInitialValue(curr.Notes)
		metadata.AnnotationId = curr.ID
		inputHint = nil
		titleText = "Update Annotation"
		submitText = "Update"
	}

	blockSet := []slack.Block{
		rosterBlock,
		slack.NewDividerBlock(),
		slack.NewRichTextBlock("anno_msg", messageUserDetails, messageContentsDetails),
		slack.NewDividerBlock(),
		slack.NewInputBlock("notes_input", plainTextBlock("Notes"), inputHint, inputBlock),
	}

	jsonMetadata, jsonErr := json.Marshal(metadata)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	return &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      createAnnotationConfirmCallbackID,
		PrivateMetadata: string(jsonMetadata),
		Title:           plainTextBlock(titleText),
		Close:           plainTextBlock("Cancel"),
		Submit:          plainTextBlock(submitText),
		Blocks:          slack.Blocks{BlockSet: blockSet},
	}, nil
}

func (h *webhookHandler) handleCreateAnnotationModalBlockAction(ctx context.Context, ic *slack.InteractionCallback) error {
	return nil
}

func (h *webhookHandler) handleCreateAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	var meta createAnnotationMetadata
	if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	var notes string
	rosterId := meta.RosterId

	if state := ic.View.State; state != nil {
		if notesInput, ok := state.Values["notes_input"]; ok {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}

		if rosterId == uuid.Nil {
			if optionsBlock, ok := state.Values["roster_select"]; ok {
				if selectBlock, optOk := optionsBlock["roster_select"]; optOk {
					id, uuidErr := uuid.Parse(selectBlock.SelectedOption.Value)
					if uuidErr != nil {
						return fmt.Errorf("invalid roster id selected: %w", uuidErr)
					}
					rosterId = id
				}
			}
		}
	}

	msgEvent := &ent.OncallEvent{
		ProviderID: meta.MsgId,
		Kind:       "message",
		Timestamp:  meta.MsgTimestamp,
		// TODO: add more message details
	}
	msgAnno := &ent.OncallAnnotation{
		ID:       meta.AnnotationId,
		RosterID: rosterId,
		Notes:    notes,
		Edges:    ent.OncallAnnotationEdges{Event: msgEvent},
	}
	_, createErr := h.messageAnnotator.CreateAnnotation(ctx, msgAnno)
	if createErr != nil {
		return fmt.Errorf("failed to create annotation: %w", createErr)
	}

	return nil
}

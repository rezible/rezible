package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rezible/rezible/ent/oncallannotation"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
)

var (
	createAnnotationCallbackID        = "create_annotation"
	createAnnotationConfirmCallbackID = "create_annotation_confirm"
)

func (p *ChatProvider) handleInteractionWebhook(w http.ResponseWriter, r *http.Request) {
	if verifyErr := p.verifyWebhook(w, r); verifyErr != nil {
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
		handlerErr = p.handleMessageAction(ctx, &ic)
	case slack.InteractionTypeBlockActions:
		handlerErr = p.handleBlockActions(ctx, &ic)
	case slack.InteractionTypeViewSubmission:
		handlerErr = p.handleViewSubmission(ctx, &ic)
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

func (p *ChatProvider) handleMessageAction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationCallbackID:
		return p.handleCreateAnnotationAction(ctx, ic)
	}
	return fmt.Errorf("unknown message action: %s", ic.CallbackID)
}

func (p *ChatProvider) handleBlockActions(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationConfirmCallbackID:
		return p.handleCreateAnnotationModalBlockAction(ctx, ic)
	}
	return nil
}

func (p *ChatProvider) handleViewSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case createAnnotationConfirmCallbackID:
		return p.handleCreateAnnotationModalSubmission(ctx, ic)
	}
	return nil
}

func (p *ChatProvider) handleCreateAnnotationAction(ctx context.Context, ic *slack.InteractionCallback) error {
	view, viewErr := p.createAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	resp, respErr := p.client.OpenViewContext(ctx, ic.TriggerID, *view)
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

/*
func (p *ChatProvider) getUserOncallRosters(ctx context.Context, id string) ([]*ent.OncallRoster, error) {
	user, userErr := p.lookupUserFn(ctx, id)
	if userErr != nil {
		return nil, fmt.Errorf("failed to get user: %w", userErr)
	}

	rosters, rostersErr := user.QueryOncallSchedules().QuerySchedule().QueryRoster().All(ctx)
	if rostersErr != nil && !ent.IsNotFound(rostersErr) {
		return nil, fmt.Errorf("failed to query oncall rosters for user: %w", rostersErr)
	}

	return rosters, nil
}
*/

type createAnnotationMetadata struct {
	MsgId        string    `json:"mid"`
	MsgTimestamp time.Time `json:"mts"`
	RosterId     uuid.UUID `json:"rid"`
	AnnotationId uuid.UUID `json:"aid"`
}

func (p *ChatProvider) createAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback) (*slack.ModalViewRequest, error) {
	msgId := fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)

	rosters, event, infoErr := p.annos.QueryUserChatMessageEventDetails(ctx, ic.User.ID, msgId)
	if infoErr != nil {
		return nil, fmt.Errorf("failed to get annotation information: %w", infoErr)
	}

	if len(rosters) == 0 {
		return &slack.ModalViewRequest{
			Type:  "modal",
			Title: plainText("No Oncall Rosters"),
			Blocks: slack.Blocks{BlockSet: []slack.Block{
				slack.NewSectionBlock(plainText("You are not a member of any oncall rosters"), nil, nil),
			}},
			Close:      plainText("Close"),
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
				shiftOptions[i] = slack.NewOptionBlockObject(sh.ID.String(), plainText(rosterName), nil)
			}

			shiftSelectElement := slack.NewOptionsSelectBlockElement(
				slack.OptTypeStatic,
				plainText("Select the roster shift to annotate"),
				"shift_roster_select",
				shiftOptions...)

			shiftBlock = slack.NewSectionBlock(plainText("Oncall Shift Rosters"), nil,
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
	inputHint := plainText("You can edit this later")

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
		slack.NewInputBlock("notes_input", plainText("Notes"), inputHint, inputBlock),
	}

	jsonMetadata, jsonErr := json.Marshal(metadata)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	return &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      createAnnotationConfirmCallbackID,
		PrivateMetadata: string(jsonMetadata),
		Title:           plainText(titleText),
		Close:           plainText("Cancel"),
		Submit:          plainText(submitText),
		Blocks:          slack.Blocks{BlockSet: blockSet},
	}, nil
}

func (p *ChatProvider) handleCreateAnnotationModalBlockAction(ctx context.Context, ic *slack.InteractionCallback) error {
	return nil
}

func (p *ChatProvider) handleCreateAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
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
	_, createErr := p.annos.CreateAnnotation(ctx, msgAnno)
	if createErr != nil {
		return fmt.Errorf("failed to create annotation: %w", createErr)
	}

	return nil
}

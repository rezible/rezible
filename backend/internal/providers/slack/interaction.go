package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallusershiftannotation"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	// fmt.Printf("block actions: %+v\n", ic)
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
	if resp != nil && !resp.Ok && len(resp.ResponseMetadata.Messages) > 0 {
		log.Debug().
			Strs("messages", resp.ResponseMetadata.Messages).
			Msg("message action: open view failed")
	}

	if respErr != nil {
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

func (p *ChatProvider) createAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback) (*slack.ModalViewRequest, error) {
	_, shifts, userErr := p.lookupAnnotationUser(ctx, ic.User.ID)
	if userErr != nil {
		return nil, fmt.Errorf("failed to get user: %w", userErr)
	}
	//
	//shiftIsActive := oncallusershift.And(oncallusershift.EndAtGT(time.Now()), oncallusershift.StartAtLT(time.Now()))
	//shifts, shiftsErr := user.QueryOncallShifts().WithRoster().Where(shiftIsActive).All(ctx)
	//if shiftsErr != nil && !ent.IsNotFound(shiftsErr) {
	//	return nil, fmt.Errorf("failed to get active oncall shifts: %w", shiftsErr)
	//}

	if len(shifts) == 0 {
		return &slack.ModalViewRequest{
			Type:  "modal",
			Title: plainText("No Active Shift"),
			Blocks: slack.Blocks{BlockSet: []slack.Block{
				slack.NewSectionBlock(plainText("You do not have a current oncall shift to annotate"), nil, nil),
			}},
			Close:      plainText("Cancel"),
			CallbackID: createAnnotationConfirmCallbackID,
		}, nil
	}

	view := &slack.ModalViewRequest{
		Type:       "modal",
		CallbackID: createAnnotationConfirmCallbackID,
		Title:      plainText("Create Annotation"),
		Close:      plainText("Cancel"),
		Submit:     plainText("Create"),
	}

	msgId := fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)

	showShiftSelect := len(shifts) == 1
	var shiftBlock slack.Block
	if showShiftSelect {
		view.PrivateMetadata = fmt.Sprintf("%s,%s", msgId, shifts[0].ID)
		shiftBlock = slack.NewRichTextBlock("anno_shift", slack.NewRichTextSection(
			slack.NewRichTextSectionTextElement("Annotating your active shift in ", nil),
			slack.NewRichTextSectionTextElement(shifts[0].Edges.Roster.Name, &slack.RichTextSectionTextStyle{Bold: true})))
	} else {
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

	msgTimeFormat := " - {date_short_pretty} at {time}"
	messageDetailsBlock := slack.NewRichTextBlock("anno_msg",
		slack.NewRichTextSection(
			slack.NewRichTextSectionUserElement(ic.Message.User, nil),
			slack.NewRichTextSectionDateElement(convertSlackTs(ic.MessageTs).Unix(), msgTimeFormat, nil, nil)),
		slack.NewRichTextSection(
			slack.NewRichTextSectionTextElement(ic.Message.Text, &slack.RichTextSectionTextStyle{Italic: true})))

	notesInputBlock := slack.NewInputBlock(
		"notes_input",
		plainText("Notes"),
		plainText("You can edit this later"),
		slack.NewPlainTextInputBlockElement(nil, "notes_input_text"))

	view.Blocks = slack.Blocks{BlockSet: []slack.Block{
		shiftBlock,
		slack.NewDividerBlock(),
		messageDetailsBlock,
		slack.NewDividerBlock(),
		notesInputBlock,
	}}

	return view, nil
}

func (p *ChatProvider) handleCreateAnnotationModalBlockAction(ctx context.Context, ic *slack.InteractionCallback) error {

	return nil
}

func (p *ChatProvider) handleCreateAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	var shiftId uuid.UUID
	var notes string

	mdParts := strings.Split(ic.View.PrivateMetadata, ",")
	if len(mdParts) != 2 {
		return fmt.Errorf("invalid metadata: %s", ic.View.PrivateMetadata)
	}

	msgId := mdParts[0]

	var uuidErr error
	shiftId, uuidErr = uuid.Parse(mdParts[1])
	if uuidErr != nil {
		return fmt.Errorf("invalid shift id: %w", uuidErr)
	}

	if state := ic.View.State; state != nil {
		if notesInput, ok := state.Values["notes_input"]; ok {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}

		if shiftId == uuid.Nil {
			if optionsBlock, ok := state.Values["shift_select"]; ok {
				if selectBlock, optOk := optionsBlock["shift_roster_select"]; optOk {
					shiftId, uuidErr = uuid.Parse(selectBlock.SelectedOption.Value)
					if uuidErr != nil {
						return fmt.Errorf("invalid shift id selected: %w", uuidErr)
					}
				}
			}
		}
	}

	anno := &ent.OncallUserShiftAnnotation{
		ShiftID:         shiftId,
		EventID:         msgId,
		EventKind:       oncallusershiftannotation.EventKindPing,
		Title:           "Slack Message",
		MinutesOccupied: 0,
		Notes:           notes,
		Pinned:          false,
	}

	if idParts := strings.Split(msgId, "_"); len(idParts) >= 2 {
		anno.OccurredAt = convertSlackTs(idParts[1])
	}

	return p.annotationCreated(ctx, anno)
}

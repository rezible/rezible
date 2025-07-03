package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rezible/rezible/ent/oncallannotation"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

var (
	createAnnotationModalViewCallbackID = "create_annotation_confirm"
)

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
	AnnotationId uuid.UUID `json:"aid"`
}

func getCreateAnnotationModalViewAnnotation(view slack.View) (*ent.OncallAnnotation, error) {
	var meta createAnnotationMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	var notes string
	if state := view.State; state != nil {
		if notesInput, ok := state.Values["notes_input"]; ok {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}

	}

	_, rosterId := getAnnotationSelectedRoster(view.State)
	if rosterId == uuid.Nil {
		return nil, fmt.Errorf("no roster id selected")
	}

	anno := &ent.OncallAnnotation{
		ID:       meta.AnnotationId,
		RosterID: rosterId,
		Notes:    notes,
		Edges: ent.OncallAnnotationEdges{
			Event: &ent.OncallEvent{
				ProviderID: meta.MsgId,
				Kind:       "message",
				Timestamp:  meta.MsgTimestamp,
				// TODO: add more message details
			},
		},
	}

	return anno, nil
}

func getAnnotationSelectedRoster(state *slack.ViewState) (string, uuid.UUID) {
	if optionsBlock, ok := state.Values["roster_select_block"]; ok {
		if selectBlock, optOk := optionsBlock["roster_select"]; optOk {
			id, uuidErr := uuid.Parse(selectBlock.SelectedOption.Value)
			if uuidErr == nil {
				return selectBlock.SelectedOption.Text.Text, id
			}
		}
	}
	return "", uuid.Nil
}

func makeCreateAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback, msgId string, rosters []*ent.OncallRoster, event *ent.OncallEvent) (*slack.ModalViewRequest, error) {
	if len(rosters) == 0 {
		return &slack.ModalViewRequest{
			Type:  "modal",
			Title: plainTextBlock("No Oncall Rosters"),
			Blocks: slack.Blocks{BlockSet: []slack.Block{
				slack.NewSectionBlock(plainTextBlock("You are not a member of any oncall rosters"), nil, nil),
			}},
			Close:      plainTextBlock("Close"),
			CallbackID: createAnnotationModalViewCallbackID,
		}, nil
	}

	msgTime := convertSlackTs(ic.MessageTs)

	metadata := createAnnotationMetadata{
		MsgId:        msgId,
		MsgTimestamp: msgTime,
	}

	// TODO: allow selecting roster
	roster := rosters[0]

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

	rosterOptions := make([]*slack.OptionBlockObject, len(rosters))
	for i, r := range rosters {
		rosterOptions[i] = slack.NewOptionBlockObject(r.ID.String(), plainTextBlock(r.Name), nil)
	}

	rosterSelectElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		plainTextBlock("Roster Name"),
		"roster_select",
		rosterOptions...).
		WithInitialOption(rosterOptions[0])

	rosterBlock := slack.NewSectionBlock(
		plainTextBlock("Annotating roster: "+roster.Name), nil,
		slack.NewAccessory(rosterSelectElement),
		slack.SectionBlockOptionBlockID("roster_select_block"))

	messageUserDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(ic.Message.User, nil),
		slack.NewRichTextSectionDateElement(msgTime.Unix(), " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(ic.Message.Text, &slack.RichTextSectionTextStyle{Italic: true}))

	inputBlock := slack.NewPlainTextInputBlockElement(nil, "notes_input_text").
		WithMinLength(1)
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
		CallbackID:      createAnnotationModalViewCallbackID,
		PrivateMetadata: string(jsonMetadata),
		Title:           plainTextBlock(titleText),
		Close:           plainTextBlock("Cancel"),
		Submit:          plainTextBlock(submitText),
		Blocks:          slack.Blocks{BlockSet: blockSet},
		ExternalID:      msgId,
	}, nil
}

func makeUpdatedCreateAnnotationModalView(ic *slack.InteractionCallback) slack.ModalViewRequest {
	rosterName, _ := getAnnotationSelectedRoster(ic.View.State)

	blockSet := ic.View.Blocks.BlockSet
	for idx, block := range ic.View.Blocks.BlockSet {
		if block.ID() == "roster_select_block" {
			if sb, ok := block.(*slack.SectionBlock); ok {
				sb.Text = plainTextBlock("Annotating roster: " + rosterName)
				blockSet[idx] = sb
			}
		}
	}

	return slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      createAnnotationModalViewCallbackID,
		PrivateMetadata: ic.View.PrivateMetadata,
		Title:           ic.View.Title,
		Close:           ic.View.Close,
		Submit:          ic.View.Submit,
		Blocks:          slack.Blocks{BlockSet: blockSet},
		ExternalID:      ic.View.ExternalID,
	}
}

func makeUserHomeView(ctx context.Context) (*slack.HomeTabViewRequest, string, error) {
	var blocks []slack.Block
	blocks = append(blocks, slack.NewSectionBlock(plainTextBlock("Home Tab"), nil, nil))
	homeView := slack.HomeTabViewRequest{
		Type:            slack.VTHomeTab,
		CallbackID:      "user_home",
		PrivateMetadata: "foo",
		Blocks:          slack.Blocks{BlockSet: blocks},
	}
	hash := time.Now().String()
	return &homeView, hash, nil
}

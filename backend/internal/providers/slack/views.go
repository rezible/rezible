package slack

import (
	"context"
	"encoding/json"
	"fmt"
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

type createAnnotationMessageMetadata struct {
	UserId       string    `json:"uid"`
	MsgId        string    `json:"mid"`
	MsgTimestamp time.Time `json:"mts"`
	MsgText      string    `json:"mtx"`
	AnnotationId uuid.UUID `json:"aid"`
}

func getCreateAnnotationModalViewAnnotation(view slack.View) (*ent.OncallAnnotation, error) {
	var meta createAnnotationMessageMetadata
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

	_, rosterId := getSelectedRoster(view.State)
	if rosterId == uuid.Nil {
		return nil, fmt.Errorf("no roster id selected")
	}

	anno := &ent.OncallAnnotation{
		ID:       meta.AnnotationId,
		RosterID: rosterId,
		Notes:    notes,
		Edges: ent.OncallAnnotationEdges{
			Event: &ent.OncallEvent{
				ProviderID:  meta.MsgId,
				Kind:        "message",
				Timestamp:   meta.MsgTimestamp,
				Source:      "slack",
				Title:       "message",
				Description: "",
			},
		},
	}

	return anno, nil
}

func getSelectedRoster(state *slack.ViewState) (string, uuid.UUID) {
	if state != nil {
		if optionsBlock, ok := state.Values["roster_select_block"]; ok {
			if selectBlock, optOk := optionsBlock["roster_select"]; optOk {
				id, uuidErr := uuid.Parse(selectBlock.SelectedOption.Value)
				if uuidErr == nil {
					return selectBlock.SelectedOption.Text.Text, id
				}
			}
		}
	}
	return "", uuid.Nil
}

type createAnnotationModalDetails struct {
	metadata       createAnnotationMessageMetadata
	rosters        []*ent.OncallRoster
	selectedRoster *ent.OncallRoster
	currAnnotation *ent.OncallAnnotation
}

func makeCreateAnnotationModalView(details *createAnnotationModalDetails) (*slack.ModalViewRequest, error) {
	if len(details.rosters) == 0 {
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

	var blockSet []slack.Block

	messageUserDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(details.metadata.UserId, nil),
		slack.NewRichTextSectionDateElement(details.metadata.MsgTimestamp.Unix(), " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(details.metadata.MsgText, &slack.RichTextSectionTextStyle{Italic: true}))

	blockSet = append(blockSet, slack.NewRichTextBlock("anno_msg", messageUserDetails, messageContentsDetails))

	var selectedRoster *ent.OncallRoster
	if details.selectedRoster != nil {
		selectedRoster = details.selectedRoster
	} else if len(details.rosters) == 1 {
		selectedRoster = details.rosters[0]
	}

	selectedOptIdx := -1
	rosterOptions := make([]*slack.OptionBlockObject, len(details.rosters))
	for i, r := range details.rosters {
		rosterOptions[i] = slack.NewOptionBlockObject(r.ID.String(), plainTextBlock(r.Name), nil)
		if selectedRoster != nil && r.ID == selectedRoster.ID {
			selectedOptIdx = i
		}
	}

	rosterSelectElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		plainTextBlock("Roster Name"),
		"roster_select",
		rosterOptions...)

	if selectedOptIdx >= 0 {
		rosterSelectElement.WithInitialOption(rosterOptions[selectedOptIdx])
	}

	rosterText := "Select a Roster to annotate"
	if selectedRoster != nil {
		rosterText = "Annotating roster: " + selectedRoster.Name
	}
	rosterBlock := slack.NewSectionBlock(
		plainTextBlock(rosterText), nil,
		slack.NewAccessory(rosterSelectElement),
		slack.SectionBlockOptionBlockID("roster_select_block"))

	blockSet = append(blockSet, slack.NewDividerBlock(), rosterBlock)

	if selectedRoster != nil {
		inputBlock := slack.NewPlainTextInputBlockElement(nil, "notes_input_text")
		//inputBlock.WithMinLength(1)
		inputHint := plainTextBlock("You can edit this later")
		if details.currAnnotation != nil {
			inputBlock.WithInitialValue(details.currAnnotation.Notes)
			inputHint = nil
		}

		blockSet = append(blockSet,
			slack.NewDividerBlock(),
			slack.NewInputBlock("notes_input", plainTextBlock("Notes"), inputHint, inputBlock))
	}

	titleText := "Create Annotation"
	submitText := "Create"

	if details.currAnnotation != nil {
		details.metadata.AnnotationId = details.currAnnotation.ID
		titleText = "Update Annotation"
		submitText = "Update"
	}

	jsonMetadata, jsonErr := json.Marshal(details.metadata)
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
		//ExternalID:      details.messageId,
	}, nil
}

//func makeUpdatedCreateAnnotationModalView(ic *slack.InteractionCallback) slack.ModalViewRequest {
//	rosterName, rosterId := getAnnotationSelectedRoster(ic.View.State)
//
//	blockSet := ic.View.Blocks.BlockSet
//	for idx, block := range ic.View.Blocks.BlockSet {
//		if block.ID() == "roster_select_block" {
//			if sb, ok := block.(*slack.SectionBlock); ok {
//				sb.Text = plainTextBlock("Annotating roster: " + rosterName)
//				blockSet[idx] = sb
//			}
//		}
//	}
//
//	return slack.ModalViewRequest{
//		Type:            "modal",
//		CallbackID:      createAnnotationModalViewCallbackID,
//		PrivateMetadata: ic.View.PrivateMetadata,
//		Title:           ic.View.Title,
//		Close:           ic.View.Close,
//		Submit:          ic.View.Submit,
//		Blocks:          slack.Blocks{BlockSet: blockSet},
//		ExternalID:      ic.View.ExternalID,
//	}
//}

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

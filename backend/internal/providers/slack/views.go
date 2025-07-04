package slack

import (
	"context"
	"encoding/json"
	"fmt"
	rez "github.com/rezible/rezible"
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

type (
	messageId             string
	annotationViewContext struct {
		meta             annotationViewMetadata
		rosters          []*ent.OncallRoster
		selectedRosterId uuid.UUID
		currAnnotation   *ent.OncallAnnotation
	}
	annotationViewMetadata struct {
		UserId       string    `json:"uid"`
		MsgId        messageId `json:"mid"`
		MsgText      string    `json:"mtx"`
		AnnotationId uuid.UUID `json:"aid,omitempty"`
	}
)

func (m messageId) getTimestamp() time.Time {
	_, ts, _ := strings.Cut(m.String(), "_")
	return convertSlackTs(ts)
}

func (m messageId) String() string {
	return string(m)
}

func getMessageId(ic *slack.InteractionCallback) messageId {
	return messageId(fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp))
}

func getSelectedAnnotationRoster(state *slack.ViewState) (string, uuid.UUID) {
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

func makeAnnotationViewContext(
	ctx context.Context,
	ic *slack.InteractionCallback,
	lookupUser rez.LookupChatUserFn,
	lookupMessageEvent rez.LookupChatMessageEventFn,
) (*annotationViewContext, error) {
	d := &annotationViewContext{}
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &d.meta); jsonErr != nil {
			return nil, jsonErr
		}
	} else {
		d.meta = annotationViewMetadata{
			MsgId:   getMessageId(ic),
			UserId:  ic.Message.User,
			MsgText: ic.Message.Text,
		}
	}

	user, userErr := lookupUser(ctx, d.meta.UserId)
	if userErr != nil {
		return nil, userErr
	}

	// TODO: probably a faster way to query this
	rosters, rostersErr := user.QueryOncallSchedules().QuerySchedule().QueryRoster().All(ctx)
	if rostersErr != nil && !ent.IsNotFound(rostersErr) {
		return nil, fmt.Errorf("failed to query oncall rosters for user: %w", rostersErr)
	}
	d.rosters = rosters
	if len(rosters) == 0 {
		return d, nil
	} else if len(rosters) == 1 {
		d.selectedRosterId = rosters[0].ID
	} else {
		_, d.selectedRosterId = getSelectedAnnotationRoster(ic.View.State)
	}
	if d.selectedRosterId != uuid.Nil {
		event, eventErr := lookupMessageEvent(ctx, d.meta.MsgId.String())
		if eventErr != nil && !ent.IsNotFound(eventErr) {
			return nil, fmt.Errorf("failed to lookup message event: %w", eventErr)
		}
		if event != nil {
			anno, annoErr := event.QueryAnnotations().Where(oncallannotation.RosterID(d.selectedRosterId)).First(ctx)
			if annoErr != nil && !ent.IsNotFound(annoErr) {
				return nil, fmt.Errorf("failed to lookup existing event annotation: %w", annoErr)
			}
			d.currAnnotation = anno
		}
	}

	return d, nil
}

func makeAnnotationModalView(c *annotationViewContext) (*slack.ModalViewRequest, error) {
	if len(c.rosters) == 0 {
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
		slack.NewRichTextSectionUserElement(c.meta.UserId, nil),
		slack.NewRichTextSectionDateElement(c.meta.MsgId.getTimestamp().Unix(), " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(c.meta.MsgText, &slack.RichTextSectionTextStyle{Italic: true}))

	blockSet = append(blockSet, slack.NewRichTextBlock("anno_msg", messageUserDetails, messageContentsDetails))

	selectedRosterIdx := -1
	rosterOptions := make([]*slack.OptionBlockObject, len(c.rosters))
	for i, r := range c.rosters {
		rosterOptions[i] = slack.NewOptionBlockObject(r.ID.String(), plainTextBlock(r.Name), nil)
		if r.ID == c.selectedRosterId {
			selectedRosterIdx = i
		}
	}

	rosterSelectElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		plainTextBlock("Roster Name"),
		"roster_select",
		rosterOptions...)

	rosterText := "Select a Roster to annotate"
	if selectedRosterIdx != -1 {
		rosterSelectElement.WithInitialOption(rosterOptions[selectedRosterIdx])
		rosterText = "Annotating roster: " + c.rosters[selectedRosterIdx].Name
	}
	rosterBlock := slack.NewSectionBlock(
		plainTextBlock(rosterText), nil,
		slack.NewAccessory(rosterSelectElement),
		slack.SectionBlockOptionBlockID("roster_select_block"))

	blockSet = append(blockSet, slack.NewDividerBlock(), rosterBlock)

	if selectedRosterIdx != -1 {
		inputBlock := slack.NewPlainTextInputBlockElement(nil, "notes_input_text")
		//inputBlock.WithMinLength(1)
		inputHint := plainTextBlock("You can edit this later")
		if c.currAnnotation != nil {
			inputBlock.WithInitialValue(c.currAnnotation.Notes)
			inputHint = nil
		}

		blockSet = append(blockSet,
			slack.NewDividerBlock(),
			slack.NewInputBlock("notes_input", plainTextBlock("Notes"), inputHint, inputBlock))
	}

	titleText := "Create Annotation"
	submitText := "Create"

	if c.currAnnotation != nil {
		c.meta.AnnotationId = c.currAnnotation.ID
		titleText = "Update Annotation"
		submitText = "Update"
	}

	jsonMetadata, jsonErr := json.Marshal(c.meta)
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
	}, nil
}

func getAnnotationModalAnnotation(ctx context.Context, view slack.View, lookupUser rez.LookupChatUserFn) (*ent.OncallAnnotation, error) {
	if lookupUser == nil {
		return nil, fmt.Errorf("no lookupUser func")
	}

	var meta annotationViewMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	user, userErr := lookupUser(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	var notes string
	if view.State != nil {
		if notesInput, inputOk := view.State.Values["notes_input"]; inputOk {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}
	}

	_, rosterId := getSelectedAnnotationRoster(view.State)
	if rosterId == uuid.Nil {
		return nil, fmt.Errorf("no roster id selected")
	}

	anno := &ent.OncallAnnotation{
		ID:        meta.AnnotationId,
		RosterID:  rosterId,
		CreatorID: user.ID,
		Notes:     notes,
		Edges: ent.OncallAnnotationEdges{
			Event: &ent.OncallEvent{
				ProviderID:  meta.MsgId.String(),
				Kind:        "message",
				Timestamp:   meta.MsgId.getTimestamp(),
				Source:      "slack",
				Title:       "Slack Message",
				Description: meta.MsgText,
			},
		},
	}

	return anno, nil
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

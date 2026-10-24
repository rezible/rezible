package slackincidents

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const viewCallbackIdIncidentDetailsModal = "incident_details_modal"

type incidentDetailsModalViewMetadata struct {
	// CommandChannelId is where /incident ran, for the declarer's confirmation.
	CommandChannelId string    `json:"cid"`
	IncidentId       uuid.UUID `json:"iid,omitempty"`
}

// makeIncidentDetailsModalView builds the modal that declares an incident, or updates the one in the metadata.
func (a *App) makeIncidentDetailsModalView(ctx context.Context, meta *incidentDetailsModalViewMetadata) (*slack.ModalViewRequest, error) {
	var curr *ent.Incident
	if meta.IncidentId != uuid.Nil {
		inc, incErr := a.incidents.GetIncident(ctx, incident.ID(meta.IncidentId))
		if incErr != nil {
			return nil, fmt.Errorf("get incident: %w", incErr)
		}
		curr = inc
	}

	incMeta, incMetaErr := a.incidents.GetIncidentMetadata(ctx)
	if incMetaErr != nil {
		return nil, fmt.Errorf("get incident metadata: %w", incMetaErr)
	}

	builder := &incidentDetailsModalViewBuilder{incident: curr}
	blockSet := builder.Build(incMeta)

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("marshal metadata: %w", jsonErr)
	}

	titleText := "Declare incident"
	submitText := "Declare"
	if curr != nil {
		titleText = "Update incident"
		submitText = "Save"
	}

	view := &slack.ModalViewRequest{
		Type:            slack.VTModal,
		CallbackID:      viewCallbackIdIncidentDetailsModal,
		Title:           slackintegration.PlainTextBlock(titleText),
		Submit:          slackintegration.PlainTextBlock(submitText),
		Close:           slackintegration.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}
	return view, nil
}

type incidentDetailsModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
}

func (b *incidentDetailsModalViewBuilder) Build(im *rez.IncidentMetadata) slack.Blocks {
	b.blocks = make([]slack.Block, 0)
	b.makeTitleInput()
	b.makeSummaryInput()
	if len(im.Severities) > 0 {
		b.makeSeveritySelect(im.Severities)
	}
	if len(im.Types) > 0 {
		b.makeTypeSelect(im.Types)
	}
	if len(im.Tags) > 0 {
		b.makeTagsSelect(im.Tags)
	}
	if len(im.Fields) > 0 {
		b.makeCustomFieldSelect(im.Fields)
	}
	if b.incident != nil && b.incident.ResponseState != incident.ResponseStateResolved {
		b.makeOpenMilestoneModalButton()
	}
	return slack.Blocks{BlockSet: b.blocks}
}

var (
	incidentModalTitleIds    = slackintegration.BlockActionIds{Block: "title", Input: "title_input"}
	incidentModalSummaryIds  = slackintegration.BlockActionIds{Block: "incident_summary", Input: "summary_input"}
	incidentModalSeverityIds = slackintegration.BlockActionIds{Block: "incident_severity", Input: "severity_select"}
	incidentModalTypeIds     = slackintegration.BlockActionIds{Block: "incident_type", Input: "type_select"}
	incidentModalTagIds      = slackintegration.BlockActionIds{Block: "incident_tags", Input: "tags_select"}
)

const incidentModalFieldInputPrefix = "incident_field_select_"

const (
	// maxOptionRunes is Slack's limit on an option's text and description.
	maxOptionRunes = 75
	// maxInputRunes is Slack's default limit on a text input's value.
	maxInputRunes = 3000
)

// optionText is an option's text or description, within Slack's limit. Empty text gives nil, which is omitted.
func optionText(text string) *slack.TextBlockObject {
	return slackintegration.PlainTextBlock(slackintegration.TruncateRunes(text, maxOptionRunes))
}

func incidentModalFieldOptionIds(fieldId string) slackintegration.BlockActionIds {
	return slackintegration.BlockActionIds{Block: "incident_field_" + fieldId, Input: incidentModalFieldInputPrefix + fieldId}
}

// incidentDetailsModalInput reads every field the modal shows. A select left empty clears its value.
func incidentDetailsModalInput(state *slack.ViewState) rez.UpdateIncidentParams {
	title := incidentModalTitleIds.GetStateValue(state)
	summary := incidentModalSummaryIds.GetStateValue(state)
	params := rez.UpdateIncidentParams{Title: &title, Summary: &summary}

	if slackintegration.GetViewStateBlockAction(state, incidentModalSeverityIds) != nil {
		if sevId, sevErr := uuid.Parse(incidentModalSeverityIds.GetStateSelectedValue(state)); sevErr == nil {
			params.SeverityID = &sevId
		} else {
			params.ClearSeverity = true
		}
	}

	if slackintegration.GetViewStateBlockAction(state, incidentModalTypeIds) != nil {
		if typeId, typeErr := uuid.Parse(incidentModalTypeIds.GetStateSelectedValue(state)); typeErr == nil {
			params.TypeID = &typeId
		} else {
			params.ClearType = true
		}
	}

	if slackintegration.GetViewStateBlockAction(state, incidentModalTagIds) != nil {
		tagIDs := make([]uuid.UUID, 0)
		for _, selectedTagId := range incidentModalTagIds.GetStateSelectedValues(state) {
			if tagId, tagErr := uuid.Parse(selectedTagId); tagErr == nil {
				tagIDs = append(tagIDs, tagId)
			}
		}
		params.TagIDs = &tagIDs
	}

	for _, actions := range state.Values {
		for actionID, blockAction := range actions {
			if !strings.HasPrefix(actionID, incidentModalFieldInputPrefix) {
				continue
			}
			if params.FieldOptionIDs == nil {
				params.FieldOptionIDs = &[]uuid.UUID{}
			}
			if fieldOptionID, fieldOptionErr := uuid.Parse(blockAction.SelectedOption.Value); fieldOptionErr == nil {
				*params.FieldOptionIDs = append(*params.FieldOptionIDs, fieldOptionID)
			}
		}
	}
	return params
}

func (b *incidentDetailsModalViewBuilder) makeTitleInput() {
	titleInput := slack.NewPlainTextInputBlockElement(nil, incidentModalTitleIds.Input)
	if b.incident != nil {
		titleInput.WithInitialValue(slackintegration.TruncateRunes(b.incident.Title, maxInputRunes))
	}
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTitleIds.Block, slackintegration.PlainTextBlock("Title"), nil, titleInput))
}

func (b *incidentDetailsModalViewBuilder) makeSummaryInput() {
	summaryInput := slack.NewPlainTextInputBlockElement(nil, incidentModalSummaryIds.Input).WithMultiline(true)
	if b.incident != nil && b.incident.Summary != "" {
		summaryInput.WithInitialValue(slackintegration.TruncateRunes(b.incident.Summary, maxInputRunes))
	}
	block := slack.NewInputBlock(incidentModalSummaryIds.Block, slackintegration.PlainTextBlock("Summary"), nil, summaryInput)
	b.blocks = append(b.blocks, block.WithOptional(true))
}

func (b *incidentDetailsModalViewBuilder) makeOpenMilestoneModalButton() {
	statusButton := slack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, b.incident.ID.String(),
		slackintegration.PlainTextBlock("Record status"))
	b.blocks = append(b.blocks, slack.NewActionBlock("incident_actions", statusButton))
}

// makeSeveritySelect offers severities in rank order; a new incident starts with none.
func (b *incidentDetailsModalViewBuilder) makeSeveritySelect(sevs ent.IncidentSeverities) {
	ranked := slices.Clone(sevs)
	slices.SortFunc(ranked, func(x, y *ent.IncidentSeverity) int {
		return cmp.Or(cmp.Compare(x.Rank, y.Rank), bytes.Compare(x.ID[:], y.ID[:]))
	})
	options := make([]*slack.OptionBlockObject, len(ranked))
	var initial *slack.OptionBlockObject
	for i, sev := range ranked {
		options[i] = slack.NewOptionBlockObject(sev.ID.String(), optionText(sev.Name), optionText(sev.Description))
		if b.incident != nil && b.incident.SeverityID != nil && *b.incident.SeverityID == sev.ID {
			initial = options[i]
		}
	}
	severitySelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentModalSeverityIds.Input, options...)
	if initial != nil {
		severitySelect.WithInitialOption(initial)
	}
	block := slack.NewInputBlock(incidentModalSeverityIds.Block, slackintegration.PlainTextBlock("Severity"), nil, severitySelect)
	b.blocks = append(b.blocks, block.WithOptional(true))
}

func (b *incidentDetailsModalViewBuilder) makeTypeSelect(types ent.IncidentTypes) {
	options := make([]*slack.OptionBlockObject, len(types))
	var initial *slack.OptionBlockObject
	for i, t := range types {
		options[i] = slack.NewOptionBlockObject(t.ID.String(), optionText(t.Name), nil)
		if b.incident != nil && b.incident.TypeID != nil && *b.incident.TypeID == t.ID {
			initial = options[i]
		}
	}
	typeSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentModalTypeIds.Input, options...)
	if initial != nil {
		typeSelect.WithInitialOption(initial)
	}
	block := slack.NewInputBlock(incidentModalTypeIds.Block, slackintegration.PlainTextBlock("Type"), nil, typeSelect)
	b.blocks = append(b.blocks, block.WithOptional(true))
}

func (b *incidentDetailsModalViewBuilder) makeTagsSelect(tags ent.IncidentTags) {
	options := make([]*slack.OptionBlockObject, len(tags))
	initialOptions := make([]*slack.OptionBlockObject, 0, len(tags))
	selectedTagIds := mapset.NewSet[uuid.UUID]()
	if b.incident != nil {
		for _, tag := range b.incident.Edges.TagAssignments {
			selectedTagIds.Add(tag.ID)
		}
	}
	for i, tag := range tags {
		label := tag.Value
		if tag.Key != "" {
			label = tag.Key + ": " + tag.Value
		}
		options[i] = slack.NewOptionBlockObject(tag.ID.String(), optionText(label), nil)
		if selectedTagIds.Contains(tag.ID) {
			initialOptions = append(initialOptions, options[i])
		}
	}

	tagSelect := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeStatic, slackintegration.PlainTextBlock("Select tags"),
		incidentModalTagIds.Input, options...)
	if len(initialOptions) > 0 {
		tagSelect.WithInitialOptions(initialOptions...)
	}
	block := slack.NewInputBlock(incidentModalTagIds.Block, slackintegration.PlainTextBlock("Tags"), nil, tagSelect)
	b.blocks = append(b.blocks, block.WithOptional(true))
}

func (b *incidentDetailsModalViewBuilder) makeCustomFieldSelect(fields ent.IncidentFields) {
	b.blocks = append(b.blocks, slack.NewDividerBlock())
	selectedOptionIds := mapset.NewSet[uuid.UUID]()
	if b.incident != nil {
		for _, selection := range b.incident.Edges.FieldSelections {
			selectedOptionIds.Add(selection.ID)
		}
	}
	for _, field := range fields {
		if len(field.Edges.Options) == 0 {
			continue
		}
		fieldOptions := make([]*slack.OptionBlockObject, len(field.Edges.Options))
		var initial *slack.OptionBlockObject
		for i, opt := range field.Edges.Options {
			fieldOptions[i] = slack.NewOptionBlockObject(opt.ID.String(), optionText(opt.Value), nil)
			if selectedOptionIds.Contains(opt.ID) {
				initial = fieldOptions[i]
			}
		}
		ids := incidentModalFieldOptionIds(field.ID.String())
		fieldOptionSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, slackintegration.PlainTextBlock("Select an option"), ids.Input, fieldOptions...)
		if initial != nil {
			fieldOptionSelect.WithInitialOption(initial)
		}
		block := slack.NewInputBlock(ids.Block, slackintegration.PlainTextBlock(field.Name), nil, fieldOptionSelect)
		b.blocks = append(b.blocks, block.WithOptional(true))
	}
}

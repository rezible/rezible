package slack

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

func TestSetIncidentDetailsModalInputMutationFieldsAppliesMetadataSelections(t *testing.T) {
	severityID := uuid.New()
	typeID := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()
	fieldOptionID := uuid.New()

	state := &slack.ViewState{
		Values: map[string]map[string]slack.BlockAction{
			incidentModalTitleIds.Block: {
				incidentModalTitleIds.Input: {Value: "API outage"},
			},
			incidentModalSeverityIds.Block: {
				incidentModalSeverityIds.Input: {SelectedOption: slack.OptionBlockObject{Value: severityID.String()}},
			},
			incidentModalTypeIds.Block: {
				incidentModalTypeIds.Input: {SelectedOption: slack.OptionBlockObject{Value: typeID.String()}},
			},
			incidentModalTagIds.Block: {
				incidentModalTagIds.Input: {
					SelectedOptions: []slack.OptionBlockObject{
						{Value: tagID1.String()},
						{Value: tagID2.String()},
					},
				},
			},
			"incident_field_environment": {
				"incident_field_select_environment": {SelectedOption: slack.OptionBlockObject{Value: fieldOptionID.String()}},
			},
		},
	}

	mutation := ent.NewClient().Incident.Create().Mutation()
	setIncidentDetailsModalInputMutationFields(mutation, state)

	title, ok := mutation.Title()
	require.True(t, ok)
	require.Equal(t, "API outage", title)

	gotSeverityID, ok := mutation.SeverityID()
	require.True(t, ok)
	require.Equal(t, severityID, gotSeverityID)

	gotTypeID, ok := mutation.TypeID()
	require.True(t, ok)
	require.Equal(t, typeID, gotTypeID)

	require.True(t, mutation.TagAssignmentsCleared())
	require.ElementsMatch(t, []uuid.UUID{tagID1, tagID2}, mutation.TagAssignmentsIDs())

	require.True(t, mutation.FieldSelectionsCleared())
	require.ElementsMatch(t, []uuid.UUID{fieldOptionID}, mutation.FieldSelectionsIDs())
}

func TestSetIncidentDetailsModalInputMutationFieldsLeavesAbsentMetadataUntouched(t *testing.T) {
	state := &slack.ViewState{
		Values: map[string]map[string]slack.BlockAction{
			incidentModalTitleIds.Block: {
				incidentModalTitleIds.Input: {Value: "API outage"},
			},
		},
	}

	mutation := ent.NewClient().Incident.Create().Mutation()
	setIncidentDetailsModalInputMutationFields(mutation, state)

	require.False(t, mutation.TagAssignmentsCleared())
	require.False(t, mutation.FieldSelectionsCleared())
}

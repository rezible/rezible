package slack

import "github.com/slack-go/slack"

func getViewStateBlockAction(state *slack.ViewState, ids blockActionIds) *slack.BlockAction {
	if block, blockOk := state.Values[ids.Block]; blockOk {
		if action, inputOk := block[ids.Input]; inputOk {
			return &action
		}
	}
	return nil
}

type blockActionIds struct {
	Block string
	Input string
}

func (ids blockActionIds) GetStateValue(state *slack.ViewState) string {
	action := getViewStateBlockAction(state, ids)
	if action == nil {
		return ""
	}
	return action.Value
}

func (ids blockActionIds) GetStateSelectedValue(state *slack.ViewState) string {
	action := getViewStateBlockAction(state, ids)
	if action == nil {
		return ""
	}
	return action.SelectedOption.Value
}

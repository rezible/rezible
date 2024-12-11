package slack

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io"
	"net/http"
)

func (p *ChatProvider) handleEventsWebhook(w http.ResponseWriter, r *http.Request) {
	if verifyErr := p.verifyWebhook(w, r); verifyErr != nil {
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
		go p.handleCallbackEvent(ev)
		w.WriteHeader(http.StatusOK)
	} else {
		log.Warn().Str("type", ev.Type).Msg("failed to handle event")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (p *ChatProvider) handleCallbackEvent(ev slackevents.EventsAPIEvent) {
	if ev.Type == slackevents.CallbackEvent {
		innerEvent := ev.InnerEvent
		switch data := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			p.handleAppMentionEvent(ev, data)
		}
	}
}

func (p *ChatProvider) handleAppMentionEvent(e slackevents.EventsAPIEvent, data *slackevents.AppMentionEvent) {
	fmt.Printf("mention event: %+v\n", data)
	_, _, msgErr := p.client.PostMessage(data.Channel, slack.MsgOptionText("hello", false))
	if msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to message")
	}
}

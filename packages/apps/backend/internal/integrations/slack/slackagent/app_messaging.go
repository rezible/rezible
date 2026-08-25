package slackagent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	asb "github.com/rezible/rezible/ent/agentsessionbinding"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	slackAgentBindingSource             = "slack"
	slackAgentBindingResourceKindThread = "thread"
)

type agentThreadBindingResource struct {
	ChannelId string
	ThreadTs  string
}

func (r *agentThreadBindingResource) makeRef() string {
	return r.ChannelId + ":" + r.ThreadTs
}

func (a *App) getBoundThreadResource(binding *ent.AgentSessionBinding) (*agentThreadBindingResource, error) {
	channelID, threadTs, ok := strings.Cut(binding.ResourceRef, ":")
	if !ok || channelID == "" || threadTs == "" {
		return nil, fmt.Errorf("invalid slack thread resource ref %q", binding.ResourceRef)
	}
	return &agentThreadBindingResource{ChannelId: channelID, ThreadTs: threadTs}, nil
}

func (a *App) lookupSlackThreadBinding(ctx context.Context, integrationID uuid.UUID, res *agentThreadBindingResource) (*ent.AgentSessionBinding, error) {
	return a.agents.LookupAgentSessionBinding(ctx,
		asb.IntegrationID(integrationID),
		asb.Source(slackAgentBindingSource),
		asb.ResourceKind(slackAgentBindingResourceKindThread),
		asb.ResourceRef(res.makeRef()),
	)
}

func (a *App) onAgentMentionedByUser(ctx context.Context, usr *ent.User, intg *ent.Integration, data *slackevents.AppMentionEvent) error {
	res := &agentThreadBindingResource{ChannelId: data.Channel, ThreadTs: data.TimeStamp}
	if data.ThreadTimeStamp != "" {
		res.ThreadTs = data.ThreadTimeStamp
	}

	binding, bindingErr := a.lookupSlackThreadBinding(ctx, intg.ID, res)
	if bindingErr != nil && !ent.IsNotFound(bindingErr) {
		return fmt.Errorf("failed to lookup agent session binding: %w", bindingErr)
	}

	cleanedText := strings.TrimSpace(mentionRe.ReplaceAllString(data.Text, ""))

	if binding == nil {
		return a.startBoundAgentThread(ctx, intg, usr.ID, res, cleanedText)
	}

	slog.Debug("continuing existing agent session in thread")
	params := &rez.RequestAgentTurnParams{
		Input: &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(cleanedText)},
	}
	_, requestErr := a.agents.RequestAgentTurn(ctx, binding.AgentSessionID, params)
	if requestErr != nil {
		slog.Error("failed to request chat agent turn", "error", requestErr)
	}
	return nil
}

const agentSessionMetadataIntegrationKey = "source_integration"

func (a *App) startBoundAgentThread(ctx context.Context, intg *ent.Integration, userId uuid.UUID, res *agentThreadBindingResource, msg string) error {
	bindingParams := rez.AgentSessionBindingParams{
		IntegrationID: new(intg.ID),
		Source:        slackAgentBindingSource,
		ResourceKind:  slackAgentBindingResourceKindThread,
		ResourceRef:   res.makeRef(),
	}
	createSessionParams := rez.CreateAgentSessionParams{
		AgentName: rezai.ChatAgent.Name,
		Input:     rezai.ChatAgentInput{UserId: userId, Message: msg},
		Bindings:  []rez.AgentSessionBindingParams{bindingParams},
		Metadata: map[string]any{
			agentSessionMetadataIntegrationKey: integrationName,
		},
	}
	if _, sessionErr := a.agents.CreateAgentSession(ctx, createSessionParams); sessionErr != nil {
		slog.Error("failed to create chat agent session", "error", sessionErr)
	}
	return nil
}

func (a *App) onSlackAgentResponseEvent(ctx context.Context, ev *rezai.EventOnAgentTurnFinished) error {
	binding, bindingErr := a.lookupAiAgentSessionBinding(ctx, ev.AgentSessionId)
	if bindingErr != nil {
		return fmt.Errorf("session binding: %w", bindingErr)
	}
	if binding != nil {
		return a.onBoundAgentThreadAgentResponse(ctx, binding, ev.Response)
	}
	slog.Debug("slack agent response with no session binding", "event", ev)
	return nil
}

func (a *App) onBoundAgentThreadAgentResponse(ctx context.Context, binding *ent.AgentSessionBinding, response *ai.Message) error {
	tr, resErr := a.getBoundThreadResource(binding)
	if resErr != nil {
		return resErr
	}

	args := SendMessageJobArgs{
		Message:       response.Text(),
		IntegrationID: *binding.IntegrationID,
		Channel:       tr.ChannelId,
		ReplyTs:       tr.ThreadTs,
	}
	if _, cmdErr := a.jobs.Insert(ctx, args, nil); cmdErr != nil {
		return fmt.Errorf("insert job: %w", cmdErr)
	}

	return nil
}

func (a *App) onBoundAgentThreadUserMessage(ctx context.Context, binding *ent.AgentSessionBinding, msg *slack.Msg) error {
	args := HandleBoundAgentThreadMessagedArgs{
		BindingId: binding.ID,
		MessageTs: msg.Timestamp,
	}
	if _, jobErr := a.jobs.Insert(ctx, args, nil); jobErr != nil {
		slog.Error("failed to insert check agent thread job", "error", jobErr)
	}
	return nil
}

type HandleBoundAgentThreadMessagedArgs struct {
	BindingId uuid.UUID `json:"binding_id" river:"unique"`
	MessageTs string    `json:"message_ts" river:"unique"`
}

func (a HandleBoundAgentThreadMessagedArgs) Kind() string {
	return "slack-agent-handle-bound-thread-response"
}

func (HandleBoundAgentThreadMessagedArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

func (a *App) getThreadMessageContext(ctx context.Context, client *slack.Client, res *agentThreadBindingResource, msgTs string, limit int) ([]slack.Message, error) {
	params := &slack.GetConversationRepliesParameters{
		ChannelID: res.ChannelId,
		Timestamp: res.ThreadTs,
		Cursor:    "",
		Latest:    msgTs,
		Inclusive: true,
		//Limit:              100,
	}
	prevBatch := make([]slack.Message, 0, limit)
	for {
		replies, hasMore, nextCursor, err := client.GetConversationRepliesContext(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get thread replies: %w", err)
		}

		batch := replies[max(0, len(replies)-limit):]
		if !hasMore {
			if len(batch) < limit && len(prevBatch) > 0 {
				numShort := limit - len(batch)
				batch = append(prevBatch[max(0, len(prevBatch)-numShort):], batch...)
			}
			return batch, nil
		}

		params.Cursor = nextCursor
		prevBatch = batch
	}
}

func (a *App) getUserNames(ctx context.Context, client *slack.Client, msgs []slack.Message) (map[string]string, error) {
	botIds := mapset.NewSet[string]()
	ids := mapset.NewSet[string]()
	for _, msg := range msgs {
		if msg.BotProfile != nil {
			botIds.Add(msg.User)
		}
		ids.Add(msg.User)
	}

	users, usersErr := client.GetUsersInfoContext(ctx, ids.ToSlice()...)
	if usersErr != nil {
		return nil, fmt.Errorf("failed to get users: %w", usersErr)
	}
	profiles := map[string]string{}
	if users != nil {
		for _, u := range *users {
			name := u.Profile.RealNameNormalized
			//if botIds.Contains(u.ID) {
			//	name += "(bot)"
			//}
			profiles[u.ID] = name
		}
	}
	return profiles, nil
}

func (a *App) handleBoundAgentThreadMessagedJob(ctx context.Context, args HandleBoundAgentThreadMessagedArgs) error {
	binding, bindingErr := a.agents.LookupAgentSessionBinding(ctx, asb.ID(args.BindingId))
	if bindingErr != nil {
		return fmt.Errorf("lookup slack agent session binding: %w", bindingErr)
	}

	res, resErr := a.getBoundThreadResource(binding)
	if resErr != nil {
		return fmt.Errorf("bound thread resource: %w", resErr)
	}

	intg, intgErr := binding.Edges.IntegrationOrErr()
	if intgErr != nil {
		return fmt.Errorf("no integration for binding: %w", intgErr)
	}

	cw, cwErr := slackintegration.NewClientWrapper(intg)
	if cwErr != nil {
		return fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	client := cw.Client()
	msgs, msgsErr := a.getThreadMessageContext(ctx, client, res, args.MessageTs, 3)
	if msgsErr != nil {
		return fmt.Errorf("failed to get thread message context: %w", msgsErr)
	}
	if len(msgs) == 0 {
		return fmt.Errorf("no messages")
	}

	usernamesMap, usernamesErr := a.getUserNames(ctx, client, msgs)
	if usernamesErr != nil {
		return fmt.Errorf("failed to get user profiles: %w", usernamesErr)
	}

	usrMsg := msgs[len(msgs)-1]

	workflowInput := rezai.ClassifyAgentThreadResponseInput{
		PreviousMessages: make([]string, max(0, len(msgs)-1)),
	}
	for i, msg := range msgs {
		name := msg.User
		if username, ok := usernamesMap[msg.User]; ok {
			name = username
		}
		msgText := fmt.Sprintf("%s: %s", name, msg.Text)
		if i < len(msgs)-1 {
			workflowInput.PreviousMessages[i] = msgText
		} else {
			workflowInput.UserMessage = msgText
		}
	}

	output, workflowErr := a.responseClassifier.Run(ctx, workflowInput)
	if workflowErr != nil {
		return fmt.Errorf("check response required: %w", workflowErr)
	}
	if !output.ShouldReply {
		return nil
	}
	params := &rez.RequestAgentTurnParams{
		Input: &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(usrMsg.Text)},
	}
	if _, requestErr := a.agents.RequestAgentTurn(ctx, binding.AgentSessionID, params); requestErr != nil {
		return fmt.Errorf("request agent turn: %w", requestErr)
	}

	return nil
}

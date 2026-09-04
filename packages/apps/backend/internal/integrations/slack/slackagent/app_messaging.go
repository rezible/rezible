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
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	slackAgentBindingProvider = "slack"
)

type agentThreadBindingResource struct {
	ChannelId string
	ThreadTs  string
}

func (r *agentThreadBindingResource) makeRef() string {
	return "thread:" + r.ChannelId + ":" + r.ThreadTs
}

func (a *App) getBoundThreadResource(binding *ent.AgentSessionBinding) (*agentThreadBindingResource, error) {
	resourceRef := strings.TrimPrefix(binding.ProviderResourceRef, "thread:")
	channelID, threadTs, ok := strings.Cut(resourceRef, ":")
	if !ok || channelID == "" || threadTs == "" {
		return nil, fmt.Errorf("invalid slack thread resource ref %q", binding.ProviderResourceRef)
	}
	return &agentThreadBindingResource{ChannelId: channelID, ThreadTs: threadTs}, nil
}

func (a *App) lookupSlackThreadBinding(ctx context.Context, integrationID uuid.UUID, res *agentThreadBindingResource) (*ent.AgentSessionBinding, error) {
	return a.agents.LookupAgentSessionBinding(ctx,
		asb.IntegrationID(integrationID),
		asb.Provider(slackAgentBindingProvider),
		asb.ProviderResourceRef(res.makeRef()),
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
	credentials, credentialsErr := slackintegration.GetValidatedConfig(intg.InstallationConfig)
	if credentialsErr != nil {
		return fmt.Errorf("decode slack installation config: %w", credentialsErr)
	}
	providerNamespace := ""
	if credentials.Team != nil {
		providerNamespace = credentials.Team.Id
	} else if credentials.Enterprise != nil {
		providerNamespace = credentials.Enterprise.Id
	}
	bindingRef := rez.ProviderResourceRef{
		Provider:          slackAgentBindingProvider,
		ProviderNamespace: providerNamespace,
		ResourceRef:       res.makeRef(),
	}
	bindingParams := rez.AgentSessionBindingParams{
		ProviderResourceRef: bindingRef,
		IntegrationID:       new(intg.ID),
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

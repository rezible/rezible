import { z } from "zod";

import type { IntegrationSettingsDefinition } from "$features/settings/lib/integrationSettingsForms.svelte";

// The backend checks that a chosen channel is a Slack channel ID.
const slackAgentSettingsSchema = z.object({
	situationNoticeChannelId: z.string(),
});

export type SlackAgentSettings = z.infer<typeof slackAgentSettingsSchema>;

// Missing and explicitly null settings choose no channel.
const readChannelId = (value: unknown) => {
	if (typeof value === "string") {
		return value;
	}
	return "";
};

export const slackAgentSettings: IntegrationSettingsDefinition<SlackAgentSettings> = {
	schema: slackAgentSettingsSchema,

	read: (userSettings) => ({
		situationNoticeChannelId: readChannelId(userSettings.situation_notice_channel_id),
	}),

	write: (values, userSettings) => ({
		...userSettings,
		situation_notice_channel_id: values.situationNoticeChannelId,
	}),
};

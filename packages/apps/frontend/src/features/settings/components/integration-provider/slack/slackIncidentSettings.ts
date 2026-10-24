import { z } from "zod";

import {
	type IntegrationSettingsDefinition,
	readBoolean,
	readSettingsObject,
	readString,
} from "$features/settings/lib/integrationSettingsForms.svelte";

const slackIncidentSettingsSchema = z.object({
	AnnouncementChannelID: z.string().trim(),
	ChannelNamePattern: z
		.string()
		.trim()
		.min(1, "Enter a channel name pattern.")
		.refine((pattern) => pattern.includes("{slug}") || pattern.includes("{id}"), {
			error: "Include {slug} or {id}, so each incident gets its own channel.",
		}),
	AutoCreateVideoConference: z.boolean(),
});

export type SlackIncidentSettings = z.infer<typeof slackIncidentSettingsSchema>;

const defaultSettings: SlackIncidentSettings = {
	AnnouncementChannelID: "",
	ChannelNamePattern: "incident-{slug}",
	AutoCreateVideoConference: false,
};

export const slackIncidentSettings: IntegrationSettingsDefinition<SlackIncidentSettings> = {
	schema: slackIncidentSettingsSchema,

	read: (userSettings) => {
		const incidents = readSettingsObject(userSettings.incidents);
		return {
			AnnouncementChannelID: readString(
				incidents.announcement_channel_id,
				defaultSettings.AnnouncementChannelID
			),
			ChannelNamePattern: readString(
				incidents.channel_name_pattern,
				defaultSettings.ChannelNamePattern
			),
			AutoCreateVideoConference: readBoolean(
				incidents.auto_create_video_conference,
				defaultSettings.AutoCreateVideoConference
			),
		};
	},

	// The incidents settings are written whole: the backend rejects keys it doesn't know.
	write: (values, userSettings) => {
		return {
			...userSettings,
			incidents: {
				announcement_channel_id: values.AnnouncementChannelID,
				channel_name_pattern: values.ChannelNamePattern,
				auto_create_video_conference: values.AutoCreateVideoConference,
			},
		};
	},
};

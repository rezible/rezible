import { z } from "zod";

import {
	type IntegrationSettingsDefinition,
	readBoolean,
	readSettingsObject,
	readString,
} from "$features/settings/lib/integrationSettingsForms.svelte";

export const inviteModeOptions = [
	{ value: "assigned_users", label: "Assigned users" },
	{ value: "all_users", label: "All users" },
] as const;

const slackIncidentSettingsSchema = z.object({
	AnnouncementChannelID: z.string().trim(),
	ChannelNamePattern: z.string().trim().min(1, "Enter a channel name pattern."),
	AutoCreateVideoConference: z.boolean(),
	InviteMode: z.enum(["assigned_users", "all_users"], { error: "Choose who to invite." }),
});

export type SlackIncidentSettings = z.infer<typeof slackIncidentSettingsSchema>;

const defaultSettings: SlackIncidentSettings = {
	AnnouncementChannelID: "",
	ChannelNamePattern: "incident-{slug}",
	AutoCreateVideoConference: false,
	InviteMode: "assigned_users",
};

const readInviteMode = (value: unknown): SlackIncidentSettings["InviteMode"] => {
	if (value === "assigned_users" || value === "all_users") {
		return value;
	}
	return defaultSettings.InviteMode;
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
			InviteMode: readInviteMode(incidents.invite_mode),
		};
	},

	write: (values, userSettings) => {
		const incidents = readSettingsObject(userSettings.incidents);
		return {
			...userSettings,
			incidents: {
				...incidents,
				announcement_channel_id: values.AnnouncementChannelID,
				channel_name_pattern: values.ChannelNamePattern,
				auto_create_video_conference: values.AutoCreateVideoConference,
				invite_mode: values.InviteMode,
			},
		};
	},
};

import { z } from "zod";

import {
	type IntegrationSettingsDefinition,
	readBoolean,
} from "$features/settings/lib/integrationSettingsForms.svelte";

const googleSettingsSchema = z.object({
	EnableVideoConference: z.boolean(),
});

export type GoogleSettings = z.infer<typeof googleSettingsSchema>;

export const googleSettings: IntegrationSettingsDefinition<GoogleSettings> = {
	schema: googleSettingsSchema,

	read: (userSettings) => ({
		EnableVideoConference: readBoolean(userSettings.EnableVideoConference, false),
	}),

	write: (values, userSettings) => ({
		...userSettings,
		...values,
	}),
};

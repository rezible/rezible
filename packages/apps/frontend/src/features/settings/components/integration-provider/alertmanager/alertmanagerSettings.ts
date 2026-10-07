import { z } from "zod";

import type {
	IntegrationSettingsDefinition,
	IntegrationUserSettings,
} from "$features/settings/lib/integrationSettingsForms.svelte";

// The backend's labels when an installation has no service_labels setting.
export const defaultServiceLabels = ["service", "service_name", "app", "app_kubernetes_io_name"];

const labelNamePattern = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

// Labels are edited as text: names separated by commas, spaces or new lines.
export const parseServiceLabels = (text: string): string[] => {
	return text.split(/[\s,]+/).filter((label) => label !== "");
};

const serviceLabelsText = z.string().superRefine((text, ctx) => {
	const seen = new Set<string>();
	for (const label of parseServiceLabels(text)) {
		if (!labelNamePattern.test(label)) {
			ctx.addIssue({ code: "custom", message: `"${label}" is not a valid label name.` });
			return;
		}
		if (seen.has(label)) {
			ctx.addIssue({ code: "custom", message: `"${label}" is listed more than once.` });
			return;
		}
		seen.add(label);
	}
});

const alertmanagerSettingsSchema = z.object({
	serviceLabels: serviceLabelsText,
});

export type AlertmanagerSettings = z.infer<typeof alertmanagerSettingsSchema>;

const readServiceLabels = (userSettings: IntegrationUserSettings): string[] => {
	const value = userSettings.service_labels;
	if (!Array.isArray(value)) {
		return defaultServiceLabels;
	}
	return value.filter((label): label is string => typeof label === "string");
};

export const alertmanagerSettings: IntegrationSettingsDefinition<AlertmanagerSettings> = {
	schema: alertmanagerSettingsSchema,

	read: (userSettings) => ({
		serviceLabels: readServiceLabels(userSettings).join(", "),
	}),

	// An empty list is saved explicitly: it turns service attachment off rather than restoring the defaults.
	write: (values, userSettings) => ({
		...userSettings,
		service_labels: parseServiceLabels(values.serviceLabels),
	}),
};

const webhookUrlPlaceholder = "<webhook url>";

// The routing example wraps the existing routing tree so existing notifications keep their receivers.
export const routingExample = (webhookUrl?: string): string => {
	const url = webhookUrl ?? webhookUrlPlaceholder;
	return `receivers:
  - name: rezible
    webhook_configs:
      - url: ${url}
        send_resolved: true
        max_alerts: 0
route:
  receiver: existing-default
  # Copy the old root's group_by, group_wait, group_interval and repeat_interval
  # here, preserving any explicit values. Omitted values keep their defaults.
  routes:
    - receiver: rezible
      continue: true
      repeat_interval: 4h
    # Move the ENTIRE previous root route here, including its receiver,
    # settings and child routes. This example shows a root with no children.
    - receiver: existing-default
`;
};

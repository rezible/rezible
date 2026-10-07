import { z } from "zod";

import type {
	IntegrationSettingsDefinition,
	IntegrationUserSettings,
} from "$features/settings/lib/integrationSettingsForms.svelte";

// The backend's label when a service label setting is empty or missing.
export const defaultServiceLabel = "service_name";

// The backend validates each value, so a rejected save shows its error.
const grafanaSettingsSchema = z.object({
	logsDataSourceUid: z.string(),
	metricsDataSourceUid: z.string(),
	logServiceLabel: z.string(),
	metricServiceLabel: z.string(),
});

export type GrafanaSettings = z.infer<typeof grafanaSettingsSchema>;

// Missing and explicitly null settings are unset.
const readText = (value: unknown) => {
	if (typeof value === "string") {
		return value;
	}
	return "";
};

const readServiceLabel = (value: unknown) => {
	const label = readText(value);
	if (label === "") {
		return defaultServiceLabel;
	}
	return label;
};

export const grafanaSettings: IntegrationSettingsDefinition<GrafanaSettings> = {
	schema: grafanaSettingsSchema,

	read: (userSettings) => ({
		logsDataSourceUid: readText(userSettings.logs_data_source_uid),
		metricsDataSourceUid: readText(userSettings.metrics_data_source_uid),
		logServiceLabel: readServiceLabel(userSettings.log_service_label),
		metricServiceLabel: readServiceLabel(userSettings.metric_service_label),
	}),

	// An empty service label saves as empty, which the backend reads as the default.
	write: (values, userSettings): IntegrationUserSettings => ({
		...userSettings,
		logs_data_source_uid: values.logsDataSourceUid.trim(),
		metrics_data_source_uid: values.metricsDataSourceUid.trim(),
		log_service_label: values.logServiceLabel.trim(),
		metric_service_label: values.metricServiceLabel.trim(),
	}),
};

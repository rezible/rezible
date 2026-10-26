import type { AlertMetrics } from "$lib/api";

export type MetricRow = {
	key: string;
	label: string;
	value: string;
};

const percentFormat = new Intl.NumberFormat(undefined, { style: "percent", maximumFractionDigits: 0 });

/** A share of feedback responses; a dash when there was no feedback to compare against. */
function feedbackShare(count: number, feedbacks: number) {
	if (feedbacks <= 0) {
		return "—";
	}
	return percentFormat.format(count / feedbacks);
}

export function signalMetricRows(metrics: AlertMetrics): MetricRow[] {
	const notAccurate = metrics.feedbacks - metrics.accurate - metrics.accurateUnknown;

	return [
		{ key: "triggers", label: "Trigger events", value: String(metrics.triggers) },
		{ key: "interrupts", label: "Interrupts", value: String(metrics.interrupts) },
		{ key: "night-interrupts", label: "Night interrupts", value: String(metrics.nightInterrupts) },
		{ key: "feedbacks", label: "Feedback given", value: String(metrics.feedbacks) },
		{
			key: "actionable",
			label: "Actionable",
			value: feedbackShare(metrics.actionable, metrics.feedbacks),
		},
		{
			key: "accurate",
			label: "Accurate (yes / no / unknown)",
			value: `${metrics.accurate} / ${notAccurate} / ${metrics.accurateUnknown}`,
		},
		{
			key: "docs-available",
			label: "Documentation available",
			value: feedbackShare(metrics.docsAvailable, metrics.feedbacks),
		},
	];
}

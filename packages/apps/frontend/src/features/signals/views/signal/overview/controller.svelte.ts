import { Context } from "runed";
import { createQuery } from "@tanstack/svelte-query";
import { getLocalTimeZone, today } from "@internationalized/date";
import { getAlertMetricsOptions } from "$lib/api";
import { signalMetricRows } from "$features/signals/lib/metrics";
import { useAlertViewController } from "../controller.svelte";

const metricsPeriodDays = 7;

export class AlertOverviewController {
	private view = useAlertViewController();

	definition = $derived(this.view.alert?.attributes);
	title = $derived(this.definition?.title ?? "");
	description = $derived(this.definition?.description ?? "");
	definitionText = $derived(this.definition?.definition ?? "");
	rosterName = $derived(this.definition?.roster?.attributes?.name);
	rosterId = $derived(this.definition?.roster?.id);

	metricsPeriodLabel = `Last ${metricsPeriodDays} days`;

	metricsQuery = createQuery(() => {
		const to = today(getLocalTimeZone());
		const from = to.subtract({ days: metricsPeriodDays });
		return getAlertMetricsOptions({
			path: { id: this.view.alertId },
			query: { from: from.toString(), to: to.toString() },
		});
	});

	metricRows = $derived.by(() => {
		const metrics = this.metricsQuery.data?.data;
		if (!metrics) {
			return [];
		}
		return signalMetricRows(metrics);
	});
}

const ctx = new Context<AlertOverviewController>("AlertOverviewController");
export const initAlertOverviewController = () => ctx.set(new AlertOverviewController());

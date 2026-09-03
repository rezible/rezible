import { useAlertViewController } from "$features/alerts/views/alert";
import { listAlertIncidentLinksOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { getLocalTimeZone, now } from "@internationalized/date";

const defaultDateRange = () => {
	return {
		from: now(getLocalTimeZone()).subtract({ days: 7 }).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: "day",
	};
};

export class AlertIncidentsViewController {
	view = useAlertViewController();

	rosterId = $state<string>();
	dateRange = $state(defaultDateRange());

	query = createQuery(() => listAlertIncidentLinksOptions({ path: { id: this.view.alertId } }));
}

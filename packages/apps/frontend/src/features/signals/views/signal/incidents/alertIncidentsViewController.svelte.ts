import { useAlertViewController } from "$features/signals/views/signal";
import { listAlertIncidentLinksOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";

export class AlertIncidentsViewController {
	view = useAlertViewController();

	query = createQuery(() => listAlertIncidentLinksOptions({ path: { id: this.view.alertId } }));
}

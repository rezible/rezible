import { WebSocketStatus } from "@hocuspocus/provider";
import { Context } from "runed";
import { useIncidentCollaboration } from "../collaboration.svelte";
import { useIncidentView } from "../controller.svelte";

type Drawer = "add-component";

export class IncidentSidebarController {
	private view = useIncidentView();
	private collab = useIncidentCollaboration();

	drawer = $state<Drawer>();
	drawerOpen = $derived(!!this.drawer);

	ctxColor = $derived.by(() => {
		if (this.collab.error) return "fill-danger";
		switch (this.collab.status) {
			case WebSocketStatus.Connecting:
				return "fill-default";
			case WebSocketStatus.Connected:
				return "fill-green-500";
			case WebSocketStatus.Disconnected:
				return "fill-warning";
		}
	});
	connectionError = $derived(this.collab.error);

	incidentId = $derived(this.view.incidentId);

}

const ctx = new Context<IncidentSidebarController>("IncidentSidebarController");
export const initIncidentSidebarController = () => ctx.set(new IncidentSidebarController());
export const useIncidentSidebarController = () => ctx.get();

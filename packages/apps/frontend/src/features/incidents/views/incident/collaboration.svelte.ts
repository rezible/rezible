import {
	HocuspocusProvider,
	WebSocketStatus,
	type StatesArray,
} from "@hocuspocus/provider";
import { onMount } from "svelte";
import { Context, watch } from "runed";
import type { IncidentViewController } from "./controller.svelte";

export class IncidentCollaborationController {
	incidentView: IncidentViewController;

	provider = $state<HocuspocusProvider>();
	awareness = $state<StatesArray>([]);
	status = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state<Error>();

	constructor(incidentView: IncidentViewController) {
		this.incidentView = incidentView;
		watch(() => incidentView.documentId, id => { this.connect(id) });
		onMount(() => (() => { this.cleanup() }));
	};

	private cleanup() {
		// https://github.com/ueberdosis/hocuspocus/issues/845
		try {
			if (this.provider?.isSynced) this.provider.disconnect();
			this.provider?.destroy();
			this.provider = undefined;
		} catch (e) {
			console.error("failed to disconnect collaboration provider ", e);
		}
		this.awareness = [];
		this.status = WebSocketStatus.Disconnected;
		this.error = undefined;
	};

	async connect(id?: string) {
		this.cleanup();
		if (!id) return;

		this.provider = new HocuspocusProvider({
			url: "/api/documents",
			token: "foobar",
			name: id,
			onAwarenessChange: ({states}) => {
				this.awareness = states;
			},
			onStatus: ({status}) => {
				this.status = status;
			},
			onAuthenticated: () => {
				console.log("authed")
				this.error = undefined;
			},
			onAuthenticationFailed: ({reason}) => {
				console.log("auth failed", reason);
				this.error = new Error(reason);
			},
		});
	};
}

const ctx = new Context<IncidentCollaborationController>("IncidentCollaborationController");
export const initIncidentCollaborationController = (vc: IncidentViewController) => ctx.set(new IncidentCollaborationController(vc));
export const useIncidentCollaboration = () => ctx.get();
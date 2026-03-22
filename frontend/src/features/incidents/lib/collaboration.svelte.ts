import {
	HocuspocusProvider,
	WebSocketStatus,
	type HocuspocusProviderConfiguration,
	type onAuthenticationFailedParameters,
	type onAwarenessChangeParameters,
	type onStatusParameters,
	type StatesArray,
} from "@hocuspocus/provider";
import { onMount } from "svelte";
import { Context, watch } from "runed";
import { useIncidentView } from "$features/incidents/views/incident";

export class IncidentCollaborationController {
	incidentView = useIncidentView();
	documentId = $derived(this.incidentView.documentId);
	provider = $state<HocuspocusProvider>();
	awareness = $state<StatesArray>([]);
	status = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state<Error>();

	constructor() {
		watch(() => this.documentId, id => { this.connect(id) });
		onMount(() => (() => { this.cleanup() }));
	};

	cleanup() {
		// https://github.com/ueberdosis/hocuspocus/issues/845
		if (this.provider) {
			try {
				if (this.provider.isSynced) this.provider.disconnect();
				this.provider.destroy();
				this.provider = undefined;
			} catch (e) {
				console.error("failed to disconnect collaboration provider ", e);
			}
		}
		this.awareness = [];
		this.status = WebSocketStatus.Disconnected;
		this.error = undefined;
	};

	private onAwarenessChange({ states }: onAwarenessChangeParameters) {
		this.awareness = states;
	}

	private onConnectionStatusChange({ status }: onStatusParameters) {
		console.log("status", status);
		this.status = status;
	}

	private onAuthenticated() {
		this.error = undefined;
	}

	private onAuthenticationFailed({ reason }: onAuthenticationFailedParameters) {
		this.error = new Error(reason);
	}

	async connect(id?: string) {
		this.cleanup();
		console.log("connect", id);
		if (!id) return;

		this.provider = new HocuspocusProvider({
			url: "/ws",
			token: "",
			name: id,
			onAwarenessChange: (e) => this.onAwarenessChange(e),
			onStatus: (e) => this.onConnectionStatusChange(e),
			onAuthenticated: () => this.onAuthenticated(),
			onAuthenticationFailed: e => this.onAuthenticationFailed(e),
		});
	};
}

const ctx = new Context<IncidentCollaborationController>("IncidentCollaborationController");
export const initIncidentCollaborationController = () => ctx.set(new IncidentCollaborationController());
export const useIncidentCollaboration = () => ctx.get();
import {
	HocuspocusProvider,
	WebSocketStatus,
	type HocuspocusProviderConfiguration,
	type onAuthenticationFailedParameters,
	type onAwarenessChangeParameters,
	type onStatusParameters,
	type StatesArray,
} from "@hocuspocus/provider";
import { requestDocumentEditorSession } from "$lib/api/oapi.gen";
import { onMount } from "svelte";
import { Context, watch } from "runed";

export class IncidentCollaborationState {
	documentName = $state<string>();
	provider = $state<HocuspocusProvider>();
	awareness = $state<StatesArray>([]);
	connectionStatus = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state<Error>();

	constructor(retroIdFn: () => (string | undefined)) {
		watch(retroIdFn, retroId => { this.connect(retroId) });
		onMount(() => (() => { this.cleanup() }));
	};

	cleanup() {
		// https://github.com/ueberdosis/hocuspocus/issues/845
		// if (collab.provider && collab.provider.isConnected) {}
		try {
			this.documentName = undefined;
			if (this.provider) {
				if (this.provider.isConnected) this.provider.disconnect();
				this.provider.destroy();
				this.provider = undefined;
			}
			this.awareness = [];
			this.connectionStatus = WebSocketStatus.Disconnected;
			this.error = undefined;
		} catch (e) {
			console.error("failed to disconnect collaboration provider ", e);
		}
	};

	private onAwarenessChange({ states }: onAwarenessChangeParameters) {
		this.awareness = states;
	}

	private onConnectionStatusChange({ status }: onStatusParameters) {
		this.connectionStatus = status;
	}

	private onAuthenticated() {
		this.error = undefined;
	}

	private onAuthenticationFailed({ reason }: onAuthenticationFailedParameters) {
		this.error = new Error(reason);
	}

	async connect(retrospectiveId?: string) {
		if (this.documentName === retrospectiveId) return;

		this.cleanup();

		if (!retrospectiveId) return;
		this.documentName = retrospectiveId;

		const { data: body, error: reqErr } = await requestDocumentEditorSession({
			body: { attributes: { documentName: retrospectiveId } },
			throwOnError: false,
		});

		if (reqErr) {
			console.error("connection error", reqErr);
			this.error = new Error("failed to connect");
			return;
		}

		const sess = body.data;

		const config: HocuspocusProviderConfiguration = {
			url: sess.connectionUrl,
			token: sess.token,
			name: sess.documentName,
			preserveConnection: false,
			onAwarenessChange: (e) => this.onAwarenessChange(e),
			onStatus: (e) => this.onConnectionStatusChange(e),
			onAuthenticated: () => this.onAuthenticated(),
			onAuthenticationFailed: e => this.onAuthenticationFailed(e),
		};
		this.provider = new HocuspocusProvider(config);
	};
}

const collabCtx = new Context<IncidentCollaborationState>("incidentCollaboration");
export const setIncidentCollaboration = (s: IncidentCollaborationState) => collabCtx.set(s);
export const useIncidentCollaboration = () => collabCtx.get();

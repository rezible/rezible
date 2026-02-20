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
import { watch } from "runed";
import type { Getter } from "$src/lib/utils.svelte";

export class RetrospectiveCollaborationState {
	documentId = $state<string>();
	provider = $state<HocuspocusProvider>();
	awareness = $state<StatesArray>([]);
	connectionStatus = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state<Error>();

	constructor(idFn: Getter<string | undefined>) {
		watch(idFn, id => { this.connect(id) });
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
		this.documentId = undefined;
		this.awareness = [];
		this.connectionStatus = WebSocketStatus.Disconnected;
		this.error = undefined;
	};

	private onAwarenessChange({ states }: onAwarenessChangeParameters) {
		this.awareness = states;
	}

	private onConnectionStatusChange({ status }: onStatusParameters) {
		console.log("status", status);
		this.connectionStatus = status;
	}

	private onAuthenticated() {
		this.error = undefined;
	}

	private onAuthenticationFailed({ reason }: onAuthenticationFailedParameters) {
		this.error = new Error(reason);
	}

	async connect(id?: string) {
		if (this.documentId === id) return;

		this.cleanup();

		if (!id) return;
		this.documentId = id;

		const { data: body, error: reqErr } = await requestDocumentEditorSession({
			path: { id },
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
			token: sess.sessionToken,
			name: sess.documentId,
			onAwarenessChange: (e) => this.onAwarenessChange(e),
			onStatus: (e) => this.onConnectionStatusChange(e),
			onAuthenticated: () => this.onAuthenticated(),
			onAuthenticationFailed: e => this.onAuthenticationFailed(e),
		};
		this.provider = new HocuspocusProvider(config);
	};
}


import { HocuspocusProvider, WebSocketStatus, type StatesArray } from "@hocuspocus/provider";
import type { DocumentSessionAuth } from "@rezible/api-client-ts";
import { requestDocumentSessionAuthMutation } from "@rezible/api-client-ts/svelte-query";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { onMount } from "svelte";

const shouldRetryStatus = (status?: number) => {
	return !(status === undefined || status === 408 || status === 429 || (status >= 500 && status < 600))
}

export class IncidentCollaborationController {
	private documentId = $state.raw<string>();
	private token = $state.raw<string>();

	provider = $state.raw<HocuspocusProvider>();
	awareness = $state.raw<StatesArray>([]);
	status = $state.raw<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state.raw<Error>();

	initialSynced = $state(false);
	unsyncedChanges = $state(0);

	constructor(idFn: Getter<string | undefined>) {
		watch(idFn, (id) => {
			this.connect(id);
		});
		onMount(() => {
			return () => {
				this.cleanup();
			};
		});
	}

	private getToken = async () => {
		
	}

	private createProvider({serverUrl: url, name, token}: DocumentSessionAuth) {
		this.provider = new HocuspocusProvider({
			url,
			name,
			token: async () => {
				let t = token;
				if (!!this.token) {
					const { data } = await this.requestSessionAuthMut.mutateAsync({
						path: {id: name}
					});
					t = data.token;
				}
				this.token = t;
				return t;
			},
			onAwarenessChange: ({ states }) => {
				console.log("awareness", states);
				this.awareness = states;
			},
			onStatus: ({ status }) => {
				this.status = status;
			},
			onAuthenticated: () => {
				this.error = undefined;
			},
			onAuthenticationFailed: ({ reason }) => {
				console.log("auth failed", reason);
				if (this.documentId !== name) return;
				if (!this.error) this.error = new Error(reason);
				this.provider?.disconnect();
			},
			onSynced: () => {
				this.initialSynced = true;
			},
			onUnsyncedChanges: ({ number }) => {
				this.unsyncedChanges = number;
			},
		});
	}

	private requestSessionAuthMut = createMutation(() => ({
		...requestDocumentSessionAuthMutation(),
		retryDelay: 250,
		retry: (failureCount, error) => {
			return failureCount < 2 && shouldRetryStatus(error.status);
		},
		onError: (error, variables) => {
			if (variables.path.id !== this.documentId) return;
			const prefix = error.status ? `(HTTP ${error.status}) ` : "";
			this.error = new Error(prefix + (error.detail ?? "Error"));
		},
	}));

	async connect(id?: string) {
		if (!id) {
			this.cleanup();
			return;
		}
		if (id === this.documentId && this.provider) return;
		if (id !== this.documentId) this.cleanup();
		this.documentId = id;
		this.error = undefined;
		try {
			const { data: auth } = await this.requestSessionAuthMut.mutateAsync({ path: { id } });
			if (this.documentId === id && !this.provider) this.createProvider(auth);
		} catch {
			// The mutation error handler exposes the failure to the report.
		}
	}

	retry = () => {
		this.error = undefined;
		if (this.provider) {
			void this.provider.connect();
			return;
		}
		void this.connect(this.documentId);
	};

	cleanup() {
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
		this.initialSynced = false;
		this.unsyncedChanges = 0;
		this.documentId = "";
	}
}

const ctx = new Context<IncidentCollaborationController>("IncidentCollaborationController");
export const initIncidentCollaborationController = (docIdFn: Getter<string | undefined>) =>
	ctx.set(new IncidentCollaborationController(docIdFn));
export const useIncidentCollaboration = () => ctx.get();

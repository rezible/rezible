import { HocuspocusProvider, WebSocketStatus, type StatesArray } from "@hocuspocus/provider";
import type { DocumentSessionAuth } from "@rezible/api-client-ts";
import { requestDocumentSessionAuthMutation } from "@rezible/api-client-ts/svelte-query";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { onMount } from "svelte";

export class IncidentCollaborationController {
	private documentId = $state.raw<string>();

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

	private createProvider({ serverUrl, token, name }: DocumentSessionAuth, documentId: string) {
		if (this.provider && this.documentId === documentId) {
			this.provider.configuration.token = token;
			this.provider.configuration.name = name;
			this.provider.disconnect();
			this.provider.connect();
			return;
		}
		this.documentId = documentId;
		this.provider = new HocuspocusProvider({
			url: serverUrl,
			token: token,
			name: name,
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
				this.error = new Error(reason);
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
		onSuccess: ({ data: auth }, variables) => {
			if (variables.path.id === this.documentId) this.createProvider(auth, variables.path.id);
		},
		onError: (error, variables) => {
			if (variables.path.id === this.documentId)
				this.error =
					error instanceof Error ? error : new Error("Unable to authorize report connection");
		},
	}));

	async connect(id?: string) {
		if (!id) {
			this.cleanup();
			return;
		}
		if (
			id === this.documentId &&
			this.provider &&
			!this.error &&
			this.status !== WebSocketStatus.Disconnected
		)
			return;
		if (id !== this.documentId) this.cleanup();
		this.documentId = id;
		this.requestSessionAuthMut.mutate({ path: { id } });
	}

	retry = () => this.connect(this.documentId);

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

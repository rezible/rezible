import { HocuspocusProvider, WebSocketStatus, type StatesArray } from "@hocuspocus/provider";
import { requestDocumentSessionAuthMutation, type DocumentSessionAuth } from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { onMount } from "svelte";
import { Doc } from "yjs";

export class IncidentCollaborationController {
	private documentId?: string;
	private document?: Doc;
	private generation = 0;
	private connecting = false;
	private initialToken?: string;

	provider = $state.raw<HocuspocusProvider>();
	awareness = $state.raw<StatesArray>([]);
	status = $state.raw<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state.raw<Error>();
	canEdit = $state(false);
	initialSynced = $state(false);
	unsyncedChanges = $state(0);

	constructor(idFn: Getter<string | undefined>) {
		watch(idFn, (id) => void this.connect(id));
		onMount(() => () => this.cleanup());
	}

	private requestSessionAuth = createMutation(() => ({
		...requestDocumentSessionAuthMutation(),
		retryDelay: 250,
		retry: (failureCount, error) => {
			const status = error.status;
			return (
				failureCount < 2 &&
				(status === undefined ||
					status === 0 ||
					status === 408 ||
					status === 429 ||
					(status >= 500 && status < 600))
			);
		},
	}));

	private createProvider(auth: DocumentSessionAuth, generation: number) {
		this.initialToken = auth.token;
		const current = () => this.generation === generation;
		this.provider = new HocuspocusProvider({
			url: auth.serverUrl,
			name: auth.name,
			document: this.document,
			token: async () => {
				if (!current()) return "";
				if (this.initialToken) {
					const token = this.initialToken;
					this.initialToken = undefined;
					return token;
				}
				try {
					const { data } = await this.requestSessionAuth.mutateAsync({ path: { id: auth.name } });
					return current() ? data.token : "";
				} catch (error) {
					if (current()) {
						this.setError(error);
						this.provider?.configuration.websocketProvider.disconnect();
					}
					return "";
				}
			},
			onAwarenessChange: ({ states }) => {
				if (current()) this.awareness = states;
			},
			onStatus: ({ status }) => {
				if (current()) this.status = status;
			},
			onAuthenticated: ({ scope }) => {
				if (!current()) return;
				this.error = undefined;
				this.canEdit = scope === "read-write";
			},
			onAuthenticationFailed: ({ reason }) => {
				if (!current()) return;
				this.setError(new Error(reason));
				this.provider?.configuration.websocketProvider.disconnect();
			},
			onSynced: ({ state }) => {
				if (current()) this.initialSynced = state;
			},
			onUnsyncedChanges: ({ number }) => {
				if (current()) this.unsyncedChanges = number;
			},
		});
	}

	private setError(error: unknown) {
		this.canEdit = false;
		if (error instanceof Error) {
			this.error = error;
		} else if (
			error &&
			typeof error === "object" &&
			"detail" in error &&
			typeof error.detail === "string"
		) {
			this.error = new Error(error.detail);
		} else {
			this.error = new Error("Unable to connect to the report.");
		}
	}

	async connect(id?: string) {
		if (id !== this.documentId) this.cleanup();
		if (!id || this.provider || this.connecting) return;
		this.documentId = id;
		this.document ??= new Doc();
		const generation = this.generation;
		this.connecting = true;
		this.error = undefined;
		try {
			const { data } = await this.requestSessionAuth.mutateAsync({ path: { id } });
			if (generation === this.generation) this.createProvider(data, generation);
		} catch (error) {
			if (generation === this.generation) this.setError(error);
		} finally {
			if (generation === this.generation) this.connecting = false;
		}
	}

	retry = () => {
		if (this.connecting) return;
		this.initialToken = undefined;
		this.error = undefined;
		if (this.provider) {
			// Reuse the document and provider; the token callback obtains fresh credentials.
			const socket = this.provider.configuration.websocketProvider;
			socket.disconnect();
			const generation = this.generation;
			void socket.connect().catch((error: unknown) => {
				if (generation === this.generation) this.setError(error);
			});
		} else {
			void this.connect(this.documentId);
		}
	};

	cleanup() {
		this.generation += 1;
		const provider = this.provider;
		this.provider = undefined;
		provider?.configuration.websocketProvider.disconnect();
		provider?.destroy();
		this.document?.destroy();
		this.document = undefined;
		this.documentId = undefined;
		this.initialToken = undefined;
		this.connecting = false;
		this.awareness = [];
		this.status = WebSocketStatus.Disconnected;
		this.error = undefined;
		this.canEdit = false;
		this.initialSynced = false;
		this.unsyncedChanges = 0;
	}
}

const ctx = new Context<IncidentCollaborationController>("IncidentCollaborationController");
export const initIncidentCollaborationController = (docIdFn: Getter<string | undefined>) =>
	ctx.set(new IncidentCollaborationController(docIdFn));
export const useIncidentCollaboration = () => ctx.get();

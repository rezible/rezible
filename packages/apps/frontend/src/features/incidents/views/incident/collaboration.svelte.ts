import {
	HocuspocusProvider,
	WebSocketStatus,
	type StatesArray,
} from "@hocuspocus/provider";
import type { DocumentSessionAuth } from "@rezible/api-client-ts";
import { requestDocumentSessionAuthMutation } from "@rezible/api-client-ts/svelte-query";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { onMount } from "svelte";

export class IncidentCollaborationController {
	provider = $state<HocuspocusProvider>();
	awareness = $state<StatesArray>([]);
	status = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state<Error>();

	constructor(docIdFn: Getter<string | undefined>) {
		watch(docIdFn, documentId => {
			this.connect(documentId);
		});
		onMount(() => {
			return () => {
				this.cleanup();
			}
		});
	}

	private createProvider({serverUrl, token, name}: DocumentSessionAuth) {
		console.log("creating collaboration provider");
		this.provider = new HocuspocusProvider({
			url: serverUrl,
			token: token,
			name: name,
			onAwarenessChange: ({states}) => {
				this.awareness = states;
			},
			onStatus: ({status}) => {
				this.status = status;
			},
			onAuthenticated: () => {
				console.log("authed");
				this.error = undefined;
			},
			onAuthenticationFailed: ({reason}) => {
				console.log("auth failed", reason);
				this.error = new Error(reason);
			},
		});
	}

	private requestSessionAuthMut = createMutation(() => ({
		...requestDocumentSessionAuthMutation(),
		onSuccess: ({data: auth}) => {
			this.createProvider(auth);
		}
	}));

	async connect(id?: string) {
		this.cleanup();
		if (!!id && id !== this.requestSessionAuthMut.variables?.path.id) {
			this.requestSessionAuthMut.mutate({path: {id}});
		}
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
	};
}

const ctx = new Context<IncidentCollaborationController>("IncidentCollaborationController");
export const initIncidentCollaborationController = (docIdFn: Getter<string | undefined>) => ctx.set(new IncidentCollaborationController(docIdFn));
export const useIncidentCollaboration = () => ctx.get();
import {
	HocuspocusProvider,
	WebSocketStatus,
	type StatesArray,
} from "@hocuspocus/provider";
import type { DocumentSession } from "@rezible/api-client-ts";
import { requestDocumentEditorSessionMutation } from "@rezible/api-client-ts/svelte-query";
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

	private createProvider(sess: DocumentSession) {
		console.log("creating collaboration provider", sess);
		this.provider = new HocuspocusProvider({
			url: sess.serverUrl,
			token: sess.token,
			name: sess.name,
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

	private requestSessionMut = createMutation(() => ({
		...requestDocumentEditorSessionMutation(),
		onSuccess: ({data: sess}) => {
			this.createProvider(sess);
		}
	}));

	async connect(id?: string) {
		this.cleanup();
		if (!!id && id !== this.requestSessionMut.variables?.path.id) {
			this.requestSessionMut.mutate({path: {id}});
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
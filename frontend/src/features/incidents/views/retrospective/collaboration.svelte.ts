import * as Y from 'yjs';
import { HocuspocusProvider, WebSocketStatus, type StatesArray } from '@hocuspocus/provider';
import { requestDocumentEditorSession, type GetRetrospectiveForIncidentResponseBody, type Retrospective } from '$src/lib/api/oapi.gen';
import { QueryObserver, useQueryClient, type CreateQueryResult } from '@tanstack/svelte-query';
import { onMount } from 'svelte';

export type CollaborationState = {
	documentName?: string;
	provider: HocuspocusProvider | null;
	awareness: StatesArray;
	status: WebSocketStatus;
	error?: Error;
};

const createCollaborationState = () => {
	const emptyState: CollaborationState = {
		documentName: undefined,
		provider: null,
		awareness: [],
		status: WebSocketStatus.Disconnected,
		error: undefined,
	};

	let collab = $state<CollaborationState>(emptyState);

	const connect = async (documentName: string) => {
		if (collab.documentName === documentName) return;
		collab.documentName = documentName;
		
		if (collab.provider) collab.provider.destroy();

		const attributes = {documentName: $state.snapshot(documentName)};
		
		const res = await requestDocumentEditorSession({
			body: {attributes},
			throwOnError: false,
		});

		if (res.error) {
			console.error("connection error", res.error);
			collab.error = new Error("failed to connect");
			return;
		}

		const sess = res.data.data;
		collab.provider = new HocuspocusProvider({
			document: new Y.Doc(),
			url: sess.connectionUrl,
			token: sess.token,
			name: sess.documentName,
			connect: true,
			preserveConnection: false,
			onAwarenessChange: ({ states }) => {
				collab.awareness = states;
			},
			onStatus({ status }) {
				collab.status = status;
			},
			onAuthenticated: () => {
				collab.error = undefined;
			},
			onDestroy() {
				// console.log("destroying provider");
			},
			onAuthenticationFailed: ({ reason }) => {
				collab.error = new Error(reason);
			}
		});
	}

	const cleanup = () => {
		console.log("cleanup");
		collab.provider?.destroy();
		collab = emptyState;
	};

	return {
		get awareness() { return collab.awareness },
		get provider() { return collab.provider },
		get status() { return collab.status },
		get error() { return collab.error },
		connect,
		cleanup,
	};
};
export const collaborationState = createCollaborationState();

export const mountCollaboration = (query: CreateQueryResult<GetRetrospectiveForIncidentResponseBody, Error>) => {
	const documentName = $derived(query.data?.data.attributes.documentName);
	$effect(() => {
		if (documentName) collaborationState.connect(documentName);
	});
	onMount(() => {
		return () => {collaborationState.cleanup();}
	});
}
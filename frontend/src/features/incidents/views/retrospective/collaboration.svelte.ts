import * as Y from 'yjs';
import { HocuspocusProvider, WebSocketStatus, type HocuspocusProviderConfiguration, type StatesArray } from '@hocuspocus/provider';
import { requestDocumentEditorSession } from '$lib/api/oapi.gen';

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
		
		const res = await requestDocumentEditorSession({
			body: {attributes: {documentName}},
			throwOnError: false,
		});

		if (res.error) {
			console.error("connection error", res.error);
			collab.error = new Error("failed to connect");
			return;
		}

		const sess = res.data.data;
		const config: HocuspocusProviderConfiguration = {
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
			onAuthenticationFailed: ({ reason }) => {
				collab.error = new Error(reason);
			}
		};

		if (collab.provider) {
			// collab.provider.disconnect();
			collab.provider.setConfiguration(config);
			// collab.provider.connect();
			collab.provider.forceSync();
		} else {
			collab.provider = new HocuspocusProvider(config);
		}
	}

	const cleanup = () => {
		collab.provider?.disconnect();
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

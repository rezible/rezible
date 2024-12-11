import * as Y from 'yjs';
import { HocuspocusProvider, WebSocketStatus, type StatesArray } from '@hocuspocus/provider';
import { requestDocumentEditorSession } from '$src/lib/api/oapi.gen';

export type CollaborationState = {
	provider: HocuspocusProvider | null;
	awareness: StatesArray;
	status: WebSocketStatus;
	error?: Error;
};

const createCollaborationState = () => {
	const emptyState = {
		provider: null,
		awareness: [],
		status: WebSocketStatus.Disconnected,
		error: undefined,
	};

	let collab = $state<CollaborationState>(emptyState);

	const connect = async (documentName: string) => {
		const attributes = {documentName};
		
		const res = await requestDocumentEditorSession({
			body: {attributes},
			throwOnError: false,
		});

		if (res.error) {
			console.log("connection error", res.error);
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

	const disconnect = () => {
		if (collab.provider) collab.provider.destroy();
		collab = emptyState;
	}

	return {
		get awareness() { return collab.awareness },
		get provider() { return collab.provider },
		get status() { return collab.status },
		get error() { return collab.error },
		connect,
		disconnect,
	};
};
export const collaborationState = createCollaborationState();

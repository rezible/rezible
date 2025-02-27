import {
	HocuspocusProvider,
	WebSocketStatus,
	type HocuspocusProviderConfiguration,
	type StatesArray,
} from "@hocuspocus/provider";
import { requestDocumentEditorSession, type DocumentEditorSession } from "$lib/api/oapi.gen";
import { onMount } from "svelte";
import { watch } from "runed";

const createCollaborationState = () => {
	let documentName = $state<string>();
	let provider = $state<HocuspocusProvider>();
	let awareness = $state<StatesArray>([]);
	let connectionStatus = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	let error = $state<Error>();

	const cleanup = () => {
		// https://github.com/ueberdosis/hocuspocus/issues/845
		// if (collab.provider && collab.provider.isConnected) {}
		try {
			documentName = undefined;
			if (provider?.isConnected) provider?.disconnect();
			provider?.destroy();
			provider = undefined;
			awareness = [];
			connectionStatus = WebSocketStatus.Disconnected;
			error = undefined;
		} catch (e) {
			console.error("failed to disconnect collaboration provider ", e);
		}
	};

	const createSessionProvider = (sess: DocumentEditorSession): HocuspocusProvider => {
		const config: HocuspocusProviderConfiguration = {
			url: sess.connectionUrl,
			token: sess.token,
			name: sess.documentName,
			connect: true,
			preserveConnection: false,
			onAwarenessChange: ({ states }) => {
				awareness = states;
			},
			onStatus({ status }) {
				console.log("status", status, sess.documentName)
				connectionStatus = status;
			},
			onAuthenticated: () => {
				error = undefined;
			},
			onAuthenticationFailed: ({ reason }) => {
				error = new Error(reason);
			},
		};
		return new HocuspocusProvider(config);
	}

	const connect = async (retrospectiveId: string) => {
		if (documentName === retrospectiveId) return;
		documentName = retrospectiveId;

		cleanup();
		/*
		if (provider) {
			collab.provider.setConfiguration(config);
			collab.provider.forceSync();
		}
		*/

		const { data: body, error: reqErr } = await requestDocumentEditorSession({
			body: { attributes: { documentName: retrospectiveId } },
			throwOnError: false,
		});

		if (reqErr) {
			console.error("connection error", reqErr);
			error = new Error("failed to connect");
			return;
		}

		provider = createSessionProvider(body.data);
	};

	const setup = (retroId: string) => {
		// const retrospectiveId = retrospectiveCtx.get().id;
		watch(() => retroId, (id) => { connect(id) });
		onMount(() => (() => { cleanup() }));
	};

	return {
		setup,
		get awareness() {
			return awareness;
		},
		get provider() {
			return provider;
		},
		get connectionStatus() {
			return connectionStatus;
		},
		get error() {
			return error;
		},
	};
};
export const collaboration = createCollaborationState();

import * as Y from "yjs";
import {
  HocuspocusProvider,
  WebSocketStatus,
  type HocuspocusProviderConfiguration,
  type StatesArray,
} from "@hocuspocus/provider";
import { requestDocumentEditorSession } from "$lib/api/oapi.gen";
import { onMount } from "svelte";
import { watch } from "runed";
import { retrospectiveCtx } from "./context";

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

  const connect = async (retrospectiveId: string) => {
    if (collab.documentName === retrospectiveId) return;
    collab.documentName = retrospectiveId;

    const res = await requestDocumentEditorSession({
      body: { attributes: { documentName: retrospectiveId } },
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
      },
    };

    if (collab.provider) {
      collab.provider.destroy();
      // collab.provider.setConfiguration(config);
      // collab.provider.forceSync();
    }
    collab.provider = new HocuspocusProvider(config);
  };

  const cleanup = () => {
    if (collab.provider && collab.provider.isConnected) {
      // https://github.com/ueberdosis/hocuspocus/issues/845
      try {
        collab.provider?.destroy();
      } catch (e) {
        console.error("failed to disconnect collaboration provider ", e);
      }
    }
    collab = emptyState;
  };

  const componentMount = () => {
	const retrospectiveId = retrospectiveCtx.get().id;
	console.log(retrospectiveId);
	watch(() => retrospectiveId, id => {connect(id)});
	onMount(() => {return () => cleanup()});
  }

  return {
	componentMount,
    get awareness() {
      return collab.awareness;
    },
    get provider() {
      return collab.provider;
    },
    get status() {
      return collab.status;
    },
    get error() {
      return collab.error;
    },
  };
};
export const collaborationState = createCollaborationState();

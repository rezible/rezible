import {
	HocuspocusProvider,
	WebSocketStatus,
	type StatesArray,
} from "@hocuspocus/provider";
import { Context } from "runed";

export class IncidentCollaborationController {
	provider = $state<HocuspocusProvider>();
	awareness = $state<StatesArray>([]);
	status = $state<WebSocketStatus>(WebSocketStatus.Disconnected);
	error = $state<Error>();

	async connect(id?: string) {
		this.cleanup();
		if (!id) return;

		const documentsUrl = new URL(window.location.href);
		documentsUrl.pathname = "/api/documents";

		console.log("documents", documentsUrl);
		this.provider = new HocuspocusProvider({
			url: documentsUrl.toString(),
			token: "foobar",
			name: id,
			onAwarenessChange: ({states}) => {
				this.awareness = states;
			},
			onStatus: ({status}) => {
				this.status = status;
			},
			onAuthenticated: () => {
				console.log("authed")
				this.error = undefined;
			},
			onAuthenticationFailed: ({reason}) => {
				console.log("auth failed", reason);
				this.error = new Error(reason);
			},
		});
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
export const initIncidentCollaborationController = () => ctx.set(new IncidentCollaborationController());
export const useIncidentCollaboration = () => ctx.get();
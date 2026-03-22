import type {
	Extension,
	onAuthenticatePayload,
	onChangePayload,
	onLoadDocumentPayload,
} from "@hocuspocus/server";
import { Forbidden } from "@hocuspocus/common";
import * as Y from "yjs";

import { client } from "./lib/api/oapi.gen/client.gen";
import { Documents } from "./lib/api/oapi.gen";

type AuthContext = {
	
};

export class RezibleServerProxy implements Extension {
	extensionName = "Rezible Proxy";

	constructor(apiUrl: string) {
		if (!apiUrl) throw new Error("missing proxy config");

		client.setConfig({baseUrl: apiUrl});
	}

	getUserAuth(data: any): string {

		return "";
	}

	async onAuthenticate(payload: onAuthenticatePayload): Promise<AuthContext> {
		const token = payload.token;

		console.log(payload);
		
		payload.connectionConfig.readOnly = true;

		return {};
	}

    async onLoadDocument(data: onLoadDocumentPayload): Promise<any> {
		// const res = await Documents.loadDocument({
		// 	auth: this.getUserAuth(data),
		// 	path: { id: data.documentName },
		// });
        // if (res.error) throw new Error("Failed to load document");

		// const doc = res.data.data;
        // const state = JSON.parse(doc.attributes.content);
        // const update = new Uint8Array(state.data);
        // if (update) Y.applyUpdate(data.document, update);
	}

	async onStoreDocument(data: onChangePayload) {
        // const content = Buffer.from(Y.encodeStateAsUpdate(data.document));
		// const res = await Documents.updateDocument({
		// 	auth: this.getUserAuth(data),
		// 	path: { id: data.documentName },
		// 	body: { attributes: { content }}
		// })
		// if (res.error) console.log("failed to update document");
	}
}
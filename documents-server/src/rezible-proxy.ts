import type {
	Extension,
	onAuthenticatePayload,
	onChangePayload,
	onLoadDocumentPayload,
} from "@hocuspocus/server";
import { Forbidden } from "@hocuspocus/common";
import * as Y from "yjs";

import { client } from "./lib/api/oapi.gen/client.gen";
import { Documents } from "./lib/api/oapi.gen/sdk.gen";
import type { DocumentEditorSessionAuth } from "./lib/api/oapi.gen/types.gen";

export class RezibleServerProxy implements Extension {
	extensionName = "Rezible Proxy";

	constructor(apiUrl: string, apiSecret: string) {
		if (!apiUrl || !apiSecret) throw new Error("missing proxy config");

		client.setConfig({
			baseUrl: apiUrl,
			auth: apiSecret,
		});
	}

	async onAuthenticate(data: onAuthenticatePayload): Promise<DocumentEditorSessionAuth> {
        const documentId = data.documentName;
		const token = data.token;

		const res = await Documents.verifyDocumentSessionAuth({
			path: { id: documentId },
			body: { attributes: {token} }
		});
        if (res.error) throw new Error("Authentication Failed");
		
		const sessionAuth = res.data.data;
		data.connectionConfig.readOnly = sessionAuth.readOnly;

		return sessionAuth;
	}

    async onLoadDocument(data: onLoadDocumentPayload): Promise<any> {
        if (!data.context.token) throw Forbidden;


		const res = await Documents.loadDocument({
			path: { id: data.documentName },
		});
        if (res.error) throw new Error("Authentication Failed");

		const doc = res.data.data;
        const state = JSON.parse(doc.attributes.content);
        const update = new Uint8Array(state.data);
        if (update) Y.applyUpdate(data.document, update);
	}

	async onStoreDocument(data: onChangePayload) {
        if (!data.context.token) throw Forbidden;

        const state = Buffer.from(Y.encodeStateAsUpdate(data.document));

		const res = await Documents.updateDocument({
			path: { id: data.documentName },
			body: { attributes: { content: state }}
		})
		if (res.error) {
			console.log("failed to update document");
		}
	}
}
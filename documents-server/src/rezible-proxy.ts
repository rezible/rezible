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
import type { AuthToken } from "./lib/api/oapi.gen/core/auth.gen";

type AuthContext = {
	token: AuthToken;
	sessionAuth: DocumentEditorSessionAuth;
};

export class RezibleServerProxy implements Extension {
	extensionName = "Rezible Proxy";

	constructor(apiUrl: string, apiSecret: string) {
		if (!apiUrl || !apiSecret) throw new Error("missing proxy config");

		client.setConfig({
			baseUrl: apiUrl,
			auth: apiSecret,
		});
	}

	async onAuthenticate(data: onAuthenticatePayload): Promise<AuthContext> {
		const token = data.token as AuthToken;
		const res = await Documents.verifyDocumentSessionAuth({
			auth: token,
			path: { id: data.documentName },
		});
        if (res.error) throw new Error("Authentication Failed");
		
		const sessionAuth = res.data.data;
		data.connectionConfig.readOnly = sessionAuth.readOnly;

		return {token, sessionAuth};
	}

	getAuthContext(data: {context: AuthContext}): AuthContext {
		const ctx = data.context as AuthContext | undefined;
		console.log("load context", ctx);
        if (!ctx?.token) throw Forbidden;
		return ctx;
	}

    async onLoadDocument(data: onLoadDocumentPayload): Promise<any> {
		const ctx = this.getAuthContext(data);
		const res = await Documents.loadDocument({
			auth: ctx.token,
			path: { id: data.documentName },
		});
        if (res.error) throw new Error("Authentication Failed");

		const doc = res.data.data;
        const state = JSON.parse(doc.attributes.content);
        const update = new Uint8Array(state.data);
        if (update) Y.applyUpdate(data.document, update);
	}

	async onStoreDocument(data: onChangePayload) {
        const ctx = this.getAuthContext(data);
        const content = Buffer.from(Y.encodeStateAsUpdate(data.document));
		const res = await Documents.updateDocument({
			auth: ctx.token,
			path: { id: data.documentName },
			body: { attributes: { content }}
		})
		if (res.error) console.log("failed to update document");
	}
}
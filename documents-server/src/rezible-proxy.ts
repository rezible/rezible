import {
	type Extension,
	type onAuthenticatePayload,
	type onChangePayload,
	type onLoadDocumentPayload,
	type onUpgradePayload,
} from "@hocuspocus/server";
import { Forbidden } from "@hocuspocus/common";
import * as Y from "yjs";

import { client } from "./lib/api/oapi.gen/client.gen";
import { Documents } from "./lib/api/oapi.gen";
import type { IncomingMessage } from "http";
import { documentTransformer, extensions } from "./transformer";

const authCookieName = "rez_access_token";
type AuthContext = {
	token: string;
};

export class RezibleServerProxy implements Extension {
	extensionName = "Rezible Proxy";

	constructor(apiUrl: string) {
		if (!apiUrl) throw new Error("missing proxy config");
		console.log("api url", apiUrl);

		client.setConfig({baseUrl: apiUrl});
	}

	getUserAuth(payload: {request: IncomingMessage}): string {
		const authCookie = payload.request.headers.cookie
			?.split(';')
			.find(c => c.trim().startsWith(`${authCookieName}=`));
			
		const authToken = authCookie?.split('=')[1];
		
		if (!authToken) {
			throw new Error('missing or invalid auth token');
		}
		return authToken;
	}

	async onUpgrade(payload: onUpgradePayload): Promise<void> {
		
		return;
	}

	async onAuthenticate(payload: onAuthenticatePayload): Promise<AuthContext> {
		const token = this.getUserAuth(payload);
		
		payload.connectionConfig.readOnly = true;

		return {token};
	}

    async onLoadDocument(data: onLoadDocumentPayload): Promise<any> {
		const res = await Documents.loadDocument({
			auth: data.context.token,
			path: { id: data.documentName },
		});
        if (res.error) throw new Error("Failed to load document: " + res.error);

		const content = res.data.data.attributes.content || "[]";
        if (!!content) {
			const state = JSON.parse(content);
			const update = new Uint8Array(state.data);
			Y.applyUpdate(data.document, update);
			return data.document;
		}
		return documentTransformer.toYdoc("{}", "default", extensions);
	}

	async onStoreDocument(data: onChangePayload) {
        const content = Buffer.from(Y.encodeStateAsUpdate(data.document));
		const res = await Documents.updateDocument({
			auth: data.context.token,
			path: { id: data.documentName },
			body: { attributes: { content }}
		})
		if (res.error) console.log("failed to update document");
	}
}
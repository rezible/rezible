import type {
	Extension,
	onAuthenticatePayload,
	onChangePayload,
	onLoadDocumentPayload,
} from "@hocuspocus/server";
import { Forbidden } from "@hocuspocus/common";
import * as Y from "yjs";
import { createHmac } from "crypto";

type SessionUser = {
	id: string;
	username: string;
}

export type AuthContext = {
	user: SessionUser;
	token: string;
}

export class RezibleServerProxy implements Extension {
	extensionName: string;
    apiUrl: string;
    apiSecret: string;

	constructor(apiUrl: string, apiSecret: string) {
		this.extensionName = "Rezible Proxy";
		this.apiUrl = apiUrl;
		this.apiSecret = apiSecret;

		if (!apiUrl || !apiSecret) throw new Error("missing proxy config");
	}

    createRequestSignature(body: string): string {
		const hmac = createHmac("sha256", this.apiSecret);
		return `sha256=${hmac.update(body).digest("hex")}`;
	}

    async apiRequest(endpoint: string, token: string, reqBody: any) {
		const body = JSON.stringify(reqBody);
		return fetch(`${this.apiUrl}/${endpoint}`, {
            method: "POST",
            headers: [
				["Content-Type", "application/json"],
                ["X-Rez-Signature-256", this.createRequestSignature(body)],
                ["Authorization", "Bearer " + token],
            ],
            body,
        });
	}

	async onAuthenticate(data: onAuthenticatePayload): Promise<AuthContext> {
        const documentId = data.documentName;
        const res = await this.apiRequest("auth", data.token, { documentId });
        if (!res.ok || res.status != 200) throw new Error("Authentication Failed");

        const { readOnly, user } = await res.json();
		data.connectionConfig.readOnly = !!readOnly;
		return { user, token: data.token } as AuthContext;
	}

    async onLoadDocument(data: onLoadDocumentPayload): Promise<any> {
        if (!data.context.token) throw Forbidden;

        const documentId = data.documentName;
        const res = await this.apiRequest("load", data.context.token, { documentId });

        if (res.status !== 200) throw new Error("failed to load");

        const state = await res.json();
        const update = new Uint8Array(state.data);
        if (update) Y.applyUpdate(data.document, update);
	}

	async onStoreDocument(data: onChangePayload) {
        if (!data.context.token) throw Forbidden;

        const documentId = data.documentName;
        const state = Buffer.from(Y.encodeStateAsUpdate(data.document));
        await this.apiRequest("update", data.context.token, { state, documentId });
	}
}
import {
	type Extension,
	type onAuthenticatePayload,
	type onLoadDocumentPayload,
	type onDestroyPayload,
	type onStoreDocumentPayload,
	type storePayload,
	type fetchPayload
} from "@hocuspocus/server";
import { Database } from "@hocuspocus/extension-database";

import { SQL } from "bun";
import { validate as isUuid } from "uuid";
import { verify } from "paseto-ts/v4";

import { emptyDocument } from "./transformer";
import type { Config } from "./config.ts";

type SessionTokenClaims = {
	sub: string;
	tenant_id: string;
	document_id: string;
	can_edit: boolean;
};

export type SessionContext = {
	tenantId: string;
	userId: string;
	documentId: string;
	canEdit: boolean;
};

const createDatabase = (db: SQL): Database => {
	return new Database({
		fetch: async ({context, documentName}: fetchPayload<SessionContext>) => {
			const rows = await db`SELECT content FROM documents 
			WHERE tenant_id=${context.tenantId}::INT AND id=${documentName}::UUID 
			LIMIT 1`.raw();

			if (rows.length === 0) return null;
			if (rows[0].length === 0) return emptyDocument;
			const content = rows[0][0];
			if (content.length === 0) return emptyDocument;
			return new Uint8Array(content);
		}, 
		store: async ({documentName, state, lastContext}: storePayload<SessionContext>) => {
			await db`INSERT INTO documents ("tenant_id", "id", "content") 
				VALUES (${lastContext.tenantId}::INT, ${documentName}::UUID, ${state}::BYTEA)
			ON CONFLICT(id) DO UPDATE
				SET content = EXCLUDED.content`;
		},
	});
}

export class DocumentsServerExtension implements Extension<SessionContext> {
	extensionName = "Rezible Documents Server";
	db: SQL;
	database: Database;
	sessionPublicKey: string;

	constructor({ dbUrl, sessionPublicKey }: Config) {
		this.sessionPublicKey = sessionPublicKey;
		this.db = new SQL({
			url: dbUrl,
		});
		this.database = createDatabase(this.db);
	}

	async onDestroy(data: onDestroyPayload): Promise<void> {
		await this.db?.close();
	}

	verifySessionToken(token: string, documentName: string): SessionContext {
		if (!token) throw new Error("missing document session token");

		const { payload } = verify<SessionTokenClaims>(this.sessionPublicKey, token);

		if (payload.iss !== "rezible-backend" || payload.aud !== "rezible-documents-server") {
			throw new Error("invalid document session issuer or audience");
		}
		if (typeof payload.sub !== "string" || !isUuid(payload.sub)) throw new Error("invalid user");
		if (typeof payload.tenant_id !== "string" || !/^[1-9][0-9]*$/.test(payload.tenant_id)) {
			throw new Error("invalid tenant");
		}
		if (!isUuid(documentName) || payload.document_id !== documentName) throw new Error("invalid document");
		if (typeof payload.can_edit !== "boolean") throw new Error("invalid document permission");
		if (!payload.iat || !payload.nbf || !payload.exp) throw new Error("missing document session timestamps");

		return {
			tenantId: payload.tenant_id,
			userId: payload.sub,
			documentId: payload.document_id,
			canEdit: payload.can_edit,
		};
	};

	// TODO: Enforce token expiry for established document connections.
	async onAuthenticate(data: onAuthenticatePayload<SessionContext>): Promise<SessionContext> {
		const session = this.verifySessionToken(data.token, data.documentName);

		data.connectionConfig = {
			isAuthenticated: true,
			readOnly: !session.canEdit,
		};

		return session;
	}

    async onLoadDocument(data: onLoadDocumentPayload<SessionContext>) {
		await this.database.onLoadDocument(data);
	}

	async onStoreDocument(data: onStoreDocumentPayload<SessionContext>) {
		return this.database.onStoreDocument(data);
	}
}

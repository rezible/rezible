import type {
	Extension,
	onAuthenticatePayload,
	onLoadDocumentPayload,
	onDestroyPayload,
	onStoreDocumentPayload,
	onDisconnectPayload,
	beforeHandleMessagePayload,
	connectedPayload,
	storePayload,
	fetchPayload
} from "@hocuspocus/server";
import { Database } from "@hocuspocus/extension-database";

import { SQL } from "bun";
import { validate as isUuid } from "uuid";
import { decrypt } from "paseto-ts/v4";

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
	expiresAt: number;
};

const createDatabase = (db: SQL): Database => {
	return new Database({
		fetch: async ({context, documentName}: fetchPayload<SessionContext>) => {
			const rows = await db`SELECT content FROM documents 
			WHERE tenant_id=${context.tenantId}::INT AND id=${documentName}::UUID 
			LIMIT 1`.raw();

			if (rows.length === 0) throw new Error("document not found");
			const content = rows[0][0];
			if (!content || content.length === 0) return emptyDocument;
			return new Uint8Array(content);
		}, 
		store: async ({documentName, state, lastContext}: storePayload<SessionContext>) => {
			const rows = await db`UPDATE documents
				SET content = ${state}::BYTEA
				WHERE tenant_id = ${lastContext.tenantId}::INT AND id = ${documentName}::UUID
				RETURNING id`;
			if (rows.length === 0) throw new Error("document not found");
		},
	});
}

export class DocumentsServerExtension implements Extension<SessionContext> {
	extensionName = "Rezible Documents Server";
	db: SQL;
	database: Database;
	sessionKey: string;
	private expiryTimers = new Map<string, ReturnType<typeof setTimeout>>();

	constructor({ dbUrl, sessionKey }: Config) {
		this.sessionKey = sessionKey;
		this.db = new SQL({ url: dbUrl });
		this.database = createDatabase(this.db);
	}

	async onDestroy(data: onDestroyPayload): Promise<void> {
		this.expiryTimers.values().forEach(clearTimeout);
		this.expiryTimers.clear();
		await this.db?.close();
	}

	verifySessionToken(token: string, documentName: string): SessionContext {
		if (!token) throw new Error("missing document session token");

		const { payload } = decrypt<SessionTokenClaims>(this.sessionKey, token);

		if (payload.iss !== "rezible-backend" || payload.aud !== "rezible-documents-server") {
			throw new Error("invalid document session issuer or audience");
		}
		if (typeof payload.sub !== "string" || !isUuid(payload.sub)) throw new Error("invalid user");
		if (typeof payload.tenant_id !== "string" || !/^[1-9][0-9]*$/.test(payload.tenant_id)) {
			throw new Error("invalid tenant");
		}
		if (typeof payload.can_edit !== "boolean") throw new Error("invalid document permission");
		if (!payload.iat || !payload.nbf || !payload.exp || !payload.jti) {
			throw new Error("missing document session timestamps or jti");
		}

		const expiresAt = Date.parse(payload.exp);
		if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) throw new Error("expired document session");
		if (!Number.isFinite(Date.parse(payload.iat)) || !Number.isFinite(Date.parse(payload.nbf))) {
			throw new Error("invalid document session timestamps");
		}

		if (!isUuid(documentName) || payload.document_id !== documentName) throw new Error("invalid document");

		return {
			tenantId: payload.tenant_id,
			userId: payload.sub,
			documentId: payload.document_id,
			canEdit: payload.can_edit,
			expiresAt,
		};
	};

	async onAuthenticate(data: onAuthenticatePayload<SessionContext>): Promise<SessionContext> {
		const session = this.verifySessionToken(data.token, data.documentName);

		data.connectionConfig = {
			isAuthenticated: true,
			readOnly: !session.canEdit,
		};

		return session;
	}

	async connected(data: connectedPayload<SessionContext>): Promise<void> {
		const existingTimer = this.expiryTimers.get(data.socketId);
		if (existingTimer) clearTimeout(existingTimer);
		this.expiryTimers.set(data.socketId, setTimeout(() => {
			this.expiryTimers.delete(data.socketId);
			data.connection.webSocket.close(4001, "document session expired");
		}, Math.max(0, data.context.expiresAt - Date.now())));
	}

	async beforeHandleMessage(data: beforeHandleMessagePayload<SessionContext>): Promise<void> {
		if (data.context.expiresAt > Date.now()) return;
		data.connection.webSocket.close(4001, "document session expired");
		throw { code: 4001, reason: "document session expired" };
	}

	async onDisconnect(data: onDisconnectPayload<SessionContext>): Promise<void> {
		const timer = this.expiryTimers.get(data.socketId);
		if (timer) clearTimeout(timer);
		this.expiryTimers.delete(data.socketId);
	}

    async onLoadDocument(data: onLoadDocumentPayload<SessionContext>) {
		await this.database.onLoadDocument(data);
	}

	async onStoreDocument(data: onStoreDocumentPayload<SessionContext>) {
		return this.database.onStoreDocument(data);
	}
}

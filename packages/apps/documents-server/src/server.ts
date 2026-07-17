import {
	type Extension,
	type onAuthenticatePayload,
	type onLoadDocumentPayload,
	type onDestroyPayload,
	type onStoreDocumentPayload,
	Server,
	type storePayload,
	type fetchPayload
} from "@hocuspocus/server";
import { Database } from "@hocuspocus/extension-database";

import { SQL } from "bun";
import { decrypt } from "paseto-ts/v4";

import { emptyDocument } from "./transformer";
import type { Config } from "./config.ts";

type SessionTokenClaims = {
	user_id: string;
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
			console.log(rows);

			if (rows.length === 0) return null;
			if (rows[0].length === 0) return emptyDocument;
			console.log(rows[0]);
			return new Uint8Array(rows[0][0]);
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
	sessionTokenKey: Uint8Array;

	constructor({ dbUrl, sessionTokenSecretKey }: Config) {
		this.sessionTokenKey = sessionTokenSecretKey;
		this.db = new SQL({
			url: dbUrl,
			onconnect: client => {
				console.log("Connected to PostgreSQL");
			},
			onclose: client => {
				console.log("PostgreSQL connection closed");
			},
		});
		this.database = createDatabase(this.db);
	}

	async onDestroy(data: onDestroyPayload): Promise<void> {
		await this.db?.close();
	}

	decryptSessionToken(token: string, documentName: string): SessionContext {
		if (!token) throw new Error("missing document session token");

		const { payload } = decrypt<SessionTokenClaims>(this.sessionTokenKey, token);

		if (payload.document_id !== documentName) throw new Error("invalid document");

		return {
			tenantId: payload.tenant_id,
			userId: payload.sub || payload.user_id,
			documentId: payload.document_id,
			canEdit: payload.can_edit,
		};
	};

	async onAuthenticate(data: onAuthenticatePayload<SessionContext>): Promise<SessionContext> {
		const session = this.decryptSessionToken(data.token, data.documentName);

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

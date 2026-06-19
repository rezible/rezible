import type {
	Extension,
	onAuthenticatePayload,
	onChangePayload,
	onLoadDocumentPayload,
	fetchPayload,
	storePayload,
	onDestroyPayload
} from "@hocuspocus/server";
import { Database } from "@hocuspocus/extension-database";
import pg from "pg";
import { emptyDocument } from "./transformer";

import { decrypt } from "paseto-ts/v4";
import type { Config } from "./config.ts";

const docsTableName = "documents";

const selectDocumentQuery = `
	SELECT content FROM ${docsTableName} 
	WHERE tenant_id=$1::INT AND id=$2::UUID 
	LIMIT 1`
const upsertDocumentQuery = `
	INSERT INTO ${docsTableName} ("tenant_id", "id", "content") 
		VALUES ($1::INT, $2::UUID, $3::BYTEA)
	ON CONFLICT(id) DO UPDATE
		SET content = EXCLUDED.content`

type AuthSessionContext = {
	tenantId: string;
	userId: string;
	documentId: string;
	canEdit: boolean;
};
const getTenantId = (data: {context: any}) => (data.context as AuthSessionContext).tenantId;

type DocumentEditorSessionClaims = {
	user_id: string;
	tenant_id: string;
	document_id: string;
	can_edit: boolean;
};

export const decryptDocumentEditorSessionToken = (key: Uint8Array, token: string, documentName: string): AuthSessionContext => {
	const { payload } = decrypt<DocumentEditorSessionClaims>(key, token);

	return {
		tenantId: String(payload.tenant_id),
		userId: payload.sub || payload.user_id,
		documentId: payload.document_id,
		canEdit: payload.can_edit,
	};
};

export class DocumentsServerExtension implements Extension {
	extensionName = "Rezible Documents Server";
	database: Database;
	dbPool: pg.Pool;
	sessionTokenKey: Uint8Array;

	constructor({ dbUrl, sessionTokenSecretKey }: Config) {
		this.dbPool = new pg.Pool({connectionString: dbUrl});
		this.database = new Database({
			fetch: data => this.fetchDocument(data), 
			store: data => this.storeDocument(data),
		});
		this.sessionTokenKey = sessionTokenSecretKey;
	}

	async onDestroy(data: onDestroyPayload): Promise<void> {
		await this.dbPool?.end();
	}

	async onAuthenticate(data: onAuthenticatePayload): Promise<AuthSessionContext> {
		if (!data.token) {
			throw new Error("missing document session token");
		}
		const session = decryptDocumentEditorSessionToken(this.sessionTokenKey, data.token, data.documentName);

		data.connectionConfig = {
			isAuthenticated: true,
			readOnly: !session.canEdit,
		};

		return session;
	}

	private async fetchDocument(data: fetchPayload) {
		const params = [getTenantId(data), data.documentName];
		const res = await this.dbPool.query(selectDocumentQuery, params);
		if (res.rowCount === 0) {
			return null;
		}
		const contentBuffer = res.rows[0].content as Buffer;
		if (contentBuffer.length > 0) {
			return new Uint8Array(contentBuffer);
		};
		return emptyDocument;
	}

    async onLoadDocument(data: onLoadDocumentPayload) {
		await this.database.onLoadDocument(data);
	}

	private async storeDocument(data: storePayload) {
		const params = [getTenantId(data), data.documentName, data.state];
		await this.dbPool.query(upsertDocumentQuery, params);
	}

	async onStoreDocument(data: onChangePayload) {
		return this.database.onStoreDocument(data);
	}
}

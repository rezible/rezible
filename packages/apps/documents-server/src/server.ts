import type { IncomingHttpHeaders } from "http";
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

import { client, getCurrentAuthSession, getDocumentAccess, type User, type Options } from "@rezible/api-client-ts";
import type {Config} from "./config.ts";

const tenantIdHeader = "x-rez-tenant-id";
const authCookieName = "rez_access_token";
const reqSecurity: Options["security"] = [{ scheme: 'bearer', type: 'http' }];

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
	user: User;
	expiresAt: string;
};
const getTenantId = (data: {context: any}) => (data.context as AuthSessionContext).tenantId;

export class DocumentsServerExtension implements Extension {
	extensionName = "Rezible Documents Server";
	database: Database;
	dbPool: pg.Pool;

	constructor({ apiUrl, dbUrl }: Config) {
		client.setConfig({
			baseUrl: `${apiUrl}/v1`,
			headers: {
				["X-Rez-DocumentsServer"]: true,
			}
		});

		this.dbPool = new pg.Pool({connectionString: dbUrl});
		this.database = new Database({
			fetch: data => this.fetchDocument(data), 
			store: data => this.storeDocument(data),
		});
	}

	async onDestroy(data: onDestroyPayload): Promise<void> {
		await this.dbPool?.end();
	}

	getRequestAuth(data: {requestHeaders: IncomingHttpHeaders}): {auth: string, security: Options["security"]} {
		const authToken = data.requestHeaders.cookie
			?.split(';')
			.find(c => c.trim().startsWith(`${authCookieName}=`))
			?.split("=")
			.at(1);
		if (!authToken) throw new Error('missing or invalid auth token');
		return {auth: authToken, security: reqSecurity};
	}

	async onAuthenticate(data: onAuthenticatePayload): Promise<AuthSessionContext> {
		const tenantId = data.request.headers[tenantIdHeader];
		if (!tenantId || Array.isArray(tenantId)) {
			throw new Error("invalid auth headers");
		}

		const authRes = await getCurrentAuthSession(this.getRequestAuth(data));
		if (authRes.error) {
			throw new Error("failed to check auth session");
		};
		const user = authRes.data.data.user;
		const expiresAt = authRes.data.data.expiresAt;

		const res = await getDocumentAccess({
			...this.getRequestAuth(data),
			path: { id: data.documentName },
		});
        if (res.error) {
			throw new Error("Failed to get access: " + JSON.stringify(res.error));
		};

		data.connectionConfig = {
			isAuthenticated: true,
			readOnly: !res.data.data.canEdit,
		};

		return {tenantId, user, expiresAt};
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
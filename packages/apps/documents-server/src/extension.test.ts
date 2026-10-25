import { expect, spyOn, test } from "bun:test";
import type { onAuthenticatePayload } from "@hocuspocus/server";
import { encrypt } from "paseto-ts/v4";
import { pasetoLocalKeyFromHex } from "./config";
import { DocumentsServerExtension, type SessionContext } from "./extension";

const documentId = "13b89cc9-6544-44b6-a718-bdc8793b8bbc";
const context: SessionContext = {
	tenantId: "1",
	userId: "6dd4db2d-c647-46ca-a189-a08b60c9d019",
	documentId,
	canEdit: false,
	expiresAt: Date.now() + 60000,
};

test("authentication mutates the original connection configuration", async () => {
	const server = new DocumentsServerExtension({
		name: "test",
		host: "localhost",
		port: 1,
		dbUrl: "postgresql://unused@localhost/unused",
		sessionKey: "unused",
		allowedOrigins: [],
	});
	spyOn(server, "verifySessionToken").mockReturnValue(context);
	const connectionConfig = { readOnly: false, isAuthenticated: false };
	const payload = {
		token: "token",
		documentName: documentId,
		connectionConfig,
	} as onAuthenticatePayload<SessionContext>;
	await server.onAuthenticate(payload);
	expect(payload.connectionConfig).toBe(connectionConfig);
	expect(connectionConfig.readOnly).toBe(true);
});

test("a session token authorizes only the document it names", async () => {
	const sessionKey = pasetoLocalKeyFromHex("ab".repeat(32));
	const server = new DocumentsServerExtension({
		name: "test",
		host: "localhost",
		port: 1,
		dbUrl: "postgresql://unused@localhost/unused",
		sessionKey,
		allowedOrigins: [],
	});
	const now = new Date();
	const token = encrypt(sessionKey, {
		iss: "rezible-backend",
		aud: "rezible-documents-server",
		sub: context.userId,
		tenant_id: context.tenantId,
		document_id: documentId,
		can_edit: true,
		iat: now.toISOString(),
		nbf: now.toISOString(),
		exp: new Date(now.getTime() + 60000).toISOString(),
		jti: crypto.randomUUID(),
	});
	const connect = (documentName: string) => server.onAuthenticate({
		token,
		documentName,
		connectionConfig: { readOnly: true, isAuthenticated: false },
	} as onAuthenticatePayload<SessionContext>);

	expect((await connect(documentId)).documentId).toBe(documentId);
	await expect(connect("0b0f7d4e-34a1-4d0b-9f0e-2a4f8e0c1d2b")).rejects.toThrow("invalid document");
});

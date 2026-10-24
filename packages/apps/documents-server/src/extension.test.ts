import { expect, spyOn, test } from "bun:test";
import type { onAuthenticatePayload } from "@hocuspocus/server";
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

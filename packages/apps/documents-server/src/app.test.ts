import { expect, test } from "bun:test";
import type { SQL } from "bun";
import { createDocumentsServer } from "./app";
import type { SessionContext } from "./extension";

const documentId = "13b89cc9-6544-44b6-a718-bdc8793b8bbc";
const context: SessionContext = {
	tenantId: "1",
	userId: "6dd4db2d-c647-46ca-a189-a08b60c9d019",
	documentId,
	canEdit: true,
	expiresAt: Date.now() + 60000,
};

const config = {
	name: "test",
	host: "localhost",
	port: 1,
	dbUrl: "postgresql://unused@localhost/unused",
	sessionKey: "unused",
	allowedOrigins: ["https://app.example.com"],
};

// A tagged-template stand-in for Bun's SQL, answering each query by its text.
const stubSql = (answer: (query: string) => Promise<unknown[]>, events: string[] = []) => {
	const sql = (strings: TemplateStringsArray) => {
		const query = strings.join("?");
		if (query.includes("UPDATE documents")) events.push("store");
		const result = answer(query);
		return Object.assign(result, { raw: () => result });
	};
	return Object.assign(sql, {
		close: async () => {
			events.push("close");
		},
	}) as unknown as SQL;
};

const documentRows = (query: string): Promise<unknown[]> => {
	if (query.includes("SELECT content")) return Promise.resolve([[null]]);
	if (query.includes("UPDATE documents")) return Promise.resolve([{ id: documentId }]);
	return Promise.resolve([{ "?column?": 1 }]);
};

const request = (path: string, headers?: HeadersInit) =>
	new Request(`http://localhost${path}`, { headers });

const status = async (server: ReturnType<typeof createDocumentsServer>, req: Request) =>
	(await server.fetch(req, undefined as any))?.status;

const openWithPendingWrite = async (server: ReturnType<typeof createDocumentsServer>) => {
	const connection = await server.hocuspocus.openDirectConnection(documentId, context);
	await connection.transact((document) => document.getText("test").insert(0, "edit"));
	await connection.disconnect({ unloadImmediately: false });
	expect(server.hocuspocus.getDocumentsCount()).toBe(1);
};

test("ready reflects the database ping", async () => {
	const healthy = createDocumentsServer(config, { db: stubSql(documentRows) });
	expect(await status(healthy, request("/ready"))).toBe(200);

	const unreachable = createDocumentsServer(config, {
		db: stubSql(() => Promise.reject(new Error("connection refused"))),
	});
	expect(await status(unreachable, request("/ready"))).toBe(503);
	expect(await status(unreachable, request("/health"))).toBe(204);
});

test("shutdown makes the server unready and rejects websocket upgrades", async () => {
	const server = createDocumentsServer(config, { db: stubSql(documentRows) });
	const upgrade = request("/", { upgrade: "websocket" });
	expect(await status(server, upgrade)).toBe(403);

	const shutdown = server.shutdown();
	expect(await status(server, request("/ready"))).toBe(503);
	expect(await status(server, upgrade)).toBe(503);
	expect(await status(server, request("/health"))).toBe(204);
	expect(await shutdown).toBe(0);
});

test("shutdown stores pending writes before closing the database", async () => {
	const events: string[] = [];
	const server = createDocumentsServer(config, { db: stubSql(documentRows, events) });
	await openWithPendingWrite(server);

	expect(await server.shutdown()).toBe(0);
	expect(events).toEqual(["store", "close"]);
	expect(server.hocuspocus.getDocumentsCount()).toBe(0);
});

test("shutdown fails when a pending write cannot be stored", async () => {
	const events: string[] = [];
	const db = stubSql((query) => {
		if (query.includes("UPDATE documents")) return Promise.reject(new Error("write failed"));
		return documentRows(query);
	}, events);
	const server = createDocumentsServer(config, { db });
	await openWithPendingWrite(server);

	expect(await server.shutdown()).toBe(1);
	expect(events).toEqual(["store", "close"]);
});

test("shutdown fails when pending writes outlast the grace period", async () => {
	const db = stubSql((query) => {
		if (query.includes("UPDATE documents")) return new Promise(() => {});
		return documentRows(query);
	});
	const server = createDocumentsServer(config, { db, shutdownTimeoutMs: 100 });
	await openWithPendingWrite(server);

	expect(await server.shutdown()).toBe(1);
});

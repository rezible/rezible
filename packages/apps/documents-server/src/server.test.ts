import { afterEach, describe, expect, test } from "bun:test";
import { DocumentsServerExtension } from "./server";

const extensions: DocumentsServerExtension[] = [];

afterEach(async () => {
	await Promise.all(extensions.map(extension => extension.onDestroy({} as never)));
	extensions.length = 0;
});

const createExtension = () => {
	const extension = new DocumentsServerExtension({
		name: "documents-server",
		host: "127.0.0.1",
		port: 7002,
		apiUrl: "http://api.local",
		dbUrl: "postgresql://documents:secret@localhost:5432/rezible",
	});
	extensions.push(extension);
	return extension;
};

describe("DocumentsServerExtension.getRequestAuth", () => {
	test("extracts the access token from the rezible auth cookie", () => {
		const extension = createExtension();

		expect(extension.getRequestAuth({
			requestHeaders: {
				cookie: "other=value; rez_access_token=test-token; another=value",
			},
		})).toEqual({
			auth: "test-token",
			security: [{ scheme: "bearer", type: "http" }],
		});
	});

	test("throws when the auth cookie is missing", () => {
		const extension = createExtension();

		expect(() => extension.getRequestAuth({
			requestHeaders: {
				cookie: "other=value",
			},
		})).toThrow("missing or invalid auth token");
	});
});

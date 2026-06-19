import { afterEach, describe, expect, test } from "bun:test";
import { DocumentsServerExtension } from "./server";
import { pasetoLocalKeyFromHex } from "./config";

const extensions: DocumentsServerExtension[] = [];
const sessionTokenSecretHex = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f";
const sessionTokenSecretKey = pasetoLocalKeyFromHex(sessionTokenSecretHex);
const documentName = "018f0b4f-7a4d-7f86-b4d3-13b95a4a0131";

afterEach(async () => {
	await Promise.all(extensions.map(extension => extension.onDestroy({} as never)));
	extensions.length = 0;
});

const createExtension = () => {
	const extension = new DocumentsServerExtension({
		name: "documents-server",
		host: "127.0.0.1",
		port: 7002,
		dbUrl: "postgresql://documents:secret@localhost:5432/rezible",
		sessionTokenSecretKey,
	});
	extensions.push(extension);
	return extension;
};

import { Server } from "@hocuspocus/server";
import { Logger } from "@hocuspocus/extension-logger";
import { DocumentsServerExtension } from "./documents";

type Config = {
	name: string;
	host: string;
	port: number;
	apiUrl: string;
	dbUrl: string;
}

const loadConfig = (): Config => {
	const name = process.env.NAME ?? "documents-server";

	const host = process.env.HOST ?? "0.0.0.0";
	let port = Number.parseInt(process.env.PORT ?? "7002", 10);
	if (port < 1024) port = 7003;

	const apiUrl = process.env.API_URL ?? "";
	const dbUrl = process.env.DOCUMENTS_DB_URL ?? process.env.DB_URL ?? "";

	return { name, host, port, apiUrl, dbUrl };
}

const createServer = () => {
    const cfg = loadConfig();
	const server = new Server({
		name: cfg.name,
		address: cfg.host,
		port: cfg.port,
		timeout: 30000,
		debounce: 1000,
		maxDebounce: 30000,
		quiet: false,
		extensions: [
			new Logger(),
            new DocumentsServerExtension(cfg.apiUrl, cfg.dbUrl),
		],
	});
	return server;
}

const runServer = async () => {
	try {
		await createServer().listen();
	} catch (e: unknown) {
		if (e instanceof Error) {
			console.error("Failed to create server: %s", e.message);
		} else {
			console.error("Failed to create server: %s", e);
		}
	}
}

runServer();

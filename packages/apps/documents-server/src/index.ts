import { Server } from "@hocuspocus/server";
import { Logger } from "@hocuspocus/extension-logger";
import { loadConfig } from "./config";
import { DocumentsServerExtension } from "./server";

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

import { Server } from "@hocuspocus/server";
import { Logger } from "@hocuspocus/extension-logger";
import { RezibleServerProxy } from "./rezible-proxy";

type Config = {
	name: string;
	host: string;
	port: number;

	apiUrl: string;
	apiSecret: string;
}

const loadConfig = (): Config => {
	const name = process.env.NAME ?? "document-server";

	let host = process.env.HOST ?? "localhost";
	let port = Number.parseInt(process.env.PORT ?? "8889", 10);
	if (port < 1024) port = 8889;

	const apiUrl = process.env.DOCUMENTS_API_URL ?? "http://localhost:8888/api/documents";
	const apiSecret = process.env.DOCUMENTS_API_SECRET ?? "foo";

	return { name, host, port, apiUrl, apiSecret };
}

const createServer = async (cfg: Config) => {
	const logger = new Logger();
    const rezProxy = new RezibleServerProxy(cfg.apiUrl, cfg.apiSecret);
	
	const server = new Server({
		name: cfg.name,
		address: cfg.host,
		port: cfg.port,
		timeout: 30000,
		debounce: 1000,
		maxDebounce: 30000,
		quiet: false,
		extensions: [
			logger,
            rezProxy,
		],
	});

	return server;
}

const runServer = async () => {
	try {
		const cfg = loadConfig();
		const server = await createServer(cfg);
		await server.listen();
	} catch (e: unknown) {
		if (e instanceof Error) {
			console.error("Failed to create server: %s", e.message);
		} else {
			console.error("Failed to create server: %s", e);
		}
	}
}

runServer();
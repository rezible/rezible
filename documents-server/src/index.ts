import { Server } from "@hocuspocus/server";
import { Logger } from "@hocuspocus/extension-logger";
import { RezibleServerProxy } from "./rezible-proxy";

type Config = {
	name: string;
	host: string;
	port: number;

	apiUrl: string;
}

const loadConfig = (): Config => {
	const name = process.env.NAME ?? "document-server";

	let host = process.env.HOST ?? "0.0.0.0";
	let port = Number.parseInt(process.env.PORT ?? "7003", 10);
	if (port < 1024) port = 7003;

	return { name, host, port };
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
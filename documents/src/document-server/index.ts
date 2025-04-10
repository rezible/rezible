import {Server} from "@hocuspocus/server";
import { Events, Webhook } from "@hocuspocus/extension-webhook";
import { Logger } from "@hocuspocus/extension-logger";
import { createDatabase } from "./database";
import { documentTransformer } from "./transformer";
import { SessionAuthentication } from "./auth";
import { RezibleDocumentsApi } from "./documents-api";

type Config = {
	name: string;
	host: string;
	port: number;

	dbUrl: string;

	apiUrl: string;
	apiWebhookSecret: string;
}

const loadConfig = (): Config => {
	const name = process.env.NAME ?? "document-server";

	let host = process.env.HOST ?? "localhost";
	let port = Number.parseInt(process.env.PORT ?? "", 10);
	if (port < 1024) port = 8889;

	const apiUrl = process.env.API_URL ?? "http://localhost:8888/api";

	const apiWebhookSecret = process.env.API_WEBHOOK_SECRET ?? "foo-bar-baz";

	const dbUrl = process.env.DB_URL;
	if (!dbUrl) {
		throw new Error("DB_URL env variable not supplied")
	}

	return { name, host, port, dbUrl, apiUrl, apiWebhookSecret };
}

const createServer = async (cfg: Config) => {
	const logger = new Logger();
	const database = await createDatabase(cfg.dbUrl);

	const webhooks = new Webhook({
		url: `${cfg.apiUrl}/webhooks/documents`,
		secret: cfg.apiWebhookSecret,
		transformer: documentTransformer,
		events: [Events.onChange],
		debounce: 5000,
		debounceMaxWait: 10000,
	});

	const sessionAuth = new SessionAuthentication({
		verifySessionUrl: `${cfg.apiUrl}/v1/documents/session/verify`,
	});

	const docsApi = new RezibleDocumentsApi({

	})

	const server = Server.configure({
		name: cfg.name,
		address: cfg.host,
		port: cfg.port,
		timeout: 30000,
		debounce: 1000,
		maxDebounce: 30000,
		quiet: false,
		extensions: [
			logger,
			sessionAuth,
			database,
			docsApi,
			// webhooks
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
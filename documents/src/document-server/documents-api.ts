import type { Extension, onRequestPayload } from "@hocuspocus/server";
import type { IncomingHttpHeaders, IncomingMessage } from "http";
import { generateHTML } from '@tiptap/html'
import { extensions, handoverSchema, schema, transformSchemaSpec } from "./transformer";
import type { JSONContent } from "@tiptap/core";
import type { SchemaSpec } from "@tiptap/pm/model";

export interface RezApiConfiguration {
	
}

const readRequestBody = (req: IncomingMessage): Promise<string> => {
	return new Promise((resolve, reject) => {
		const parts: Buffer[] = [];
		req.on("data", chunk => parts.push(chunk));
		req.on("end", () => resolve(Buffer.concat(parts).toString()));
		req.on("error", err => reject(err));
	});
}

type RezApiHandlerFunc = (body: string) => Promise<any>;

export class RezibleDocumentsApi implements Extension {
	extensionName: string;
	configuration: RezApiConfiguration = {};
	handlers: Record<string, RezApiHandlerFunc> = {};

	constructor(configuration: Partial<RezApiConfiguration>) {
		this.extensionName = "Rezible API";
		this.configuration = { ...this.configuration, ...configuration };
		this.handlers = {
			"transform": this.handleTransformDocument,
			"schema-spec": this.handleGetSchemaSpec,
		}
	}

	async onRequest(data: onRequestPayload) {
		return new Promise<void>(async (continueRequest, handled) => {
			const { request, response } = data;

			const pathname = request.url?.split("/api/")[1];
			if (!pathname) {
				return continueRequest();
			} else if (!(pathname in this.handlers)) {
				console.log("unknown endpoint ", pathname);
				return continueRequest();
			}

			const authed = await this.authenticateRequest(request.headers);
			if (authed) {
				try {
					const rawBody = await readRequestBody(request);
					const respBody = await this.handlers[pathname](rawBody);
					response.writeHead(200, {"content-type": "application/json"})
					response.write(JSON.stringify(respBody));
				} catch (e) {
					response.writeHead(500);
					console.error("failed to handle api request: ", e);
				}
			} else {
				response.writeHead(401);
			}

			response.end();
			handled();
		})
	}

	async authenticateRequest(headers: IncomingHttpHeaders): Promise<boolean> {
		return true
	}

	async handleTransformDocument(rawBody: string) {
		const data = JSON.parse(rawBody) as {content: string; format?: "text" | "html"};
		
		const doc = JSON.parse(data.content) as JSONContent;
		if (data.format === "html") {
			return generateHTML(doc, extensions);
		}
		
		const content = schema.nodeFromJSON(doc);
		return content.toString();
	}

	async handleGetSchemaSpec(rawBody: string) {
		const {name} = JSON.parse(rawBody) as {name: string};
		
		if (name === "handover") {
			const transformed = transformSchemaSpec(handoverSchema.spec);
			return {spec: transformed};
		}
		
		return null;
	}
}

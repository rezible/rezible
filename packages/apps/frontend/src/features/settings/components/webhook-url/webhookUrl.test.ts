import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";

import type { IntegrationInstallation } from "$lib/api";

import type { WebhookUrlController } from "./webhookUrl.svelte";

// Bun does not compile runes. Compile the real controller and supply its external dependencies locally,
// avoiding module mocks that would affect other tests.
const source = await Bun.file(new URL("./webhookUrl.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "webhookUrl.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const controllerCode = compiled.replace(/^import .*;$/gm, "").replace(/^export /gm, "");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

type IssueRequest = { path: { id: string } };

function setup(issue: (request: IssueRequest) => Promise<{ data: { url: string } }>) {
	const issued = mock(issue);
	const copied = mock(async (_text: string) => {});
	const toast = { success: mock(() => {}), error: mock(() => {}) };
	Object.defineProperty(globalThis, "navigator", {
		value: { clipboard: { writeText: copied } },
		configurable: true,
	});

	const Controller = new Function(
		"$",
		"createMutation",
		"issueIntegrationWebhookTokenMutation",
		"toast",
		`${controllerCode}\nreturn WebhookUrlController;`
	)(
		runtime,
		() => ({ mutateAsync: issued }),
		() => ({}),
		toast
	) as new () => WebhookUrlController;

	const controller = new Controller();
	return { controller, issued, copied };
}

const installationId = "installation-1";
const installation = (webhookUrlIssued: boolean) =>
	({ id: installationId, attributes: { webhookUrlIssued } }) as IntegrationInstallation;

describe("WebhookUrlController", () => {
	test("the first URL is generated at once and shown with copy; replacing it asks first", async () => {
		let count = 0;
		const { controller, issued, copied } = setup(async () => {
			count += 1;
			return { data: { url: `https://api.test/webhooks/alertmanager/token-${count}` } };
		});
		const unissued = installation(false);
		expect(controller.hasUrl(unissued)).toBe(false);

		await controller.requestGenerate(unissued);
		expect(controller.confirmingId).toBeUndefined();
		expect(issued).toHaveBeenCalledWith({ path: { id: installationId } });
		const firstUrl = controller.issuedUrl(installationId);
		expect(firstUrl).toBe("https://api.test/webhooks/alertmanager/token-1");
		expect(controller.hasUrl(unissued)).toBe(true);

		await controller.copy(firstUrl!, "Webhook URL");
		expect(copied).toHaveBeenCalledWith(firstUrl);

		// The page's installation still says no URL was issued; this controller knows better.
		await controller.requestGenerate(unissued);
		expect(controller.confirmingId).toBe(installationId);
		expect(issued).toHaveBeenCalledTimes(1);

		controller.cancelGenerate();
		expect(controller.confirmingId).toBeUndefined();
		expect(issued).toHaveBeenCalledTimes(1);

		await controller.requestGenerate(unissued);
		await controller.confirmGenerate();
		expect(controller.confirmingId).toBeUndefined();
		expect(controller.issuedUrl(installationId)).toBe("https://api.test/webhooks/alertmanager/token-2");
	});

	test("replacing a URL issued before this page asks first", async () => {
		const { controller, issued } = setup(async () => ({
			data: { url: "https://api.test/webhooks/webhook/token" },
		}));
		const issuedBefore = installation(true);
		expect(controller.hasUrl(issuedBefore)).toBe(true);

		await controller.requestGenerate(issuedBefore);
		expect(controller.confirmingId).toBe(installationId);
		expect(issued).not.toHaveBeenCalled();

		await controller.confirmGenerate();
		expect(issued).toHaveBeenCalledTimes(1);
		expect(controller.issuedUrl(installationId)).toBe("https://api.test/webhooks/webhook/token");
	});

	test("a failed replacement keeps the confirmation open with its error and shows no URL", async () => {
		let fail = false;
		const { controller } = setup(async () => {
			if (fail) {
				throw { title: "Unavailable", status: 503, detail: "" };
			}
			return { data: { url: "https://api.test/webhooks/alertmanager/token" } };
		});
		await controller.requestGenerate(installation(false));

		fail = true;
		await controller.requestGenerate(installation(false));
		await controller.confirmGenerate();
		expect(controller.issuedUrl(installationId)).toBeUndefined();
		expect(controller.issueError(installationId)?.status).toBe(503);
		expect(controller.confirmingId).toBe(installationId);
		expect(controller.confirmingError?.status).toBe(503);

		controller.cancelGenerate();
		expect(controller.confirmingError).toBeUndefined();
	});

	test("a failed first generation shows its error, and the next attempt asks first", async () => {
		const { controller, issued } = setup(async () => {
			throw { title: "Unavailable", status: 503, detail: "" };
		});

		await controller.requestGenerate(installation(false));
		expect(controller.issueError(installationId)?.status).toBe(503);
		expect(controller.issuedUrl(installationId)).toBeUndefined();

		// The failed attempt may still have stored a token, so the next one is treated as a replacement.
		await controller.requestGenerate(installation(false));
		expect(controller.confirmingId).toBe(installationId);
		expect(issued).toHaveBeenCalledTimes(1);
	});

	test("a new page starts without a URL", () => {
		const { controller } = setup(async () => ({ data: { url: "unused" } }));
		expect(controller.issuedUrl(installationId)).toBeUndefined();
	});
});

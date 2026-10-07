import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";

import type { AlertmanagerProviderController } from "./alertmanagerProvider.svelte";

// Bun does not compile runes. Compile the real controller and supply its external dependencies locally,
// avoiding module mocks that would affect other tests.
const source = await Bun.file(new URL("./alertmanagerProvider.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "alertmanagerProvider.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const controllerCode = compiled.replace(/^import .*;$/gm, "").replace(/^export /gm, "");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

type IssueRequest = { path: { id: string } };

function setup(issue: (request: IssueRequest) => Promise<{ data: { url: string } }>) {
	const install = mock(async (_name: string, _config: Record<string, unknown>) => true);
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
		`${controllerCode}\nreturn AlertmanagerProviderController;`
	)(
		runtime,
		() => ({ mutateAsync: issued }),
		() => ({}),
		toast
	) as new (provider: { install: typeof install }) => AlertmanagerProviderController;

	const controller = new Controller({ install });
	return { controller, install, issued, copied };
}

const installationId = "installation-1";

describe("AlertmanagerProviderController", () => {
	test("installing sends no configuration", async () => {
		const { controller, install } = setup(async () => ({ data: { url: "unused" } }));
		await controller.install();
		expect(install).toHaveBeenCalledWith("alertmanager", {});
	});

	test("generating warns first, then shows the new URL once with copy", async () => {
		let count = 0;
		const { controller, issued, copied } = setup(async () => {
			count += 1;
			return { data: { url: `https://api.test/webhooks/alertmanager/token-${count}` } };
		});

		controller.requestGenerate(installationId);
		expect(controller.confirmingId).toBe(installationId);
		expect(issued).not.toHaveBeenCalled();

		controller.cancelGenerate();
		expect(controller.confirmingId).toBeUndefined();
		expect(issued).not.toHaveBeenCalled();

		controller.requestGenerate(installationId);
		await controller.confirmGenerate();
		expect(issued).toHaveBeenCalledWith({ path: { id: installationId } });
		expect(controller.confirmingId).toBeUndefined();
		const firstUrl = controller.issuedUrl(installationId);
		expect(firstUrl).toBe("https://api.test/webhooks/alertmanager/token-1");

		await controller.copy(firstUrl!, "Webhook URL");
		expect(copied).toHaveBeenCalledWith(firstUrl);

		// Generating again asks again, then replaces the shown URL.
		controller.requestGenerate(installationId);
		expect(issued).toHaveBeenCalledTimes(1);
		await controller.confirmGenerate();
		expect(controller.issuedUrl(installationId)).toBe("https://api.test/webhooks/alertmanager/token-2");
	});

	test("a failed generation keeps the confirmation open with its error and shows no URL", async () => {
		let fail = false;
		const { controller } = setup(async () => {
			if (fail) {
				throw { title: "Unavailable", status: 503, detail: "" };
			}
			return { data: { url: "https://api.test/webhooks/alertmanager/token" } };
		});
		controller.requestGenerate(installationId);
		await controller.confirmGenerate();

		fail = true;
		controller.requestGenerate(installationId);
		await controller.confirmGenerate();
		expect(controller.issuedUrl(installationId)).toBeUndefined();
		expect(controller.issueError(installationId)?.status).toBe(503);
		expect(controller.confirmingId).toBe(installationId);
		expect(controller.confirmingError?.status).toBe(503);

		controller.cancelGenerate();
		expect(controller.confirmingError).toBeUndefined();
	});

	test("a new page starts without a URL", () => {
		const { controller } = setup(async () => ({ data: { url: "unused" } }));
		expect(controller.issuedUrl(installationId)).toBeUndefined();
	});
});

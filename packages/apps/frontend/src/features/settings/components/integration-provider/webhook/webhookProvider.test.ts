import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";

import { defaultWebhookPreset, webhookIntegrationName, webhookPresets } from "./deploymentPreset";
import type { WebhookProviderController } from "./webhookProvider.svelte";

// Bun does not compile runes. Compile the real controller and supply its imports locally.
const source = await Bun.file(new URL("./webhookProvider.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "webhookProvider.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const controllerCode = compiled.replace(/^import .*;$/gm, "").replace(/^export /gm, "");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

function setup() {
	const install = mock(async (_name: string, _config: Record<string, unknown>) => true);
	const Controller = new Function(
		"$",
		"defaultWebhookPreset",
		"webhookIntegrationName",
		"webhookPresets",
		`${controllerCode}\nreturn WebhookProviderController;`
	)(runtime, defaultWebhookPreset, webhookIntegrationName, webhookPresets) as new (provider: {
		install: typeof install;
	}) => WebhookProviderController;

	const controller = new Controller({ install });
	return { controller, install };
}

describe("WebhookProviderController", () => {
	test("installing asks for a preset, whose only option is Deployments", () => {
		const { controller } = setup();

		expect(controller.presets).toEqual([{ value: "deployment", label: "Deployments" }]);
		expect(controller.preset).toBe("deployment");
		expect(controller.presetLabel).toBe("Deployments");
	});

	test("installing sends the chosen preset", async () => {
		const { controller, install } = setup();

		await controller.install();

		expect(install).toHaveBeenCalledWith("webhook", { preset: "deployment" });
	});
});

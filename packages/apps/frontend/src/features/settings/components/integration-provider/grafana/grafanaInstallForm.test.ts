import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";

import type { GrafanaInstallForm } from "./grafanaInstallForm.svelte";

// Bun does not compile runes. Compile the real form and supply its runtime locally, avoiding module mocks
// that would affect other tests.
const source = await Bun.file(new URL("./grafanaInstallForm.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "grafanaInstallForm.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const formCode = compiled.replace(/^import .*;$/gm, "").replace(/^export /gm, "");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

type Install = (name: string, config: Record<string, unknown>, successMessage?: string) => Promise<boolean>;

const Form = new Function("$", `${formCode}\nreturn GrafanaInstallForm;`)(runtime) as new (
	provider: { install: Install },
	replaceUrl?: string
) => GrafanaInstallForm;

function setup(replaceUrl?: string, result = true) {
	const install = mock<Install>(async () => result);
	const form = new Form({ install }, replaceUrl);
	return { form, install };
}

describe("Grafana install form", () => {
	test("installing requires a URL and a token", async () => {
		const { form, install } = setup();

		expect(await form.install()).toBe(false);
		expect(form.urlError).toBeDefined();
		expect(form.tokenError).toBeDefined();

		form.url = "https://grafana.example.com";
		form.token = "  ";
		expect(await form.install()).toBe(false);
		expect(form.tokenError).toBeDefined();
		expect(install).not.toHaveBeenCalled();

		form.token = " glsa_token ";
		expect(await form.install()).toBe(true);
		expect(install).toHaveBeenCalledWith(
			"grafana",
			{ url: "https://grafana.example.com", token: "glsa_token" },
			undefined
		);
	});

	test("the token is cleared after installing", async () => {
		const { form } = setup();
		form.url = "https://grafana.example.com";
		form.token = "glsa_token";

		await form.install();

		expect(form.token).toBe("");
		expect(form.url).toBe("");
	});

	test("a failed install keeps what was entered", async () => {
		const { form } = setup(undefined, false);
		form.url = "https://grafana.example.com";
		form.token = "glsa_token";

		expect(await form.install()).toBe(false);
		expect(form.token).toBe("glsa_token");
	});

	test("replacing the token installs with the same URL", async () => {
		const url = "https://grafana.example.com";
		const { form, install } = setup(url);
		expect(form.replacing).toBe(true);
		expect(form.url).toBe(url);

		form.token = "glsa_new";
		await form.install();

		const [name, config] = install.mock.calls[0]!;
		expect(name).toBe("grafana");
		expect(config).toEqual({ url, token: "glsa_new" });
		expect(form.url).toBe(url);
		expect(form.token).toBe("");
	});
});

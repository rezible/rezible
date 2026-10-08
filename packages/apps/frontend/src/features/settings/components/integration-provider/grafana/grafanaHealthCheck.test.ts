import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";

import type { IntegrationHealthCheck, IntegrationInstallation } from "$lib/api";

import type { GrafanaHealthCheck } from "./grafanaHealthCheck.svelte";

// Bun does not compile runes. Compile the real controller and supply its external dependencies locally,
// avoiding module mocks that would affect other tests.
const source = await Bun.file(new URL("./grafanaHealthCheck.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "grafanaHealthCheck.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const controllerCode = compiled.replace(/^import [^;]*;$/gm, "").replace(/^export /gm, "");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

type CheckRequest = { path: { id: string } };
type Check = (request: CheckRequest) => Promise<{ data: IntegrationHealthCheck }>;

function setup(check: Check) {
	const checked = mock(check);
	const Controller = new Function(
		"$",
		"createMutation",
		"checkIntegrationHealthMutation",
		`${controllerCode}\nreturn GrafanaHealthCheck;`
	)(
		runtime,
		() => ({ mutateAsync: checked }),
		() => ({})
	) as new () => GrafanaHealthCheck;

	return { health: new Controller(), checked };
}

const installationWith = (userSettings: Record<string, unknown>) =>
	({
		id: "installation-1",
		attributes: { name: "grafana", userSettings },
	}) as unknown as IntegrationInstallation;

const installation = installationWith({ logs_data_source_uid: "loki" });

describe("Grafana connection check", () => {
	test("shows that the check worked", async () => {
		const { health, checked } = setup(async () => ({ data: { ok: true } }));

		await health.check(installation);

		expect(checked).toHaveBeenCalledWith({ path: { id: installation.id } });
		expect(health.resultFor(installation)).toEqual({ kind: "ok" });
		expect(health.checking).toBe(false);
	});

	test("shows the error a failed check returned", async () => {
		const message = "logs: no logs data source configured";
		const { health } = setup(async () => ({ data: { ok: false, error: message } }));

		await health.check(installation);

		expect(health.resultFor(installation)).toEqual({ kind: "failed", message });
	});

	test("shows a request that could not run", async () => {
		const error = { code: "forbidden", title: "Forbidden", status: 403, detail: "" } as const;
		const { health } = setup(async () => {
			throw error;
		});

		await health.check(installation);

		expect(health.resultFor(installation)).toEqual({ kind: "error", error });
	});

	test("a result is not shown once the settings change", async () => {
		const { health } = setup(async () => ({ data: { ok: true } }));

		await health.check(installation);

		const changed = installationWith({ logs_data_source_uid: "" });
		expect(health.resultFor(changed)).toBeUndefined();
	});

	test("replacing the token clears the result, and a check already running does not put its result back", async () => {
		let finish: (value: { data: IntegrationHealthCheck }) => void = () => {};
		const { health } = setup(
			() =>
				new Promise((resolve) => {
					finish = resolve;
				})
		);

		const running = health.check(installation);
		expect(health.checking).toBe(true);

		health.clear();
		expect(health.checking).toBe(false);

		finish({ data: { ok: true } });
		await running;

		expect(health.resultFor(installation)).toBeUndefined();
		expect(health.checking).toBe(false);
	});

	test("installing again clears a finished result, even for an installation with the same settings", async () => {
		const { health } = setup(async () => ({ data: { ok: true } }));

		await health.check(installation);
		health.clear();

		const reinstalled = { ...installation, id: "installation-2" } as IntegrationInstallation;
		expect(health.resultFor(installation)).toBeUndefined();
		expect(health.resultFor(reinstalled)).toBeUndefined();
	});
});

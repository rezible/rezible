import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";
import { SvelteMap } from "svelte/reactivity";

import type { ErrorModel, IntegrationInstallation } from "$lib/api";
import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

import { grafanaSettings, type GrafanaSettings } from "./grafanaSettings";

// Bun does not compile runes. Compile the real settings forms and drafts, and supply their external
// dependencies locally, avoiding module mocks that would affect other tests.
const compileRunes = async (relativePath: string) => {
	const source = await Bun.file(new URL(relativePath, import.meta.url)).text();
	const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
	const options = { filename: relativePath, generate: "client", dev: false } as const;
	const compiled = compileModule(javascript, options).js.code;
	return compiled.replace(/^import [^;]*;$/gm, "").replace(/^export /gm, "");
};

const runtime = await import(import.meta.resolve("svelte/internal/client"));
const draftCode = await compileRunes("../../../lib/settings-draft.svelte.ts");
const formsCode = await compileRunes("../../../lib/integrationSettingsForms.svelte.ts");

const { SettingsDraftList } = new Function("$", "SvelteMap", `${draftCode}\nreturn { SettingsDraftList };`)(
	runtime,
	SvelteMap
);

type UpdateSettings = (id: string, userSettings: Record<string, unknown>) => Promise<IntegrationInstallation>;

function setup(updateSettings: UpdateSettings, userSettings: Record<string, unknown> = {}) {
	const installation = {
		id: "installation-1",
		attributes: { name: "grafana", userSettings },
	} as unknown as IntegrationInstallation;

	const updated = mock(updateSettings);
	const integrations = { updateSettings: updated };
	// The forms read installations once; later updates arrive through save.
	const watch = (getter: () => unknown, callback: (value: unknown) => void) => callback(getter());
	const toast = { success: () => {} };

	const Forms = new Function(
		"$",
		"watch",
		"SvelteMap",
		"toast",
		"SettingsDraftList",
		"useIntegrationsController",
		`${formsCode}\nreturn IntegrationSettingsForms;`
	)(runtime, watch, SvelteMap, toast, SettingsDraftList, () => integrations) as new (
		...args: ConstructorParameters<typeof IntegrationSettingsForms<GrafanaSettings>>
	) => IntegrationSettingsForms<GrafanaSettings>;

	const forms = new Forms(grafanaSettings, () => [installation]);
	return { forms, installation, updated };
}

describe("Grafana settings", () => {
	test("missing settings show empty UIDs and the default service labels", () => {
		expect(grafanaSettings.read({})).toEqual({
			logsDataSourceUid: "",
			metricsDataSourceUid: "",
			logServiceLabel: "service_name",
			metricServiceLabel: "service_name",
		});
	});

	test("saving writes the backend's keys and keeps other settings", async () => {
		const { forms, installation, updated } = setup(
			async (_id, userSettings) => ({
				...installation,
				attributes: { ...installation.attributes, userSettings },
			}),
			{ other: true }
		);

		const draft = forms.draftFor(installation.id)!;
		draft.value.values.logsDataSourceUid = " loki ";
		draft.value.values.metricsDataSourceUid = "prometheus";
		draft.value.values.metricServiceLabel = "job";
		expect(draft.dirty).toBe(true);

		await forms.save(installation);

		expect(updated).toHaveBeenCalledWith(installation.id, {
			other: true,
			logs_data_source_uid: "loki",
			metrics_data_source_uid: "prometheus",
			log_service_label: "service_name",
			metric_service_label: "job",
		});
		expect(forms.saveError(installation.id)).toBeUndefined();
		expect(forms.draftFor(installation.id)!.dirty).toBe(false);
	});

	test("a rejected save shows its error and keeps the edit", async () => {
		const rejection: ErrorModel = {
			title: "Bad Request",
			status: 400,
			detail: `invalid input: logs_data_source_uid "a/b" is not a Grafana data source UID`,
		};
		const { forms, installation } = setup(async () => {
			throw rejection;
		});

		const draft = forms.draftFor(installation.id)!;
		draft.value.values.logsDataSourceUid = "a/b";

		await forms.save(installation);

		expect(forms.saveError(installation.id)).toEqual(rejection);
		expect(forms.draftFor(installation.id)!.value.values.logsDataSourceUid).toBe("a/b");
		expect(forms.draftFor(installation.id)!.dirty).toBe(true);
	});
});

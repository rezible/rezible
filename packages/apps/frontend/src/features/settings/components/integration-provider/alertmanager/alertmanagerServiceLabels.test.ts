import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";
import { SvelteMap } from "svelte/reactivity";

import type { ApiError, IntegrationInstallation } from "$lib/api";
import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

import { alertmanagerSettings, type AlertmanagerSettings } from "./alertmanagerSettings";

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

function setup(updateSettings: UpdateSettings) {
	const installation = {
		id: "installation-1",
		attributes: { name: "alertmanager", userSettings: { other: true } },
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
		...args: ConstructorParameters<typeof IntegrationSettingsForms<AlertmanagerSettings>>
	) => IntegrationSettingsForms<AlertmanagerSettings>;

	const forms = new Forms(alertmanagerSettings, () => [installation]);
	return { forms, installation, updated };
}

describe("alertmanager service labels form", () => {
	test("saving an edited list writes it and keeps other settings", async () => {
		const { forms, installation, updated } = setup(async (id, userSettings) => ({
			...installation,
			attributes: { ...installation.attributes, userSettings },
		}));

		const draft = forms.draftFor(installation.id)!;
		draft.value.values.serviceLabels = "job, service";
		expect(draft.dirty).toBe(true);

		await forms.save(installation);

		expect(updated).toHaveBeenCalledWith(installation.id, {
			other: true,
			service_labels: ["job", "service"],
		});
		expect(forms.saveError(installation.id)).toBeUndefined();
		expect(forms.draftFor(installation.id)!.dirty).toBe(false);
	});

	test("a list the server rejects shows its error and keeps the edit", async () => {
		const rejection: ApiError = {
			code: "invalid_input",
			title: "Bad Request",
			status: 400,
			detail: 'invalid user settings: service_labels contains "job" more than once',
		};
		const { forms, installation } = setup(async () => {
			throw rejection;
		});

		const draft = forms.draftFor(installation.id)!;
		draft.value.values.serviceLabels = "job";

		await forms.save(installation);

		expect(forms.saveError(installation.id)).toEqual(rejection);
		expect(forms.draftFor(installation.id)!.value.values.serviceLabels).toBe("job");
		expect(forms.draftFor(installation.id)!.dirty).toBe(true);
	});
});

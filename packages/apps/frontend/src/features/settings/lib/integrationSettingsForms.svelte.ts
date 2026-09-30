import type { ErrorModel, IntegrationInstallation } from "$lib/api";

import { watch, type Getter } from "runed";
import { SvelteMap } from "svelte/reactivity";
import { toast } from "svelte-sonner";
import { z } from "zod";

import { type SettingsDraft, SettingsDraftList } from "./settings-draft.svelte";
import { useIntegrationsController } from "./integrationsController.svelte";

export type IntegrationUserSettings = Record<string, unknown>;

// Maps between an installation's stored user settings and the values edited by its form.
export type IntegrationSettingsDefinition<T> = {
	schema: z.ZodType<T>;
	// Read saved settings, filling missing fields with defaults.
	read: (userSettings: IntegrationUserSettings) => T;
	// Settings writes replace the whole object, so keep saved fields the form does not edit.
	write: (values: T, userSettings: IntegrationUserSettings) => IntegrationUserSettings;
};

type SettingsFormItem<T> = {
	id: string;
	values: T;
};

export type IntegrationSettingsFieldErrors = Partial<Record<string, string>>;

// Keeps a separate settings draft for each installation.
export class IntegrationSettingsForms<T> {
	private integrations = useIntegrationsController();
	private definition: IntegrationSettingsDefinition<T>;
	private drafts = new SettingsDraftList<SettingsFormItem<T>>();
	private saveErrors = new SvelteMap<string, ErrorModel>();

	savingId = $state<string>();

	constructor(
		definition: IntegrationSettingsDefinition<T>,
		installations: Getter<IntegrationInstallation[]>
	) {
		this.definition = definition;

		watch(installations, (current) => {
			this.drafts.receive(current.map((installation) => this.toFormItem(installation)));
		});
	}

	private toFormItem(installation: IntegrationInstallation): SettingsFormItem<T> {
		const userSettings = installation.attributes.userSettings ?? {};
		return { id: installation.id, values: this.definition.read(userSettings) };
	}

	draftFor(id: string): SettingsDraft<SettingsFormItem<T>> | undefined {
		return this.drafts.get(id);
	}

	fieldErrors(id: string): IntegrationSettingsFieldErrors {
		const draft = this.draftFor(id);
		if (!draft) {
			return {};
		}

		const result = this.definition.schema.safeParse(draft.value.values);
		if (result.success) {
			return {};
		}

		const errors: IntegrationSettingsFieldErrors = {};
		for (const issue of result.error.issues) {
			const field = String(issue.path.at(0) ?? "");
			errors[field] ??= issue.message;
		}
		return errors;
	}

	saveError(id: string) {
		return this.saveErrors.get(id);
	}

	save = async (installation: IntegrationInstallation) => {
		const draft = this.draftFor(installation.id);
		if (!draft || this.savingId) return;

		const parsed = this.definition.schema.safeParse(draft.value.values);
		if (!parsed.success) return;

		const savedSettings = installation.attributes.userSettings ?? {};
		const userSettings = this.definition.write(parsed.data, savedSettings);

		this.savingId = installation.id;
		this.saveErrors.delete(installation.id);
		try {
			const updated = await this.integrations.updateSettings(installation.id, userSettings);
			this.drafts.accept(this.toFormItem(updated));
			toast.success("Settings saved.");
		} catch (e) {
			this.saveErrors.set(installation.id, e as ErrorModel);
		} finally {
			this.savingId = undefined;
		}
	};

	cancel = (id: string) => {
		this.draftFor(id)?.cancel();
		this.saveErrors.delete(id);
	};
}

export const readSettingsObject = (value: unknown): IntegrationUserSettings => {
	if (!value || typeof value !== "object" || Array.isArray(value)) {
		return {};
	}
	return value as IntegrationUserSettings;
};

export const readString = (value: unknown, fallback: string) => {
	if (typeof value === "string") {
		return value;
	}
	return fallback;
};

export const readBoolean = (value: unknown, fallback: boolean) => {
	if (typeof value === "boolean") {
		return value;
	}
	return fallback;
};

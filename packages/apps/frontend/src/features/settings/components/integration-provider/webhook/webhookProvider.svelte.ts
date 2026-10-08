import type { IntegrationProviderController } from "../controller.svelte";

import { defaultWebhookPreset, webhookIntegrationName, webhookPresets } from "./deploymentPreset";

type ProviderInstaller = Pick<IntegrationProviderController, "install">;

// Installs webhook integrations. Each installation accepts one preset, chosen here and fixed afterwards.
export class WebhookProviderController {
	private provider: ProviderInstaller;

	readonly presets = webhookPresets;
	preset = $state(defaultWebhookPreset);

	presetLabel = $derived(this.presets.find((preset) => preset.value === this.preset)?.label ?? "");

	constructor(provider: ProviderInstaller) {
		this.provider = provider;
	}

	install = async () => {
		await this.provider.install(webhookIntegrationName, { preset: this.preset });
	};
}

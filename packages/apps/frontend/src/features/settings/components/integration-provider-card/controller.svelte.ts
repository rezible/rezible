import { Context, watch, type Getter } from "runed";

import {
	type ErrorModel,
	type AvailableIntegration,
	type InstalledIntegration,
	requestIntegrationEventSyncMutation,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";

import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

export class IntegrationProviderCardController {
	integrations = useIntegrationsController();

	private provider = $state<string>();
	available = $derived<AvailableIntegration[]>(this.integrations.availableByProvider.get(this.provider || "") ?? []);

	constructor(providerFn: Getter<string>) {
		watch(providerFn, provider => {this.provider = provider});
	}

	single = $derived(this.available.length === 1 ? this.available.at(0) : undefined);
	
	// supportedCapabilities = $derived(this.integration?.supportedCapabilities ?? []);

	installations = $derived<Map<string, InstalledIntegration[]>>(
		new Map(this.available.map(a => [a.name, this.integrations.installationsByName.get(a.name) ?? []])));
	// maxInstallReached = $derived<Map<string, number>>(new Map(this.available.map(a => [a.name, ])
	// 	typeof this.integration?.maxInstalls === "number" && this.installations.length >= this.integration.maxInstalls
	// );
	// canInstall = $derived(!this.maxInstallsReached);

	// enabledCapabilities = $derived(getEnabledCapabilties(this.installations));
}

const ctx = new Context<IntegrationProviderCardController>("IntegrationProviderCardController");
export const initIntegrationProviderCardController = (provider: Getter<string>) =>
	ctx.set(new IntegrationProviderCardController(provider));
export const useAvailableIntegrationCardController = () => ctx.get();

import {
	configureIntegrationMutation,
	type ConfigureIntegrationRequestBody,
	type ConfiguredIntegration,
	type ErrorModel,
	listConfiguredIntegrationsOptions,
	listAvailableIntegrationsOptions,
} from "$lib/api";
import { useAuthSessionState } from "$src/lib/auth-session.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { SvelteMap } from "svelte/reactivity";

export class IntegrationsController {
	session = useAuthSessionState();

	inOAuthFlow = $state(false);

	private listAvailableQuery = createQuery(() => listAvailableIntegrationsOptions());
	available = $derived(this.listAvailableQuery.data?.data || []);

	private listConfiguredQuery = createQuery(() => listConfiguredIntegrationsOptions());
	configured = $derived(this.listConfiguredQuery.data?.data || []);
	configuredById = $derived(new SvelteMap(this.configured.map((intg) => [intg.id, intg])));
	configuredByProvider = $derived.by(() => {
		const grouped = new SvelteMap<string, ConfiguredIntegration[]>();
		for (const intg of this.configured) {
			const curr = grouped.get(intg.provider) ?? [];
			grouped.set(intg.provider, [...curr, intg]);
		}
		return grouped;
	});

	private configureMut = createMutation(() => ({
		...configureIntegrationMutation({}),
		onSuccess: () => {
			void this.listConfiguredQuery.refetch();
		},
	}));

	configuringProvider = $derived(this.configureMut.variables?.path?.provider ?? "");
	configuringError = $derived(this.configureMut.error?.detail || this.configureMut.error?.title || "");
	isConfiguring = $derived(this.configureMut.isPending);

	async configure(provider: string, attributes: ConfigureIntegrationRequestBody["attributes"]) {
		await this.configureMut.mutateAsync({
			path: { provider },
			body: { attributes },
		});
	}

	errorFor(provider: string) {
		if (this.configuringProvider !== provider) return "";
		return this.configuringError;
	}

	isSaving(provider: string) {
		if (!this.isConfiguring) return false;
		return this.configuringProvider === provider;
	}

	loading = $derived(this.listAvailableQuery.isPending || this.listConfiguredQuery.isPending);
	error = $derived((this.listAvailableQuery.error ?? this.listConfiguredQuery.error) as ErrorModel | null);
}

const ctx = new Context<IntegrationsController>("IntegrationsController");
export const initIntegrationsController = () => ctx.set(new IntegrationsController());
export const useIntegrationsController = () => ctx.get();

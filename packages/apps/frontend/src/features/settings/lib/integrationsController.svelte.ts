import { configureIntegrationMutation, type ConfigureIntegrationRequestBody, type ErrorModel, listConfiguredIntegrationsOptions, listAvailableIntegrationsOptions } from "$lib/api";
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
	configuredMap = $derived(new SvelteMap(this.configured.map((intg) => [intg.name, intg])));

	private configureMut = createMutation(() => ({
		...configureIntegrationMutation({}),
		onSuccess: () => {
			void this.listConfiguredQuery.refetch();
		},
	}));

	configuringName = $derived(this.configureMut.variables?.path?.name ?? "");
	configuringError = $derived(this.configureMut.error?.detail || this.configureMut.error?.title || "");
	isConfiguring = $derived(this.configureMut.isPending);

	async configure(name: string, attributes: ConfigureIntegrationRequestBody["attributes"]) {
		await this.configureMut.mutateAsync({
			path: { name },
			body: { attributes },
		});
	}

	errorFor(name: string) {
		if (this.configuringName !== name) return "";
		return this.configuringError;
	}

	isSaving(name: string) {
		if (!this.isConfiguring) return false;
		return this.configuringName === name;
	}

	loading = $derived(this.listAvailableQuery.isPending || this.listConfiguredQuery.isPending);
	error = $derived((this.listAvailableQuery.error ?? this.listConfiguredQuery.error) as ErrorModel | null);
}

const ctx = new Context<IntegrationsController>("IntegrationsController");
export const initIntegrationsController = () => ctx.set(new IntegrationsController());
export const useIntegrationsController = () => ctx.get();

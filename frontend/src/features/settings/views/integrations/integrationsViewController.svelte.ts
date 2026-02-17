import { configureIntegrationMutation, type ConfigureIntegrationRequestBody, type ErrorModel, listConfiguredIntegrationsOptions, listSupportedIntegrationsOptions, completeIntegrationOauthFlowMutation, startIntegrationOauthFlowMutation } from "$lib/api";
import { useAuthSessionState } from "$lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { watch } from "runed";
import { useSearchParams } from "runed/kit";
import { SvelteMap } from "svelte/reactivity";
import { z } from "zod";

const oauthCallbackParamsSchema = z.object({
	name: z.string().default(""),
	code: z.string().default(""),
	state: z.string().default(""),
});

class IntegrationOAuthController {
	private callbackParams = useSearchParams(oauthCallbackParamsSchema);
	private callbackName = $derived(this.callbackParams.name);
	private onCompleted: () => void;

	constructor(onCompleted: () => void) {
		this.onCompleted = onCompleted;
		watch(() => this.callbackName, (name) => {
			void this.onCallbackSet(name);
		});
	}

	private startFlowMut = createMutation(() => startIntegrationOauthFlowMutation({}));
	loadingFlowUrl = $derived(this.startFlowMut.isPending);
	startFlowErr = $derived(this.startFlowMut.error?.detail || this.startFlowMut.error?.title || "");

	async startFlow(name: string) {
		try {
			const resp = await this.startFlowMut.mutateAsync({ path: { name } });
			window.location.assign(new URL(resp.data.flow_url));
		} catch {
			// surfaced via startFlowErr
		}
	}

	private completeFlowMut = createMutation(() => completeIntegrationOauthFlowMutation({}));
	completingFlow = $derived(this.completeFlowMut.isPending);
	completeFlowErr = $derived(this.completeFlowMut.error?.detail || this.completeFlowMut.error?.title || "");

	private async onCallbackSet(name?: string) {
		if (!name || this.completingFlow) return;

		const { state, code } = $state.snapshot(this.callbackParams);
		this.callbackParams.reset();

		if (!state || !code) return;

		try {
			await this.completeFlowMut.mutateAsync({
				path: { name },
				body: { attributes: { state, code } },
			});
			this.onCompleted();
		} catch {
			// surfaced via completeFlowErr
		}
	}
}

export class IntegrationsViewController {
	session = useAuthSessionState();
	oauth: IntegrationOAuthController;

	constructor() {
		this.oauth = new IntegrationOAuthController(() => {
			void this.listConfiguredQuery.refetch();
		});
	}

	private listSupportedQuery = createQuery(() => listSupportedIntegrationsOptions());
	supported = $derived(this.listSupportedQuery.data?.data || []);

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

	loading = $derived(this.listSupportedQuery.isPending || this.listConfiguredQuery.isPending);
	queryErr = $derived((this.listSupportedQuery.error ?? this.listConfiguredQuery.error) as ErrorModel | null);
	queryErrorMessage = $derived(this.queryErr?.detail || this.queryErr?.title || "");
}

const ctx = new Context<IntegrationsViewController>("IntegrationsViewController");
export const initIntegrationsViewController = () => ctx.set(new IntegrationsViewController());
export const useIntegrationsViewController = () => ctx.get();

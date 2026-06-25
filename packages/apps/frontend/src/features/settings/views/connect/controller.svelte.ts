import { page } from "$app/state";
import {
	completeIntegrationOauthFlowMutation,
	type ErrorModel,
	type IntegrationOAuthInstallResult,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { onMount } from "svelte";

import { postIntegrationOAuthCompleteMessage } from "$features/settings/lib/integrationsOAuthController.svelte";

const missingParamsError: ErrorModel = {
	title: "Integration Setup Failed",
	detail: "The OAuth provider did not return the required code and state parameters.",
	status: 400,
};

export class ConnectIntegrationController {
	private completeMut = createMutation(() => completeIntegrationOauthFlowMutation({}));
	private currentName = $state.raw("");

	name = $derived(this.currentName);

	constructor(nameFn: Getter<string>) {
		watch(nameFn, name => {
			this.currentName = name;
		});
		onMount(() => {
			this.complete();
		});
	}

	private finish(result?: IntegrationOAuthInstallResult, error?: ErrorModel) {
		postIntegrationOAuthCompleteMessage({ name: this.name, result, error })
		setTimeout(() => window.close(), 50);
	}

	async complete() {
		const code = page.url.searchParams.get("code");
		const state = page.url.searchParams.get("state");

		if (!code || !state) {
			this.finish(undefined, missingParamsError);
			return;
		}

		try {
			const resp = await this.completeMut.mutateAsync({
				path: { name: this.name },
				body: { attributes: { code, state } },
			});
			this.finish(resp.data);
		} catch (e) {
			this.finish(undefined, e as ErrorModel);
		} finally {
			this.finish();
		}
	}

	close() {
		window.close();
	}
}

const ctx = new Context<ConnectIntegrationController>("ConnectIntegrationController");
export const initConnectIntegrationController = (nameFn: Getter<string>) =>
	ctx.set(new ConnectIntegrationController(nameFn));
export const useConnectIntegrationController = () => ctx.get();

import { page } from "$app/state";
import {
	completeIntegrationOauthFlowMutation,
	type ErrorModel,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { onMount } from "svelte";

import { postIntegrationOAuthCompleteMessage } from "$features/settings/lib/integrationsOAuthController.svelte";

export class ConnectIntegrationController {
	private completeMut = createMutation(() => completeIntegrationOauthFlowMutation({}));
	private currentName = $state("");

	done = $state(false);
	notifiedOpener = $state(false);
	error = $state<ErrorModel>();
	name = $derived(this.currentName);

	constructor(nameFn: Getter<string>) {
		watch(nameFn, name => {
			this.currentName = name;
		});
		onMount(() => {
			this.complete();
		});
	}

	private missingParamsError(): ErrorModel {
		return {
			title: "Integration Setup Failed",
			detail: "The OAuth provider did not return the required code and state parameters.",
			status: 400,
		};
	}

	private notifyError(error: ErrorModel) {
		this.error = error;
		this.notifiedOpener = postIntegrationOAuthCompleteMessage({ name: this.name, error });
	}

	private finish() {
		this.done = true;
		setTimeout(() => window.close(), 750);
	}

	async complete() {
		const code = page.url.searchParams.get("code");
		const state = page.url.searchParams.get("state");

		if (!code || !state) {
			this.notifyError(this.missingParamsError());
			this.finish();
			return;
		}

		try {
			const resp = await this.completeMut.mutateAsync({
				path: { name: this.name },
				body: { attributes: { code, state } },
			});
			this.notifiedOpener = postIntegrationOAuthCompleteMessage({
				name: this.name,
				result: resp.data,
			});
		} catch (e) {
			this.notifyError(e as ErrorModel);
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

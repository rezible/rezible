import {
	startIntegrationOauthFlowMutation,
	type ApiError,
	type IntegrationOAuthInstallResult,
} from "$lib/api";

import { browser } from "$app/environment";
import { createMutation } from "@tanstack/svelte-query";
import { onDestroy } from "svelte";

const OAuthMessageType = "rezible.integration-oauth-complete";

type IntegrationOAuthMessage = {
	type: typeof OAuthMessageType;
	name: string;
	result?: IntegrationOAuthInstallResult;
	error?: ApiError;
};

const popupBlockedError: ApiError = {
	code: "unprocessable",
	detail: "Allow pop-ups for Rezible and try again.",
};

export const postIntegrationOAuthCompleteMessage = (message: Omit<IntegrationOAuthMessage, "type">) => {
	if (!browser || !window.opener) return false;
	window.opener.postMessage({ type: OAuthMessageType, ...message }, window.location.origin);
	return true;
};

type OAuthSuccessHandler = (name: string, result: IntegrationOAuthInstallResult) => void;

// The same integration can be installed from several buttons, such as one per Slack workspace.
// The origin identifies which button started a flow, so only that button shows its progress and errors.
export class IntegrationOAuthController {
	pendingName = $state<string>();
	private pendingOrigin = $state<string>();
	inFlow = $derived(!!this.pendingName);

	private failedOrigin = $state<string>();
	private error = $state.raw<ApiError>();

	private popup: Window | undefined;
	private stopPopupCloseCheck: VoidFunction | undefined;
	private onSuccess: OAuthSuccessHandler;

	private startOAuthFlowMut = createMutation(() => startIntegrationOauthFlowMutation());

	constructor(onSuccess: OAuthSuccessHandler) {
		this.onSuccess = onSuccess;

		if (browser) {
			window.addEventListener("message", this.handleOAuthMessage);
			onDestroy(() => {
				window.removeEventListener("message", this.handleOAuthMessage);
				this.endFlow();
			});
		}
	}

	isPending(origin: string) {
		return this.pendingOrigin === origin;
	}

	errorFor(origin: string) {
		if (this.failedOrigin !== origin) {
			return undefined;
		}
		return this.error;
	}

	clearError = () => {
		this.failedOrigin = undefined;
		this.error = undefined;
	};

	startFlowFor = async (name: string, origin = name) => {
		if (!browser || this.inFlow) return;
		this.clearError();

		// Open the popup synchronously with the click so browsers do not block it.
		const popup = window.open("about:blank", `rezible-oauth-${name}`, "popup,width=640,height=760");
		if (!popup) {
			this.fail(origin, popupBlockedError);
			return;
		}

		this.pendingName = name;
		this.pendingOrigin = origin;
		this.popup = popup;
		this.watchPopupClosed();

		try {
			const resp = await this.startOAuthFlowMut.mutateAsync({ path: { name } });
			popup.location.assign(new URL(resp.data.flow_url));
		} catch (e) {
			popup.close();
			this.endFlow();
			this.fail(origin, e as ApiError);
		}
	};

	private fail(origin: string, error: ApiError) {
		this.failedOrigin = origin;
		this.error = error;
	}

	private endFlow() {
		this.pendingName = undefined;
		this.pendingOrigin = undefined;
		this.popup = undefined;
		this.stopPopupCloseCheck?.();
		this.stopPopupCloseCheck = undefined;
	}

	// Closing the popup without finishing sign-in cancels the flow without an error.
	private watchPopupClosed() {
		const check = setInterval(() => {
			if (this.popup?.closed) {
				this.endFlow();
			}
		}, 500);
		this.stopPopupCloseCheck = () => clearInterval(check);
	}

	private handleOAuthMessage = (event: MessageEvent<IntegrationOAuthMessage>) => {
		if (event.origin !== window.location.origin) return;
		if (event.data?.type !== OAuthMessageType) return;

		const name = this.pendingName;
		const origin = this.pendingOrigin;
		if (!name || !origin || event.data.name !== name) return;

		this.endFlow();
		if (event.data.error) {
			this.fail(origin, event.data.error);
		} else if (event.data.result) {
			this.onSuccess(name, event.data.result);
		}
	};
}

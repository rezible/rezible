import {
	startIntegrationOauthFlowMutation, 
	type ErrorModel,
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
    error?: ErrorModel;
};

export const postIntegrationOAuthCompleteMessage = (message: Omit<IntegrationOAuthMessage, "type">) => {
    if (!browser || !window.opener) return false;
    window.opener.postMessage({ type: OAuthMessageType, ...message }, window.location.origin);
    return true;
};

export class IntegrationOAuthController {
    inFlowForName = $state<string>();
    error = $state<ErrorModel>();
    popup = $state.raw<Window>();
    private clearPopupInterval: VoidFunction | undefined;
    private onSuccess: (res: IntegrationOAuthInstallResult) => void;
    private onCancel: VoidFunction | undefined;

    constructor(onSuccess: (res: IntegrationOAuthInstallResult) => void, onCancel?: () => void) {
        this.onSuccess = onSuccess;
        this.onCancel = onCancel;
        if (browser) {
            window.addEventListener("message", this.handleOAuthMessage);
            onDestroy(() => window.removeEventListener("message", this.handleOAuthMessage));
        }
    }

    private setError(err: unknown) {
        this.error = {
            title: "Integration Setup Failed",
            detail: err instanceof Error ? err.message : "An unknown issue occurred",
        };
    };

    clearPopup() {
        this.inFlowForName = undefined;
        this.error = undefined;
        this.popup = undefined;
        this.clearPopupInterval?.();
    }

    createPopupCloseCheck() {
        const check = setInterval(() => {
            if (!!this && this.popup?.closed) {
                if (!!this.inFlowForName) {
                    this.onCancel?.();
                }
                this.clearPopup();
            }
        }, 500);
        this.clearPopupInterval = () => {
            clearInterval(check);
            if (this) this.clearPopupInterval = undefined;
        };
    }

    private startOAuthFlowMut = createMutation(() => ({
        ...startIntegrationOauthFlowMutation({}),
    }));

    async startFlowFor(name: string) {
        if (!browser) return;
        this.clearPopup();

        this.inFlowForName = name;

        const popup = window.open("about:blank", `rezible-oauth-${name}`, "popup,width=640,height=760");
        if (!popup) {
            this.setError(new Error("Popup blocked. Allow popups and try again."));
            this.inFlowForName = undefined;
            return;
        }

        this.popup = popup;
        this.createPopupCloseCheck();
        try {
            const resp = await this.startOAuthFlowMut.mutateAsync({path: { name }});
            popup.location.assign(new URL(resp.data.flow_url));
        } catch (e) {
            popup.close();
            this.setError(e);
            this.inFlowForName = undefined;
            this.popup = undefined;
        }
    }

    private handleOAuthMessage = (event: MessageEvent<IntegrationOAuthMessage>) => {
        if (event.origin !== window.location.origin) return;
        if (event.data?.type !== OAuthMessageType) return;
        if (this.inFlowForName && event.data.name !== this.inFlowForName) return;

        this.clearPopup();
        if (event.data.error) {
            this.error = event.data.error;
        } else if (event.data.result) {
            this.onSuccess?.(event.data.result);
        }
    }

    inFlow = $derived(this.startOAuthFlowMut.isPending || !!this.inFlowForName);
};

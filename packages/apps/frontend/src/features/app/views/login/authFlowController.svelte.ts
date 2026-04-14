import { createMutation } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { useAuthSessionState } from "$lib/auth.svelte";
import { OidcClient, WebStorageStateStore } from "oidc-client-ts";
// import { completeAuthSessionFlowMutation, type AuthSessionConfig, type CompleteAuthSessionFlowRequestAttributes, type ErrorModel } from "$lib/api";
import type { ErrorModel } from "$lib/api";
import { resolve } from "$app/paths";
import { watch } from "runed";
import z from "zod";
import { useSearchParams } from "runed/kit";

const transformFlowCompletionError = (err: ErrorModel): ErrorModel => {
    const title = "Login Failed";
    let detail = err.detail;
    if (err.detail === "domain_not_allowed") {
        detail = "Signup is not currently available for your domain";
    }
    return { title, detail };
};

const oidcStateStore = new WebStorageStateStore({ 
    prefix: "rez_auth.",
    store: window.sessionStorage, 
});

export const callbackParamsSchema = z.object({
	code: z.string().default(""),
	error: z.string().default(""),
	error_description: z.string().default(""),
});

export class AuthFlowController {
    private authSession = useAuthSessionState();

    private inFlow = $state(false);
    error = $state<ErrorModel>();

    private client: OidcClient;
    private callbackParams = useSearchParams(callbackParamsSchema);
    private callbackCode = $derived(this.callbackParams.code);
    private callbackError = $derived(this.callbackParams.error);
    private isCallback = $derived(!!this.callbackCode || !!this.callbackError);

    constructor(cfg: AuthSessionConfig) {
        const redirectUri = `https://${page.url.host}${resolve("/login/callback")}`
        this.client = new OidcClient({
            authority: cfg.issuer,
            client_id: cfg.app_client_id,
            scope: cfg.app_client_scopes.join(" "),
            response_type: "code",
            redirect_uri: redirectUri,
            stateStore: oidcStateStore,
        });

        watch(() => this.isCallback, isCallback => {
            if (!this.inFlow && isCallback) this.onCallbackParamsSet();
        })
    }

    private setError(err: unknown) {
        this.inFlow = false;
        this.error = {
            title: "Login Error",
            detail: (err instanceof Error) ? err.message : "An unknown issue occurred",
        };
    }
    clearError() {
        this.inFlow = false;
        this.error = undefined;
    }

    async doSignIn() {
        this.inFlow = true;
        this.client.createSigninRequest({})
            .then(({url}) => window.location.assign(url))
            .catch(err => {
                console.log("creating sign in request", err);
                this.setError(err);
            });
    }

    private completeAuthFlowMut = createMutation(() => ({
        ...completeAuthSessionFlowMutation(),
        onMutate: () => {
            this.inFlow = true;
        },
        onSuccess: () => {
            this.clearError();
            this.authSession.refetch();
        },
        onError: (err) => {
            this.setError(transformFlowCompletionError(err));
        },
    }));

    private async getCompleteFlowAttributes() {
        const { response, state } = await this.client.readSigninResponseState(page.url.toString());
        const code = response.code;
        const verifier = state.code_verifier;
        if (!code || !verifier) {
            throw new Error("missing sign in response params");
        }
        return {code, verifier} as CompleteAuthSessionFlowRequestAttributes
    }

    private async onCallbackParamsSet() {
        if (!!this.callbackParams.code) {
            await this.onSuccessCallback();
        } else if (!!this.callbackParams.error) {
            this.onErrorCallback();
        }
    }
    
    private async onSuccessCallback() {
        this.inFlow = true;
        try {
            const attributes = await this.getCompleteFlowAttributes();
            this.completeAuthFlowMut.mutateAsync({body: {attributes}});
        } catch (err) {
            this.setError(err);
        }
    }

    private onErrorCallback() {
        const error = this.callbackParams.error || "unknown";
        const description = this.callbackParams.error_description || "unknown";
        console.log("auth callback error", {error, description});
        this.setError(new Error("Auth provider responded with an error"));
        this.callbackParams.reset();
    }

    loading = $derived(this.inFlow);
}

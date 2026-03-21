import { Context, watch } from "runed";
import { useAuthSessionState, type SessionErrorCategory } from "$lib/auth.svelte";
import { OidcClient, WebStorageStateStore, type CreateSigninRequestArgs } from "oidc-client-ts";
import { APP_URL, AUTH_ISSUER_URL, AUTH_CLIENT_ID } from "$lib/config";
import { page } from "$app/state";
import { createMutation } from "@tanstack/svelte-query";
import { completeAuthSessionFlowMutation } from "$src/lib/api";
import { z } from "zod";
import { useSearchParams } from "runed/kit";

const errorDisplayText: Record<SessionErrorCategory, string> = {
    unknown: "An unknown error occurred",
    invalid: "Auth session is invalid",
    session_expired: "Your session has expired",
    invalid_user: "You signed in successfully, but Rezible does not have your details.",
    no_session: "",
};

const oidcCallbackParamsSchema = z.object({
    code: z.string().default(""),
    state: z.string().default(""),
});

const scopes = [
    "openid",
    "offline_access",
    "profile",
    "email",
    "groups",
];

export class AuthScreenController {
    private session = useAuthSessionState();
    private callbackParams = useSearchParams(oidcCallbackParamsSchema);
    private callbackCode = $derived(this.callbackParams.code);

    oidc = new OidcClient({
        authority: AUTH_ISSUER_URL,
        client_id: AUTH_CLIENT_ID,
        redirect_uri: `${APP_URL}/auth/callback`,
        response_type: "code",
        scope: scopes.join(" "),
        stateStore: new WebStorageStateStore({ 
            prefix: "rez_auth.",
            store: window.sessionStorage, 
        })
    });

	errorCategory = $derived(this.session.error?.category);
    showError = $derived(!!this.session.error && this.errorCategory !== "no_session");
    showLogoutButton = $derived(this.errorCategory === "invalid_user");
    errorText = $derived(errorDisplayText[this.errorCategory ?? "unknown"]);

    constructor() {
        watch(() => this.callbackCode, code => {this.onCallbackParamsSet(page.url)});
    }

    startingFlow = $state(false);
    async startLoginFlow() {
        this.startingFlow = true;
        const opts: CreateSigninRequestArgs = {};
        try {
            const req = await this.oidc.createSigninRequest(opts);
            window.location.assign(req.url);
        } finally {
            this.startingFlow = false;
        }
    }

    private completeMut = createMutation(() => ({
        ...completeAuthSessionFlowMutation(),
        onSuccess: () => {this.session.refetch()}
    }));
    responseStateError = $state<string>();
    readingResponseState = $state(false);
    async onCallbackParamsSet(url: URL) {
        if (!this.callbackCode) return;

        this.readingResponseState = true;
        this.responseStateError = undefined;

        const urlStr = url.toString();
        this.callbackParams.reset();

        let code = "";
        let verifier = "";
        try {
            const { state, response } = await this.oidc.readSigninResponseState(urlStr, true);
            if (response.code && state.code_verifier) {
                code = response.code;
                verifier = state.code_verifier;
            } else {
                this.responseStateError = "missing code or verifier";
            }
        } finally {
            this.readingResponseState = false;
        }
        await this.oidc.clearStaleState();

        if (!!code && !!verifier) {
            this.completeMut.mutate({body: {attributes: {code, verifier}}});
        }
    }

    loading = $derived(this.startingFlow || this.readingResponseState || this.completeMut.isPending);
}

const ctx = new Context<AuthScreenController>("AuthScreenController");
export const initAuthScreenController = () => ctx.set(new AuthScreenController());
export const useAuthScreenController = () => ctx.get();
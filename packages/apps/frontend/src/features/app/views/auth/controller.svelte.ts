import { Context, watch } from "runed";
import { z } from "zod";
import { useSearchParams } from "runed/kit";
import { createMutation } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { useAuthSessionState, AuthSessionErrorCategory } from "$lib/auth.svelte";
import { OidcClient, WebStorageStateStore } from "oidc-client-ts";
import { AUTH_OIDC_ISSUER_PATH, AUTH_OIDC_CLIENT_ID, AUTH_OIDC_CLIENT_SCOPES, APP_AUTH_ROUTE_BASE } from "$lib/config";
import { completeAuthSessionFlowMutation, type ErrorModel } from "$lib/api";

const authSessionErrorDisplayText: Record<AuthSessionErrorCategory, string> = {
    [AuthSessionErrorCategory.NoSession]: "",
    [AuthSessionErrorCategory.SessionExpired]: "Your session has expired",
    [AuthSessionErrorCategory.SessionInvalid]: "Your session is invalid",
    [AuthSessionErrorCategory.ServerError]: "Something went wrong while authenticating you",
    [AuthSessionErrorCategory.Unknown]: "Something went wrong while authenticating you",
};

const transformFlowCompletionError = (err: ErrorModel): ErrorModel => {
    const title = "Login Failed";
    let detail = err.detail;
    if (err.detail === "domain_not_allowed") {
        detail = "Signup is not currently available for your domain";
    }
    return { title, detail };
}

const makeOidcClient = (): OidcClient => {
    const appUrl = `https://${window.location.host}`;
    return new OidcClient({
        authority: AUTH_OIDC_ISSUER_PATH,
        client_id: AUTH_OIDC_CLIENT_ID,
        scope: AUTH_OIDC_CLIENT_SCOPES,
        redirect_uri: `${appUrl}${APP_AUTH_ROUTE_BASE}/callback`,
        response_type: "code",
        stateStore: new WebStorageStateStore({ 
            prefix: "rez_auth.",
            store: window.sessionStorage, 
        })
    });
}

export class AuthViewController {
    private callbackParams = useSearchParams(z.object({
        code: z.string().default(""),
        state: z.string().default(""),
    }));
    private callbackCode = $derived(this.callbackParams.code);
    private authSession = useAuthSessionState();

    private oidcClient = makeOidcClient();
    
    private inAuthFlow = $state(false);
    showLogoutButton = $derived(this.authSession.error === AuthSessionErrorCategory.SessionInvalid);
    authSessionError = $state<ErrorModel>();
    authFlowErr = $state<ErrorModel>();

    private completeAuthFlowMut = createMutation(() => ({
        ...completeAuthSessionFlowMutation(),
        onSuccess: () => {
            this.authSession.refetch();
        },
        onError: (err) => {
            this.inAuthFlow = false;
            this.authFlowErr = transformFlowCompletionError(err);
        },
    }));

    loading = $derived(this.inAuthFlow || this.completeAuthFlowMut.isPending);

    constructor() {
        watch(() => this.callbackCode, _ => {this.onCallbackParamsSet()});
        watch(() => this.authSession.error, c => {this.onAuthSessionError(c)});
    }

    async doSignOut() {
        await this.authSession.logout();
    }

    async doSignIn() {
        this.inAuthFlow = true;
        try {
            const req = await this.oidcClient.createSigninRequest({});
            window.location.assign(req.url);
        } catch (err) {
            this.inAuthFlow = false;
            console.error(err);
        }
    }

    private async onAuthSessionError(cat?: AuthSessionErrorCategory) {
        if (!cat || cat == AuthSessionErrorCategory.NoSession) {
            this.authSessionError = undefined;
        } else {
            this.authSessionError = {
                title: "Auth Session Invalid",
                detail: authSessionErrorDisplayText[cat] || "Unknown",
            };
        }
    }

    private async onCallbackParamsSet() {
        if (!this.callbackCode) return;

        const urlStr = page.url.toString();

        this.callbackParams.reset();
        this.inAuthFlow = true;

        try {
            const { state, response } = await this.oidcClient.readSigninResponseState(urlStr, true);
            if (!response.code || !state.code_verifier) {
                throw new Error(`missing code`);
            }
            const attributes = {code: response.code, verifier: state.code_verifier};
            this.completeAuthFlowMut.mutate({body: {attributes}});
        } catch (err) {
            this.inAuthFlow = false;
            let detail = "Unknown Issue";
            if (err instanceof Error) detail = err.message;
            this.authFlowErr = {title: "Invalid Auth Response", detail};
        } finally {
            await this.oidcClient.clearStaleState();
        }
    }
}

const ctx = new Context<AuthViewController>("AuthViewController");
export const initAuthViewController = () => ctx.set(new AuthViewController());
export const useAuthViewController = () => ctx.get();
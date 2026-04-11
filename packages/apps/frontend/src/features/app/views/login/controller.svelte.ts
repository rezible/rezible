import { Context, watch } from "runed";
import { createMutation } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { useAuthSessionState, AuthSessionErrorCategory } from "$lib/auth.svelte";
import { OidcClient, WebStorageStateStore } from "oidc-client-ts";
import { AUTH_ISSUER, AUTH_OIDC_CLIENT_ID, AUTH_OIDC_CLIENT_SCOPES, APP_LOGIN_ROUTE } from "$lib/config";
import { completeAuthSessionFlowMutation, type ErrorModel } from "$lib/api";
import { replaceState } from "$app/navigation";

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
};

const translateAuthSessionError = (cat?: AuthSessionErrorCategory) => {
    if (!cat || cat === AuthSessionErrorCategory.NoSession) return;
    return {
        title: "Auth Session Invalid",
        detail: authSessionErrorDisplayText[cat] || "Unknown",
    };
}

const makeOidcClient = (): OidcClient => {
    const host = `https://${window.location.host}`;
    const authority = AUTH_ISSUER.startsWith("/") ? `${host}${AUTH_ISSUER}` : AUTH_ISSUER;
    return new OidcClient({
        authority,
        client_id: AUTH_OIDC_CLIENT_ID,
        scope: AUTH_OIDC_CLIENT_SCOPES,
        redirect_uri: `${host}/${APP_LOGIN_ROUTE}/callback`,
        response_type: "code",
        stateStore: new WebStorageStateStore({ 
            prefix: "rez_auth.",
            store: window.sessionStorage, 
        }),
    });
};

export class LoginViewController {
    private pageUrl = $derived(page.url);
    private authSession = useAuthSessionState();
    private oidcClient = makeOidcClient();
    private inAuthFlow = $state(false);

    showLogoutButton = $derived(this.authSession.error === AuthSessionErrorCategory.SessionInvalid);
    authSessionError = $state<ErrorModel>();
    authFlowErr = $state<ErrorModel>();

    private setAuthFlowError(err?: unknown) {
        this.inAuthFlow = false;
        if (err === undefined) {
            this.authFlowErr = err;
        } else {
            this.authFlowErr = {
                title: "Login Failed",
                detail: (err instanceof Error) ? err.message : "An unknown issue occurred",
            };
        }
    }

    private completeAuthFlowMut = createMutation(() => ({
        ...completeAuthSessionFlowMutation(),
        onSuccess: () => {
            this.authSession.refetch();
        },
        onError: (err) => {
            this.setAuthFlowError(transformFlowCompletionError(err));
        },
    }));

    loading = $derived(this.inAuthFlow || this.completeAuthFlowMut.isPending);

    constructor() {
        watch(() => this.authSession.error, c => {
            this.authSessionError = translateAuthSessionError(c);
        });
        watch(() => this.pageUrl, () => {
            this.onPageUrlChanged();
        });
    }

    async doSignOut() {
        await this.authSession.logout();
    }

    async doSignIn() {
        this.inAuthFlow = true;
        this.oidcClient.createSigninRequest({})
            .then(({url}) => window.location.assign(url))
            .catch(err => this.setAuthFlowError(err));
    }

    private async onPageUrlChanged() {
        this.inAuthFlow = true;
        let callbackErr: unknown | undefined = undefined;
        const url = new URL(window.location.href);
        try {
            if (url.searchParams.has("code")) {
                await this.onLoginSuccessCallback(url);
            } else if (url.searchParams.has("error")) {
                this.onLoginErrorCallback(url);
            }
        } catch (err) {
            callbackErr = err;
        } finally {
            this.setAuthFlowError(callbackErr);
            await this.oidcClient.clearStaleState();
            url.search = '';
            replaceState(url.toString(), page.state);
        }
    }
    
    private async onLoginSuccessCallback(url: URL) {
        const { state, response } = await this.oidcClient.readSigninResponseState(url.toString(), true);
        const code = response.code;
        const verifier = state.code_verifier;
        if (!code || !verifier) throw new Error(`missing ${!code ? "code" : "code_verifier"}`);
        await this.completeAuthFlowMut.mutateAsync({body: {attributes: {code, verifier}}});
    }

    private onLoginErrorCallback(url: URL) {
        const error = url.searchParams.get("error") || "unknown";
        const description = url.searchParams.get("error_description") || "unknown";
        console.log("auth callback error", {error, description});
        throw new Error(`Auth provider responded with an error`);
    }
}

const ctx = new Context<LoginViewController>("LoginViewController");
export const initLoginViewController = () => ctx.set(new LoginViewController());
export const useLoginViewController = () => ctx.get();
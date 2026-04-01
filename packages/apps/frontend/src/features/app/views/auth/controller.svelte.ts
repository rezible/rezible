import { Context, watch } from "runed";
import { z } from "zod";
import { useSearchParams } from "runed/kit";
import { createMutation } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { useAuthSessionState, type AuthSessionError, type AuthSessionErrorCategory } from "$lib/auth.svelte";
import { OidcClient, WebStorageStateStore, type CreateSigninRequestArgs } from "oidc-client-ts";
import { AUTH_OIDC_ISSUER_PATH, AUTH_OIDC_CLIENT_ID, AUTH_OIDC_CLIENT_SCOPES, AUTH_OIDC_CLIENT_REDIRECT_PATH } from "$lib/config";
import { completeAuthSessionFlowMutation, type ErrorModel } from "$lib/api";

const authSessionErrorDisplayText = new Map<AuthSessionErrorCategory, string>([
    ["unknown", "An unknown error occurred"],
    ["invalid", "Auth session is invalid"],
    ["session_expired", "Your session has expired"],
    ["invalid_user", "You signed in successfully, but Rezible does not have your details."],
    // ["no_session", ""],
]);

const showLogoutButtonErrorCategories = new Set<AuthSessionErrorCategory>(["invalid_user"]);

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
        authority: `${appUrl}${AUTH_OIDC_ISSUER_PATH}`,
        client_id: AUTH_OIDC_CLIENT_ID,
        scope: AUTH_OIDC_CLIENT_SCOPES,
        redirect_uri: `${appUrl}${AUTH_OIDC_CLIENT_REDIRECT_PATH}`,
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
    private authSessionErrorCategory = $derived(this.authSession.error?.category);

    private oidcClient = makeOidcClient();
    
    private inAuthFlow = $state(false);
    showLogoutButton = $derived(!!this.authSessionErrorCategory 
        && showLogoutButtonErrorCategories.has(this.authSessionErrorCategory));
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
        watch(() => this.authSessionErrorCategory, c => {this.onAuthSessionError(c)});
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
        if (!cat || cat == "no_session") {
            this.authSessionError = undefined;
        } else {
            this.authSessionError = {
                title: "Auth Session Invalid",
                detail: authSessionErrorDisplayText.get(cat) || "Unknown",
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
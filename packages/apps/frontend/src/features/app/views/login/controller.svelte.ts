import z from "zod";
import { goto } from "$app/navigation";
import { useAuthSessionState, AuthSessionErrorCategory } from "$lib/auth.svelte";
import type { ErrorModel } from "$lib/api";
import { page } from "$app/state";

const authSessionErrorDisplayText: Record<AuthSessionErrorCategory, string> = {
    [AuthSessionErrorCategory.NoSession]: "",
    [AuthSessionErrorCategory.SessionExpired]: "Your session has expired",
    [AuthSessionErrorCategory.SessionInvalid]: "Your session is invalid",
    [AuthSessionErrorCategory.ServerError]: "Something went wrong while authenticating you",
    [AuthSessionErrorCategory.Unknown]: "Something went wrong while authenticating you",
};
const transformAuthSessionError = (cat?: AuthSessionErrorCategory) => {
    if (!cat || cat === AuthSessionErrorCategory.NoSession) return;
    const title = "Auth Session Invalid";
    const detail = authSessionErrorDisplayText[cat] || "Unknown";
    return { title, detail } as ErrorModel;
};

const loginErrorDisplayText: Record<string, string> = {
    ["create_redirect"]: "Failed to redirect to identity provider",
    ["write_auth_session"]: "Failed to write auth session",
    ["write_auth_state"]: "Failed to write auth state",
    ["read_auth_state"]: "Failed to read auth state",
    ["callback_exchange"]: "Failed to perform callback exchange with identity provider",
    ["identity_sync"]: "Failed to sync user & organization information",
};
const transformLoginErrorCode = (code: string | null) => {
    if (!code) return;
    const title = "Login Error";
    const detail = loginErrorDisplayText[code] || "An unknown problem occurred";
    return {title, detail} as ErrorModel;
};

export class LoginViewController {
    private session = useAuthSessionState();

    loaded = $state(false);
	inFlow = $state(false);

	authSessionError = $derived(transformAuthSessionError(this.session.error));
	showLogout = $derived(this.session.error === AuthSessionErrorCategory.SessionInvalid);

	loginError = $state<ErrorModel>();
    constructor() {
        const params = page.url.searchParams;

        this.loginError = transformLoginErrorCode(params.get("error"));

        if (!this.loginError && params.has("flow", "true")) {
            this.doLogin();
        } else {
            goto(window.location.pathname, { replaceState: true, noScroll: true })
                .then(() => {this.loaded = true});
        }
    }

    async doLogin() {
        this.inFlow = true;
        await goto("/api/auth/login");
    }

    async doLogout() {
        this.inFlow = true;
        await this.session.logout();
    }

	titleText = $derived("Authentication Required");
	descriptionText = $derived("Continue with your identity provider");
}
import {
	type GetCurrentAuthSessionResponse,
	type User,
	getCurrentAuthSessionOptions,
	type ErrorModel,
	type Organization,
} from "$lib/api";
import { APP_URL, AUTH_URL, AUTH_APP_CLIENT_ID } from "$lib/config";
import { UserManager, WebStorageStateStore, User as OidcUser } from "oidc-client-ts";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { onMount } from "svelte";

export type SessionErrorCategory = "unknown" | "invalid" | "session_expired" | "no_session" | "invalid_user";

export type SessionError = {
	category: SessionErrorCategory;
	code?: string;
};

const OIDC_CONFIG = {
    authority: AUTH_URL,
    client_id: AUTH_APP_CLIENT_ID,
    redirect_uri: APP_URL + "/auth/callback/login",
    post_logout_redirect_uri: APP_URL + "/auth/callback/logout",
    response_type: 'code',
    scope: 'openid profile email',
    code_challenge_method: 'S256'
};

export class AuthSessionState {
	private oidc: UserManager;
	constructor() {
		this.oidc = new UserManager({
			userStore: new WebStorageStateStore({ store: window.localStorage }),
			...OIDC_CONFIG,
		});
	}

	loaded = $derived(true);
	user = $state<OidcUser | null>(null);
	isAuthenticated = $derived(!!this.user);
	isSetup = $derived(false);

	async onSignInCallback(url?: string) {
		const user = await this.oidc.signinCallback(url);
		if (user) this.user = user;
		else this.user = null;
	}

	async signIn() {
		const state = "a2123a67ff11413fa19217a9ea0fbad5";
		this.oidc.signinRedirect({state});
	}

	async signOut() {
		this.oidc.signoutRedirect();
	}
};

const sessionCtx = new Context<AuthSessionState>("authSession");
export const setAuthSessionState = (s: AuthSessionState) => sessionCtx.set(s);
export const useAuthSessionState = () => sessionCtx.get();

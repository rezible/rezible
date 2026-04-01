import {
	type GetCurrentAuthSessionResponse,
	type User,
	getCurrentAuthSessionOptions,
	type ErrorModel,
	type Organization,
	refreshAuthSessionMutation,
	clearAuthSessionMutation,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { onMount } from "svelte";

export type AuthSessionErrorCategory = "unknown" | "invalid" | "session_expired" | "no_session" | "invalid_user";

export type AuthSessionError = {
	category: AuthSessionErrorCategory;
	code?: string;
};

const parseSessionError = (err: ErrorModel): AuthSessionError => {
	const status = err.status ?? 503;
	const detail = err.detail;
	let category: AuthSessionErrorCategory = "unknown";
	if (status === 401) {
		if (detail === "session_expired") {
			category = "session_expired";
		} else if (detail === "no_session") {
			category = "no_session";
		} else if (detail === "invalid_user") {
			category = "invalid_user";
		}
	} else if (status === 403) {
		category = "invalid";
	} else if (status === 404) {
		category = "invalid_user";
	} else if (status >= 500) {
		// TODO
		console.error("failed to get auth session", status, err);
	}
	return {category: category, code: detail} as AuthSessionError;
}

type AuthSession = {
	expiresAt: Date;
	user: User;
	organization: Organization;
};

const parseUserAuthSessionResponse = ({data}: GetCurrentAuthSessionResponse): AuthSession => {
	return {
		user: data.user,
		organization: data.organization,
		expiresAt: parseAbsoluteToLocal(data.expiresAt).toDate(),
	};
};

export class AuthSessionState {
	constructor() {
		onMount(() => (this.runSessionExpiryCheck()))
	}
	
	private query = createQuery(() => getCurrentAuthSessionOptions());
	private queryData = $derived(this.query.data);

	session = $derived(!!this.queryData ? parseUserAuthSessionResponse(this.queryData) : null);
	loaded = $derived(this.query.isFetched);
	user = $derived(this.session?.user);
	org = $derived(this.session?.organization);
	
	error = $derived.by<AuthSessionError | undefined>(() => {
		if (this.session && this.session.expiresAt < new Date(Date.now())) {
			return {category: "session_expired"};
		}
		if (this.query.error) {
			return parseSessionError(this.query.error as ErrorModel);
		}
	});

	isAuthenticated = $derived(!!this.session && !this.error);
	isSetup = $derived(this.isAuthenticated && !this.org?.attributes.setupRequired);

	refetch() {
		this.query.refetch();
	}

	private logoutMut = createMutation(() => ({
		...clearAuthSessionMutation(),
		onSuccess: () => {this.refetch()}
	}));
	async logout() {
		this.logoutMut.mutate({});
	}

	private runSessionExpiryCheck() {
		const CheckIntervalMs = 10_000;
		const checkExpiry = () => {
			if (!this.session) return;
			const timeLeft = this.session.expiresAt.valueOf() - new Date(Date.now()).valueOf();
			if (timeLeft <= 0) {
				this.error = {category: "session_expired"};
			} else if (timeLeft <= CheckIntervalMs * 100) {
				this.refreshSession(timeLeft);
			}
		}
		const i = setInterval(checkExpiry, CheckIntervalMs);
		return () => clearInterval(i);
	}

	private refreshSessionMut = createMutation(() => ({
		...refreshAuthSessionMutation(),
		onSuccess: () => {
			console.log("auth session refreshed");
		}
	}));
	private async refreshSession(timeLeft: number) {
		console.log("auth session expiring soon", timeLeft);
		await this.refreshSessionMut.mutateAsync({});
		this.refetch();
	}
};

const sessionCtx = new Context<AuthSessionState>("authSession");
export const initAuthSessionState = () => sessionCtx.set(new AuthSessionState());
export const useAuthSessionState = () => sessionCtx.get();
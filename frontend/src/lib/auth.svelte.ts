import { onNavigate } from "$app/navigation";
import {
	type UserNotification,
	type GetCurrentAuthSessionResponse,
	type User,
	getCurrentAuthSessionOptions,
	type ErrorModel,
	type Organization,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { createQuery, QueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { onMount } from "svelte";

export type SessionErrorCategory = "unknown" | "invalid" | "expired" | "no_session" | "no_user";

export type SessionError = {
	category: SessionErrorCategory;
	code?: string;
};

const parseSessionError = (err: ErrorModel): SessionError => {
	let errCategory: SessionErrorCategory = "unknown";
	const status = err.status ?? 503;
	const errCode = err.detail;
	if (status === 401) {
		if (errCode === "session_expired") {
			errCategory = "expired";
		} else if (errCode === "no_session") {
			errCategory = "no_session";
		} else if (errCode === "missing_user") {
			errCategory = "no_user";
		}
	} else if (status === 404) {
		errCategory = "no_user";
	} else if (status >= 500) {
		// TODO
		console.error("failed to get auth session", status, err);
	}
	return {category: errCategory, code: errCode} as SessionError;
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

const SessionExpiryCheckIntervalMs = 10_000;

export class AuthSessionState {
	private query = createQuery(() => getCurrentAuthSessionOptions());

	session = $derived(this.query.data ? parseUserAuthSessionResponse(this.query.data) : null);
	loaded = $derived(this.query.isFetched);
	user = $derived(this.session?.user);
	org = $derived(this.session?.organization);
	
	error = $derived.by<SessionError | undefined>(() => {
		if (this.session && this.session.expiresAt < new Date(Date.now())) {
			return {category: "expired"};
		}
		if (this.query.error) {
			return parseSessionError(this.query.error as ErrorModel);
		}
	});

	isAuthenticated = $derived(!!this.session && !this.error);
	isSetup = $derived(this.isAuthenticated && !this.org?.requiresInitialSetup);

	refetch() {
		this.query.refetch();
	}

	checkSessionExpiry() {
		if (!this.session) return;
		const timeLeft = this.session.expiresAt.valueOf() - new Date(Date.now()).valueOf();
		if (timeLeft <= 0) {
			this.error = {category: "expired"};
		} else if (timeLeft <= SessionExpiryCheckIntervalMs * 100) {
			this.refreshSession(timeLeft);
		}
	}

	private refreshSession(timeLeft: number) {
		console.log("auth session expiring soon", timeLeft);
	}

	constructor() {
		onMount(() => {
			const i = setInterval(() => {
				this.checkSessionExpiry();
			}, SessionExpiryCheckIntervalMs);
			return () => {
				clearInterval(i);
			}
		})
	}
};

const sessionCtx = new Context<AuthSessionState>("authSession");
export const setAuthSessionState = (s: AuthSessionState) => sessionCtx.set(s);
export const useAuthSessionState = () => sessionCtx.get();

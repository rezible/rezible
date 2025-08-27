import {
	type UserNotification,
	type GetCurrentUserAuthSessionResponse,
	type User,
	getCurrentUserAuthSessionOptions,
	type ErrorModel,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { createQuery, QueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { onMount } from "svelte";

export type SessionErrorCategory = "unknown" | "invalid" | "expired" | "no_session" | "no_user";

export type SessionError = {
	category: SessionErrorCategory;
	code?: string;
};

type AuthSession = {
	expiresAt: Date;
	user: User;
};

const parseUserAuthSessionResponse = (resp: GetCurrentUserAuthSessionResponse): AuthSession => {
	return {
		user: resp.data.user,
		expiresAt: parseAbsoluteToLocal(resp.data.expiresAt).toDate(),
	};
};

const SessionExpiryCheckIntervalMs = 10_000;

export class AuthSessionState {
	private query = createQuery(() => getCurrentUserAuthSessionOptions());

	session = $derived(this.query.data ? parseUserAuthSessionResponse(this.query.data) : null);
	loaded = $derived(this.query.isFetched);
	user = $derived(this.session?.user);
	
	error = $derived.by<SessionError | undefined>(() => {
		if (this.session && this.session.expiresAt < new Date(Date.now())) {
			return {category: "expired"};
		}
		if (!this.query.error) return;
		const err = this.query.error as ErrorModel;
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
	});

	isAuthenticated = $derived(!!this.session && !this.error);

	constructor() {
		onMount(() => {
			const i = setInterval(() => {
				if (!this || !this.session) return;
				const timeLeft = this.session.expiresAt.valueOf() - new Date(Date.now()).valueOf();
				if (timeLeft <= 0) this.error = {category: "expired"};
				if (timeLeft <= SessionExpiryCheckIntervalMs * 100) {
					console.log("auth session expiring soon", timeLeft);	
				}
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

const createNotifications = () => {
	const notifications = $state<UserNotification[]>([]);
	let queryClient = $state<QueryClient>();

	return {
		get inbox() {
			return notifications;
		},
		setQueryClient: (c: QueryClient) => {
			queryClient = c;
		},
	};
};
export const notifications = createNotifications();

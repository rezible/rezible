import { goto, onNavigate } from "$app/navigation";
import { page } from "$app/state";
import {
	type UserNotification,
	type GetCurrentUserAuthSessionResponse,
	type User,
	getCurrentUserAuthSessionOptions,
	type ErrorModel,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { redirect } from "@sveltejs/kit";
import { createQuery, QueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
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
};

const sessionCtx = new Context<AuthSessionState>("authSession");
export const setAuthSessionState = (s: AuthSessionState) => sessionCtx.set(s);
export const useAuthSessionState = () => sessionCtx.get();

/*
const refreshWindowSecs = 60 * 3;

const startRefetchQuery = (client: QueryClient) => {
	const refetchInterval = 1000 * 60; // 1 minute
	const opts = queryOptions({
		...getCurrentUserAuthSessionOptions(),
		refetchInterval,
	});
	const observer = new QueryObserver(client, opts);
	observer.subscribe((status) => {
		if (status.isFetching) return;

		if (status.isError) {
			console.log("auth error", status.error);
			return;
		}

		if (status.isSuccess) {
			const sess = parseUserSessionResponse(status.data);
			const now = new Date(Date.now());
			if (sess.expiresAt && differenceInSeconds(sess.expiresAt, now) < refreshWindowSecs) {
				console.log(`less than ${refreshWindowSecs / 60} minutes left until auth expires`);
			}
		}
	});

	return () => {
		if (observer) observer.destroy();
	};
};
*/

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

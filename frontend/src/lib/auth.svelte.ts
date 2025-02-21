import { dev } from "$app/environment";
import {
	client,
	getCurrentUserSessionOptions,
	type UserNotification,
	type GetCurrentUserSessionResponse,
	type User,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { getCurrentUserSession } from "./api/oapi.gen";
import { QueryClient, QueryObserver, queryOptions } from "@tanstack/svelte-query";
import { differenceInSeconds } from "date-fns/differenceInSeconds";

// TODO: load this
export const AUTH_REDIRECT_URL = dev ? "http://localhost:8888/auth" : "/auth";
const refreshWindowSecs = 60 * 3;

type SessionError = "unknown" | "invalid" | "expired" | "no_user";

type AuthSession = {
	expiresAt: Date;
	user: User;
};
const parseUserSessionResponse = ({ data }: GetCurrentUserSessionResponse): AuthSession => {
	return {
		user: data.user,
		expiresAt: parseAbsoluteToLocal(data.expiresAt).toDate(),
	};
};

const createSession = () => {
	let session = $state<AuthSession>();
	let error = $state<SessionError>();

	const set = (s: AuthSession) => {
		session = s;

		if (s.expiresAt < new Date(Date.now())) {
			error = "expired";
		} else {
			error = undefined;
		}
	};

	const clear = () => {
		session = undefined;
		error = undefined;
	};

	const load = async (_fetch?: typeof fetch) => {
		const {
			data,
			error: respError,
			response,
		} = await getCurrentUserSession({
			client,
			fetch: _fetch,
			throwOnError: false,
		});

		if (data) {
			set(parseUserSessionResponse(data));
			return;
		}

		clear();

		let newError: SessionError = "unknown";
		const status = response.status;
		if (status === 401) {
			const authErr = respError.detail;

			if (authErr == "no_session" || authErr == "session_expired") {
				return AUTH_REDIRECT_URL;
			}

			if (authErr == "missing_user") {
				newError = "no_user";
			}
		} else if (status === 404) {
			newError = "no_user";
		} else if (status >= 500) {
			// TODO
			console.error("failed to get auth session", status, respError);
		}

		error = newError;

		return;
	};

	return {
		load,
		get user() {
			return session?.user;
		},
		get userId() {
			return session?.user.id;
		},
		get username() {
			return session?.user.attributes.name || "<username>";
		},
		get email() {
			return session?.user.attributes.email || "<email>";
		},
		get accentColor() {
			return "#a33333";
		},
		get error() {
			return error;
		},
		get expiresAt() {
			return session?.expiresAt;
		},
	};
};
export const session = createSession();

const startRefetchQuery = (client: QueryClient) => {
	const refetchInterval = 1000 * 60; // 1 minute
	const opts = queryOptions({
		...getCurrentUserSessionOptions(),
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

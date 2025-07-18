import {
	client,
	type UserNotification,
	type GetCurrentUserAuthSessionResponse,
	type User,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { getCurrentUserAuthSession } from "./api/oapi.gen";
import { QueryClient } from "@tanstack/svelte-query";

export type SessionErrorCategory = "unknown" | "invalid" | "expired" | "no_session" | "no_user";

type SessionError = {
	category: SessionErrorCategory;
	code?: string;
};

type AuthSession = {
	expiresAt: Date;
	user: User;
};

const parseUserAuthSessionResponse = ({ data }: GetCurrentUserAuthSessionResponse): AuthSession => {
	return {
		user: data.user,
		expiresAt: parseAbsoluteToLocal(data.expiresAt).toDate(),
	};
};

const createAuthSessionState = () => {
	let session = $state<AuthSession>();
	let loaded = $state(false);
	let error = $state<SessionError>();

	const set = (s: AuthSession) => {
		session = s;

		if (s.expiresAt < new Date(Date.now())) {
			error = {category: "expired"};
		} else {
			error = undefined;
		}
	};

	const clear = () => {
		session = undefined;
		error = undefined;
		loaded = false;
	};

	const load = async (_fetch?: typeof fetch) => {
		if (loaded) return !!session;

		const {
			data,
			error: respError,
			response,
		} = await getCurrentUserAuthSession({
			client,
			fetch: _fetch,
			throwOnError: false,
		});

		loaded = true;

		if (data) {
			set(parseUserAuthSessionResponse(data));
			return true;
		}

		clear();

		let errCategory: SessionErrorCategory = "unknown";
		const status = response.status;
		const errCode = respError.detail;
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
			console.error("failed to get auth session", status, respError);
		}

		error = {category: errCategory, code: errCode};

		return false;
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
export const session = createAuthSessionState();

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

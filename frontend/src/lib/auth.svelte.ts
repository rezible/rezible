import { dev } from '$app/environment';
import { client, getCurrentUserSessionOptions, type UserNotification, type GetCurrentUserSessionResponse, type User } from '$lib/api';
import { getCurrentUserSession } from "./api/oapi.gen/services.gen"
import { QueryClient, QueryObserver, queryOptions } from '@tanstack/svelte-query';
import { differenceInSeconds } from 'date-fns/differenceInSeconds';

// TODO: load this
export const AUTH_REDIRECT_URL = dev ? "http://localhost:8888/auth" : "/auth";
const refreshWindowSecs = 60 * 3;

type Session = {
	expiresAt?: Date;
	user?: User;
};
const parseUserSessionResponse = ({data}: GetCurrentUserSessionResponse): Session => {
	return {
		user: data.user,
		expiresAt: new Date(Date.parse(data.expires_at)),
	}

}
const createSession = () => {
	let session = $state<Session>({});
	let isValid = $state(false);

	const clear = () => {session = {}}
	const setUser = (u: User) => {session.user = u};
	const setExpiresAt = (d: Date) => {session.expiresAt = d};
	const checkIsValid = () => {
		isValid = !!session.expiresAt && session.expiresAt > new Date(Date.now())
	}
	const setSession = (s: Session) => {
		session = s;
		checkIsValid();
	}

	const fetchSession = async (_fetch?: typeof fetch) => {
		const { data, error, response } = await getCurrentUserSession({client, fetch: _fetch, throwOnError: false});
		console.log(data, error, response);
		if (data) {
			setSession(parseUserSessionResponse(data));
			return;
		}
		clear();
		return response.status;
	}

	return {
		fetchSession,
		get userId() { return session.user?.id },
		get user() { return session.user },
		get username() { return session.user?.attributes.name || "<username>" },
		get email() { return session.user?.attributes.email || "<email>" },
		get accentColor() { return "#a33333" },
		get isValid() { return isValid },
		get expiresAt() { return session.expiresAt },
		clear,
   };
}
export const session = createSession();

const startRefetchQuery = (client: QueryClient) => {
	const refetchInterval = 1000 * 60; // 1 minute
	const opts = queryOptions({
		...getCurrentUserSessionOptions(),
		refetchInterval,
	});
	const observer = new QueryObserver(client, opts);
	observer.subscribe(status => {
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

	return () => {if (observer) observer.destroy()};
}

const createNotifications = () => {
	const notifications = $state<UserNotification[]>([]);
	let queryClient = $state<QueryClient>();

	return {
		get inbox () { return notifications },
		setQueryClient: (c: QueryClient) => {queryClient = c}
	}
}
export const notifications = createNotifications();
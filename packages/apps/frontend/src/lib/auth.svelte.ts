import {
	type GetCurrentAuthSessionResponse,
	type User,
	getCurrentAuthSessionOptions,
	type ErrorModel,
	type Organization,
	refreshAuthSessionMutation,
	clearAuthSessionMutation,
	type GetCurrentAuthSessionResponseBody,
} from "$lib/api";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { createMutation, createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { Context } from "runed";
import { onMount } from "svelte";

export enum AuthSessionErrorCategory {
	NoSession = "auth_session_missing",
	SessionExpired = "auth_session_expired",
	SessionInvalid = "auth_session_invalid",
	ServerError = "server_error",
	Unknown = "unknown",
};

type AuthSession = {
	expiresAt: Date;
	user: User;
	organization: Organization;
};

const parseAuthSessionQueryResponseError = (err: ErrorModel): AuthSessionErrorCategory => {
	const status = err.status ?? 503;
	let category: AuthSessionErrorCategory = AuthSessionErrorCategory.Unknown;
	if (status === 401) {
		const mappedCategory = err.detail as AuthSessionErrorCategory;
		if (Object.values(AuthSessionErrorCategory).includes(mappedCategory)) {
			category = mappedCategory;
		}
	} else if (status >= 500) {
		category = AuthSessionErrorCategory.ServerError;
	}
	return category;
};

const parseAuthSessionQueryResponseData = ({data}: GetCurrentAuthSessionResponseBody): AuthSession => {
	return {
		user: data.user,
		organization: data.organization,
		expiresAt: parseAbsoluteToLocal(data.expiresAt).toDate(),
	};
};

type AuthSessionQueryResult = CreateQueryResult<GetCurrentAuthSessionResponseBody, ErrorModel>;
type ParsedAuthSessionQueryResult = {
	session?: AuthSession,
	error?: AuthSessionErrorCategory,
};
const parseUserAuthSessionQueryResponse = ({data: body, error}: AuthSessionQueryResult): ParsedAuthSessionQueryResult => {
	if (!!error) {
		return {error: parseAuthSessionQueryResponseError(error)};
	} else if (!!body) {
		return {session: parseAuthSessionQueryResponseData(body)};
	}
	return {};
};

export class AuthSessionState {
	constructor() {
		onMount(() => (this.startSessionExpiryCheck()))
	}
	
	private query = createQuery(() => getCurrentAuthSessionOptions());
	private parsedResponse = $derived(parseUserAuthSessionQueryResponse(this.query));

	loaded = $derived(this.query.isFetched);
	session = $derived(this.parsedResponse.session);
	error = $derived(this.parsedResponse.error);

	user = $derived(this.session?.user);
	org = $derived(this.session?.organization);

	isAuthenticated = $derived(!!this.session && !this.error);
	isSetup = $derived(this.isAuthenticated && !this.org?.attributes.setupRequired);

	refetch() {
		this.query.refetch();
	}

	private logoutMut = createMutation(() => ({
		...clearAuthSessionMutation(),
		onSuccess: () => {
			this.refetch();
		}
	}));
	async logout() {
		this.logoutMut.mutate({});
	}

	private refreshSessionMut = createMutation(() => ({
		...refreshAuthSessionMutation(),
		onSuccess: () => {
			console.log("auth session refreshed");
			this.refetch();
		}
	}));
	private startSessionExpiryCheck() {
		const CheckIntervalMs = 10_000;
		const checkExpiry = () => {
			if (!this.session || this.refreshSessionMut.isPending) return;
			const timeLeft = this.session.expiresAt.valueOf() - new Date(Date.now()).valueOf();
			if (timeLeft <= 0) {
				// TODO: handle this better
				this.refetch();
			} else if (timeLeft <= CheckIntervalMs * 100) {
				console.log("auth session expiring soon", timeLeft);
				this.refreshSessionMut.mutate({});
			}
		}
		const i = setInterval(checkExpiry, CheckIntervalMs);
		return () => clearInterval(i);
	};
};

const ctx = new Context<AuthSessionState>("AuthSessionState");
export const initAuthSessionState = () => ctx.set(new AuthSessionState());
export const useAuthSessionState = () => ctx.get();
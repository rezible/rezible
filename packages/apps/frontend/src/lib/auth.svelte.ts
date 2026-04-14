import { navigating, page } from "$app/state";
import {
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
import { Context, watch } from "runed";
import { onMount } from "svelte";
import { beforeNavigate, goto } from "$app/navigation";
import type { RouteId } from "$app/types";
import { resolve } from "$app/paths";

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

const LoginRoute = resolve("/login");
const SettingsRouteId = resolve("/settings");
const getAuthRedirect = (routeId: RouteId | null, isAuthenticated: boolean, isSetup: boolean) => {
	if (!routeId) return null;

	const isLoginRoute = routeId?.startsWith(LoginRoute);
	if (!isAuthenticated) return isLoginRoute ? null : LoginRoute;

	const isSettingsRoute = routeId?.startsWith(SettingsRouteId);
	if (!isSetup) return isSettingsRoute ? null : SettingsRouteId;

	return isLoginRoute ? "/" : null;
}

export class AuthSessionState {
	private query = createQuery(() => getCurrentAuthSessionOptions());
	private parsedResponse = $derived(parseUserAuthSessionQueryResponse(this.query));

	private loaded = $derived(this.query.isFetched);
	error = $derived(this.parsedResponse.error);
	private session = $derived(this.parsedResponse.session);
	user = $derived(this.session?.user);
	org = $derived(this.session?.organization);

	isAuthenticated = $derived(!!this.session && !this.error);
	isSetup = $derived(this.isAuthenticated && !this.org?.attributes.setupRequired);

	constructor() {
		this.startSessionExpiryCheck();
		this.guardNavigation();
	};

	private redirectTo = $derived(this.loaded ? getAuthRedirect(page.route.id, this.isAuthenticated, this.isSetup) : undefined);
	ready = $derived(this.loaded && !this.redirectTo);

	private guardNavigation() {
		watch(() => this.redirectTo, route => {
			if (!!route && route !== navigating.to?.route.id) {
				goto(route);
			}
		});
		beforeNavigate(async nav => {
			if (nav.willUnload || !nav.to) return;
			const wouldRedirectTo = getAuthRedirect(nav.to.route.id, this.isAuthenticated, this.isSetup);
			if (!!wouldRedirectTo && nav.to.route.id !== wouldRedirectTo) {
				nav.cancel();
			}
		});
	}

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
		onMount(() => {
			const i = setInterval(checkExpiry, CheckIntervalMs);
			return () => clearInterval(i);
		});
	};
};

const ctx = new Context<AuthSessionState>("AuthSessionState");
export const initAuthSessionState = () => ctx.set(new AuthSessionState());
export const useAuthSessionState = () => ctx.get();
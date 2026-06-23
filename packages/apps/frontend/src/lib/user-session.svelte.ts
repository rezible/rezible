import { navigating, page } from "$app/state";
import {
	getUserSessionOptions,
	type ErrorModel,
	type GetUserSessionResponseBody,
	type UserSession,
} from "$lib/api";
import { parseAbsoluteToLocal, ZonedDateTime } from "@internationalized/date";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { onMount, tick } from "svelte";
import { beforeNavigate, goto } from "$app/navigation";
import type { RouteId } from "$app/types";
import { resolve } from "$app/paths";

export enum ApiAuthErrorCategory {
	NoSession = "auth_session_missing",
	SessionExpired = "auth_session_expired",
	SessionInvalid = "auth_session_invalid",
	ServerError = "server_error",
	Unknown = "unknown",
};
const authErrCategories = Object.values(ApiAuthErrorCategory);

const parseUserSessionResponseError = (err: ErrorModel): ApiAuthErrorCategory => {
	if (err.status === 401) {
		const mappedCategory = err.detail as ApiAuthErrorCategory;
		if (authErrCategories.includes(mappedCategory)) {
			return mappedCategory;
		}
	} else if (!err.status || err.status >= 500) {
		return ApiAuthErrorCategory.ServerError;
	}
	return ApiAuthErrorCategory.Unknown;
};

type AuthSessionQueryResult = CreateQueryResult<GetUserSessionResponseBody, ErrorModel>;
type ParsedAuthSessionQueryResult = {
	session?: Omit<UserSession, "expiresAt"> & { expiresAt: ZonedDateTime },
	error?: ApiAuthErrorCategory,
};
const parseUserSessionQueryResponse = ({data: body, error}: AuthSessionQueryResult): ParsedAuthSessionQueryResult => {
	let res: ParsedAuthSessionQueryResult = {};
	if (!!error) {
		res.error = parseUserSessionResponseError(error);
	} else if (!!body) {
		res.session = {
			user: body.data.user,
			organization: body.data.organization,
			expiresAt: parseAbsoluteToLocal(body.data.expiresAt),
		};
	}
	return res;
};

const LoginRoute = resolve("/login");
const InitialSetupRouteId = resolve("/settings/initial-setup");
const ConnectRoutePrefix = resolve("/connect");
const getAuthRedirect = (routeId: RouteId | null, isAuthenticated: boolean, isSetup: boolean) => {
	if (!routeId) return null;

	const isLoginRoute = routeId?.startsWith(LoginRoute);
	if (!isAuthenticated) return isLoginRoute ? null : LoginRoute;

	const isInitialSetupRoute = routeId?.startsWith(InitialSetupRouteId);
	const isConnectRoute = routeId?.startsWith(ConnectRoutePrefix);
	if (!isSetup) return isInitialSetupRoute || isConnectRoute ? null : InitialSetupRouteId;
	if (isSetup && isInitialSetupRoute) return "/settings";

	return isLoginRoute ? "/" : null;
}

export class UserSessionState {
	private query = createQuery(() => getUserSessionOptions());
	private loaded = $derived(this.query.isFetched);

	private parsedResponse = $derived(parseUserSessionQueryResponse(this.query));
	
	error = $derived(this.parsedResponse.error);

	private session = $derived(this.parsedResponse.session);
	private sessionExpiresAt = $derived(!!this.session ? this.session.expiresAt.toDate() : null);
	user = $derived(this.session?.user);
	org = $derived(this.session?.organization);

	isAuthenticated = $derived(!!this.session && !this.error);
	isSetup = $derived(this.isAuthenticated && !this.org?.attributes.setupRequired);

	constructor() {
		this.startSessionExpiryCheck();
		this.addNavigationGuards();
	};

	private redirectTo = $derived(this.loaded ? getAuthRedirect(page.route.id, this.isAuthenticated, this.isSetup) : undefined);

	private addNavigationGuards() {
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

	signingOut = $state(false);
	async logout() {
		this.signingOut = true;
		await tick()
		await goto("/api/auth/logout");
		this.signingOut = false;
	}

	ready = $derived(this.loaded && !this.signingOut && !this.redirectTo);

	private startSessionExpiryCheck() {
		const CheckIntervalMs = 10_000;
		const checkExpiry = () => {
			// if (!this.session || this.refreshSessionMut.isPending) return;
			if (!this.sessionExpiresAt) return;
			const timeLeft = this.sessionExpiresAt.valueOf() - Date.now();
			if (timeLeft <= 0) {
				// TODO: handle this better
				// this.refetch();
				console.log("user auth session expired");
			} else if (timeLeft <= CheckIntervalMs * 100) {
				console.log("auth session expiring soon", timeLeft);
				// this.refreshSessionMut.mutate({});
			}
		}
		onMount(() => {
			const i = setInterval(checkExpiry, CheckIntervalMs);
			return () => clearInterval(i);
		});
	};
};

const ctx = new Context<UserSessionState>("UserSessionState");
export const initUserSessionState = () => ctx.set(new UserSessionState());
export const useUserSessionState = () => ctx.get();

import { navigating, page } from "$app/state";
import {
	getUserSessionOptions,
	type ApiError,
	type GetUserSessionResponseBody,
	type UserSession,
} from "$lib/api";
import { parseAbsoluteToLocal, ZonedDateTime } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { onMount, tick } from "svelte";
import { beforeNavigate, goto } from "$app/navigation";
import type { RouteId } from "$app/types";
import { resolve } from "$app/paths";

type ParsedUserSession = Omit<UserSession, "expiresAt"> & { expiresAt: ZonedDateTime };

// A 401 ends the session even while the last loaded session is still cached.
const parseUserSession = (data?: GetUserSessionResponseBody, error?: ApiError | null) => {
	if (!data || error?.status === 401) {
		return undefined;
	}
	const res = data.data;
	return {
		user: res.user,
		organization: res.organization,
		organizationRole: res.organizationRole,
		expiresAt: parseAbsoluteToLocal(res.expiresAt),
	} satisfies ParsedUserSession;
};

const DefaultRoute = resolve("/");
const LoginRoute = resolve("/login");
const InitialSetupRoute = resolve("/settings/initial-setup");
const getAuthRedirect = (routeId: RouteId | null, isAuthenticated: boolean, isSetup: boolean) => {
	if (!routeId) return null;

	const isLoginRoute = routeId.startsWith(LoginRoute);
	if (!isAuthenticated) {
		return isLoginRoute ? null : LoginRoute;
	}

	const isInitialSetupRoute = routeId.startsWith(InitialSetupRoute);
	if (!isSetup) {
		return isInitialSetupRoute ? null : InitialSetupRoute;
	}
	if (isSetup && isInitialSetupRoute) {
		return resolve("/settings/integrations");
	}

	return isLoginRoute ? consumeReturnLocation() : null;
};

/**
 * Safe return-location storage for authentication recovery.
 *
 * The last valid product location is kept in sessionStorage so a user who
 * re-authenticates mid-session lands back where they were. Stored locations
 * are validated before use; anything invalid or unsafe falls back to the
 * dashboard.
 */
const returnLocationSessionStorageKey = "rezible.return-location";

const isSafeReturnLocation = (location: string | null): location is string => {
	if (!location) return false;
	if (!location.startsWith("/") || location.startsWith("//")) return false;
	if (location.startsWith("/api/") || location.startsWith("/login")) return false;
	return true;
};

const recordReturnLocation = (pathname: string | null) => {
	if (!isSafeReturnLocation(pathname)) return;
	try {
		sessionStorage.setItem(returnLocationSessionStorageKey, pathname);
	} catch {}
};

const consumeReturnLocation = (): string => {
	try {
		const stored = sessionStorage.getItem(returnLocationSessionStorageKey);
		sessionStorage.removeItem(returnLocationSessionStorageKey);
		if (isSafeReturnLocation(stored)) return stored;
	} catch {} // storage unavailable; the dashboard fallback applies
	return DefaultRoute;
};

export class UserSessionState {
	private query = createQuery(() => getUserSessionOptions());
	private loaded = $derived(this.query.isFetched);

	error = $derived(this.query.error ?? undefined);

	private session = $derived(parseUserSession(this.query.data, this.query.error));

	/** Whether a session has loaded in this tab. It stays true after the session ends, until a reload. */
	hadSession = $state(false);

	private sessionExpiresAt = $derived(!!this.session ? this.session.expiresAt.toDate() : null);
	user = $derived(this.session?.user);
	org = $derived(this.session?.organization);
	orgPreferences = $derived(this.org?.attributes.preferences);
	organizationRole = $derived(this.session?.organizationRole);
	isAdmin = $derived(this.organizationRole === "admin");

	isAuthenticated = $derived(!!this.session);
	isSetup = $derived(this.isAuthenticated && !this.org?.attributes.setupRequired);
	private isReady = $derived(this.isAuthenticated && this.isSetup);
	private returnLocation = $derived(this.isReady ? page.url.pathname : null);

	constructor() {
		this.trackHadSession();
		this.startSessionExpiryCheck();
		this.addNavigationGuards();
		this.trackReturnLocation();
	}

	private trackHadSession() {
		watch(
			() => this.session,
			(session) => {
				if (!!session) {
					this.hadSession = true;
				}
			}
		);
	}

	private trackReturnLocation() {
		watch(() => this.returnLocation, recordReturnLocation);
	}

	private redirectTo = $derived(
		this.loaded ? getAuthRedirect(page.route.id, this.isAuthenticated, this.isSetup) : undefined
	);

	private addNavigationGuards() {
		watch(
			() => this.redirectTo,
			(route) => {
				if (!!route && route !== navigating.to?.route.id) {
					goto(route);
				}
			}
		);
		beforeNavigate(async (nav) => {
			if (nav.willUnload || !nav.to) return;
			const wouldRedirectTo = getAuthRedirect(nav.to.route.id, this.isAuthenticated, this.isSetup);
			if (!!wouldRedirectTo && nav.to.route.id !== wouldRedirectTo) {
				nav.cancel();
			}
		});
	}

	async refetch() {
		await this.query.refetch();
	}

	signingOut = $state(false);
	async logout() {
		this.signingOut = true;
		await tick();
		await goto("/api/auth/logout");
		this.signingOut = false;
	}

	ready = $derived(this.loaded && !this.signingOut && !this.redirectTo);

	private refreshing = false;
	// Refetching after a refresh loads the new expiry, or ends the session when the refresh was refused (401).
	// Any other failure is retried by the next check.
	private async refreshSession() {
		this.refreshing = true;
		try {
			const response = await fetch("/api/auth/refresh", { method: "POST", credentials: "include" });
			if (response.status === 204 || response.status === 401) {
				await this.refetch();
			}
		} catch {
		} finally {
			this.refreshing = false;
		}
	}

	private startSessionExpiryCheck() {
		const CheckIntervalMs = 10_000;
		const RefreshWindowMs = 5 * 60_000;
		const checkExpiry = () => {
			if (!this.sessionExpiresAt) return;
			const timeLeft = this.sessionExpiresAt.valueOf() - Date.now();
			if (timeLeft <= 0) {
				this.refetch();
			} else if (timeLeft <= RefreshWindowMs && !this.refreshing) {
				this.refreshSession();
			}
		};
		onMount(() => {
			const i = setInterval(checkExpiry, CheckIntervalMs);
			return () => clearInterval(i);
		});
	}
}

const ctx = new Context<UserSessionState>("UserSessionState");
export const initUserSessionState = () => ctx.set(new UserSessionState());
export const useUserSessionState = () => ctx.get();

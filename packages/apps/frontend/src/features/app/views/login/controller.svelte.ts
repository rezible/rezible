import { goto } from "$app/navigation";
import { useUserSessionState } from "$src/lib/user-session.svelte";
import type { ErrorDisplay } from "$lib/api";
import { page } from "$app/state";
import { sessionNotice } from "./sessionNotice";

const loginErrorDisplayText: Record<string, string> = {
	["create_redirect"]: "Failed to redirect to identity provider",
	["write_auth_session"]: "Failed to write auth session",
	["write_auth_state"]: "Failed to write auth state",
	["read_auth_state"]: "Failed to read auth state",
	["callback_exchange"]: "Failed to perform callback exchange with identity provider",
	["identity_sync"]: "Failed to sync user & organization information",
};
const transformLoginErrorCode = (code: string | null): ErrorDisplay | undefined => {
	if (!code) return;
	const title = "Login Error";
	const detail = loginErrorDisplayText[code] || "An unknown problem occurred";
	return { title, detail };
};

export class LoginViewController {
	private session = useUserSessionState();

	loaded = $state(false);
	inFlow = $state(false);

	sessionNotice = $derived(sessionNotice(this.session.error, this.session.hadSession));

	loginError = $state<ErrorDisplay>();
	constructor() {
		const params = page.url.searchParams;

		this.loginError = transformLoginErrorCode(params.get("error"));

		if (!this.loginError && params.has("flow", "true")) {
			this.doLogin();
		} else {
			goto(window.location.pathname, { replaceState: true, noScroll: true }).then(() => {
				this.loaded = true;
			});
		}
	}

	async doLogin() {
		this.inFlow = true;
		await goto("/api/auth/login");
	}

	titleText = $derived("Authentication Required");
	descriptionText = $derived("Continue with your identity provider");
}

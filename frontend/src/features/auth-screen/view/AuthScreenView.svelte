<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { dev } from "$app/environment";
	import { getAuthSessionConfigOptions } from "$lib/api";
	import { session, type SessionErrorCategory } from "$lib/auth.svelte";
	import Button from "$components/button/Button.svelte";
	import Header from "$components/header/Header.svelte";

	// TODO: load this
	const AUTH_URL_BASE = dev ? "http://localhost:8888/auth" : "/auth";

	const configQuery = createQuery(() => getAuthSessionConfigOptions());
	const config = $derived(configQuery.data?.data);

	const errorCategory = $derived(session.error?.category);

	// redirect to logout if user is not found
	const authPath = $derived(errorCategory === "no_user" ? "/logout" : "");
	const authUrl = $derived(`${AUTH_URL_BASE}${authPath}`);

	const errorDisplayText: Record<SessionErrorCategory, string> = {
		unknown: "An unknown error occurred",
		invalid: "Auth session is invalid",
		expired: "Your session has expired",
		no_user: "You signed in successfully, but Rezible does not have your details.",
		no_session: "",
	};
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Authentication Required" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if session.error && errorCategory !== "no_session"}
			<div class="bg-danger-900/50 border-danger/20 border rounded p-2">
				<span class="">{errorDisplayText[errorCategory ?? "unknown"]}</span>
			</div>
		{/if}

		{#if errorCategory === "no_user"}
			<Button href="{AUTH_URL_BASE}/logout" loading={configQuery.isLoading} color="primary" variant="fill">Logout</Button>
		{:else}
			<Button href="{AUTH_URL_BASE}/saml" loading={configQuery.isLoading} color="primary" variant="fill">Continue with SAML</Button>
			<Button href="{AUTH_URL_BASE}/openid-connect" loading={configQuery.isLoading} color="primary" variant="fill">Continue with Open ID Connect</Button>
			<Button href="{AUTH_URL_BASE}/github" loading={configQuery.isLoading} color="primary" variant="fill">Continue with Github</Button>
		{/if}
	</div>
</div>

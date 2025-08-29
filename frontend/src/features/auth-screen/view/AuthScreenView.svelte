<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { BACKEND_URL, getAuthSessionConfigOptions } from "$lib/api";
	import { useAuthSessionState, type SessionErrorCategory } from "$lib/auth.svelte";
	import Button from "$components/button/Button.svelte";
	import Header from "$components/header/Header.svelte";
	import { mdiGoogle, mdiKey } from "@mdi/js";
	import Icon from "$src/components/icon/Icon.svelte";

	const session = useAuthSessionState();

	const configQuery = createQuery(() => getAuthSessionConfigOptions());
	const config = $derived(configQuery.data?.data);

	const providers = $derived(config?.providers.filter(p => p.enabled));

	const errorCategory = $derived(session.error?.category);

	type ProviderDisplay = {label: string; icon: string};
	const providerDisplay = new Map<string, ProviderDisplay>([
		["saml", {label: "SSO", icon: mdiKey}],
		["google", {label: "Google", icon: mdiGoogle}],
	]);

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
			<Button href="{BACKEND_URL}/logout" loading={configQuery.isLoading} color="primary" variant="fill">Logout</Button>
		{:else if !!providers}
			{#each providers as p}
				{@const display = providerDisplay.get(p.name.toLowerCase())}
				<Button href="{BACKEND_URL}{p.startFlowEndpoint}" color="primary" variant="fill">
					<span class="flex items-center gap-2">
					Continue with {display?.label ?? p.name}
					{#if display?.icon}
						<Icon data={display.icon} />
					{/if}
					</span>
				</Button>
			{/each}
		{/if}
	</div>
</div>

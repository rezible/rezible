<script lang="ts">
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import { useAuthSessionState, AuthSessionErrorCategory } from "$lib/auth.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiKey } from "@mdi/js";
	import { useQueryClient } from "@tanstack/svelte-query";
	import { useSearchParams } from "runed/kit";
	import z from "zod";

	const session = useAuthSessionState();

	const authSessionErrorDisplayText: Record<AuthSessionErrorCategory, string> = {
		[AuthSessionErrorCategory.NoSession]: "",
		[AuthSessionErrorCategory.SessionExpired]: "Your session has expired",
		[AuthSessionErrorCategory.SessionInvalid]: "Your session is invalid",
		[AuthSessionErrorCategory.ServerError]: "Something went wrong while authenticating you",
		[AuthSessionErrorCategory.Unknown]: "Something went wrong while authenticating you",
	};
	let authSessionError = $derived.by(() => {
		if (!session.error || session.error === AuthSessionErrorCategory.NoSession) return;
		return {
			title: "Auth Session Invalid",
			detail: authSessionErrorDisplayText[session.error] || "Unknown",
		};
	});
	const showLogout = $derived(session.error === AuthSessionErrorCategory.SessionInvalid);

	const params = useSearchParams(z.object({
		error: z.string().default(""),
	}));
	const loginError = $derived(params.error);
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Authentication Required" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if !!authSessionError}
			<InlineAlert 
				error={authSessionError}
				onDismiss={() => {authSessionError = undefined}}
			/>
		{/if}

		{#if !!loginError}
			<InlineAlert 
				error={{title: "Login Error", detail: loginError}}
				onDismiss={() => {params.reset()}}
			/>
		{/if}

        {#if showLogout}
			<Button onclick={() => {session.logout()}} color="primary">Logout</Button>
		{/if}

		<Button href="/api/auth/login" color="primary">
			<span class="flex items-center gap-2">
				Sign In
				<Icon data={mdiKey} />
			</span>
		</Button>
	</div>
</div>
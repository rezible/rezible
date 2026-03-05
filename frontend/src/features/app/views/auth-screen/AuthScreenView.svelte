<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";

	const session = useAuthSessionState();

</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Authentication Required" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if session.isAuthenticated}
			<Button onclick={() => {session.signOut()}}>sign out</Button>
		{:else}
			<Button onclick={() => {session.signIn()}}>sign in</Button>
		{/if}
	</div>
</div>

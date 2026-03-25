<script lang="ts">
	import { mdiKey } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import { initAuthScreenController } from "./controller.svelte";
	import { useAuthSessionState } from "$lib/auth.svelte";
	import InlineAlert from "$src/components/inline-alert/InlineAlert.svelte";

	const session = useAuthSessionState();
	const view = initAuthScreenController();
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Authentication Required" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if view.showSessionError}
			<div class="bg-danger-900/50 border-danger/20 border rounded p-2">
				<span class="">{view.sessionErrorText}</span>
			</div>
		{/if}

		{#if !!view.completeFlowErr}
			<InlineAlert error={{title: "Signing In Failed", detail: view.completeFlowErrText}} />
		{/if}
        
        {#if view.showLogoutButton}
			<Button onclick={() => {session.logout()}} color="primary">Logout</Button>
		{:else}
			<Button onclick={() => {view.startLoginFlow()}} color="primary">
				<span class="flex items-center gap-2">
					Continue
					<Icon data={mdiKey} />
				</span>
			</Button>
		{/if}
	</div>
</div>
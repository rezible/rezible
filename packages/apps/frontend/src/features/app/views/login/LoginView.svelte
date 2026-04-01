<script lang="ts">
	import { mdiKey } from "@mdi/js";
	import { initLoginViewController } from "./controller.svelte";

	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";

	const controller = initLoginViewController();
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Authentication Required" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if !!controller.authSessionError}
			<InlineAlert 
				error={controller.authSessionError}
				onDismiss={() => {controller.authSessionError = undefined}}
			/>
		{/if}

		{#if !!controller.authFlowErr}
			<InlineAlert 
				error={controller.authFlowErr}
				onDismiss={() => {controller.authFlowErr = undefined}}
			/>
		{/if}
        
        {#if controller.showLogoutButton}
			<Button onclick={() => {controller.doSignOut()}} color="primary">Logout</Button>
		{:else}
			<Button onclick={() => {controller.doSignIn()}} color="primary">
				<span class="flex items-center gap-2">
					Sign In
					<Icon data={mdiKey} />
				</span>
			</Button>
		{/if}
	</div>
</div>
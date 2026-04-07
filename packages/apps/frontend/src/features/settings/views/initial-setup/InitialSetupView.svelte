<script lang="ts">
	import { appShell } from "$features/app";
	import Header from "$components/header/Header.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import RequiredIntegrationsSetup from "./RequiredIntegrationsSetup.svelte";
	import { initInitialSetupViewController } from "./initialSetupViewController.svelte";

	type Props = {};
	const {}: Props = $props();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const view = initInitialSetupViewController();
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Organization Setup" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if view.step === "required_integrations"}
			{#if view.integrationsLoading}
				<LoadingIndicator />
			{:else}
				<RequiredIntegrationsSetup />
			{/if}
		{/if}
	</div>
</div>
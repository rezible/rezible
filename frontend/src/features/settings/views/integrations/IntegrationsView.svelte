<script lang="ts">
	import { appShell } from "$features/app";
	import * as Alert from "$components/ui/alert";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";
	import { initIntegrationsViewController } from "./integrationsViewController.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Integrations" }
	]);

	const view = initIntegrationsViewController();
</script>

<div class="flex flex-col gap-3 p-1">
	{#if view.oauth.loadingFlowUrl}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Starting OAuth flow...</span>
		</div>
	{/if}

	{#if view.oauth.completingFlow}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Completing OAuth flow...</span>
		</div>
	{/if}

	{#if view.oauth.startFlowErr}
		<Alert.Root variant="destructive">
			<Alert.Title>OAuth flow could not start</Alert.Title>
			<Alert.Description>{view.oauth.startFlowErr}</Alert.Description>
		</Alert.Root>
	{/if}

	{#if view.oauth.completeFlowErr}
		<Alert.Root variant="destructive">
			<Alert.Title>OAuth flow could not complete</Alert.Title>
			<Alert.Description>{view.oauth.completeFlowErr}</Alert.Description>
		</Alert.Root>
	{/if}

	{#if view.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading integrations...</span>
		</div>
	{:else if view.queryErrorMessage}
		<Alert.Root variant="destructive">
			<Alert.Title>Could not load integrations</Alert.Title>
			<Alert.Description>{view.queryErrorMessage}</Alert.Description>
		</Alert.Root>
	{:else}
		<div class="grid gap-3 md:grid-cols-2">
			{#each view.supported as integration}
				{@const name = integration.name}
				{@const configured = view.configuredMap.get(name)}
				{#key name}
					<IntegrationConfigCard
						{integration}
						{configured}
						startOAuthFlow={() => view.oauth.startFlow(name)}
						configureIntegration={(attrs) => view.configure(name, attrs)}
						isSaving={view.isSaving(name)}
						errorMessage={view.errorFor(name)}
					/>
				{/key}
			{/each}
		</div>
	{/if}
</div>

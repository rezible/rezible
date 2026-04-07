<script lang="ts">
	import { appShell } from "$features/app";
	import * as Alert from "$components/ui/alert";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";
	import { initIntegrationsController } from "./controller.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Integrations" }
	]);

	const controller = initIntegrationsController();
</script>

<div class="flex flex-col gap-3 p-1">
	{#if controller.oauth.loadingFlowUrl}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Starting OAuth flow...</span>
		</div>
	{:else if controller.oauth.completingFlow}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Completing OAuth flow...</span>
		</div>
	{:else if controller.oauth.startFlowErr}
		<Alert.Root variant="destructive">
			<Alert.Title>OAuth flow could not start</Alert.Title>
			<Alert.Description>{controller.oauth.startFlowErr}</Alert.Description>
		</Alert.Root>
	{:else if controller.oauth.completeFlowErr}
		<Alert.Root variant="destructive">
			<Alert.Title>OAuth flow could not complete</Alert.Title>
			<Alert.Description>{controller.oauth.completeFlowErr}</Alert.Description>
		</Alert.Root>
	{:else if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading integrations...</span>
		</div>
	{:else if controller.queryErrorMessage}
		<Alert.Root variant="destructive">
			<Alert.Title>Could not load integrations</Alert.Title>
			<Alert.Description>{controller.queryErrorMessage}</Alert.Description>
		</Alert.Root>
	{:else}
		<div class="grid gap-3 md:grid-cols-2">
			{#each controller.available as integration}
				{@const name = integration.name}
				{@const configured = controller.configuredMap.get(name)}
				{#key name}
					<IntegrationConfigCard
						{integration}
						{configured}
						startOAuthFlow={() => controller.oauth.startFlow(name)}
						configureIntegration={(attrs) => controller.configure(name, attrs)}
						isSaving={controller.isSaving(name)}
						errorMessage={controller.errorFor(name)}
					/>
				{/key}
			{/each}
		</div>
	{/if}
</div>

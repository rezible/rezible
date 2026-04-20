<script lang="ts">
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";
	import { initIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";

	const controller = initIntegrationsController();
</script>

<div class="flex flex-col gap-3 p-1">
	{#if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading integrations...</span>
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} />
	{:else if controller.oauth.loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Starting OAuth flow...</span>
		</div>
	{:else if controller.oauth.error}
		<InlineAlert error={controller.oauth.error} />
	{:else}
		<div class="grid gap-3 md:grid-cols-2">
			{#each controller.available as integration}
				{@const name = integration.name}
				{#key name}
					<IntegrationConfigCard {integration}
						configured={controller.configuredMap.get(name)}
						isSaving={controller.isSaving(name)}
						errorMessage={controller.errorFor(name)}
					/>
				{/key}
			{/each}
		</div>
	{/if}
</div>

<script lang="ts">
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationProviderCard from "$src/features/settings/components/integration-provider-card/IntegrationProviderCard.svelte";
	import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import { useIntegrationOAuthController } from "$src/features/settings/lib/integrationOAuthController.svelte";

	const controller = useIntegrationsController();
	const oauth = useIntegrationOAuthController();
</script>

<div class="flex flex-col gap-3 p-1">
	{#if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading integrations...</span>
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} />
	{:else if oauth.inFlow}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Processing OAuth...</span>
		</div>
	{:else}
		<div class="grid gap-3 md:grid-cols-1">
			{#each controller.availableByProvider as [provider, integrations]}
				{#key provider}
					<IntegrationProviderCard {provider} {integrations} />
				{/key}
			{/each}
		</div>
	{/if}
</div>

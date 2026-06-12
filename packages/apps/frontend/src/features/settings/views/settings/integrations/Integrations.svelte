<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";

	import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
	import IntegrationProviderCard from "$features/settings/components/integration-provider-card/IntegrationProviderCard.svelte";
	import IntegrationInstallDialog from "$features/settings/components/configure-integration-dialog/ConfigureIntegrationDialog.svelte";
	import IntegrationDataSyncDialog from "$features/settings/components/integration-datasync-dialog/IntegrationDataSyncDialog.svelte";

	const controller = useIntegrationsController();
</script>

<IntegrationInstallDialog />
<IntegrationDataSyncDialog />

<div class="flex flex-col gap-4 p-1">
	{#if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading...</span>
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} />
	{:else}
		<div class="flex flex-col gap-4">
			{#each controller.availableProviders as provider}
				{#key provider}
					<IntegrationProviderCard {provider} />
				{/key}
			{/each}
		</div>
	{/if}
</div>

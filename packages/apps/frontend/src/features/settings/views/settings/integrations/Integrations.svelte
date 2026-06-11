<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";

	import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
	import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";
	import AvailableIntegrationCard from "$features/settings/components/available-integration-card/AvailableIntegrationCard.svelte";
	import IntegrationInstallDialog from "$features/settings/components/configure-integration-dialog/ConfigureIntegrationDialog.svelte";

	const controller = useIntegrationsController();
	const oauth = useIntegrationOAuthController();
</script>

<IntegrationInstallDialog />

<div class="flex flex-col gap-4 p-1">
	{#if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading...</span>
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} />
	{:else if oauth.inFlow}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
		</div>
	{:else}
		<div class="flex flex-col gap-4">
			{#each controller.availableByProvider as [provider, integrations]}
				{#key provider}
					<section class="flex flex-col gap-3 border border-border p-4">
						<div class="flex items-center justify-between gap-3">
							<div class="min-w-0">
								<h2 class="truncate text-base font-semibold capitalize">{provider}</h2>
								<p class="text-sm text-muted-foreground">
									{integrations.length} {integrations.length === 1 ? "integration" : "integrations"} available
								</p>
							</div>
							<Badge variant="outline" class="capitalize">{provider}</Badge>
						</div>

						<div class="grid gap-3">
							{#each integrations as integration}
								<AvailableIntegrationCard {integration} />
							{/each}
						</div>
					</section>
				{/key}
			{/each}
		</div>
	{/if}
</div>

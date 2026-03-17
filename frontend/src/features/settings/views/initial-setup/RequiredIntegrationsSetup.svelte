<script lang="ts">
	import { useInitialSetupViewController } from "./initialSetupViewController.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";
	import InlineAlert from "$src/components/inline-alert/InlineAlert.svelte";

	const view = useInitialSetupViewController();
</script>

<div class="flex gap-2">
	{#if view.integrations.oauth.loadingFlowUrl}
		<LoadingIndicator />
	{:else if view.integrations.oauth.completingFlow}
		<LoadingIndicator />
	{:else if view.integrations.oauth.completeFlowErr}
		<InlineAlert error={view.integrations.oauth.completeFlowErr} />
	{:else}
	<div class="flex flex-col gap-2">
        {#if !!view.integrations.nextRequiredDataKind}
			{#each view.integrations.availableDataKindIntegrations as integration}
                {@const name = integration.name}
				{#key name}
				    {@const configured = view.integrations.configuredMap.get(name)}
					<IntegrationConfigCard {integration} {configured}
						startOAuthFlow={() => {view.integrations.oauth.startFlow(name)}}
						configureIntegration={attrs => {view.integrations.doConfigure(name, attrs)}}
					/>
				{/key}
			{:else}
				<div class="p-2 border-error-300 border-2">
					<span>No integrations available for this data</span>
				</div>
			{/each}
		{/if}
	</div>
	{/if}
</div>
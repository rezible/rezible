<script lang="ts">
	import { useInitialSetupViewController } from "./initialSetupViewController.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";

	const view = useInitialSetupViewController();

    const nextRequiredDataKind = $derived(view.integrations.nextRequiredDataKind);
</script>

<div class="flex gap-2">
	{#if view.integrations.oauth.loadingFlowUrl}
		<span>loading oauth flow</span>
		<LoadingIndicator />
	{:else if view.integrations.oauth.completingFlow}
		<span>completing oauth flow</span>
		<LoadingIndicator />
	{:else if view.integrations.oauth.completeFlowErr}
		<span>complete integration error</span>
	{:else}
	<div class="flex flex-col gap-2">
        {#if !!nextRequiredDataKind}
			<div class="w-full flex gap-2">
				{#each view.integrations.remainingRequiredDataKinds as kind}
					<div class="p-2 border" 
						class:bg-info-100={kind === nextRequiredDataKind}
					>{kind}</div>
				{/each}
			</div>

			<span>Integrations supporting {nextRequiredDataKind}</span>
			{#each view.integrations.nextRequiredSupportedIntegrations as integration}
                {@const name = integration.name}
				{#key name}
				    {@const configured = view.integrations.configuredMap.get(name)}
					<IntegrationConfigCard {integration} {configured}
						{nextRequiredDataKind}
						startOAuthFlow={() => {view.integrations.oauth.startFlow(name)}}
						configureIntegration={attrs => {view.integrations.doConfigure(name, attrs)}}
					/>
				{/key}
			{:else}
				<div class="p-2 border-error-300 border-2">
					<span>No supported integrations available for this data</span>
				</div>
			{/each}
		{/if}
	</div>
	{/if}
</div>
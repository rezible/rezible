<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import { initInitialSetupViewController } from "./controller.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";

	const ctrl = initInitialSetupViewController();
	const integrations = $derived(ctrl.integrations);
</script>

{#snippet requiredIntegrations()}
<div class="flex gap-2">
	{#if integrations.oauth.loading}
		<LoadingIndicator />
	{:else if integrations.oauth.error}
		<InlineAlert error={integrations.oauth.error} />
	{:else}
	<div class="flex flex-col gap-2">
		{#each integrations.availableDataKindIntegrations as integration}
			{@const name = integration.name}
			{#key name}
				{@const configured = integrations.configuredMap.get(name)}
				<IntegrationConfigCard {integration} {configured}
					startOAuthFlow={() => {integrations.oauth.startFlow(name)}}
					configureIntegration={attrs => {integrations.doConfigure(name, attrs)}}
				/>
			{/key}
		{:else}
			<div class="p-2 border-error-300 border-2">
				<span>No integrations available for this data</span>
			</div>
		{/each}
	</div>
	{/if}
</div>
{/snippet}

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Quick Setup" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if ctrl.loading}
			<LoadingIndicator />
		{:else}
			{#if ctrl.step === "required_integrations"}
				{@render requiredIntegrations()}
			{/if}
		{/if}
	</div>
</div>
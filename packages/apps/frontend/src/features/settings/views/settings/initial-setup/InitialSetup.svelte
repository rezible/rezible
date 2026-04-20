<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import { initInitialSetupViewController } from "./controller.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";

	const ctrl = initInitialSetupViewController();
	const oauth = $derived(ctrl.integrations.oauth);
</script>

{#snippet requiredIntegrations()}
<div class="flex gap-2">
	{#if oauth.loading}
		<Spinner />
	{:else if oauth.error}
		<InlineAlert error={oauth.error} />
	{:else}
	<div class="flex flex-col gap-2">
		{#each ctrl.availableDataKindIntegrations as integration}
			{@const name = integration.name}
			{#key name}
				{@const configured = ctrl.integrations.configuredMap.get(name)}
				<IntegrationConfigCard {integration} {configured} />
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
		<Header title="Quick Setup" classes={{ root: "gap-2", title: "text-2xl" }} />

		{#if ctrl.loading}
			<Spinner />
		{:else}
			{#if ctrl.step === "required_integrations"}
				{@render requiredIntegrations()}
			{/if}
		{/if}
	</div>
</div>
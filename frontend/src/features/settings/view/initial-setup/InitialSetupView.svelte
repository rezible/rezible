<script lang="ts">
	import { appShell } from "$features/app-shell";
	import Button from "$components/button/Button.svelte";
	import Header from "$components/header/Header.svelte";
	import { useInitialSetupViewState } from "$features/settings/lib/initialSetupViewState.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";
	import type { SupportedIntegration } from "$lib/api";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const view = useInitialSetupViewState();
</script>

{#snippet integrationCard(integration: SupportedIntegration)}
	{@const configured = view.configuredIntegrationsMap.get(integration.name)}
	{#key integration.name}
		<IntegrationConfigCard {integration} {configured}
			startOAuthFlow={() => {view.oauth.startFlow(integration.name)}}
			configureIntegration={attrs => {view.doConfigureIntegration(integration.name, attrs)}}
		/>
	{/key}
{/snippet}

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Organization Setup" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if view.loadingIntegrations}
			<span>loading integrations</span>
			<LoadingIndicator />
		{:else if view.configuringIntegration}
			<span>configuring integration</span>
			<LoadingIndicator />
		{:else if view.oauth.loadingFlowUrl}
			<span>loading oauth flow</span>
			<LoadingIndicator />
		{:else if view.oauth.completingFlow}
			<span>completing oauth flow</span>
			<LoadingIndicator />
		{:else if view.oauth.completeFlowErr}
			<span>complete integration error</span>
		{:else}
			<span>Required Data Kinds:</span>
			<div class="w-full flex gap-2">
				{#each view.requiredDataKinds as kind}
					<div class="p-2 border" 
						class:bg-success-100={view.configuredEnabledDataKinds.has(kind)}
						class:bg-info-100={view.nextRequiredDataKind === kind}
					>{kind}</div>
				{/each}
			</div>

			<!-- TODO: this should be tabs, with required data kinds first -->
			<div class="flex flex-col gap-2">
				{#if !!view.nextRequiredDataKind}
					<span>Integrations supporting {view.nextRequiredDataKind}</span>
					{#each view.nextRequiredSupportedIntegrations as intg}
						{@render integrationCard(intg)}
					{:else}
						<div class="p-2 border-error-300 border-2">
							<span>No supported integrations available for this data</span>
						</div>
					{/each}
				{:else}
					<span>All Integrations</span>
					{#each view.supportedIntegrations as intg}
						{@render integrationCard(intg)}
					{/each}
				{/if}
			</div>
			
			{#if !view.nextRequiredDataKind}
				<Button 
					color="secondary" 
					variant="fill"
					disabled={false}
					onclick={() => view.doFinishOrganizationSetup()} 
					loading={view.finishingSetup}
				>
					Finish setup
				</Button>
			{/if}
		{/if}
	</div>
</div>
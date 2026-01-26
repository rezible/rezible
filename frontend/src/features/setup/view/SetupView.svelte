<script lang="ts">
	import { appShell } from "$features/app-shell";
	import Button from "$src/components/button/Button.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { useSetupViewState } from "../lib/viewState.svelte";
	import LoadingIndicator from "$src/components/loading-indicator/LoadingIndicator.svelte";
	import IntegrationSetupCard from "../components/IntegrationSetupCard.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const view = useSetupViewState();
</script>


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
			<span>Data Kinds:</span>
			<div class="w-full flex gap-2">
				{#each view.requiredDataKinds as kind}
					<div class="p-2 border" class:bg-success-100={view.configuredEnabledDataKinds.has(kind)}>{kind}</div>
				{/each}
			</div>

			<!-- TODO: this should be tabs, with required data kinds first -->
			{#if !!view.nextRequiredDataKind}
				{@const dataKind = view.nextRequiredDataKind}
				<div class="flex flex-col gap-2">
					{#each view.nextRequiredSupportedIntegrations as integration}
						<IntegrationSetupCard {integration} {dataKind} />
					{:else}
						<div class="p-2 border-error-300 border-2">
							<span>No supported integrations available for this data</span>
						</div>
					{/each}
				</div>
			{/if}
			
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
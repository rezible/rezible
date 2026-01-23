<script lang="ts">
	import { appShell } from "$features/app-shell";
	import Button from "$src/components/button/Button.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { useSetupViewState } from "../lib/viewState.svelte";
	import LoadingIndicator from "$src/components/loading-indicator/LoadingIndicator.svelte";
	import type { SupportedIntegration } from "$src/lib/api";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const view = useSetupViewState();
</script>

{#snippet oauthFlowButtonContent(name: string)}
	{#if name === "slack"}
	<img alt="Add to Slack" height="40" width="139" 
		src="https://platform.slack-edge.com/img/add_to_slack.png" 
		srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
	{:else}
		<span>Start OAuth Flow</span>
	{/if}
{/snippet}

{#snippet integrationSetupCard(intg: SupportedIntegration, dataKind: string)}
	{@const configured = view.configuredIntegrationsMap.get(intg.name)}
	{@const configValid = configured?.attributes.configValid}
	<div class="border p-2 flex flex-col gap-2">
		<span>{intg.name}</span>
		{#if !configured && intg.oauthRequired}
			<Button onclick={() => {view.oauth.startFlow(intg.name)}}>
				{@render oauthFlowButtonContent(intg.name)}
			</Button>
		{:else}
			{#if configValid}
				<span>enable support for {dataKind} button</span>
			{/if}
			<span>config form</span>
			<Button 
				variant="fill-light"
				color="secondary"
				onclick={() => {view.doConfigureIntegration(intg.name, {})}}
			>
				Do Initial Configure
			</Button>
		{/if}
	</div>
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
			<span>Data Kinds:</span>
			<div class="w-full flex gap-2">
				{#each view.requiredDataKinds as kind}
					<div class="p-2 border" class:bg-success-100={view.configuredEnabledDataKinds.has(kind)}>{kind}</div>
				{/each}
			</div>

			<!-- TODO: this should be tabs, with required data kinds first -->
			{#if !!view.nextRequiredDataKind}
				<div class="flex flex-col gap-2">
					{#each view.nextRequiredSupportedIntegrations as intg}
						{@render integrationSetupCard(intg, view.nextRequiredDataKind)}	
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
<script lang="ts">
	import { appShell } from "$features/app-shell";
	import Button from "$src/components/button/Button.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { useSetupViewState } from "../lib/viewState.svelte";
	import LoadingIndicator from "$src/components/loading-indicator/LoadingIndicator.svelte";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const view = useSetupViewState();

	const dataKinds = [
		{name: "Chat", kind: "chat", required: true},
		{name: "Users", kind: "users", required: true},
	];
	const requiredDataKinds = new Set(dataKinds.filter(k => k.required).map(k => k.kind));
	const configuredMap = $derived(new SvelteMap(view.configuredIntegrations.map(intg => ([intg.attributes.name, intg]))));
	const completedDataKinds = $derived(new SvelteSet(view.configuredIntegrations.flatMap(intg => (intg.attributes.enabledDataKinds))));
	const remainingRequiredKinds = $derived(requiredDataKinds.difference(completedDataKinds));
</script>

{#snippet oauthFlowButton(name: string)}
	<Button onclick={() => {view.oauth.startFlow(name)}}>
	{#if name === "slack"}
	<img alt="Add to Slack" height="40" width="139" 
		src="https://platform.slack-edge.com/img/add_to_slack.png" 
		srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
	{:else}
		<span>Start OAuth Flow</span>
	{/if}
	</Button>
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
		{:else if view.oauth.completeIntegrationErr}
			<span>complete integration error</span>
		{:else}
			<div class="flex flex-col gap-2">
				{#each view.supportedIntegrations as intg}
					{@const configured = configuredMap.get(intg.name)}
					<div class="border p-2">
						<span>{intg.name}</span>
						{#if !!configured?.attributes.configValid}
							<br />
							<span>valid</span>
						{:else}
							<br />
							{#if intg.oauthRequired}
								{@render oauthFlowButton(intg.name)}
							{:else}
								<Button 
									variant="fill-light"
									color="secondary"
									onclick={() => {view.doConfigureIntegration(intg.name, {})}}
								>
									Enable
								</Button>
							{/if}
						{/if}
					</div>
				{/each}
			</div>
			{#if remainingRequiredKinds.size === 0}
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
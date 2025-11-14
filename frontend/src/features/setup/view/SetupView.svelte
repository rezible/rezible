<script lang="ts">
	import { appShell } from "$features/app-shell";
	import Button from "$src/components/button/Button.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { useSetupViewState } from "../lib/viewState.svelte";
	import LoadingIndicator from "$src/components/loading-indicator/LoadingIndicator.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const view = useSetupViewState();
</script>

{#snippet flowButtonImg(id: string)}
	{#if id === "slack"}
	<img alt="Add to Slack" height="40" width="139" 
		src="https://platform.slack-edge.com/img/add_to_slack.png" 
		srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
	{:else}
		<span>unknown integration flow provider</span>
	{/if}
{/snippet}

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Organization Setup" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		{#if view.loading}
			<LoadingIndicator />
		{:else}
			{#if !!view.nextRequiredIntegrationId}
				{@const href = view.nextRequiredIntegrationFlowUrl}
				{@const flowErr = view.nextRequiredIntegrationFlowErr}

				{#if href}
					<a {href}>
						{@render flowButtonImg(view.nextRequiredIntegrationId)}
					</a>
				{:else if flowErr}
					<span>flow error: {flowErr}</span>
				{:else}
					<LoadingIndicator />
				{/if}
			{:else}
				<Button 
					color="secondary" 
					variant="fill"
					disabled={false}
					onclick={() => view.doFinishOrganizationSetup()} 
					loading={view.isFinishingSetup}
				>Finish setup</Button>
			{/if}
		{/if}
	</div>
</div>
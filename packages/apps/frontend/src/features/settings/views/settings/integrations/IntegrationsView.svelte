<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Empty from "$components/ui/empty";
	import { Skeleton } from "$components/ui/skeleton";

	import {
		type IntegrationProviderSummary,
		useIntegrationsController,
	} from "$features/settings/lib/integrationsController.svelte";

	import IntegrationProviderCard from "./integration-provider-card/IntegrationProviderCard.svelte";

	const controller = useIntegrationsController();

	registerPageDescriptor(() => ({
		title: "Integrations",
		parents: [
			{ label: "Settings", path: resolve("/settings") },
			{ label: "Organization", path: resolve("/settings/organization") },
		],
	}));

	const hasProviders = $derived(controller.providers.length > 0);
</script>

{#snippet providerSection(title: string, providers: IntegrationProviderSummary[])}
	<section class="@container flex flex-col gap-4">
		<h2 class="text-base font-semibold">{title}</h2>
		<!-- Fill the width with as many columns as fit; when only one fits, center it at a readable width. -->
		<div
			class="grid w-full grid-cols-[repeat(auto-fill,minmax(22rem,1fr))] gap-4 @max-[45rem]:mx-auto @max-[45rem]:max-w-md @max-[45rem]:grid-cols-1"
		>
			{#each providers as provider (provider.name)}
				<IntegrationProviderCard {provider} />
			{/each}
		</div>
	</section>
{/snippet}

<div class="flex w-full max-w-6xl flex-col gap-10">
	{#if controller.loading}
		<div class="grid grid-cols-[repeat(auto-fill,minmax(22rem,1fr))] gap-4">
			{#each { length: 3 }, index (index)}
				<Skeleton class="h-44 w-full" />
			{/each}
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} dismissable={false} />
	{:else if !hasProviders}
		<Empty.Root class="border">
			<Empty.Header>
				<Empty.Title>No integrations available</Empty.Title>
				<Empty.Description>This deployment does not have any integrations enabled.</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{:else}
		{#if controller.installedProviders.length > 0}
			{@render providerSection("Installed", controller.installedProviders)}
		{/if}
		{#if controller.availableProviders.length > 0}
			{@render providerSection("Available", controller.availableProviders)}
		{/if}
	{/if}
</div>

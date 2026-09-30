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

{#snippet providerSection(title: string, description: string, providers: IntegrationProviderSummary[])}
	<section class="flex flex-col gap-3">
		<div class="flex flex-col gap-0.5">
			<h2 class="text-sm font-semibold">{title}</h2>
			<p class="text-sm text-muted-foreground">{description}</p>
		</div>
		<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
			{#each providers as provider (provider.name)}
				<IntegrationProviderCard {provider} />
			{/each}
		</div>
	</section>
{/snippet}

<div class="flex max-w-5xl flex-col gap-8">
	{#if controller.loading}
		<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
			{#each { length: 3 }, index (index)}
				<Skeleton class="h-36 w-full" />
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
			{@render providerSection(
				"Installed",
				"Providers with saved connections.",
				controller.installedProviders
			)}
		{/if}
		{#if controller.availableProviders.length > 0}
			{@render providerSection(
				"Available",
				"Connect these providers to bring their data into Rezible.",
				controller.availableProviders
			)}
		{/if}
	{/if}
</div>

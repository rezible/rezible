<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";

	import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
	
	import * as Card from "$components/ui/card";
	
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { Badge } from "$src/components/ui/badge";
	import { resolve } from "$app/paths";

	const controller = useIntegrationsController();

	setPageBreadcrumbs(() => ([
		{ label: "Settings", path: "/settings" },
		{ label: "Integrations", path: "/settings/integrations"}
	]));
</script>

<div class="flex flex-col gap-4">
	{#if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading...</span>
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} />
	{:else}
		{#each controller.providers as provider}
			{@const hasInstalls = controller.installationsByProvider.has(provider.name)}
			<a href={resolve("/settings/integrations/[provider]", {provider: provider.name})}>
				<Card.Root class="min-w-xs max-w-sm gap-3 p-2">
					<Card.Header class="p-0">
						<Card.Title class="truncate">{provider.displayName}</Card.Title>
						{#if hasInstalls}
							<Card.Action>
								<Badge>Installed</Badge>
							</Card.Action>
						{/if}
					</Card.Header>

					<Card.Content class="flex flex-col gap-2 p-0">
						{provider.description}
					</Card.Content>
				</Card.Root>
			</a>
		{/each}
	{/if}
</div>

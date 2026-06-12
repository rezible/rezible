<script lang="ts">
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";

	import { initIntegrationProviderCardController } from "./controller.svelte";
	import RiCircleLine from "remixicon-svelte/icons/add-circle-line";
	import InstalledIntegrationPanel from "./InstalledIntegrationPanel.svelte";

	type Props = {
		provider: string;
		onlyName?: string;
		showInstalledDetails?: boolean;
	};
	const { provider, onlyName, showInstalledDetails = true }: Props = $props();

	const ctrl = initIntegrationProviderCardController(() => provider);
	const available = $derived(!!onlyName ? ctrl.available.filter(a => a.name === onlyName) : ctrl.available);
</script>

{#each available as integration (integration.name)}
	{@const installations = ctrl.integrations.installationsByName.get(integration.name) || []}
	<Card.Root class="min-w-xs gap-3 p-4">
		<Card.Header class="p-0">
			<div class="flex items-start justify-between gap-4">
				<div class="flex min-w-0 flex-col gap-2">
					<Card.Title class="truncate">{integration.displayName}</Card.Title>
					<Card.Description>{integration.description}</Card.Description>
				</div>
					<Button
						onclick={() => ctrl.integrations.openConfigureDialog(integration)}
						variant="outline"
						disabled={integration.maxInstalls !== undefined && integration.maxInstalls <= installations.length}
					>
						{installations.length > 0 ? "Install another" : "Install"}
						<RiCircleLine />
					</Button>
			</div>
		</Card.Header>

		{#if installations.length > 0}
		<Card.Content class="flex flex-col gap-2 p-0">
			{#each installations as installation (installation.id)}
				<InstalledIntegrationPanel {installation} openConfigDialog={() => {ctrl.integrations.openConfigureDialog(integration, installation)}} />
			{/each}
		</Card.Content>
		{/if}
	</Card.Root>
{/each}
<script lang="ts">
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";

	import { initIntegrationProviderCardController } from "./controller.svelte";
	import RiCircleLine from "remixicon-svelte/icons/add-circle-line";
	import InstalledIntegrationPanel from "./InstalledIntegrationPanel.svelte";
	import { cn } from "$src/lib/utils";

	type Props = {
		provider: string;
		onlyName?: string;
		compact?: boolean;
	};
	const { provider, onlyName, compact = false }: Props = $props();

	const ctrl = initIntegrationProviderCardController(() => provider);
	const available = $derived(!!onlyName ? ctrl.available.filter(a => a.name === onlyName) : ctrl.available);
</script>

{#each available as integration (integration.name)}
	{@const installations = ctrl.integrations.installationsByName.get(integration.name) || []}
	{@const canInstallAnother = !integration.maxInstalls || (installations.length < integration.maxInstalls)}
	<Card.Root class={cn("min-w-xs gap-3", compact ? "p-2" : "p-4")}>
		<Card.Header class="p-0">
			<Card.Title class="truncate">{integration.displayName}</Card.Title>
			<Card.Description>{integration.description}</Card.Description>
			{#if canInstallAnother}
			<Card.Action>
				<Button
					onclick={() => ctrl.integrations.openConfigureDialog(integration)}
					variant="outline"
				>
					{installations.length > 0 ? "Install another" : "Install"}
					<RiCircleLine />
				</Button>
			</Card.Action>
			{/if}
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
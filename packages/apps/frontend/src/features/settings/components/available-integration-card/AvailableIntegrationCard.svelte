<script lang="ts">
	import type { AvailableIntegration } from "$lib/api";

	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import { Badge } from "$components/ui/badge";

	import { initAvailableIntegrationCardController } from "./availableIntegrationController.svelte";

	import RiCircleLine from "remixicon-svelte/icons/add-circle-line";
	import InstalledIntegrationPanel from "./InstalledIntegrationPanel.svelte";

	type Props = {
		integration: AvailableIntegration;
		showInstalledDetails?: boolean;
	};
	const { integration, showInstalledDetails = true }: Props = $props();

	const ctrl = initAvailableIntegrationCardController(() => integration);
</script>

<Card.Root class="min-w-xs gap-3 p-4">
	<Card.Header class="p-0">
		<div class="flex items-start justify-between gap-4">
			<div class="flex min-w-0 flex-col gap-2">
				<div class="flex flex-wrap items-center gap-2">
					<Card.Title class="truncate">{ctrl.displayName}</Card.Title>
					{#if ctrl.hasInstalled}
						<Badge variant="secondary">{ctrl.installations.length} connected</Badge>
					{/if}
					{#if ctrl.supportedCapabilities.length > 0}
						<div class="flex flex-wrap gap-1">
							{#each ctrl.supportedCapabilities as capability (capability)}
								<Badge variant="outline">{capability}</Badge>
							{/each}
						</div>
					{/if}
				</div>
				{#if ctrl.description}
					<Card.Description>{ctrl.description}</Card.Description>
				{/if}
			</div>
			<Button
				onclick={() => ctrl.openConfigDialog()}
				variant="outline"
				disabled={ctrl.maxInstallsReached}
			>
				<RiCircleLine />
				{ctrl.hasInstalled ? "Connect another" : "Connect"}
			</Button>
		</div>
	</Card.Header>

	{#if showInstalledDetails && ctrl.hasInstalled}
		<Card.Content class="flex flex-col gap-2 p-0">
			{#each ctrl.installations as installation (installation.id)}
				<InstalledIntegrationPanel {installation} openConfigDialog={() => {ctrl.openConfigDialog(installation)}} />
			{/each}
		</Card.Content>
	{/if}
</Card.Root>

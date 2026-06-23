<script lang="ts">
	import { Badge } from "$src/components/ui/badge";
	import Button from "$src/components/ui/button/button.svelte";
	import * as Card from "$src/components/ui/card";
	import { providerDisplays } from "$src/features/settings/lib/integrationsController.svelte";
	import { useInitialSetupController } from "../../initialSetupController.svelte";

	const ctrl = useInitialSetupController();
</script>

<div class="flex flex-row gap-4 justify-between">
	<div class="space-y-1 max-w-xs">
		<h2 class="text-lg font-semibold">Suggested integrations</h2>
		<p class="text-sm text-muted-foreground">
			Configure commonly useful integrations, or skip this for now and install them later in Settings > Integrations.
		</p>
	</div>
	
	<div class="flex flex-col gap-3 flex-1">
		{#each ctrl.integrationSuggestions.entries() as [name, hasInstalls]}
			{@const info = providerDisplays.get(name)}
			<Card.Root class="min-w-xs max-w-sm gap-3 p-2">
				<Card.Header class="p-0">
					<Card.Title class="truncate">{info?.displayName || name}</Card.Title>
					<Card.Description>{info?.description || ""}</Card.Description>
						<Card.Action>
							{#if hasInstalls}
								<Badge>Installed</Badge>
							{/if}
							<Button variant={hasInstalls ? "secondary" : "outline"} onclick={() => {ctrl.openIntegrationProviderDialog(name)}}>
								Configure
							</Button>
						</Card.Action>
				</Card.Header>
			</Card.Root>
		{/each}
	</div>
</div>
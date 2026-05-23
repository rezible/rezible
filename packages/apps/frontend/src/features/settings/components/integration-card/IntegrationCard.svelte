<script lang="ts">
	import type { AvailableIntegration } from "$lib/api";

	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import { initIntegrationCardController } from "./controller.svelte";
	import InlineAlert from "$src/components/inline-alert/InlineAlert.svelte";
	import Spinner from "$src/components/ui/spinner/spinner.svelte";
	import IntegrationDataSync from "./IntegrationDataSync.svelte";
	import ConfiguredIntegrationPanel from "./ConfiguredIntegrationPanel.svelte";

	import RiCircleLine from "remixicon-svelte/icons/add-circle-line";
	import IntegrationOAuthSelect from "./IntegrationOAuthSelect.svelte";

	type Props = {
		integration: AvailableIntegration;
	};
	const { integration }: Props = $props();

	const ctrl = initIntegrationCardController(() => integration);

	const showContent = $derived(!!ctrl.configError || ctrl.oauthSelectionRequired || ctrl.hasConfigured || ctrl.showConfig);
</script>

<Card.Root class="gap-4 p-4 min-w-xs">
	<Card.Header class="p-0">
		<div class="flex items-center justify-between gap-4 h-fit">
			<div class="flex flex-col gap-2">
				<Card.Title class="capitalize">{integration.name}</Card.Title>
			</div>
			<div class="flex items-center gap-2">
				<Button
					onclick={() => {ctrl.startConfig()}}
					variant="outline"
					disabled={ctrl.loading}
				>
					Connect {ctrl.hasConfigured ? "another" : ""}
					<RiCircleLine />
				</Button>
			</div>
		</div>
	</Card.Header>

	<Card.Content class="p-0 flex flex-col gap-3 {showContent ? '' : 'hidden'}">
		{#if !!ctrl.configError}
			<InlineAlert bind:error={ctrl.configError} />
		{/if}

		{#if ctrl.oauthSelectionRequired}
			<IntegrationOAuthSelect />
		{:else if ctrl.showConfig}
			<ctrl.ConfigComponent />

			<div class="flex flex-wrap gap-2">
				<Button disabled={!ctrl.hasChanges || ctrl.loading} onclick={() => ctrl.saveConfig()}>
					{#if ctrl.loading}
						<Spinner />
						Saving...
					{:else}
						Save
					{/if}
				</Button>
				<Button disabled={ctrl.loading} onclick={() => ctrl.clearConfig()}>Cancel</Button>
			</div>
		{:else if ctrl.hasConfigured}
			<div class="flex flex-col gap-2">
				{#each ctrl.configured as configured (configured.id)}
					<ConfiguredIntegrationPanel {configured} />
				{/each}
			</div>
			<IntegrationDataSync />
		{:else}
			<span>No connections configured</span>
		{/if}
	</Card.Content>
</Card.Root>

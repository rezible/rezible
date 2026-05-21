<script lang="ts">
	import type { AvailableIntegration } from "$lib/api";

	import * as Card from "$components/ui/card";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { initIntegrationCardController } from "./controller.svelte";
	import InlineAlert from "$src/components/inline-alert/InlineAlert.svelte";
	import Spinner from "$src/components/ui/spinner/spinner.svelte";
	import RiGithubFill from "remixicon-svelte/icons/github-fill";
	import { Checkbox } from "$components/ui/checkbox";
	import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";
	import IntegrationDataSync from "./IntegrationDataSync.svelte";
	import ConfiguredIntegrationPanel from "./ConfiguredIntegrationPanel.svelte";

	type Props = {
		integration: AvailableIntegration;
	};
	const { integration }: Props = $props();

	const ctrl = initIntegrationCardController(() => integration);
	const oauth = useIntegrationOAuthController();

	const showContent = $derived(!!ctrl.configError || ctrl.oauthSelectionRequired || ctrl.hasConfigured || ctrl.showConfig);
</script>

<Card.Root class="gap-4 p-4 min-w-xs">
	<Card.Header class="p-0">
		<div class="flex items-center justify-between gap-4 h-fit">
			<div class="flex flex-col gap-2">
				<Card.Title class="capitalize">{integration.name}</Card.Title>
			</div>
			<div class="flex items-center gap-2">
				{#if ctrl.hasConfigured}
					<Badge variant="secondary">{ctrl.configured.length} configured</Badge>
				{:else if !ctrl.showConfig}
					<Button
						onclick={() => {
							ctrl.startConfigure();
						}}
						variant="outline"
						disabled={ctrl.loading}
					>
						Connect {ctrl.hasConfigured ? "another" : ""}
					</Button>
				{/if}
			</div>
		</div>
	</Card.Header>

	<Card.Content class="p-0 flex flex-col gap-3 {showContent ? '' : 'hidden'}">
		{#if !!ctrl.configError}
			<InlineAlert bind:error={ctrl.configError} />
		{/if}

		{#if ctrl.oauthSelectionRequired}
			<div class="flex flex-col gap-3 rounded-md border p-3">
				<div class="flex flex-col gap-1">
					<span class="text-sm font-medium">Select installations</span>
					<span class="text-sm text-muted-foreground">Choose which accounts to connect.</span>
				</div>
				<div class="flex flex-col gap-2">
					{#each oauth.selectionOptions as option (option.externalRef)}
						<label class="flex items-center gap-3 rounded-md border p-3 text-sm">
							<Checkbox
								checked={oauth.selectedExternalRefs.has(option.externalRef)}
								onCheckedChange={(checked) =>
									oauth.toggleSelection(option.externalRef, !!checked)}
							/>
							<span class="flex flex-col">
								<span class="font-medium">{option.displayName}</span>
								<span class="text-muted-foreground">{option.externalRef}</span>
							</span>
						</label>
					{/each}
				</div>
				<Button
					disabled={ctrl.loading || oauth.selectedExternalRefs.size === 0}
					onclick={() => oauth.selectOAuthOptions()}
				>
					{#if ctrl.loading}
						<Spinner />
					{/if}
					Connect selected
				</Button>
			</div>
		{:else}
			{#if ctrl.hasConfigured}
				<div class="flex flex-col gap-2">
					{#each ctrl.configured as configured (configured.id)}
						<ConfiguredIntegrationPanel {configured} />
					{/each}
				</div>
				<IntegrationDataSync />
			{/if}

			{#if ctrl.showConfig}
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
			{/if}
		{/if}
	</Card.Content>
</Card.Root>

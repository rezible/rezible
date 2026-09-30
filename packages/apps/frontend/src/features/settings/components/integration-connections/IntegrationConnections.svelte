<script lang="ts">
	import type { Snippet } from "svelte";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiArrowUpSLine from "remixicon-svelte/icons/arrow-up-s-line";

	import type { IntegrationInstallation } from "$lib/api";
	import { Button } from "$components/ui/button";

	import IntegrationConnection from "../integration-connection/IntegrationConnection.svelte";
	import { useIntegrationProviderController } from "../integration-provider/controller.svelte";

	type Props = {
		name: string;
		referenceLabel?: string;
		// Shown when there are no connections yet.
		empty?: Snippet;
		// Extra actions for each connection.
		connectionActions?: Snippet<[IntegrationInstallation]>;
		// Settings for one connection. A single connection shows them directly; in a list they open on demand.
		settings?: Snippet<[IntegrationInstallation]>;
	};
	const { name, referenceLabel, empty, connectionActions, settings }: Props = $props();

	const ctrl = useIntegrationProviderController();

	const installations = $derived(ctrl.integrations.installationsFor(name));
	const onlyInstallation = $derived(installations.length === 1 ? installations[0] : undefined);
</script>

{#if installations.length === 0}
	{@render empty?.()}
{:else if onlyInstallation}
	<IntegrationConnection installation={onlyInstallation} {referenceLabel}>
		{#snippet actions()}
			{@render connectionActions?.(onlyInstallation)}
		{/snippet}

		{#if settings}
			{@render settings(onlyInstallation)}
		{/if}
	</IntegrationConnection>
{:else}
	<div class="flex flex-col gap-2">
		{#each installations as installation (installation.id)}
			{@const open = ctrl.isOpen(installation.id)}
			<IntegrationConnection {installation} {referenceLabel}>
				{#snippet actions()}
					{@render connectionActions?.(installation)}
					{#if settings}
						<Button
							variant="ghost"
							size="sm"
							aria-expanded={open}
							onclick={() => ctrl.toggleOpen(installation.id)}
						>
							Settings
							{#if open}
								<RiArrowUpSLine data-icon="inline-end" />
							{:else}
								<RiArrowDownSLine data-icon="inline-end" />
							{/if}
						</Button>
					{/if}
				{/snippet}

				{#if settings && open}
					{@render settings(installation)}
				{/if}
			</IntegrationConnection>
		{/each}
	</div>
{/if}

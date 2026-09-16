<script lang="ts">
	import { resolve } from "$app/paths";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import * as Popover from "$components/ui/popover";
	import * as Tooltip from "$components/ui/tooltip";
	import SituationStatus from "$features/situations/components/situation-status/SituationStatus.svelte";
	import type { SituationController } from "./controller.svelte";
	import type { Incident } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";

	type Props = { controller: SituationController };
	let { controller }: Props = $props();

	const situation = $derived(controller.situation);
</script>

{#snippet incidentLink(incident: Incident)}
	{@const attrs = incident.attributes}
	<Tooltip.Root>
		<Tooltip.Trigger>
			{#snippet child({ props })}
				<Button
					{...props}
					variant="outline"
					size="sm"
					class="min-w-0 max-w-full shrink"
					href={resolve("/incidents/[slug]/[[view=incidentView]]", {
						slug: attrs.slug,
					})}
				>
					<span class="min-w-0 truncate">{attrs.title}</span>
					<span class="shrink-0 text-status-warning-foreground">
						{attrs.currentStatus}
					</span>
				</Button>
			{/snippet}
		</Tooltip.Trigger>
		<Tooltip.Content>{attrs.title} - {attrs.slug}</Tooltip.Content>
	</Tooltip.Root>
{/snippet}

<div class="flex min-w-0 max-w-full flex-wrap items-center justify-end gap-2 sm:max-w-md sm:flex-nowrap">
	{#if situation}
		<div class="shrink-0">
			<SituationStatus attributes={situation.attributes} />
		</div>
		{#if controller.incidentIds.length === 0}
			<span class="text-xs text-muted-foreground">No linked incident</span>
		{:else if controller.incidentIds.length === 1}
			{@const query = controller.incidentsQuery[0]}
			<LoadingQueryWrapper {query} view={incidentLink}>
				{#snippet loading()}
					<span role="status" class="flex items-center gap-2 text-xs">
						<Spinner />
						Loading incident
					</span>
				{/snippet}
				{#snippet error()}
					<Button 
						variant="ghost" size="sm" 
						onclick={() => query?.refetch()}
					>Incident unavailable - Retry</Button>
				{/snippet}
			</LoadingQueryWrapper>
		{:else}
			<Popover.Root>
				<Popover.Trigger>
					{#snippet child({ props })}
						<Button {...props} 
							variant="outline" 
							size="sm"
						>{controller.incidentIds.length} linked incidents</Button>
					{/snippet}
				</Popover.Trigger>
				<Popover.Content align="end" class="flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-3">
					<h2 class="font-semibold">Linked incidents</h2>
					{#each controller.incidentIds as id, index (id)}
						{@const query=controller.incidentsQuery[index]}
						<LoadingQueryWrapper {query} view={incidentLink}>
							{#snippet loading()}
								<span role="status" class="flex items-center gap-2 text-xs">
									<Spinner /> Loading incident
								</span>
							{/snippet}
							{#snippet error()}
								<Button 
									variant="ghost" size="sm" 
									onclick={() => query?.refetch()}
								>Incident unavailable - Retry</Button>
							{/snippet}
						</LoadingQueryWrapper>
					{/each}
				</Popover.Content>
			</Popover.Root>
		{/if}
	{/if}
</div>

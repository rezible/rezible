<script lang="ts">
	import { resolve } from "$app/paths";
	import { createQuery } from "@tanstack/svelte-query";
	import { Badge } from "$components/ui/badge";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { listIncidentsOptions } from "$lib/api";

	type Props = {
		id?: string;
	};

	const { id }: Props = $props();
	const incidentsQuery = createQuery(() => ({
		...listIncidentsOptions({}),
		enabled: !!id,
	}));
</script>

<div class="p-4">
	<LoadingQueryWrapper query={incidentsQuery}>
		{#snippet view(incidents)}
			{#if incidents.length > 0}
				<div class="divide-y divide-border rounded border border-border">
					{#each incidents as incident (incident.id)}
						<a
							href={resolve("/incidents/[slug]/[[view=incidentView]]", { slug: incident.attributes.slug })}
							class="flex items-center justify-between gap-3 px-3 py-3 hover:bg-muted/50"
						>
							<span class="min-w-0 truncate text-sm font-medium text-foreground">
								{incident.attributes.title}
							</span>
							<Badge variant="outline">{incident.attributes.severity.attributes.name}</Badge>
						</a>
					{/each}
				</div>
			{:else}
				<div class="rounded border border-dashed border-border p-6 text-sm text-muted-foreground">
					No incident links found for this entity.
				</div>
			{/if}
		{/snippet}
	</LoadingQueryWrapper>
</div>

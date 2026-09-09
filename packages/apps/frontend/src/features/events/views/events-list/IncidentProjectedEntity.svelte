<script lang="ts">
	import type { EventProjectionEntity, IncidentAttributes } from "$lib/api";
	import { getIncidentOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { Badge } from "$components/ui/badge";
	import { resolve } from "$app/paths";

	type Props = {
		entity: EventProjectionEntity;
	};
	const { entity }: Props = $props();

	type IncidentStatus = IncidentAttributes["currentStatus"];

	const statusLabels: Record<IncidentStatus, string> = {
		started: "Started",
		mitigated: "Mitigated",
		resolved: "Resolved",
	};

	const statusClasses: Record<IncidentStatus, string> = {
		started: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
		mitigated: "border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300",
		resolved: "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
	};

	const query = createQuery(() => getIncidentOptions({ path: { id: entity.entityId } }));
	const incident = $derived(query.data?.data);

	function normalizeStatus(value: string | undefined): IncidentStatus {
		return value && value in statusLabels ? (value as IncidentStatus) : "started";
	}
</script>

{#if query.isLoading}
	<div class="rounded-md border p-2 text-xs text-muted-foreground">Loading incident...</div>
{:else if query.isError || !incident}
	<span>invalid incident</span>
{:else}
	{@const attrs = incident.attributes}
	{@const status = normalizeStatus(attrs.currentStatus)}
	<a
		href={resolve("/incidents/[slug]/[[view=incidentView]]", { slug: attrs.slug || incident.id })}
		class="grid gap-2 rounded-md border bg-muted/20 p-3 transition hover:border-primary/40 hover:bg-muted/40"
	>
		<div class="flex flex-wrap items-center gap-2">
			<Badge variant="outline" class={statusClasses[status]}>{statusLabels[status]}</Badge>
			{#if attrs.severity?.attributes.name}
				<Badge variant="secondary">{attrs.severity.attributes.name}</Badge>
			{/if}
			{#if attrs.type?.attributes.name}
				<Badge variant="outline">{attrs.type.attributes.name}</Badge>
			{/if}
			<Badge variant="outline">incident</Badge>
		</div>
		<div class="min-w-0">
			<div class="truncate text-sm font-medium">{attrs.title}</div>
			{#if attrs.summary}
				<div class="mt-1 line-clamp-2 text-xs text-muted-foreground">{attrs.summary}</div>
			{/if}
		</div>
	</a>
{/if}

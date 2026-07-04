<script lang="ts">
	import type { EventProjection } from "$lib/api";
	import { Badge } from "$components/ui/badge";
	import ProjectedEntity from "./ProjectedEntity.svelte";

	type Props = {
		projections: EventProjection[];
	};
	const { projections }: Props = $props();

	const statusClasses: Record<EventProjection["attributes"]["status"], string> = {
		pending: "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
		succeeded: "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
		failed: "border-destructive/40 bg-destructive/10 text-destructive",
	};

	function formatDateTime(value: string | undefined) {
		if (!value) return "Not started";
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return "Not started";
		return new Intl.DateTimeFormat(undefined, {
			month: "short",
			day: "numeric",
			hour: "numeric",
			minute: "2-digit",
		}).format(date);
	}
</script>

<div class="grid gap-3">
	<div class="text-sm font-medium">Projections</div>

	{#each projections as projection (projection.id)}
		{@const attrs = projection.attributes}
		<div class="grid gap-3 rounded-md border bg-background p-3">
			<div class="flex flex-wrap items-center justify-between gap-2">
				<div class="min-w-0">
					<div class="truncate text-sm font-medium">{attrs.projector}</div>
					<div class="text-xs text-muted-foreground">Started {formatDateTime(attrs.startedAt)}</div>
				</div>
				<Badge variant="outline" class={statusClasses[attrs.status]}>{attrs.status}</Badge>
			</div>

			{#if attrs.error}
				<div class="rounded-md border border-destructive/30 bg-destructive/10 p-2 text-xs text-destructive">
					{attrs.error}
				</div>
			{/if}

			<div class="grid gap-2">
				<div class="text-xs font-medium text-muted-foreground">Projected entities</div>
				{#each attrs.entities as entity (`${projection.id}-${entity.entityKind}-${entity.entityId}`)}
					<ProjectedEntity {entity} />
				{:else}
					<div class="rounded-md border border-dashed p-2 text-xs text-muted-foreground">
						No projected entities
					</div>
				{/each}
			</div>
		</div>
	{:else}
		<div class="rounded-md border border-dashed p-3 text-sm text-muted-foreground">
			No projections recorded for this event
		</div>
	{/each}
</div>

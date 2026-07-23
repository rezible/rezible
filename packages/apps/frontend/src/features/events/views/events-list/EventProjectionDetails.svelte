<script lang="ts">
	import type { EventProjection } from "$lib/api";
	import ProjectedEntity from "./ProjectedEntity.svelte";

	type Props = {
		projection?: EventProjection;
	};
	const { projection }: Props = $props();

	function formatDateTime(value: string | undefined) {
		if (!value) return "Unknown";
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return "Unknown";
		return new Intl.DateTimeFormat(undefined, {
			month: "short",
			day: "numeric",
			hour: "numeric",
			minute: "2-digit",
		}).format(date);
	}
</script>

<div class="grid gap-3">
	<div class="text-sm font-medium">Projection</div>

	{#if projection}
		{@const attrs = projection.attributes}
		<div class="grid gap-3 rounded-md border bg-background p-3">
			<div class="text-xs text-muted-foreground">
				Completed {formatDateTime(attrs.completedAt)}
			</div>

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
			No projection recorded for this event
		</div>
	{/if}
</div>

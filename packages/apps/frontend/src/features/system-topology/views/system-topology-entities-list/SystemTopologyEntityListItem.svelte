<script lang="ts">
	import { resolve } from "$app/paths";
	import { Badge } from "$components/ui/badge";
	import type { SystemTopologyEntity } from "$lib/api";

	type Props = {
		entity: SystemTopologyEntity;
	};

	const { entity }: Props = $props();

	const description = $derived(entity.attributes.description?.trim() || "No description provided.");
	// TODO: utils truncate helper
	const maxDesc = 120;
	const descSnippet = $derived(description.length > maxDesc ? `${description.slice(0, maxDesc - 3)}...` : description);
	const relationshipCount = $derived(entity.attributes.relationships?.length ?? 0);
</script>

<a
	href={resolve("/system/[id]/[[view=systemTopologyEntityView]]", { id: entity.id })}
	class="group grid gap-2 border-b border-border px-3 py-3 transition-colors hover:bg-muted/50"
>
	<div class="flex min-w-0 items-center justify-between gap-3">
		<div class="min-w-0">
			<div class="truncate text-sm font-medium text-foreground group-hover:text-primary">
				{entity.attributes.displayName}
			</div>
			<p class="mt-1 line-clamp-2 text-sm text-muted-foreground">
				{descSnippet}
			</p>
		</div>
		<div class="flex shrink-0 items-center gap-2">
			<Badge variant="secondary">{entity.attributes.kind}</Badge>
			<Badge variant="outline">{relationshipCount} links</Badge>
		</div>
	</div>
</a>

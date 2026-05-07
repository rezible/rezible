<script lang="ts">
	import { resolve } from "$app/paths";
	import { Badge } from "$components/ui/badge";
	import type { SystemComponent } from "$lib/api";

	type Props = {
		component: SystemComponent;
	};

	const { component }: Props = $props();

	const description = $derived(component.attributes.description?.trim() || "No description provided.");
	const snippet = $derived(description.length > 120 ? `${description.slice(0, 117)}...` : description);
	const kindLabel = $derived(component.attributes.kind?.attributes.label || "Unclassified");
</script>

<a
	href={resolve("/system/[id]/[[view=systemComponentView]]", { id: component.id })}
	class="group grid gap-2 border-b border-border px-3 py-3 transition-colors hover:bg-muted/50"
>
	<div class="flex min-w-0 items-center justify-between gap-3">
		<div class="min-w-0">
			<div class="truncate text-sm font-medium text-foreground group-hover:text-primary">
				{component.attributes.name}
			</div>
			<p class="mt-1 line-clamp-2 text-sm text-muted-foreground">
				{snippet}
			</p>
		</div>
		<div class="flex shrink-0 items-center gap-2">
			<Badge variant="secondary">{kindLabel}</Badge>
			<Badge variant="outline">{component.attributes.relationshipCount} links</Badge>
		</div>
	</div>
</a>

<script lang="ts">
	import { resolve } from "$app/paths";
	import type { Incident } from "$lib/api";
	import { Badge } from "$components/ui/badge";

	type Props = {
		incident: Incident;
	};

	let { incident }: Props = $props();

	const attrs = $derived(incident.attributes);
</script>

<a
	href={resolve("/incidents/[slug]/[[view=incidentView]]", { slug: attrs.slug })}
	class="block min-w-0 space-y-1 border-b border-border px-3 py-3 last:border-b-0 hover:bg-muted/50 focus-visible:outline-primary"
>
	<div class="flex flex-wrap items-center gap-2">
		<span class="min-w-0 break-words text-sm font-medium">{attrs.title}</span>
		<Badge variant="outline" class="capitalize">{attrs.currentStatus}</Badge>
		{#if attrs.severity?.attributes.name}
			<Badge variant="secondary">{attrs.severity.attributes.name}</Badge>
		{/if}
	</div>
	<div class="flex flex-wrap justify-between gap-x-3 text-xs text-muted-foreground">
		<span>{attrs.slug}</span>
		<time datetime={attrs.openedAt}>Opened {new Date(attrs.openedAt).toLocaleString()}</time>
	</div>
</a>

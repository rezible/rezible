<script lang="ts">
	import type { EventProjectionEntity } from "$lib/api";
	import { getAlertOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { Badge } from "$components/ui/badge";
	import { resolve } from "$app/paths";

	type Props = {
		entity: EventProjectionEntity;
	};
	const { entity }: Props = $props();

	const query = createQuery(() => getAlertOptions({ path: { id: entity.entityId } }));
	const alert = $derived(query.data?.data);
</script>

{#if query.isLoading}
	<div class="rounded-md border p-2 text-xs text-muted-foreground">Loading alert...</div>
{:else if query.isError || !alert}
	<span>invalid alert</span>
{:else}
	<a
		href={resolve("/alerts/[id]/[[view=alertView]]", { id: alert.id })}
		class="grid gap-2 rounded-md border bg-muted/20 p-3 transition hover:border-primary/40 hover:bg-muted/40"
	>
		<div class="flex flex-wrap items-center gap-2">
			<Badge variant="outline">alert</Badge>
		</div>
		<div class="min-w-0">
			<div class="truncate text-sm font-medium">{alert.attributes.title}</div>
			{#if alert.attributes.description}
				<div class="mt-1 line-clamp-2 text-xs text-muted-foreground">
					{alert.attributes.description}
				</div>
			{:else if alert.attributes.definition}
				<div class="mt-1 line-clamp-2 text-xs text-muted-foreground">
					{alert.attributes.definition}
				</div>
			{/if}
		</div>
	</a>
{/if}

<script lang="ts">
	import type { EventProjectionEntity } from "$lib/api";
	import { Badge } from "$components/ui/badge";
	import AlertProjectedEntity from "./AlertProjectedEntity.svelte";
	import IncidentProjectedEntity from "./IncidentProjectedEntity.svelte";

	type Props = {
		entity: EventProjectionEntity;
	};
	const { entity }: Props = $props();

	const entityKind = $derived(entity.entityKind.toLowerCase());
</script>

{#if entityKind === "incident"}
	<IncidentProjectedEntity {entity} />
{:else if entityKind === "alert"}
	<AlertProjectedEntity {entity} />
{:else}
	<div class="flex min-w-0 flex-wrap items-center gap-2 rounded-md border bg-muted/20 p-2 text-xs">
		<Badge variant="outline" class="capitalize">{entity.entityKind}</Badge>
		<span class="break-all font-mono text-muted-foreground">{entity.entityId}</span>
	</div>
{/if}

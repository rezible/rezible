<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listIncidentSystemComponentsOptions } from "$lib/api";

	import { incidentCtx } from "$features/incidents/lib/context";
	import { Button, ListItem } from "svelte-ux";

	const incidentId = incidentCtx.get().id;
	const componentsQuery = createQuery(() =>
		listIncidentSystemComponentsOptions({ path: { id: incidentId } }),
	);
	const incidentComponents = $derived(componentsQuery.data?.data ?? []);
</script>

<div>
	{#each incidentComponents as cmp (cmp.id)}
		<ListItem
			title={cmp.attributes.role}
			subheading={cmp.attributes.role}
			avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
			class="flex-1"
			noShadow
		>
			<div slot="actions">
				<Button>Do Something</Button>
			</div>
		</ListItem>
	{/each}
	{#if incidentComponents.length === 0}
		<span>No components linked to this incident</span>
	{/if}
</div>

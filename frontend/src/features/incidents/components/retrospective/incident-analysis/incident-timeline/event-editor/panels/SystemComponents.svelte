<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listSystemAnalysisComponentsOptions } from "$lib/api";

	import { incidentCtx } from "$features/incidents/lib/context";
	import { Button, Icon, ListItem } from "svelte-ux";
	import { mdiShapeSquareRoundedPlus } from "@mdi/js";
	import { getIconForComponentKind } from "$lib/systemComponents";

	const analysisId = incidentCtx.get().attributes.system_analysis_id;
	const componentsQuery = createQuery(() =>
		listSystemAnalysisComponentsOptions({ path: { id: analysisId } })
	);
	const incidentComponents = $derived(componentsQuery.data?.data ?? []);
</script>

<div>
	{#each incidentComponents as c (c.id)}
		{@const attr = c.attributes.component.attributes}
		<ListItem
			title={attr.name}
			subheading={attr.description}
			avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
			class="flex-1"
			noShadow
		>
			<div slot="avatar" class="rounded-xl size-8 grid place-content-center">
				<Icon data={getIconForComponentKind(attr.kind)} classes={{root: "size-5"}} />
			</div>
			<div slot="actions">
				<Button icon={mdiShapeSquareRoundedPlus} iconOnly />
			</div>
		</ListItem>
	{/each}
	
	{#if incidentComponents.length === 0}
		<span>No components linked to this incident</span>
	{/if}
</div>

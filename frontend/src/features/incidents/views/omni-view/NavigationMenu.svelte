<script lang="ts">
    import type { Incident } from "$lib/api";
    import type { IncidentViewRouteParam } from "$src/params/incidentView";

	type Props = {
		incidentSlug: string;
		retroType: string;
		viewParam: IncidentViewRouteParam;
	}
	const { incidentSlug, retroType, viewParam }: Props = $props();

	const fullRetroViewGroups = [
		{label: "Timeline", route: "timeline"},
		{label: "Analysis", route: "analysis"},
		{label: "Report", route: "report"},
	];
	const quickRetroViewGroups = [
		{label: "Report", route: "report"}
	];

	const viewGroups = $derived([
		{
			label: "Details",
			children: [{label: "Overview", route: ""}]
		},
		{
			label: "Retrospective",
			children: retroType === "full" ? fullRetroViewGroups : quickRetroViewGroups,
		},
	]);

	const currRoute = $derived(viewParam || "");
</script>

<div class="w-40 flex flex-col gap-2 overflow-y-auto">
	{#each viewGroups as g, i}
		<div class="border p-2 bg-surface-200 flex flex-col gap-1">
			<span class="text-surface-content/75">{g.label}</span>
			{#each g.children as v}
				{@const active = (v.route === currRoute)}
				<a href="/incidents/{incidentSlug}/{v.route}">
					<div class="p-2 rounded border" class:border-r-4={active} class:bg-primary-600={active} class:text-primary-content={active}>
						<span>{v.label}</span>
					</div>
				</a>
			{/each}
		</div>
	{/each}
</div>
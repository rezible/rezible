<script lang="ts" module>
	import type { IncidentMilestone } from "$lib/api";
	export type Props = {
		milestone?: IncidentMilestone;
		selected: boolean;
		onSelect?: (id: string, trigger: HTMLElement) => void;
	};
</script>

<script lang="ts">
	import { milestoneLabels } from "$features/incidents/lib/status";

	let { milestone, selected, onSelect }: Props = $props();

	const label = $derived(milestone ? milestoneLabels[milestone.attributes.kind] : "");
</script>

<button
	type="button"
	class="rounded px-2 py-1 text-left focus-visible:outline-2 focus-visible:outline-primary"
	class:ring-1={selected}
	class:ring-primary={selected}
	aria-pressed={selected}
	onclick={(event) => {
		event.stopPropagation();
		if (milestone) onSelect?.(milestone.id, event.currentTarget);
	}}
>
	{label}
</button>

<script lang="ts" module>
	import type { IncidentMilestone } from "$lib/api";
	export type Props = {
		milestone: IncidentMilestone;
		selected: boolean;
	};
</script>

<script lang="ts">
	let { milestone, selected }: Props = $props();
</script>

<span>{milestone.attributes.kind} selected:{selected}</span>

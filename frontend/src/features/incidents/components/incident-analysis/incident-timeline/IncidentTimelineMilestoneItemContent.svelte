<script lang="ts" module>
	import type { IncidentMilestone } from "$lib/api";
	export type Props = {
		milestone: IncidentMilestone;
	};
</script>

<script lang="ts">
	let { milestone }: Props = $props();
</script>

<span>{milestone.attributes.kind}</span>

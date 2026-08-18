<script lang="ts" module>
	import type { TimelineAnalysisEntry } from "./entry-model";

	export type Props = {
		event?: TimelineAnalysisEntry;
		selected: boolean;
	};
</script>

<script lang="ts">
	const { event, selected }: Props = $props();
</script>

<div class="p-2" class:ring-1={selected} class:ring-primary={selected}>
	<span>{event?.attributes.title || "title"}</span>
</div>

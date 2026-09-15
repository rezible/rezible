<script lang="ts" module>
	import type { TimelineAnalysisEntry } from "./entry-model";
	export type Props = {
		event?: TimelineAnalysisEntry;
		selected: boolean;
		onSelect: (id: string, trigger: HTMLElement) => void;
	};
</script>

<script lang="ts">
	let { event, selected, onSelect }: Props = $props();
</script>

<button
	type="button"
	class="rounded p-2 text-left focus-visible:outline-2 focus-visible:outline-primary"
	class:ring-1={selected}
	class:ring-primary={selected}
	aria-pressed={selected}
	onclick={(click) => {
		click.stopPropagation();
		if (event) onSelect(event.id, click.currentTarget);
	}}
>
	{event?.attributes.title}
</button>

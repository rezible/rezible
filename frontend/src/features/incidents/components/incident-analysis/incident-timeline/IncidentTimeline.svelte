<script lang="ts">
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
	import { timeline } from "./timeline.svelte";
	import EventDialog from "./event-dialog/EventDialog.svelte";
	import IncidentTimelineActionsBar from "./IncidentTimelineActionsBar.svelte";

	let containerEl = $state<HTMLElement>();
	timeline.setup(() => containerEl);
</script>

<div class="w-full h-full overflow-y-hidden" bind:this={containerEl}></div>

<IncidentTimelineActionsBar />

<EventDialog />

<style lang="postcss">
	/* the root timeline container element */
	:global(.vis-timeline) {
		@apply bg-surface-200;
		border-color: #3e3e3e;
	}

	/* horizontal timeline line (item dots connect to this) */
	:global(.vis-panel.vis-center) {
		border-color: oklch(var(--border-surface-content) * .75);
		border-color: #b1b1b7;
	}

	/* vertical grid lines */
	:global(.vis-time-axis .vis-grid.vis-minor) {
		border-color: #b1b1b7;
	}
	:global(.vis-time-axis .vis-grid.vis-major) {
		border-color: #b1b1b7;
	}

	/* item dot connecting to horizontal timeline line */
	:global(.vis-item.vis-dot) {}

	/* line from dot to item */
	:global(.vis-item, .vis-item.vis-line) {}

	/* timeline item */
	:global(.vis-item) {
		@apply bg-neutral text-neutral-content border-accent;
	}

	/* selected timeline item */
	:global(.vis-item.vis-selected) {
		@apply bg-accent text-accent-content border-accent;
	}

	/* text of time axis, eg Mon 10 */
	:global(.vis-time-axis .vis-text) {
		padding-top: 4px;
		padding-left: 4px;
	}

	/* text of major time axis, eg February 2025 */
	:global(.vis-time-axis .vis-text.vis-major) {
		font-weight: bold;
	}

	/* current time vertical line */
	:global(.vis-current-time) {
		@apply bg-secondary;
	}
</style>
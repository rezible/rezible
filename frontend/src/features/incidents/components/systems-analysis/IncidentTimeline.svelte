<script lang="ts">
	import { onMount } from "svelte";
	import { DataSet } from "vis-data/esnext";
	import {
		Timeline,
		type DataItem,
		type IdType,
		type TimelineOptions,
	} from "vis-timeline/esnext";
	import "vis-timeline/dist/vis-timeline-graph2d.min.css";
    import { createEventTemplateElement } from "./timeline-event.svelte";

	let timelineEl = $state<HTMLElement>();
	let timeline = $state<Timeline>();

	const items = $state(new DataSet<any>([
		{
			id: "A",
			content: "Period A",
			start: "2014-01-16",
			end: "2014-01-22",
			type: "background",
		},
		{
			id: "B",
			content: "Period B",
			start: "2014-01-25",
			end: "2014-01-30",
			type: "background",
			className: "negative",
		},
	]));

	onMount(() => {
		if (!timelineEl) return;

		const eventComponents = new Map<IdType, ReturnType<typeof createEventTemplateElement>>();
		const addItem = (id: IdType) => {
			const created = createEventTemplateElement(id.toString());
			items.add({ id: 1, content: created.element, start: "2014-01-23" });
			eventComponents.set(id, created);
		}
		addItem("bleh");

		const options: TimelineOptions = {
			height: "100%",
			// template: templateTimelineEvent,
		};
		timeline = new Timeline(timelineEl, items, options);

		return () => {
			timeline?.destroy();
			eventComponents.forEach(c => c.unmount());
		};
	});
</script>

<div class="h-full" bind:this={timelineEl}></div>

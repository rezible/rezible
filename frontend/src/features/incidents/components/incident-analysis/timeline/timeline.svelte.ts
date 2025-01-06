import { mount, unmount } from "svelte";
import { Timeline, type IdType, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import IncidentTimelineEvent, { type TimelineEventComponentProps } from "./IncidentTimelineEvent.svelte";

export const createTimelineEventElement = (id: string) => {
	let props = $state<TimelineEventComponentProps>({label: "initial"});
	
	const target = document.createElement("div");
	target.setAttribute("event-id", id);

	const component = mount(IncidentTimelineEvent, {target, props});
	
	return {
		get element () {return target},
		setLabel: (label: string) => {props.label = label},
		unmount: () => {unmount(component)}
	}
}

const createTimelineState = () => {
	let timeline = $state<Timeline>();

	const items = new DataSet<any>([
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
	]);

	const eventComponents = new Map<IdType, ReturnType<typeof createTimelineEventElement>>();
	const addEvent = (id: IdType) => {
		const created = createTimelineEventElement(id.toString());
		items.add({ id: 1, content: created.element, start: "2014-01-23" });
		eventComponents.set(id, created);
	}

	const mount = (container: HTMLElement) => {
		addEvent("bleh");

		const options: TimelineOptions = {
			height: "100%",
		};
		timeline = new Timeline(container, items, options);
	}

	const unmount = () => {
		timeline?.destroy();
		eventComponents.forEach(c => c.unmount());
		eventComponents.clear();
		items.clear();
	}

	return {
		mount,
		addEvent,
		unmount,
	}
}
export const timeline = createTimelineState();
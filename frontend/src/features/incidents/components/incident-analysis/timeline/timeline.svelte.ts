import { mount, unmount } from "svelte";
import { Timeline, type IdType, type TimelineOptions } from "vis-timeline/esnext";
import { DataSet } from "vis-data/esnext";

import IncidentTimelineEvent, { type TimelineEventComponentProps } from "./IncidentTimelineEvent.svelte";

export const createEventTemplateElement = (id: string) => {
	let props = $state<TimelineEventComponentProps>({label: "initial"});
	const element = document.createElement("div");
	element.setAttribute("event-id", id);
	const component = mount(IncidentTimelineEvent, {target: element, props});
	
	return {
		get element () {return element},
		setLabel: (label: string) => {props.label = label},
		unmount: () => {unmount(component)}
	}
}

const createTimelineState = () => {
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

	const eventComponents = new Map<IdType, ReturnType<typeof createEventTemplateElement>>();
	const addItem = (id: IdType) => {
		const created = createEventTemplateElement(id.toString());
		items.add({ id: 1, content: created.element, start: "2014-01-23" });
		eventComponents.set(id, created);
	}

	const mount = (container: HTMLElement) => {
		addItem("bleh");

		const options: TimelineOptions = {
			height: "100%",
			// template: templateTimelineEvent,
		};
		timeline = new Timeline(container, items, options);
	}

	const unmount = () => {
		timeline?.destroy();
		eventComponents.forEach(c => c.unmount());
		eventComponents.clear();
	}

	return {
		mount,
		addItem,
		unmount,
	}
}
export const timeline = createTimelineState();
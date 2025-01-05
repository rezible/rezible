import { mount, unmount } from "svelte";
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
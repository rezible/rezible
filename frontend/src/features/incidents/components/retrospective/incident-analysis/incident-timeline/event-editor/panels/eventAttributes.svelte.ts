import type { TimelineEvent } from "../../types";

type EventType = TimelineEvent["type"];
const createEventAttributesState = () => {
	let eventType = $state<EventType>("observation");
	
	let valid = $state(false);

	const initNew = () => {
		
	}

	const initFromEvent = (e: any) => {
		// TODO
		valid = true;
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
		valid = true;
	}

	// this is gross but oh well
	return {
		initNew,
		initFromEvent,
		get eventType() { return eventType },
		set eventType(t: EventType) { eventType = t; onUpdate(); },
		asAttributes(): any {
			return {
			}
		},
		get valid() { return valid },
	}
}

export const eventAttributes = createEventAttributesState();
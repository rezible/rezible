import type { IncidentEventAttributes } from "$lib/api";

type EventKind = IncidentEventAttributes["kind"];
const createEventAttributesState = () => {
	let eventKind = $state<EventKind>("observation");
	let isKey = $state(false);
	
	let valid = $state(false);

	const initNew = () => {
		eventKind = "observation";
		isKey = false;
		valid = false;
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
		get eventKind() { return eventKind },
		set eventKind(t: EventKind) { eventKind = t; onUpdate(); },
		get isKey() { return isKey },
		set isKey(v: boolean) { isKey = v; onUpdate() },
		asAttributes(): any {
			return {
			}
		},
		get valid() { return valid },
	}
}

export const eventAttributes = createEventAttributesState();
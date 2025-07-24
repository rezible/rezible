import { getOncallEventOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class EventViewState {
	eventId = $state<string>(null!);

	private eventQuery = createQuery(() => getOncallEventOptions({ path: { id: this.eventId } }));
	event = $derived(this.eventQuery.data?.data);
	eventTitle = $derived(this.event?.attributes.title ?? "");
	
	constructor(idFn: () => string) {
		this.eventId = idFn();
		watch(idFn, id => {this.eventId = id});
	}
}

export const eventViewStateCtx = new Context<EventViewState>("eventView");
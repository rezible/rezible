import { getEventOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class EventViewState {
	eventId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.eventId = idFn();
		watch(idFn, id => {this.eventId = id});
	}

	private eventQuery = createQuery(() => getEventOptions({ path: { id: this.eventId } }));
	event = $derived(this.eventQuery.data?.data);
	eventTitle = $derived(this.event?.attributes.title ?? "");
}

const ctx = new Context<EventViewState>("EventViewState");
export const setEventViewState = (idFn: Getter<string>) => ctx.set(new EventViewState(idFn));
export const useEventViewState = () => ctx.get();
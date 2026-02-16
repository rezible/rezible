import { getEventOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class EventViewController {
	eventId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.eventId = idFn();
		watch(idFn, id => {this.eventId = id});
	}

	private eventQuery = createQuery(() => getEventOptions({ path: { id: this.eventId } }));
	event = $derived(this.eventQuery.data?.data);
	eventTitle = $derived(this.event?.attributes.title ?? "");
}

const ctx = new Context<EventViewController>("EventViewController");
export const initEventViewController = (idFn: Getter<string>) => ctx.set(new EventViewController(idFn));
export const useEventViewController = () => ctx.get();
import { getOncallEventOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class OncallEventViewState {
	eventId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.eventId = idFn();
		watch(idFn, id => {this.eventId = id});
	}

	private eventQuery = createQuery(() => getOncallEventOptions({ path: { id: this.eventId } }));
	event = $derived(this.eventQuery.data?.data);
	eventTitle = $derived(this.event?.attributes.title ?? "");
}

const ctx = new Context<OncallEventViewState>("OncallEventView");
export const setOncallEventViewState = (idFn: Getter<string>) => ctx.set(new OncallEventViewState(idFn));
export const useOncallEventViewState = () => ctx.get();
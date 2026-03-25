import { getMeetingSessionOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class MeetingSessionViewController {
	sessionId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.sessionId = idFn();
		watch(idFn, id => {this.sessionId = id});
	}

	query = createQuery(() => getMeetingSessionOptions({ path: { id: this.sessionId } }));
	title = $derived(this.query.data?.data.attributes.title);
}

const ctx = new Context<MeetingSessionViewController>("MeetingSessionViewController");
export const initMeetingSessionViewController = (idFn: Getter<string>) => ctx.set(new MeetingSessionViewController(idFn));
export const useMeetingSessionViewController = () => ctx.get();
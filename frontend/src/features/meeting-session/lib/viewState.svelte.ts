import { getMeetingSessionOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class MeetingSessionViewState {
	sessionId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.sessionId = idFn();
		watch(idFn, id => {this.sessionId = id});
	}

	query = createQuery(() => getMeetingSessionOptions({ path: { id: this.sessionId } }));
	title = $derived(this.query.data?.data.attributes.title);
}

const ctx = new Context<MeetingSessionViewState>("meetingSessionView");
export const setMeetingSessionViewState = (idFn: Getter<string>) => ctx.set(new MeetingSessionViewState(idFn));
export const useMeetingSessionViewState = () => ctx.get();
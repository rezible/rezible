import { listMeetingSessionsOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export class MeetingsListViewController {
	query = createQuery(() => listMeetingSessionsOptions());

	monthStart = $state<Date>();
}

const ctx = new Context<MeetingsListViewController>("MeetingsListViewController");
export const initMeetingsListViewController = () => ctx.set(new MeetingsListViewController());
export const useMeetingsListViewController = () => ctx.get();

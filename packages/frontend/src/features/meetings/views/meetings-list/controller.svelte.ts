import { listMeetingSessionsOptions, type ListMeetingSessionsData } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export class MeetingsListViewController {
	searchValue = $state<string>();

	queryParams = $state<ListMeetingSessionsData["query"]>({});
	query = createQuery(() => listMeetingSessionsOptions({ query: this.queryParams }));

	monthStart = $state<Date>();
}

const ctx = new Context<MeetingsListViewController>("MeetingsListViewController");
export const initMeetingsListViewController = () => ctx.set(new MeetingsListViewController());
export const useMeetingsListViewController = () => ctx.get();
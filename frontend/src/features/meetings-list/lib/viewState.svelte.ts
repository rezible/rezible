import { listMeetingSessionsOptions, type ListMeetingSessionsData } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export class MeetingsListViewState {
	searchValue = $state<string>();

	queryParams = $state<ListMeetingSessionsData["query"]>({});
	query = createQuery(() => listMeetingSessionsOptions({ query: this.queryParams }));

	monthStart = $state<Date>();
}

const ctx = new Context<MeetingsListViewState>("meetingsListView");
export const setMeetingsListViewState = () => ctx.set(new MeetingsListViewState());
export const useMeetingsListViewState = () => ctx.get();
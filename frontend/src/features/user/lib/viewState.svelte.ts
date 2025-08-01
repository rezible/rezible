import { getUserOptions, listIncidentsOptions, listOncallRostersOptions, listOncallShiftsOptions, listTeamsOptions } from "$lib/api";
import type { Getter } from "$lib/utils.svelte";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class UserViewState {
	userId = $state<string>(null!);

	constructor(idFn: Getter<string>) {
		this.userId = idFn();
		watch(idFn, id => {this.userId = id});
	}

	userQuery = createQuery(() => getUserOptions({path: {id: this.userId}}));
	user = $derived(this.userQuery.data?.data);
	userName = $derived(this.user?.attributes.name ?? "");

	// TODO: use user timezone
	timeZone = $derived(getLocalTimeZone());
	userLocalTime = $derived(new Date().toLocaleTimeString([], {timeZone: this.timeZone, hour: "2-digit", minute: "2-digit"}));
	
	shiftsQuery = createQuery(() => listOncallShiftsOptions({query: {userId: this.userId}}))
	oncallShifts = $derived(this.shiftsQuery.data?.data);

	incidentsQuery = createQuery(() => listIncidentsOptions({query: {}}));
	incidents = $derived(this.incidentsQuery.data?.data);

	teamsQuery = createQuery(() => listTeamsOptions({query: {}}));
	teams = $derived(this.teamsQuery.data?.data);

	rostersQuery = createQuery(() => listOncallRostersOptions({query: {}}));
	rosters = $derived(this.rostersQuery.data?.data);
}

const ctx = new Context<UserViewState>("userView");
export const setUserViewState = (idFn: Getter<string>) => ctx.set(new UserViewState(idFn));
export const useUserViewState = () => ctx.get();
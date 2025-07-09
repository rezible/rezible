import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { getUserOncallInformationOptions, type GetUserOncallInformationData } from "./api";
import { session } from "./auth.svelte";
import { Context } from "runed";

export class UserOncallInformationState {
	private queryClient = useQueryClient();

	private infoQueryOptions = $derived(getUserOncallInformationOptions({
		query: {
			userId: session.userId,
			activeShifts: true,
		}}
	));
	infoQuery = createQuery(() => ({...this.infoQueryOptions, enabled: !!session.user}));
	current = $derived(this.infoQuery.data?.data);

	rosterIds = $derived(this.current?.rosters.map(r => r.id) ?? []);
	activeShifts = $derived(this.current?.activeShifts ?? []);

	loaded = $derived(this.infoQuery.isFetched);

	invalidate() {
		this.queryClient.invalidateQueries(this.infoQueryOptions);
	}
}

const ctx = new Context<UserOncallInformationState>("userOncallInformation");
export const setUserOncallInformationState = () => ctx.set(new UserOncallInformationState());
export const useUserOncallInformation = () => ctx.get();
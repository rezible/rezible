import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { getUserOncallInformationOptions, type GetUserOncallInformationData } from "./api";
import { useAuthSessionState } from "./auth.svelte";
import { Context } from "runed";

export class UserOncallInformationState {
	private session = useAuthSessionState();
	private queryClient = useQueryClient();

	private userId = $derived(this.session.user?.id || "");
	private infoQueryOptions = $derived(getUserOncallInformationOptions({
		query: {
			userId: this.userId,
			activeShifts: true,
		}}
	));
	infoQuery = createQuery(() => ({...this.infoQueryOptions, enabled: !!this.userId}));
	current = $derived(this.infoQuery.data?.data);

	rosters = $derived(this.current?.rosters ?? []);
	rosterIds = $derived(this.rosters.map(r => r.id));

	activeShifts = $derived(this.current?.activeShifts ?? []);

	loaded = $derived(this.infoQuery.isFetched);

	invalidate() {
		this.queryClient.invalidateQueries(this.infoQueryOptions);
	}
}

const ctx = new Context<UserOncallInformationState>("userOncallInformation");
export const setUserOncallInformationState = () => ctx.set(new UserOncallInformationState());
export const useUserOncallInformation = () => ctx.get();
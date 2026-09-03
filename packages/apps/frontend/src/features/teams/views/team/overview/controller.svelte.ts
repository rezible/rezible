import { listOncallRostersOptions, listUsersOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useTeamViewController } from "$features/teams/views/team";
import { Context } from "runed";

export class TeamOverviewController {
	private team = useTeamViewController();

	paginatedUsersQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) => ({
			...listUsersOptions({
				query: { teamId: this.team.teamId, ...pagination },
			}),
			enabled: !!this.team.teamId,
		}),
	});

	paginatedRostersQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) => ({
			...listOncallRostersOptions({
				query: { teamId: this.team.teamId, ...pagination },
			}),
			enabled: !!this.team.teamId,
		}),
	});

	users = $derived(this.paginatedUsersQuery.query.data?.data ?? []);
	rosters = $derived(this.paginatedRostersQuery.query.data?.data ?? []);
}

const ctx = new Context<TeamOverviewController>("TeamOverviewController");
export const initTeamOverviewController = () => ctx.set(new TeamOverviewController());
export const useTeamOverviewController = () => ctx.get();

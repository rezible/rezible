import { listTasksOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { useTeamViewController } from "$features/teams/views/team";
import { Context } from "runed";

export class TeamBacklogController {
	private team = useTeamViewController();

	paginatedTasksQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) => ({
			...listTasksOptions({ query: pagination }),
			enabled: !!this.team.teamId,
		}),
	});

	query = $derived(this.paginatedTasksQuery.query);
}

const ctx = new Context<TeamBacklogController>("TeamBacklogController");
export const initTeamBacklogController = () => ctx.set(new TeamBacklogController());
export const useTeamBacklogController = () => ctx.get();

import { getTeamOptions } from "$lib/api";
import { redirect } from "@sveltejs/kit";
import { validate as isValidUUID } from "uuid";
import type { PageLoad } from "./$types";

export const load = (async ({ params, parent, url }) => {
	const { queryClient } = await parent();

	const param = params.slug;
	const res = await queryClient.fetchQuery(getTeamOptions({ path: { id: param } }));
	const team = res.data;
	
	const slug = team.attributes.slug;
	queryClient.setQueryData(getTeamOptions({ path: { id: team.id } }).queryKey, res);
	queryClient.setQueryData(getTeamOptions({ path: { id: slug } }).queryKey, res);

	if (isValidUUID(param)) {
		throw redirect(301, url.pathname.replaceAll(param, slug) + url.search);
	}

	return {team}
}) satisfies PageLoad;

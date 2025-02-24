import { getTeamOptions } from "$lib/api";
import { redirect } from "@sveltejs/kit";
import { validate as isValidUUID } from "uuid";
import type { LayoutLoad } from "./$types";

export const load = (async ({ params, parent, url }) => {
	const { queryClient } = await parent();

	if (!isValidUUID(params.slug)) {
		return {
			id: params.slug,
		};
	}

	const id = params.slug;
	const res = await queryClient.fetchQuery(getTeamOptions({ path: { id } }));
	const slug = res.data.attributes.slug;

	queryClient.setQueryData(getTeamOptions({ path: { id: slug } }).queryKey, res);
	const slugPath = url.pathname.replaceAll(id, slug) + url.search;
	throw redirect(301, slugPath);
}) satisfies LayoutLoad;

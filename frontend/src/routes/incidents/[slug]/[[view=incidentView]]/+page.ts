import { getIncidentOptions } from "$lib/api";
import { redirect } from "@sveltejs/kit";
import { validate as isValidUUID } from "uuid";
import type { PageLoad } from "./$types";

export const load = (async ({ params, parent, url }) => {
	const { queryClient } = await parent();

	const slug = params.slug;

	if (isValidUUID(slug)) {
		const res = await queryClient.fetchQuery(getIncidentOptions({ path: { id: slug } }));
		const realSlug = res.data.attributes.slug;
		queryClient.setQueryData(getIncidentOptions({ path: { id: realSlug } }).queryKey, res);
		const slugPath = url.pathname.replaceAll(slug, realSlug) + url.search;
		throw redirect(301, slugPath);
	}

	return { slug };
}) satisfies PageLoad;

import { getIncidentOptions } from "$lib/api";
import { redirect } from "@sveltejs/kit";
import { validate as isValidUUID } from "uuid";
import type { PageLoad } from "./$types";
import type { QueryClient } from "@tanstack/svelte-query";

const validateSlugParamOrRedirect = async (param: string, url: URL, qc: QueryClient) => {
	if (!isValidUUID(param)) return param;

	const res = await qc.fetchQuery(getIncidentOptions({ path: { id: param } }));
	const slug = res.data.attributes.slug;
	qc.setQueryData(getIncidentOptions({ path: { id: slug } }).queryKey, res);
	const slugPath = url.pathname.replaceAll(slug, slug) + url.search;
	throw redirect(301, slugPath);
}

export const load = (async ({ params, parent, url }) => {
	const { queryClient } = await parent();

	const slug = await validateSlugParamOrRedirect(params.slug, url, queryClient);

	return { slug };
}) satisfies PageLoad;

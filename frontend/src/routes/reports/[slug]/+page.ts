import { redirect } from "@sveltejs/kit";
import { validate as isValidUUID } from "uuid";
import { queryOptions } from "@tanstack/svelte-query";
import type { PageLoad } from "./$types";

export const load = (async ({ params, parent, url }) => {
	const { queryClient } = await parent();

	const slugParam = params.slug;

	if (slugParam === "new") {
		return {builder: true}
	}

	if (isValidUUID(slugParam)) {
		/*
		const id = slugParam;
		const res = await queryClient.fetchQuery(getIncidentOptions({ path: { id } }));
		const slug = res.data.attributes.slug;
		queryClient.setQueryData(getIncidentOptions({ path: { id: slug } }).queryKey, res);
		const slugPath = url.pathname.replaceAll(id, slug) + url.search;
		throw redirect(301, slugPath);
		*/
	}

	return {id: slugParam, builder: false};
}) satisfies PageLoad;

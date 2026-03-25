import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load = (async ({ params, url }) => {
	const queryParams = new URLSearchParams({ name: params.name });
	url.searchParams.forEach((value, key) => queryParams.set(key, value));

	throw redirect(301, "/settings/integrations?" + queryParams.toString());
}) satisfies PageLoad;

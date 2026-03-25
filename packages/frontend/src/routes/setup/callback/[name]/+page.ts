import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load = (async ({ params, url }) => {
	const queryParams = new URLSearchParams({name: params.name});
	url.searchParams.forEach((v, k) => queryParams.set(k, v));

	throw redirect(301, "/setup?" + queryParams.toString());
}) satisfies PageLoad;

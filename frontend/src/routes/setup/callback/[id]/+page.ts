import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load = (async ({ params, url }) => {
	const id = params.id;

	const queryParams = new URLSearchParams({
		providerId: id,
	});
	url.searchParams.forEach((v, name) => {
		queryParams.set(name, v);
	});

	throw redirect(301, "/setup?" + queryParams.toString());
}) satisfies PageLoad;

import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import { resolve } from "$app/paths";

export const load = (async ({ params, url }) => {
	const p = new URLSearchParams(url.searchParams);
	p.set("name", params.name);
	throw redirect(301, `${resolve("/settings/integrations")}?${p.toString()}`);
}) satisfies PageLoad;

import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load = (async ({ url }) => {
    const queryParams = new URLSearchParams({});
    url.searchParams.forEach((v, k) => queryParams.set(k, v));
    throw redirect(301, "/login?" + queryParams.toString());
}) satisfies PageLoad;

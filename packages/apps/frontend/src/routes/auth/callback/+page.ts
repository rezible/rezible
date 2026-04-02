import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import { APP_AUTH_ROUTE_BASE } from "$lib/config.js";

export const load = (async ({ url }) => {
    const queryParams = new URLSearchParams({});
    url.searchParams.forEach((v, k) => queryParams.set(k, v));
    throw redirect(301, `${APP_AUTH_ROUTE_BASE}?${queryParams.toString()}`);
}) satisfies PageLoad;

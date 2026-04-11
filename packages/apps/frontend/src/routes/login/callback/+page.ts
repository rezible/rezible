import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import { APP_LOGIN_ROUTE } from "$lib/config";

export const load = (async ({ url }) => {
    throw redirect(301, `${APP_LOGIN_ROUTE}?${url.searchParams.toString()}`);
}) satisfies PageLoad;

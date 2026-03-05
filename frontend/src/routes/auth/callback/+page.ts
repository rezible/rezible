import { getTeamOptions } from "$lib/api";
import { redirect } from "@sveltejs/kit";
import { validate as isValidUUID } from "uuid";
import type { PageLoad } from "./$types";

export const load = (async () => {
	throw redirect(301, "/auth");
}) satisfies PageLoad;

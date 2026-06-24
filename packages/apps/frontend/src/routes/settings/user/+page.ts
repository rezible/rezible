import { redirect } from "@sveltejs/kit";
import { resolve } from "$app/paths";

export const load = ({ url }) => {
    throw redirect(301, `${resolve("/settings/user/preferences")}?${url.searchParams.toString()}`);
};

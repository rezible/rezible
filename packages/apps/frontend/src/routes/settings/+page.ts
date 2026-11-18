import { redirect } from "@sveltejs/kit";
import { resolve } from "$app/paths";

export const load = ({ url }) => {
	throw redirect(307, `${resolve("/settings/user")}?${url.searchParams.toString()}`);
};

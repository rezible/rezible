import { mdiDelta } from "@mdi/js";
import type { ListSystemComponentKindsResponse, SystemComponentAttributes } from "./api";
import type { MenuOption } from "svelte-ux";

export const systemComponentKindQueryMenuOptionSelect = (data: ListSystemComponentKindsResponse): MenuOption<string>[] => {
	return [{ label: "Service", value: "service" }];
}

export const getIconForComponentKind = (kind: SystemComponentAttributes["kind"]) => {
	return mdiDelta;
}
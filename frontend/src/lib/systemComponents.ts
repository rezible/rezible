import { mdiDelta } from "@mdi/js";
import type { ListSystemComponentKindsResponse, SystemComponentAttributes, SystemComponentKind } from "./api";
import type { MenuOption } from "svelte-ux";

export const systemComponentKindQueryMenuOptionSelect = (data: ListSystemComponentKindsResponse): MenuOption<string>[] => {
	return [{ label: "Service", value: "service" }];
}

export const getIconForComponentKind = (kind: SystemComponentKind) => {
	return mdiDelta;
}
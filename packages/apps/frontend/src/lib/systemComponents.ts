import { mdiDelta } from "@mdi/js";
import type { SystemComponentKind } from "./api";

type MenuOption<T extends any> = { label: string, value: T };

export const getSystemComponentKindMenuOptions = (kinds: SystemComponentKind[]): MenuOption<string>[] => {
	return kinds.map(k => ({ label: k.attributes.label, value: k.id }));
}

export const getIconForComponentKind = (kindId: string) => {
	// TODO
	return mdiDelta;
}
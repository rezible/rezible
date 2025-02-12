import { mdiDelta } from "@mdi/js";
import type { SystemComponentAttributes } from "./api";

export const getIconForComponentKind = (kind: SystemComponentAttributes["kind"]) => {
	return mdiDelta;
}
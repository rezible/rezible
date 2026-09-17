import { getMapCategoryDisplay } from "$features/systems/lib/system-map/category";
import type { GraphEntity } from "$features/systems/lib/system-map/graph";
import { normalized } from "$lib/utils";

export type MapNodePresentation = {
	label: string;
	categoryLabel: string;
	kindLabel?: string;
};

export const nodePresentationForEntity = ({category, kind, label}: GraphEntity): MapNodePresentation => {
	const categoryLabel = getMapCategoryDisplay(category).categoryLabel;
	const ignoreKinds = new Set<string>(["", "subject", normalized(category), normalized(categoryLabel)])
	const kindLabel = !ignoreKinds.has(normalized(kind)) ? kind.trim() : undefined;

	return { label, categoryLabel, kindLabel };
};

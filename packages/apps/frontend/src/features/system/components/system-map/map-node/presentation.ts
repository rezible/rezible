import type { GraphEntity } from "$features/system/lib/system-map/graph";
import { entityCategoryLabel, entityLabel } from "$features/system/lib/system-map/labels";

export type MapNodePresentation = {
	label: string;
	categoryLabel: string;
};

/** Shows the category and canonical kind once each. */
export const nodePresentationForEntity = (entity: GraphEntity): MapNodePresentation => ({
	label: entityLabel(entity),
	categoryLabel: entityCategoryLabel(entity.category),
});
